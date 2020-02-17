package admission

import (
	"bytes"
	"context"
	gojson "encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/Bonial-International-GmbH/ingress-monitor-controller/pkg/monitor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type mockService struct {
	mock.Mock
	monitor.Service
}

func (m *mockService) AnnotateIngress(ingress *v1beta1.Ingress) (bool, error) {
	args := m.Called(ingress)
	updated := args.Bool(0)
	if updated {
		if ingress.Annotations == nil {
			ingress.Annotations = map[string]string{}
		}

		ingress.Annotations["foobar"] = "baz"
	}

	return args.Bool(0), args.Error(1)
}

func TestIngressHandler_Handle(t *testing.T) {
	tests := []struct {
		name     string
		req      admission.Request
		expected admission.Response
		setup    func(*mockService)
	}{
		{
			name: "returns admission error if object is not a valid ingress",
			req: admission.Request{
				AdmissionRequest: admissionv1beta1.AdmissionRequest{
					Object: runtime.RawExtension{
						Raw: []byte(`iaminvalid`),
					},
				},
			},
			expected: admission.Errored(http.StatusBadRequest, errors.New(`couldn't get version/kind; json parse error: json: cannot unmarshal string into Go value of type struct { APIVersion string "json:\"apiVersion,omitempty\""; Kind string "json:\"kind,omitempty\"" }`)),
		},
		{
			name: "allows ingresses that do not need to be annotated",
			req: admission.Request{
				AdmissionRequest: admissionv1beta1.AdmissionRequest{
					Object: runtime.RawExtension{
						Raw: serializeObject(t, &v1beta1.Ingress{}),
					},
				},
			},
			setup: func(m *mockService) {
				m.On("AnnotateIngress", mock.Anything).Return(false, nil)
			},
			expected: admission.Allowed(""),
		},
		{
			name: "does not deny if annotating ingress fails",
			req: admission.Request{
				AdmissionRequest: admissionv1beta1.AdmissionRequest{
					Object: runtime.RawExtension{
						Raw: serializeObject(t, &v1beta1.Ingress{}),
					},
				},
			},
			setup: func(m *mockService) {
				m.On("AnnotateIngress", mock.Anything).Return(false, errors.New("whoops"))
			},
			expected: admission.Allowed(""),
		},
		{
			name: "creates json patch",
			req: admission.Request{
				AdmissionRequest: admissionv1beta1.AdmissionRequest{
					Object: runtime.RawExtension{
						Raw: serializeObject(t, &v1beta1.Ingress{}),
					},
				},
			},
			setup: func(m *mockService) {
				m.On("AnnotateIngress", mock.Anything).Return(true, nil)
			},
			expected: func() admission.Response {
				raw := serializeObject(t, &v1beta1.Ingress{})

				buf, err := gojson.Marshal(&v1beta1.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"foobar": "baz",
						},
					},
				})
				require.NoError(t, err)

				return admission.PatchResponseFromRaw(raw, buf)
			}(),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockSvc := &mockService{}

			if test.setup != nil {
				test.setup(mockSvc)
			}

			d, err := admission.NewDecoder(runtime.NewScheme())
			require.NoError(t, err)

			h := &IngressHandler{
				monitorService: mockSvc,
				decoder:        d,
			}

			resp := h.Handle(context.Background(), test.req)

			assert.Equal(t, test.expected, resp)
		})
	}
}

func serializeObject(t *testing.T, obj runtime.Object) []byte {
	encoder := json.NewSerializer(json.DefaultMetaFactory, nil, nil, false)

	var buf bytes.Buffer

	err := encoder.Encode(obj, &buf)
	if err != nil {
		t.Fatal(err)
	}

	return buf.Bytes()
}
