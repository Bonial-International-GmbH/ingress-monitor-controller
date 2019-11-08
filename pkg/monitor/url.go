package monitor

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/Bonial-International-GmbH/ingress-monitor-controller/pkg/config"
	"k8s.io/api/extensions/v1beta1"
)

const (
	nginxForceSSLRedirectAnnotation = "nginx.ingress.kubernetes.io/force-ssl-redirect"
)

// isSupported returns true if an ingress is supported by the controller.
// Supported ingresses must have at least one ingress rule and must not use
// wildcards in hostnames.
func isSupported(ingress *v1beta1.Ingress) bool {
	if supportsTLS(ingress) {
		return !containsWildcard(ingress.Spec.TLS[0].Hosts[0])
	}

	if len(ingress.Spec.Rules) == 0 {
		return false
	}

	return !containsWildcard(ingress.Spec.Rules[0].Host)
}

// buildURL builds the url that should be monitored on an ingress.
func buildURL(ingress *v1beta1.Ingress) (string, error) {
	host := getHost(ingress)

	u, err := url.Parse(host)
	if err != nil {
		return "", err
	}

	if path, ok := ingress.Annotations[config.AnnotationPathOverride]; ok {
		u.Path = path
	}

	return u.String(), nil
}

func getHost(ingress *v1beta1.Ingress) string {
	if supportsTLS(ingress) {
		return fmt.Sprintf("https://%s", ingress.Spec.TLS[0].Hosts[0])
	}

	if forceHTTPS(ingress) {
		return fmt.Sprintf("https://%s", ingress.Spec.Rules[0].Host)
	}

	return fmt.Sprintf("http://%s", ingress.Spec.Rules[0].Host)
}

func supportsTLS(ingress *v1beta1.Ingress) bool {
	return len(ingress.Spec.TLS) > 0 && len(ingress.Spec.TLS[0].Hosts) > 0 && len(ingress.Spec.TLS[0].Hosts[0]) > 0
}

func forceHTTPS(ingress *v1beta1.Ingress) bool {
	a := config.Annotations(ingress.Annotations)

	if a.Bool(config.AnnotationForceHTTPS) || a.Bool(nginxForceSSLRedirectAnnotation) {
		return true
	}

	return false
}

func containsWildcard(hostName string) bool {
	return strings.Contains(hostName, "*")
}
