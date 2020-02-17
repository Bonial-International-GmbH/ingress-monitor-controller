package controller

import (
	"context"
	"time"

	"github.com/Bonial-International-GmbH/ingress-monitor-controller/pkg/config"
	"github.com/Bonial-International-GmbH/ingress-monitor-controller/pkg/monitor"
	"k8s.io/api/extensions/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// IngressReconciler reconciles ingresses to their desired state.
type IngressReconciler struct {
	client.Client

	monitorService monitor.Service
	creationDelay  time.Duration
}

// NewIngressReconciler creates a new *IngressReconciler.
func NewIngressReconciler(client client.Client, monitorService monitor.Service, options *config.Options) *IngressReconciler {
	return &IngressReconciler{
		Client:         client,
		monitorService: monitorService,
		creationDelay:  options.CreationDelay,
	}
}

// Reconcile creates, updates or deletes ingress monitors whenever an ingress
// changes. It implements reconcile.Reconciler.
func (r *IngressReconciler) Reconcile(req reconcile.Request) (reconcile.Result, error) {
	ingress := &v1beta1.Ingress{}

	err := r.Get(context.Background(), req.NamespacedName, ingress)
	if apierrors.IsNotFound(err) {
		// The ingress was deleted. Construct a metadata-only ingress object
		// just for monitor deletion.
		ingress = &v1beta1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:      req.NamespacedName.Name,
				Namespace: req.NamespacedName.Namespace,
			},
		}

		err = r.monitorService.DeleteMonitor(ingress)
	} else if err == nil {
		createAfter := time.Until(ingress.CreationTimestamp.Add(r.creationDelay))

		// If a creation delay was configured, we will requeue the
		// reconciliation until after the creation delay passed.
		if createAfter > 0 {
			return reconcile.Result{RequeueAfter: createAfter}, nil
		}

		annotations := config.Annotations(ingress.Annotations)

		if annotations.BoolValue(config.AnnotationEnabled) {
			err = r.monitorService.EnsureMonitor(ingress)
		} else {
			err = r.monitorService.DeleteMonitor(ingress)
		}
	}

	return reconcile.Result{}, err
}
