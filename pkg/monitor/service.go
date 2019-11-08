package monitor

import (
	"github.com/Bonial-International-GmbH/ingress-monitor-controller/pkg/config"
	"github.com/Bonial-International-GmbH/ingress-monitor-controller/pkg/models"
	"github.com/Bonial-International-GmbH/ingress-monitor-controller/pkg/provider"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/klog"
)

// Service defines the interface for a service that takes care of creating,
// updating or deleting monitors.
type Service interface {
	CreateOrUpdateMonitor(ingress *v1beta1.Ingress) error
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

// CreateOrUpdateMonitor implements Service.
func (s *service) CreateOrUpdateMonitor(ingress *v1beta1.Ingress) error {
	if !isSupported(ingress) {
		klog.V(1).Infof(`ingress "%s/%s" is not supported, not creating monitor`, ingress.Namespace, ingress.Name)
		return nil
	}

	name, err := s.namer.Name(ingress)
	if err != nil {
		return err
	}

	url, err := buildURL(ingress)
	if err != nil {
		return err
	}

	klog.V(1).Infof("fetching monitor %q", name)

	monitor, err := s.provider.Get(name)
	if err == models.ErrMonitorNotFound {
		monitor := &models.Monitor{
			URL:         url,
			Name:        name,
			Annotations: ingress.Annotations,
		}

		klog.Infof("creating monitor %#v", monitor)

		return s.provider.Create(monitor)
	} else if err != nil {
		return err
	}

	monitor.URL = url
	monitor.Name = name
	monitor.Annotations = ingress.Annotations

	klog.Infof("updating monitor %#v", monitor)

	return s.provider.Update(monitor)
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

	klog.Infof("deleting monitor %q", name)

	err = s.provider.Delete(name)
	if err == models.ErrMonitorNotFound {
		klog.V(1).Infof("monitor %q was already deleted", name)
		return nil
	}

	return err
}
