package site24x7

import (
	"github.com/Bonial-International-GmbH/ingress-monitor-controller/pkg/config"
	"github.com/Bonial-International-GmbH/ingress-monitor-controller/pkg/models"
	site24x7 "github.com/Bonial-International-GmbH/site24x7-go"
	site24x7api "github.com/Bonial-International-GmbH/site24x7-go/api"
)

// Provider manages Site24x7 website monitors.
type Provider struct {
	client site24x7.Client
	config config.Site24x7Config
}

// NewProvider creates a new Site24x7 provider with given Site24x7Config.
func NewProvider(config config.Site24x7Config) *Provider {
	return &Provider{
		client: site24x7.New(site24x7.Config{
			ClientID:     config.ClientID,
			ClientSecret: config.ClientSecret,
			RefreshToken: config.RefreshToken,
		}),
		config: config,
	}
}

// Create implements provider.Interface.
func (p *Provider) Create(m *models.Monitor) error {
	monitor, err := p.buildMonitor(m)
	if err != nil {
		return err
	}

	_, err = p.client.Monitors().Create(monitor)

	return err
}

// Create implements provider.Interface.
func (p *Provider) Get(name string) (*models.Monitor, error) {
	monitors, err := p.client.Monitors().List()
	if err != nil {
		return nil, err
	}

	for _, monitor := range monitors {
		if monitor.DisplayName != name {
			continue
		}

		m := &models.Monitor{
			ID:   monitor.MonitorID,
			Name: monitor.DisplayName,
			URL:  monitor.Website,
		}

		return m, nil
	}

	return nil, models.ErrMonitorNotFound
}

// Create implements provider.Interface.
func (p *Provider) Update(m *models.Monitor) error {
	monitor, err := p.buildMonitor(m)
	if err != nil {
		return err
	}

	_, err = p.client.Monitors().Update(monitor)

	return err
}

// Create implements provider.Interface.
func (p *Provider) Delete(name string) error {
	monitor, err := p.Get(name)
	if err != nil {
		return err
	}

	return p.client.Monitors().Delete(monitor.ID)
}

func (p *Provider) buildMonitor(m *models.Monitor) (*site24x7api.Monitor, error) {
	return newBuilder(p.client, p.config.MonitorDefaults).
		withDefaulters(p.getDefaulters()...).
		withModel(m).
		build()
}

func (p *Provider) getDefaulters() []defaulter {
	var defaulters []defaulter

	if p.config.MonitorDefaults.AutoLocationProfile {
		defaulters = append(defaulters, withDefaultLocationProfile)
	}

	if p.config.MonitorDefaults.AutoMonitorGroup {
		defaulters = append(defaulters, withDefaultMonitorGroup)
	}

	if p.config.MonitorDefaults.AutoNotificationProfile {
		defaulters = append(defaulters, withDefaultNotificationProfile)
	}

	if p.config.MonitorDefaults.AutoThresholdProfile {
		defaulters = append(defaulters, withDefaultThresholdProfile)
	}

	if p.config.MonitorDefaults.AutoUserGroup {
		defaulters = append(defaulters, withDefaultUserGroup)
	}

	return defaulters
}
