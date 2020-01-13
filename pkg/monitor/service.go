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

	// GetProviderIPSourceRanges retrieves the IP source ranges that the
	// monitor provider is using to perform checks from. It is a list of CIDR
	// blocks. These source ranges can be used to update the IP whitelist (if
	// one is defined) of an ingress to allow checks by the monitor provider.
	GetProviderIPSourceRanges(ingress *v1beta1.Ingress) ([]string, error)
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
		klog.V(1).Infof(`ignoring unsupported ingress "%s/%s": %v`, ing.Namespace, ing.Name, err)
		return nil
	}

	newMonitor, err := s.buildMonitorModel(ing)
	if err != nil {
		return err
	}

	oldMonitor, err := s.provider.Get(newMonitor.Name)
	if err == models.ErrMonitorNotFound {
		return s.createMonitor(newMonitor)
	} else if err != nil {
		return err
	}

	return s.updateMonitor(oldMonitor, newMonitor)
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

func (s *service) createMonitor(monitor *models.Monitor) error {
	err := s.provider.Create(monitor)
	if err != nil {
		return err
	}

	klog.Infof("monitor %q created", monitor.Name)

	return nil
}

func (s *service) updateMonitor(oldMonitor, newMonitor *models.Monitor) error {
	newMonitor.ID = oldMonitor.ID

	err := s.provider.Update(newMonitor)
	if err != nil {
		return err
	}

	klog.Infof("monitor %q updated", newMonitor.Name)

	return nil
}

func (s *service) deleteMonitor(name string) error {
	err := s.provider.Delete(name)
	if err == models.ErrMonitorNotFound {
		klog.V(4).Infof("monitor %q is not present", name)
		return nil
	} else if err != nil {
		return err
	}

	klog.Infof("monitor %q deleted", name)

	return nil
}

func (s *service) buildMonitorModel(ing *v1beta1.Ingress) (*models.Monitor, error) {
	name, err := s.namer.Name(ing)
	if err != nil {
		return nil, err
	}

	url, err := ingress.BuildMonitorURL(ing)
	if err != nil {
		return nil, err
	}

	monitor := &models.Monitor{
		URL:         url,
		Name:        name,
		Annotations: ing.Annotations,
	}

	return monitor, nil
}

// GetProviderIPSourceRanges implements Service.
func (s *service) GetProviderIPSourceRanges(ing *v1beta1.Ingress) ([]string, error) {
	err := ingress.Validate(ing)
	if err != nil {
		klog.V(1).Infof(`ignoring unsupported ingress "%s/%s": %v`, ing.Namespace, ing.Name, err)
		return nil, nil
	}

	monitor, err := s.buildMonitorModel(ing)
	if err != nil {
		return nil, err
	}

	return s.provider.GetIPSourceRanges(monitor)
}
