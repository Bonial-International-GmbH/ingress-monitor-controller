package monitor

import (
	"github.com/Bonial-International-GmbH/ingress-monitor-controller/pkg/config"
	"github.com/Bonial-International-GmbH/ingress-monitor-controller/pkg/ingress"
	"github.com/Bonial-International-GmbH/ingress-monitor-controller/pkg/models"
	"github.com/Bonial-International-GmbH/ingress-monitor-controller/pkg/provider"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/klog"
)

// Service defines the interface for a service that takes care of creating,
// updating or deleting monitors.
type Service interface {
	// EnsureMonitor ensures that a monitor is in sync with the current ingress
	// configuration. If the monitor does not exist, it will be created.
	EnsureMonitor(ingress *v1beta1.Ingress) error

	// DeleteMonitor deletes the monitor for an ingress. It must not be treated
	// as an error if the monitor was already deleted.
	DeleteMonitor(ingress *v1beta1.Ingress) error
}

type service struct {
	provider provider.Interface
	namer    *Namer
	options  *config.Options
}

// NewService creates a new Service with options. Returns an error if service
// initialization fails.
func NewService(options *config.Options) (Service, error) {
	provider, err := provider.New(options.ProviderName, options.ProviderConfig)
	if err != nil {
		return nil, err
	}

	namer, err := NewNamer(options.NameTemplate)
	if err != nil {
		return nil, err
	}

	s := &service{
		provider: provider,
		namer:    namer,
		options:  options,
	}

	return s, nil
}

// EnsureMonitor implements Service.
func (s *service) EnsureMonitor(ing *v1beta1.Ingress) error {
	err := ingress.Validate(ing)
	if err != nil {
		klog.V(1).Infof(`ingress "%s/%s" is not supported: %v`, ing.Namespace, ing.Name, err)
		return nil
	}

	name, err := s.namer.Name(ing)
	if err != nil {
		return err
	}

	url, err := ingress.BuildMonitorURL(ing)
	if err != nil {
		return err
	}

	monitor, err := s.provider.Get(name)
	if err == models.ErrMonitorNotFound {
		return s.createMonitor(name, url, ing.Annotations)
	} else if err != nil {
		return err
	}

	return s.updateMonitor(monitor, name, url, ing.Annotations)
}

// DeleteMonitor implements Service.
func (s *service) DeleteMonitor(ingress *v1beta1.Ingress) error {
	name, err := s.namer.Name(ingress)
	if err != nil {
		return err
	}

	if s.options.NoDelete {
		klog.V(1).Infof("not deleting monitor %q because monitor deletion is disabled", name)
		return nil
	}

	return s.deleteMonitor(name)
}

func (s *service) createMonitor(name, url string, annotations map[string]string) error {
	monitor := &models.Monitor{
		URL:         url,
		Name:        name,
		Annotations: annotations,
	}

	err := s.provider.Create(monitor)
	if err != nil {
		return err
	}

	klog.Infof("monitor %q created", monitor.Name)

	return nil
}

func (s *service) updateMonitor(monitor *models.Monitor, name, url string, annotations map[string]string) error {
	monitor.URL = url
	monitor.Name = name
	monitor.Annotations = annotations

	err := s.provider.Update(monitor)
	if err != nil {
		return err
	}

	klog.Infof("monitor %q updated", monitor.Name)

	return nil
}

func (s *service) deleteMonitor(name string) error {
	err := s.provider.Delete(name)
	if err == models.ErrMonitorNotFound {
		klog.V(1).Infof("monitor %q was already deleted", name)
		return nil
	} else if err != nil {
		return err
	}

	klog.Infof("monitor %q deleted", name)

	return nil
}
