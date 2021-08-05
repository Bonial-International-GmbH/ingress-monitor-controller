package monitor

import (
	"bytes"
	"text/template"

	networkingv1 "k8s.io/api/networking/v1"
)

type templateArgs struct {
	IngressName string
	Namespace   string
}

// Namer builds names for ingress monitors from a name template.
type Namer struct {
	template *template.Template
}

// NewNamer creates a new *Namer with given name template string. Returns an
// error if the name template is invalid.
func NewNamer(nameTemplate string) (*Namer, error) {
	tpl, err := template.New("monitor-name").Parse(nameTemplate)
	if err != nil {
		return nil, err
	}

	n := &Namer{
		template: tpl,
	}

	return n, nil
}

// Name builds a monitor name for given ingress. Returns an error if rendering
// the name template fails.
func (n *Namer) Name(ingress *networkingv1.Ingress) (string, error) {
	var buf bytes.Buffer

	err := n.template.Execute(&buf, templateArgs{
		IngressName: ingress.Name,
		Namespace:   ingress.Namespace,
	})
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
