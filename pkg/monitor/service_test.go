package monitor

import (
	"errors"
	"testing"

	"github.com/Bonial-International-GmbH/ingress-monitor-controller/pkg/config"
	"github.com/Bonial-International-GmbH/ingress-monitor-controller/pkg/models"
	"github.com/Bonial-International-GmbH/ingress-monitor-controller/pkg/provider/fake"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestService_EnsureMonitor(t *testing.T) {
	tests := []struct {
		name     string
		ingress  *v1beta1.Ingress
		options  config.Options
		setup    func(*fake.Provider)
		validate func(*testing.T, *fake.Provider)
		expected error
	}{
		{
			name: "invalid ingress is ignored without error",
			ingress: &v1beta1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "kube-system",
				},
			},
			validate: func(t *testing.T, p *fake.Provider) {
				assert.Len(t, p.Calls, 0)
			},
		},
		{
			name: "non-existent monitor is created",
			ingress: &v1beta1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "kube-system",
					Annotations: map[string]string{
						config.AnnotationEnabled: "true",
					},
				},
				Spec: v1beta1.IngressSpec{
					Rules: []v1beta1.IngressRule{
						{Host: "foo.bar.baz"},
					},
				},
			},
			setup: func(p *fake.Provider) {
				p.On("Get", "kube-system-foo").Return(nil, models.ErrMonitorNotFound)
				p.On("Create", &models.Monitor{
					URL:  "http://foo.bar.baz",
					Name: "kube-system-foo",
					Annotations: config.Annotations{
						config.AnnotationEnabled: "true",
					},
				}).Return(nil)
			},
		},
		{
			name: "existing monitor is created",
			ingress: &v1beta1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "kube-system",
					Annotations: map[string]string{
						config.AnnotationEnabled: "true",
					},
				},
				Spec: v1beta1.IngressSpec{
					Rules: []v1beta1.IngressRule{
						{Host: "foo.bar.baz"},
					},
				},
			},
			setup: func(p *fake.Provider) {
				p.On("Get", "kube-system-foo").Return(&models.Monitor{
					ID:   "123",
					Name: "kube-system-foo",
					URL:  "http://bar.baz",
				}, nil)
				p.On("Update", &models.Monitor{
					ID:   "123",
					URL:  "http://foo.bar.baz",
					Name: "kube-system-foo",
					Annotations: config.Annotations{
						config.AnnotationEnabled: "true",
					},
				}).Return(nil)
			},
		},
		{
			name: "does not create/update monitor if lookup fails",
			ingress: &v1beta1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "kube-system",
					Annotations: map[string]string{
						config.AnnotationEnabled: "true",
					},
				},
				Spec: v1beta1.IngressSpec{
					Rules: []v1beta1.IngressRule{
						{Host: "foo.bar.baz"},
					},
				},
			},
			setup: func(p *fake.Provider) {
				p.On("Get", "kube-system-foo").Return(nil, errors.New("error"))
			},
			validate: func(t *testing.T, p *fake.Provider) {
				p.AssertNotCalled(t, "Create", mock.Anything)
				p.AssertNotCalled(t, "Update", mock.Anything)
			},
			expected: errors.New("error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			svc, provider := newTestService(t, &test.options)

			if test.setup != nil {
				test.setup(provider)
			}

			err := svc.EnsureMonitor(test.ingress)
			if test.expected != nil {
				require.Error(t, err)
				assert.Equal(t, test.expected.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			if test.validate != nil {
				test.validate(t, provider)
			}
		})
	}
}

func TestService_DeleteMonitor(t *testing.T) {
	tests := []struct {
		name     string
		ingress  *v1beta1.Ingress
		options  config.Options
		setup    func(*fake.Provider)
		validate func(*testing.T, *fake.Provider)
		expected error
	}{
		{
			name: "delete monitor for ingress",
			ingress: &v1beta1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "kube-system",
				},
			},
			setup: func(p *fake.Provider) {
				p.On("Delete", "kube-system-foo").Return(nil)
			},
			validate: func(t *testing.T, p *fake.Provider) {
				p.AssertCalled(t, "Delete", "kube-system-foo")
			},
		},
		{
			name: "deletion of nonexistant monitor does not error",
			ingress: &v1beta1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "kube-system",
				},
			},
			setup: func(p *fake.Provider) {
				p.On("Delete", "kube-system-foo").Return(models.ErrMonitorNotFound)
			},
			validate: func(t *testing.T, p *fake.Provider) {
				p.AssertCalled(t, "Delete", "kube-system-foo")
			},
		},
		{
			name:    "no deletions if NoDelete options is set",
			options: config.Options{NoDelete: true},
			ingress: &v1beta1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "kube-system",
				},
			},
			validate: func(t *testing.T, p *fake.Provider) {
				p.AssertNotCalled(t, "Delete", mock.Anything)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			svc, provider := newTestService(t, &test.options)

			if test.setup != nil {
				test.setup(provider)
			}

			err := svc.DeleteMonitor(test.ingress)
			if test.expected != nil {
				require.Error(t, err)
				assert.Equal(t, test.expected.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			if test.validate != nil {
				test.validate(t, provider)
			}
		})
	}
}

func newTestService(t *testing.T, options *config.Options) (*service, *fake.Provider) {
	namer, err := NewNamer("{{.Namespace}}-{{.IngressName}}")
	if err != nil {
		t.Fatal(err)
	}

	provider := &fake.Provider{}

	svc := &service{
		provider: provider,
		namer:    namer,
		options:  options,
	}

	return svc, provider
}
