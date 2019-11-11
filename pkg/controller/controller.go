package controller

import (
	"time"

	"github.com/Bonial-International-GmbH/ingress-monitor-controller/pkg/config"
	"github.com/Bonial-International-GmbH/ingress-monitor-controller/pkg/monitor"
	"github.com/pkg/errors"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/fields"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
)

// Controller controls monitors for ingress resources.
type Controller struct {
	client        kubernetes.Interface
	queue         workqueue.RateLimitingInterface
	informer      cache.SharedIndexInformer
	creationDelay time.Duration
	service       monitor.Service
}

// New creates a new *Controller. The controller will watch for ingress changes
// and will create, update or delete monitors as necessary.
func New(client kubernetes.Interface, svc monitor.Service, options *config.Options) *Controller {
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	listWatcher := cache.NewListWatchFromClient(client.ExtensionsV1beta1().RESTClient(), "ingresses", options.Namespace, fields.Everything())
	indexers := cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}
	informer := cache.NewSharedIndexInformer(listWatcher, &v1beta1.Ingress{}, options.ResyncInterval, indexers)

	c := &Controller{
		client:        client,
		creationDelay: options.CreationDelay,
		service:       svc,
		queue:         queue,
		informer:      informer,
	}

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.onIngressAdded,
		UpdateFunc: c.onIngressUpdated,
		DeleteFunc: c.onIngressDeleted,
	})

	return c
}

// Run starts the controller and will block until stopCh is closed.
func (c *Controller) Run(stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()

	defer c.queue.ShutDown()

	klog.Info("starting controller")

	go c.informer.Run(stopCh)

	// Wait for all involved caches to be synced, before processing items from the queue is started.
	if !cache.WaitForCacheSync(stopCh, c.informer.HasSynced) {
		return errors.Errorf("timed out waiting for caches to sync")
	}

	go wait.Until(c.worker, time.Second, stopCh)

	klog.Info("watching for changes")

	<-stopCh
	klog.Info("stopping controller")

	return nil
}

func (c *Controller) onIngressAdded(obj interface{}) {
	ingress := obj.(*v1beta1.Ingress)

	klog.V(1).Infof(`received add event for ingress "%s/%s"`, ingress.Namespace, ingress.Name)

	delay := time.Until(ingress.CreationTimestamp.Add(c.creationDelay))

	c.queue.AddAfter(updateEvent{
		oldIngress: ingress,
		newIngress: ingress,
	}, delay)
}

func (c *Controller) onIngressUpdated(old, new interface{}) {
	oldIngress := old.(*v1beta1.Ingress)
	newIngress := new.(*v1beta1.Ingress)

	klog.V(1).Infof(`received update event for ingress "%s/%s"`, newIngress.Namespace, newIngress.Name)

	delay := time.Until(newIngress.CreationTimestamp.Add(c.creationDelay))

	c.queue.AddAfter(updateEvent{
		oldIngress: oldIngress,
		newIngress: newIngress,
	}, delay)
}

func (c *Controller) onIngressDeleted(obj interface{}) {
	ingress := obj.(*v1beta1.Ingress)

	klog.V(1).Infof(`received delete event for ingress "%s/%s"`, ingress.Namespace, ingress.Name)

	c.queue.Add(deleteEvent{ingress: ingress})
}

// worker consumes the workqueue and handles every event obtained from the
// queue.
func (c *Controller) worker() {
	workFunc := func() bool {
		key, quit := c.queue.Get()
		if quit {
			return true
		}
		defer c.queue.Done(key)

		event := key.(event)

		err := event.handle(c.service)
		c.handleError(err, event)
		return false
	}

	for {
		if quit := workFunc(); quit {
			return
		}
	}
}

// handleError ensures that key gets requeued for a number of times in case err
// is non-nil until it gives up and logs the final error.
func (c *Controller) handleError(err error, key interface{}) {
	if err == nil {
		c.queue.Forget(key)
		return
	}

	if c.queue.NumRequeues(key) < 5 {
		klog.V(4).Infof("requeuing key %v for sync due to error: %v", key, err)
		c.queue.AddRateLimited(key)
		return
	}

	c.queue.Forget(key)
	utilruntime.HandleError(err)
	klog.V(1).Infof("dropping key %v out of the queue due to error: %v", key, err)
}
