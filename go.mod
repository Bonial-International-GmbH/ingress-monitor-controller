module github.com/Bonial-International-GmbH/ingress-monitor-controller

go 1.16

require (
	github.com/Bonial-International-GmbH/site24x7-go v0.0.6
	github.com/imdario/mergo v0.3.12
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.10.0
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.1.3
	github.com/stretchr/testify v1.7.0
	k8s.io/api v0.19.2
	k8s.io/apimachinery v0.19.2
	sigs.k8s.io/controller-runtime v0.7.2
	sigs.k8s.io/yaml v1.2.0
)
