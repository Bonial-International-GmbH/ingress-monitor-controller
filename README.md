ingress-monitor-controller
==========================

[![Build Status](https://travis-ci.org/Bonial-International-GmbH/ingress-monitor-controller.svg?branch=master)](https://travis-ci.org/Bonial-International-GmbH/ingress-monitor-controller)
[![Go Report Card](https://goreportcard.com/badge/github.com/Bonial-International-GmbH/ingress-monitor-controller?style=flat)](https://goreportcard.com/report/github.com/Bonial-International-GmbH/ingress-monitor-controller)
[![GoDoc](https://godoc.org/github.com/Bonial-International-GmbH/ingress-monitor-controller?status.svg)](https://godoc.org/github.com/Bonial-International-GmbH/ingress-monitor-controller)

A Kubernetes controller for automatically configuring website monitors for
ingresses. Currently the following providers are supported:

- [Site24x7](https://www.site24x7.com)
- Null provider (only useful for testing and debugging)

Building the Controller
-----------------------

Clone the repository and build a docker image by running:

```sh
make docker-build
```

This will build and tag the image as `ingress-monitor-controller:latest`. If
you want to use a different image name and tag, override the `IMAGE` and `TAG`
environment variables:

```sh
make docker-build IMAGE=my.registry/ingress-monitor-controller TAG="$(git rev-parse HEAD)"
```

Deploying the Controller
------------------------

You can find example manifests for deploying the controller in the
[`deploy/`](deploy/) directory. Be sure to customize them to your needs before
applying them to your cluster:

```sh
kubectl apply -f deploy/
```

Configuration
-------------

The provider for the website monitors can be set via the `--provider` argument
of the controller. The controller optionally reads provider specific config
from a file which can be specified using the `--provider-config` flag.

### Provider Configuration File

The config file has the following YAML format:

```yaml
name-of-the-provider:
  [...provider specific config like authentication...]
  monitorDefaults:
    [...provider specific default values for monitors if not overridden explicitly...]
```

Refer to [`pkg/config/providers.go`](pkg/config/providers.go) for all available
config keys and their documentation. For a usage example of the config file
check out the ConfigMap and Deployment manifests in the [`deploy/`](deploy/)
directory.

Example configuration for the Site24x7 provider:

```yaml
site24x7:
  clientID: the-oauth-client-id
  clientSecret: the-oauth-client-secret
  refreshToken: the-oauth-refresh-token
  monitorDefaults:
	Actions:
      - alert_type: 0
        action_id: "123"
	AuthPass: ""
	AuthUser: ""
    AutoLocationProfile: true
	AutoMonitorGroup: true
	AutoNotificationProfile: true
	AutoThresholdProfile: true
	AutoUserGroup: true
	CheckFrequency: "1"
    CustomHeaders:
      - name: X-Monitor-Created-By
        value: ingress-monitor-controller
	HTTPMethod: G
	LocationProfileID: "123"
	MatchCase: true
	MonitorGroupIDs:
      - "123"
	NotificationProfileID: "456"
	ThresholdProfileID: "678"
	Timeout: 10
	UseNameServer: true
	UserAgent: "curl/v1.33.7"
	UserGroupIDs:
      - "456"
```

### Ingress Annotations

To automatically create a website monitor for an ingress, it requires to be annotated with the `ingress-monitor.bonial.com/enabled` annotation:

```yaml
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  annotations:
    ingress-monitor.bonial.com/enabled: "true"
  name: my-ingress
  namespace: my-namespace
```

### Global Ingress Annotations

Global ingress annotations configure behaviour that is not specific to a
certain provider. The following annotations are supported:

| Annotation                                 | Description                                                                                | Default   |
| ------------                               | -------------                                                                              | --------- |
| `ingress-monitor.bonial.com/enabled`       | Controls whether a monitor should be created for the ingress or not                        | `false`   |
| `ingress-monitor.bonial.com/force-https`   | Forces the monitored URL to be HTTPS even it TLS is not configured for the ingress         | `false`   |
| `ingress-monitor.bonial.com/path-override` | By default, `/` is monitored. This can be overridden with this annotation (e.g. `/health`) | `/`       |

### Supported Third Party Annotations

The controller will honor the `nginx.ingress.kubernetes.io/force-ssl-redirect`
annotation and force website monitors to be created for HTTPS if this
annotation is set to `true`.

### Provider Specific Annotations

You can control the configuration of a website monitor via annotations that are
provider specific. For a full list of supported annotations check out the
constants and their documentation in
[`pkg/config/annotations.go`](pkg/config/annotations.go).
