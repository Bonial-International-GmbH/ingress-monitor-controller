package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Bonial-International-GmbH/ingress-monitor-controller/pkg/admission"
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	klog.Info("starting controller")

	ctrl := controller.New(client, svc, options)

	go handleSignals(cancel)
	go func() {
		defer cancel()
		err := ctrl.Run(ctx.Done())
		if err != nil {
			klog.Error(err)
		}
	}()

	if options.EnableAdmission {
		err := serveAdmissionWebhook(ctx, svc, options, cancel)
		if err != nil {
			return err
		}
	}

	<-ctx.Done()

	klog.Info("exiting")

	return nil
}

func handleSignals(cancelFunc func()) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, os.Interrupt)
	<-signals
	klog.Info("received signal, terminating...")
	cancelFunc()
}

func serveAdmissionWebhook(ctx context.Context, svc monitor.Service, options *config.Options, cancel func()) error {
	tlsConfig, err := options.TLSConfig()
	if err != nil {
		return errors.Wrapf(err, "loading tls config failed")
	}

	klog.Info("starting admission webhook")

	webhook := admission.NewWebhook(svc)

	mux := http.NewServeMux()
	mux.HandleFunc("/admit", admission.HandlerFunc(webhook.Admit))

	done := make(chan struct{})

	srv := &http.Server{
		Addr:      options.ListenAddr,
		Handler:   mux,
		TLSConfig: tlsConfig,
	}

	listenAndServeTLS(ctx, srv, done)
	cancel()

	<-done

	return nil
}

func listenAndServeTLS(ctx context.Context, srv *http.Server, doneCh chan<- struct{}) {
	go func() {
		defer close(doneCh)

		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		klog.Info("gracefully stopping server")

		err := srv.Shutdown(ctx)
		if err != nil {
			klog.Error(err)
		}
	}()

	klog.Infof("listening on %s\n", srv.Addr)

	err := srv.ListenAndServeTLS("", "")
	if err != http.ErrServerClosed {
		klog.Error(err)
	}
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
