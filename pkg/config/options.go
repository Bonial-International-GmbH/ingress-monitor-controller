package config

import (
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

const (
	// DefaultProvider is the provider that is used if none is specified
	// explicitly.
	DefaultProvider = ProviderNull

	// DefaultNameTemplate is the default template used for naming monitors.
	DefaultNameTemplate = "{{.Namespace}}-{{.IngressName}}"

	// DefaultResyncInterval is the default interval for the controller to
	// resync Ingress resources.
	DefaultResyncInterval = 5 * time.Minute
)

// Options holds the options that can be configured via cli flags.
type Options struct {
	ProviderConfigFile string
	Namespace          string
	ProviderName       string
	NameTemplate       string
	NoDelete           bool
	CreationDelay      time.Duration
	ResyncInterval     time.Duration
	ProviderConfig     ProviderConfig
}

// NewDefaultOptions creates a new *Options value with defaults set.
func NewDefaultOptions() *Options {
	return &Options{
		ResyncInterval: DefaultResyncInterval,
		ProviderName:   DefaultProvider,
		NameTemplate:   DefaultNameTemplate,
		ProviderConfig: NewDefaultProviderConfig(),
	}
}

// AddFlags adds cli flags for configurable options to the command.
func (o *Options) AddFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&o.NoDelete, "no-delete", o.NoDelete, "If set, delete actions will only be printed but not executed. This is useful for debugging.")
	cmd.Flags().DurationVar(&o.CreationDelay, "creation-delay", o.CreationDelay, "Duration to wait after an ingress is created before creating the monitor for it.")
	cmd.Flags().DurationVar(&o.ResyncInterval, "resync-interval", o.ResyncInterval, "Duration after which to recheck all ingresses.")
	cmd.Flags().StringVar(&o.NameTemplate, "name-template", o.NameTemplate, "The template to use for the monitor name.")
	cmd.Flags().StringVar(&o.Namespace, "namespace", o.Namespace, "Namespace to watch. If empty, all namespaces are watched.")
	cmd.Flags().StringVar(&o.ProviderConfigFile, "provider-config", o.ProviderConfigFile, "Location of the config file for the monitor providers.")
	cmd.Flags().StringVar(&o.ProviderName, "provider", o.ProviderName, "The provider to use for creating monitors.")
}

// Validate validates options.
func (o *Options) Validate() error {
	if o.CreationDelay < 0 {
		return errors.Errorf("--delete-after has to be greater than or equal to 0s")
	}

	if o.ResyncInterval < 0 {
		return errors.Errorf("--resync-interval has to be greater than or equal to 0s")
	}

	if o.NameTemplate == "" {
		return errors.Errorf("--name-template must not be empty")
	}

	if o.ProviderName == "" {
		return errors.Errorf("--provider must not be empty")
	}

	return nil
}
