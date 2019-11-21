package ingress

import (
	"testing"

	"github.com/Bonial-International-GmbH/ingress-monitor-controller/pkg/config"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name     string
		ingress  *v1beta1.Ingress
		expected error
	}{
		{
			name: "valid ingress without TLS",
			ingress: &v1beta1.Ingress{
				Spec: v1beta1.IngressSpec{
					Rules: []v1beta1.IngressRule{
						{Host: "foo.bar.baz"},
					},
				},
			},
		},
		{
			name: "valid ingress with TLS",
			ingress: &v1beta1.Ingress{
				Spec: v1beta1.IngressSpec{
					TLS: []v1beta1.IngressTLS{
						{Hosts: []string{"foo.bar.baz"}},
					},
					Rules: []v1beta1.IngressRule{
						{Host: "foo.bar.baz"},
					},
				},
			},
		},
		{
			name: "wildcard TLS hosts are not supported",
			ingress: &v1beta1.Ingress{
				Spec: v1beta1.IngressSpec{
					TLS: []v1beta1.IngressTLS{
						{Hosts: []string{"*.bar.baz"}},
					},
					Rules: []v1beta1.IngressRule{
						{Host: "foo.bar.baz"},
					},
				},
			},
			expected: errors.New(`ingress TLS host "*.bar.baz" contains wildcards`),
		},
		{
			name: "wildcard hosts are not supported",
			ingress: &v1beta1.Ingress{
				Spec: v1beta1.IngressSpec{
					Rules: []v1beta1.IngressRule{
						{Host: "*.bar.baz"},
					},
				},
			},
			expected: errors.New(`ingress host "*.bar.baz" contains wildcards`),
		},
		{
			name:     "ingress needs to have at least one rule",
			ingress:  &v1beta1.Ingress{},
			expected: errors.New(`ingress does not have any rules`),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := Validate(test.ingress)
			if test.expected != nil {
				require.Error(t, err)
				assert.Equal(t, test.expected.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBuildMonitorURL(t *testing.T) {
	tests := []struct {
		name     string
		ingress  *v1beta1.Ingress
		expected string
	}{
		{
			name: "simple http url",
			ingress: &v1beta1.Ingress{
				Spec: v1beta1.IngressSpec{
					Rules: []v1beta1.IngressRule{
						{Host: "foo.bar.baz"},
					},
				},
			},
			expected: "http://foo.bar.baz",
		},
		{
			name: "https url via TLS config",
			ingress: &v1beta1.Ingress{
				Spec: v1beta1.IngressSpec{
					TLS: []v1beta1.IngressTLS{
						{Hosts: []string{"foo.bar.baz"}},
					},
					Rules: []v1beta1.IngressRule{
						{Host: "foo.bar.baz"},
					},
				},
			},
			expected: "https://foo.bar.baz",
		},
		{
			name: "https url via nginx ingress redirect annotation",
			ingress: &v1beta1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						nginxForceSSLRedirectAnnotation: "true",
					},
				},
				Spec: v1beta1.IngressSpec{
					Rules: []v1beta1.IngressRule{
						{Host: "foo.bar.baz"},
					},
				},
			},
			expected: "https://foo.bar.baz",
		},
		{
			name: "https url via force https annotation",
			ingress: &v1beta1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						config.AnnotationForceHTTPS: "true",
					},
				},
				Spec: v1beta1.IngressSpec{
					Rules: []v1beta1.IngressRule{
						{Host: "foo.bar.baz"},
					},
				},
			},
			expected: "https://foo.bar.baz",
		},
		{
			name: "respect path override annotation",
			ingress: &v1beta1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						config.AnnotationForceHTTPS:   "true",
						config.AnnotationPathOverride: "health",
					},
				},
				Spec: v1beta1.IngressSpec{
					Rules: []v1beta1.IngressRule{
						{Host: "foo.bar.baz"},
					},
				},
			},
			expected: "https://foo.bar.baz/health",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			url, err := BuildMonitorURL(test.ingress)
			require.NoError(t, err)
			assert.Equal(t, test.expected, url)
		})
	}
}
