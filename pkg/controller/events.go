package controller

import (
	"github.com/Bonial-International-GmbH/ingress-monitor-controller/pkg/config"
	"github.com/Bonial-International-GmbH/ingress-monitor-controller/pkg/monitor"
	"k8s.io/api/extensions/v1beta1"
)

type event interface {
	handle(s monitor.Service) error
}

type updateEvent struct {
	oldIngress *v1beta1.Ingress
	newIngress *v1beta1.Ingress
}

func (e updateEvent) handle(s monitor.Service) error {
	a := config.Annotations(e.newIngress.Annotations)

	if a.Bool(config.AnnotationEnabled) {
		return s.EnsureMonitor(e.newIngress)
	}

	return s.DeleteMonitor(e.oldIngress)
}

type deleteEvent struct {
	ingress *v1beta1.Ingress
}

func (e deleteEvent) handle(s monitor.Service) error {
	return s.DeleteMonitor(e.ingress)
}