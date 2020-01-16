package admission

import (
	"errors"
	"testing"

	"github.com/Bonial-International-GmbH/ingress-monitor-controller/pkg/config"
	"github.com/Bonial-International-GmbH/ingress-monitor-controller/pkg/monitor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"k8s.io/api/admission/v1beta1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type mockService struct {
	monitor.Service
	mock.Mock
}

func (m *mockService) GetProviderIPSourceRanges(ingress *extensionsv1beta1.Ingress) ([]string, error) {
	args := m.Called(ingress)
	if obj, ok := args.Get(0).([]string); ok {
		return obj, args.Error(1)
	}

	return nil, args.Error(1)
}

func TestWebhook_Admit(t *testing.T) {
	tests := []struct {
		name                 string
		request              *v1beta1.AdmissionRequest
		expectedResponse     *v1beta1.AdmissionResponse
		providerSourceRanges []string
		providerErr          error
		expectError          bool
		expectedErr          error
	}{
		{
			name: "non-ingress objects are not patched",
			request: newAdmissionRequestBuilder(t).
				withResource(metav1.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}).
				build(),
			expectedResponse: newAdmissionResponseBuilder(t).build(),
		},
		{
			name: "ingress objects without monitor annotation are not patched",
			request: newAdmissionRequestBuilder(t).
				withObject(&extensionsv1beta1.Ingress{}).
				build(),
			expectedResponse: newAdmissionResponseBuilder(t).build(),
		},
		{
			name: `ingress objects with monitor annotation with value "false" are not patched`,
			request: newAdmissionRequestBuilder(t).
				withObject(&extensionsv1beta1.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							config.AnnotationEnabled: "false",
						},
					},
				}).
				build(),
			expectedResponse: newAdmissionResponseBuilder(t).build(),
		},
		{
			name: `ingress objects without source range whitelist annotation are not patched`,
			request: newAdmissionRequestBuilder(t).
				withObject(&extensionsv1beta1.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							config.AnnotationEnabled: "true",
						},
					},
				}).
				build(),
			expectedResponse: newAdmissionResponseBuilder(t).build(),
		},
		{
			name: `errors while retrieving provider source ranges do not cause object to be rejected`,
			request: newAdmissionRequestBuilder(t).
				withObject(&extensionsv1beta1.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							config.AnnotationEnabled:            "true",
							nginxWhitelistSourceRangeAnnotation: "1.2.3.4/32",
						},
					},
				}).
				build(),
			providerErr:      errors.New("whoops"),
			expectedResponse: newAdmissionResponseBuilder(t).build(),
		},
		{
			name: `empty provider source ranges do not cause the object to be patched`,
			request: newAdmissionRequestBuilder(t).
				withObject(&extensionsv1beta1.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							config.AnnotationEnabled:            "true",
							nginxWhitelistSourceRangeAnnotation: "1.2.3.4/32",
						},
					},
				}).
				build(),
			providerSourceRanges: []string{},
			expectedResponse:     newAdmissionResponseBuilder(t).build(),
		},
		{
			name: `provider source ranges are merged with the configured whitelist and produce a patch`,
			request: newAdmissionRequestBuilder(t).
				withObject(&extensionsv1beta1.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							config.AnnotationEnabled:            "true",
							nginxWhitelistSourceRangeAnnotation: "1.2.3.4/32",
						},
					},
				}).
				build(),
			providerSourceRanges: []string{"5.6.7.8/32"},
			expectedResponse: newAdmissionResponseBuilder(t).withJSONPatch([]patchOperation{
				{
					Op:    "replace",
					Path:  "/metadata/annotations/nginx.ingress.kubernetes.io~1whitelist-source-range",
					Value: "1.2.3.4/32,5.6.7.8/32",
				},
			}).build(),
		},
		{
			name: `already present source ranges are not added again`,
			request: newAdmissionRequestBuilder(t).
				withObject(&extensionsv1beta1.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							config.AnnotationEnabled:            "true",
							nginxWhitelistSourceRangeAnnotation: "1.2.3.4/32",
						},
					},
				}).
				build(),
			providerSourceRanges: []string{"5.6.7.8/32", "1.2.3.4/32"},
			expectedResponse: newAdmissionResponseBuilder(t).withJSONPatch([]patchOperation{
				{
					Op:    "replace",
					Path:  "/metadata/annotations/nginx.ingress.kubernetes.io~1whitelist-source-range",
					Value: "1.2.3.4/32,5.6.7.8/32",
				},
			}).build(),
		},
		{
			name: `if provider source ranges are already whitelisted, no patch is created`,
			request: newAdmissionRequestBuilder(t).
				withObject(&extensionsv1beta1.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							config.AnnotationEnabled:            "true",
							nginxWhitelistSourceRangeAnnotation: "5.6.7.8/32,1.2.3.4/32,9.10.11.12/32",
						},
					},
				}).
				build(),
			providerSourceRanges: []string{"1.2.3.4/32", "5.6.7.8/32"},
			expectedResponse:     newAdmissionResponseBuilder(t).build(),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			svc := &mockService{}

			svc.On("GetProviderIPSourceRanges", mock.Anything).Return(test.providerSourceRanges, test.providerErr)

			c := NewWebhook(svc)

			response, err := c.Admit(test.request)
			if test.expectError {
				require.Error(t, err)
				if test.expectedErr != nil {
					assert.Equal(t, test.expectedErr, err)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expectedResponse, response)
			}
		})
	}
}
