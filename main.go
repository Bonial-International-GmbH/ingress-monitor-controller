package main

import (
	"flag"

	"github.com/Bonial-International-GmbH/ingress-monitor-controller/pkg/admission"
	"github.com/Bonial-International-GmbH/ingress-monitor-controller/pkg/config"
	"github.com/Bonial-International-GmbH/ingress-monitor-controller/pkg/controller"
	"github.com/Bonial-International-GmbH/ingress-monitor-controller/pkg/monitor"
	"github.com/imdario/mergo"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/klog"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	restconfig "sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

func init() {
	klog.InitFlags(flag.CommandLine)
	flag.Set("logtostderr", "true")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
}

// NewRootCommand creates a new *cobra.Command that is used as the root command
// for ingress-monitor-controller.
func NewRootCommand() *cobra.Command {
	options := config.NewDefaultOptions()

	cmd := &cobra.Command{
		Use:           "ingress-monitor-controller",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := options.Validate()
			if err != nil {
				return err
			}

			return Run(options)
		},
	}

	options.AddFlags(cmd)

	return cmd
}

func main() {
	cmd := NewRootCommand()

	cmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)

	if err := cmd.Execute(); err != nil {
		klog.Fatal(err)
	}
}

// Run sets up that controller and initiates the controller loop.
func Run(options *config.Options) error {
	ctrl.SetLogger(zap.Logger(true))

	if options.ProviderConfigFile != "" {
		klog.V(1).Infof("loading provider config from %s", options.ProviderConfigFile)

		providerConfig, err := config.ReadProviderConfig(options.ProviderConfigFile)
		if err != nil {
			return errors.Wrapf(err, "failed to load provider config from file")
		}

		err = mergo.Merge(&options.ProviderConfig, providerConfig, mergo.WithOverride)
		if err != nil {
			return errors.Wrapf(err, "failed to merge provider configs")
		}
	}

	klog.V(4).Infof("running with options: %+v", options)

	mgr, err := manager.New(restconfig.GetConfigOrDie(), manager.Options{
		CertDir: options.TLSCertDir,
	})
	if err != nil {
		return errors.Wrapf(err, "failed to create controller manager")
	}

	svc, err := monitor.NewService(options)
	if err != nil {
		return errors.Wrapf(err, "failed to initialize monitor service")
	}

	reconciler := controller.NewIngressReconciler(mgr.GetClient(), svc, options)

	err = builder.
		ControllerManagedBy(mgr).
		For(&v1beta1.Ingress{}).
		Complete(reconciler)
	if err != nil {
		return errors.Wrapf(err, "failed to create controller")
	}

	if options.EnableAdmission {
		whs := mgr.GetWebhookServer()

		whs.Register("/admit", &webhook.Admission{
			Handler: admission.NewIngressHandler(svc),
		})
	}

	err = mgr.Start(signals.SetupSignalHandler())
	if err != nil {
		return errors.Wrapf(err, "unable to run manager")
	}

	return nil
}
