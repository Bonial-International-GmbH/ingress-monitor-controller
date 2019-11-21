package null

import (
	"github.com/Bonial-International-GmbH/ingress-monitor-controller/pkg/models"
)

// Provider does not perform any monitor actions. This is useful for testing.
type Provider struct{}

// Create implements provider.Interface.
func (p *Provider) Create(_ *models.Monitor) error {
	return nil
}

// Create implements provider.Interface.
func (p *Provider) Get(_ string) (*models.Monitor, error) {
	return nil, models.ErrMonitorNotFound
}

// Create implements provider.Interface.
func (p *Provider) Update(_ *models.Monitor) error {
	return nil
}

// Create implements provider.Interface.
func (p *Provider) Delete(_ string) error {
	return nil
}
