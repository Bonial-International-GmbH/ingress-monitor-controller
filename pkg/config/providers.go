package config

import (
	"io/ioutil"
	"os"

	site24x7api "github.com/Bonial-International-GmbH/site24x7-go/api"
	"sigs.k8s.io/yaml"
)

const (
	// ProviderSite24x7 uses Site24x7 for managing ingress monitors.
	ProviderSite24x7 = "site24x7"

	// ProviderNull does nothing but log create/update/delete monitor events.
	// This is intended for testing purposes only.
	ProviderNull = "null"
)

// ProviderConfig contains the configuration for all supported monitor
// providers.
type ProviderConfig struct {
	Site24x7 Site24x7Config `json:"site24x7"`
}

// Site24x7Config is the configration for the Site24x7 website monitor
// provider.
type Site24x7Config struct {
	// ClientID is the OAuth2 client ID provided by Site24x7.
	ClientID string `json:"clientID"`

	// ClientSecret is the OAuth2 client secret provided by Site24x7.
	ClientSecret string `json:"clientSecret"`

	// RefreshToken is the OAuth2 refresh token provided by Site24x7.
	RefreshToken string `json:"refreshToken"`

	// MonitorDefaults contain defaults that apply to all monitors. The
	// defaults can be overridden explicitly for each monitor via ingress
	// annotations (see annotations.go for all available annotations).
	MonitorDefaults Site24x7MonitorDefaults `json:"monitorDefaults"`
}

// Site24x7MonitorDefaults define the monitor defaults that are used for each
// monitor if not overridden explicitly via ingress annotations.
type Site24x7MonitorDefaults struct {
	AutoLocationProfile     bool                    `json:"autoLocationProfile"`
	AutoNotificationProfile bool                    `json:"autoNotificationProfile"`
	AutoThresholdProfile    bool                    `json:"autoThresholdProfile"`
	AutoMonitorGroup        bool                    `json:"autoMonitorGroup"`
	AutoUserGroup           bool                    `json:"autoUserGroup"`
	MatchCase               bool                    `json:"matchCase"`
	Timeout                 int                     `json:"timeout"`
	CheckFrequency          string                  `json:"checkFrequency"`
	HTTPMethod              string                  `json:"httpMethod"`
	AuthUser                string                  `json:"authUser"`
	AuthPass                string                  `json:"authPass"`
	UserAgent               string                  `json:"userAgent"`
	UseNameServer           bool                    `json:"useNameServer"`
	LocationProfileID       string                  `json:"locationProfileID"`
	NotificationProfileID   string                  `json:"notificationProfileID"`
	ThresholdProfileID      string                  `json:"thresholdProfileID"`
	MonitorGroupIDs         []string                `json:"monitorGroupIDs"`
	UserGroupIDs            []string                `json:"userGroupIDs"`
	Actions                 []site24x7api.ActionRef `json:"actions"`
	CustomHeaders           []site24x7api.Header    `json:"customHeaders"`
}

// NewDefaultProviderConfig creates a new default provider config.
func NewDefaultProviderConfig() ProviderConfig {
	return ProviderConfig{
		Site24x7: Site24x7Config{
			ClientID:     os.Getenv("SITE24X7_CLIENT_ID"),
			ClientSecret: os.Getenv("SITE24X7_CLIENT_SECRET"),
			RefreshToken: os.Getenv("SITE24X7_REFRESH_TOKEN"),
			MonitorDefaults: Site24x7MonitorDefaults{
				AutoLocationProfile:     true,
				AutoNotificationProfile: true,
				AutoThresholdProfile:    true,
				AutoMonitorGroup:        true,
				AutoUserGroup:           true,
				CheckFrequency:          "1",
				HTTPMethod:              "G",
				Timeout:                 10,
				UseNameServer:           true,
				CustomHeaders:           []site24x7api.Header{},
				Actions:                 []site24x7api.ActionRef{},
			},
		},
	}
}

// ReadProviderConfig reads the provider configuration from given file.
func ReadProviderConfig(filename string) (*ProviderConfig, error) {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config ProviderConfig

	err = yaml.Unmarshal(buf, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
