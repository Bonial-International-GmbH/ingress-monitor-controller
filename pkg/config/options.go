package config

import (
	"crypto/tls"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

const (
	// DefaultProvider is the provider that is used if none is specified
	// explicitly.
	DefaultProvider = ProviderSite24x7

	// DefaultNameTemplate is the default template used for naming monitors.
	DefaultNameTemplate = "{{.Namespace}}-{{.IngressName}}"

	// DefaultResyncInterval is the default interval for the controller to
	// resync Ingress resources.
	DefaultResyncInterval = 1 * time.Hour

	// DefaultListenAddr is the default listen address for the admission
	// controller.
	DefaultListenAddr = "0.0.0.0:443"
)

// Options holds the options that can be configured via cli flags.
type Options struct {
	ProviderConfigFile string
	Namespace          string
	ProviderName       string
	NameTemplate       string
	TLSCertFile        string
	TLSPrivateKeyFile  string
	ListenAddr         string
	NoDelete           bool
	EnableAdmission    bool
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
		ListenAddr:     DefaultListenAddr,
		ProviderConfig: NewDefaultProviderConfig(),
	}
}

// AddFlags adds cli flags for configurable options to the command.
func (o *Options) AddFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&o.NoDelete, "no-delete", o.NoDelete, "If set, monitors will not be deleted if the ingress is deleted.")
	cmd.Flags().DurationVar(&o.CreationDelay, "creation-delay", o.CreationDelay, "Duration to wait after an ingress is created before creating the monitor for it.")
	cmd.Flags().DurationVar(&o.ResyncInterval, "resync-interval", o.ResyncInterval, "Duration after which to recheck all ingresses.")
	cmd.Flags().StringVar(&o.NameTemplate, "name-template", o.NameTemplate, "The template to use for the monitor name. Valid fields are: .IngressName, .Namespace.")
	cmd.Flags().StringVar(&o.Namespace, "namespace", o.Namespace, "Namespace to watch. If empty, all namespaces are watched.")
	cmd.Flags().StringVar(&o.ProviderConfigFile, "provider-config", o.ProviderConfigFile, "Location of the config file for the monitor providers.")
	cmd.Flags().StringVar(&o.ProviderName, "provider", o.ProviderName, "The provider to use for creating monitors.")
	cmd.Flags().StringVar(&o.TLSCertFile, "tls-cert-file", o.TLSCertFile, "File containing the default x509 Certificate for HTTPS. (CA cert, if any, concatenated after server cert).")
	cmd.Flags().StringVar(&o.TLSPrivateKeyFile, "tls-private-key-file", o.TLSPrivateKeyFile, "File containing the default x509 private key matching --tls-cert-file.")
	cmd.Flags().StringVar(&o.ListenAddr, "listen-addr", o.ListenAddr, "The listen address for the admission controller.")
	cmd.Flags().BoolVar(&o.EnableAdmission, "enable-admission", o.EnableAdmission, "If set, an admission controller will be launched listening at the configured address. The admission controller will automatically add IP whitelistings for the provider source ranges on the ingress objects if needed.")
}

// Validate validates options.
func (o *Options) Validate() error {
	if o.CreationDelay < 0 {
		return errors.Errorf("--creation-delay has to be greater than or equal to 0s")
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

// TLSConfig creates a new *tls.Config from options.
func (o *Options) TLSConfig() (*tls.Config, error) {
	sCert, err := tls.LoadX509KeyPair(o.TLSCertFile, o.TLSPrivateKeyFile)
	if err != nil {
		return nil, err
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{sCert},
	}

	return config, nil
}
