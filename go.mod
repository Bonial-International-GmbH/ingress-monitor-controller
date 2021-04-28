module github.com/Bonial-International-GmbH/ingress-monitor-controller

go 1.13

require (
	github.com/Bonial-International-GmbH/site24x7-go v0.0.6
	github.com/imdario/mergo v0.3.6
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v1.0.0
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v0.0.7
	github.com/stretchr/testify v1.4.0
	k8s.io/api v0.17.2
	k8s.io/apimachinery v0.17.2
	sigs.k8s.io/controller-runtime v0.5.0
	sigs.k8s.io/yaml v1.1.0
)
