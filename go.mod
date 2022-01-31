module github.com/Bonial-International-GmbH/ingress-monitor-controller

go 1.16

require (
	github.com/Bonial-International-GmbH/site24x7-go v0.0.6
	github.com/imdario/mergo v0.3.12
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.11.0
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.2.1
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/stretchr/testify v1.7.0
	k8s.io/api v0.22.1
	k8s.io/apimachinery v0.23.3
	sigs.k8s.io/controller-runtime v0.10.0
	sigs.k8s.io/yaml v1.2.0
)
