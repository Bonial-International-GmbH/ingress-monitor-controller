package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/Bonial-International-GmbH/ingress-monitor-controller/pkg/config"
	"github.com/Bonial-International-GmbH/ingress-monitor-controller/pkg/controller"
	"github.com/Bonial-International-GmbH/ingress-monitor-controller/pkg/monitor"
	"github.com/imdario/mergo"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
)

func init() {
	klog.InitFlags(flag.CommandLine)
	flag.Set("logtostderr", "true")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
}

// NewRootCommand creates a new *cobra.Command that is used as the root command
// for ingress-monitor-controller.
func NewRootCommand() *cobra.Command {
	o := config.NewDefaultOptions()

	cmd := &cobra.Command{
		Use:           "ingress-monitor-controller",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := o.Validate()
			if err != nil {
				return err
			}

			return Run(o)
		},
	}

	o.AddFlags(cmd)

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

	client, err := newClient()
	if err != nil {
		return errors.Wrapf(err, "initializing kubernetes client failed")
	}

	klog.V(4).Infof("running with options: %+v", options)

	svc, err := monitor.NewService(options)
	if err != nil {
		return errors.Wrapf(err, "initializing monitor service failed")
	}

	c := controller.New(client, svc, options)

	ctx, cancel := context.WithCancel(context.Background())

	go handleSignals(cancel)

	return c.Run(ctx.Done())
}

func handleSignals(cancelFunc func()) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, os.Interrupt)
	<-signals
	klog.Info("received signal, terminating...")
	cancelFunc()
}

// newClient returns a new Kubernetes client with the default config.
func newClient() (kubernetes.Interface, error) {
	var kubeconfig string
	if _, err := os.Stat(clientcmd.RecommendedHomeFile); err == nil {
		kubeconfig = clientcmd.RecommendedHomeFile
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(config)
}
