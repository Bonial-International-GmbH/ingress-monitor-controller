package site24x7

import (
	"github.com/Bonial-International-GmbH/ingress-monitor-controller/pkg/config"
	"github.com/Bonial-International-GmbH/ingress-monitor-controller/pkg/models"
	site24x7 "github.com/Bonial-International-GmbH/site24x7-go"
)

// Provider manages Site24x7 website monitors.
type Provider struct {
	client  site24x7.Client
	config  config.Site24x7Config
	builder *builder
}

// NewProvider creates a new Site24x7 provider with given Site24x7Config.
func NewProvider(config config.Site24x7Config) *Provider {
	client := site24x7.New(site24x7.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		RefreshToken: config.RefreshToken,
	})

	return &Provider{
		client:  client,
		config:  config,
		builder: newBuilder(client, config.MonitorDefaults),
	}
}

// Create implements provider.Interface.
func (p *Provider) Create(model *models.Monitor) error {
	monitor, err := p.builder.FromModel(model)
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
func (p *Provider) Update(model *models.Monitor) error {
	monitor, err := p.builder.FromModel(model)
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
