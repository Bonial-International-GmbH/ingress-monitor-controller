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
		if ingress.Annotations[config.AnnotationEnabled] == "true" {
			createAfter := time.Until(ingress.CreationTimestamp.Add(r.creationDelay))

			// If a creation delay was configured, we will requeue the
			// reconciliation until after the creation delay passed.
			if createAfter > 0 {
				return reconcile.Result{RequeueAfter: createAfter}, nil
			}

			err = r.handleCreateOrUpdate(ingress)
		} else {
			err = r.monitorService.DeleteMonitor(ingress)
		}
	}

	return reconcile.Result{}, err
}

func (r *IngressReconciler) handleCreateOrUpdate(ingress *v1beta1.Ingress) error {
	updated, err := r.reconcileAnnotations(ingress)
	if err != nil || updated {
		// In case of an error we return it here to force requeuing of the
		// reconciliation request. If the ingress was updated, we return
		// here as well because the update will cause the creation of a new
		// ingress update event which will be consumed by Reconcile and we
		// want to avoid duplicate execution of the EnsureMonitor logic. This
		// is an optimization to avoid unnecessary API calls to the monitor
		// provider.
		return err
	}

	return r.monitorService.EnsureMonitor(ingress)
}

// reconcileAnnotations reconciles the ingress annotations, that is, it may
// update the nginx.ingress.kubernetes.io/whitelist-source-range annotation
// with ip source ranges of the monitor provider. If annotations were updated,
// it will update the ingress object on the cluster and return true and the
// first return value. The will effectively cause the creation of a new ingress
// update event which is then picked up by the reconciler.
func (r *IngressReconciler) reconcileAnnotations(ingress *v1beta1.Ingress) (updated bool, err error) {
	ingressCopy := ingress.DeepCopy()

	updated, err = r.monitorService.AnnotateIngress(ingressCopy)
	if err != nil || !updated {
		return false, err
	}

	err = r.Update(context.Background(), ingressCopy)
	if err != nil {
		return false, err
	}

	return true, nil
}
