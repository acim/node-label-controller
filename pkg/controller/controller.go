package controller

import (
	"fmt"
	"time"

	api "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	informers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typed "k8s.io/client-go/kubernetes/typed/core/v1"
	listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
)

const controllerName = "node-label-controller"

// Controller is Kubernetes controller implementation for node labeling.
type Controller struct {
	kubeClient  kubernetes.Interface
	nodesLister listers.NodeLister
	nodesSynced cache.InformerSynced
	workqueue   workqueue.RateLimitingInterface
	recorder    record.EventRecorder
	nodeMatcher func(*api.Node) bool
	labelKey    string
	labelValue  string
}

// NewController creates new node label controller.
func NewController(kubeClient kubernetes.Interface, nodesInformer informers.NodeInformer, nodeMatcher func(*api.Node) bool, labelKey, labelValue string) *Controller {
	klog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&typed.EventSinkImpl{Interface: kubeClient.CoreV1().Events("")})

	c := &Controller{
		kubeClient:  kubeClient,
		nodesLister: nodesInformer.Lister(),
		nodesSynced: nodesInformer.Informer().HasSynced,
		workqueue:   workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), controllerName),
		recorder:    eventBroadcaster.NewRecorder(scheme.Scheme, api.EventSource{Component: controllerName}),
		nodeMatcher: nodeMatcher,
		labelKey:    labelKey,
		labelValue:  labelValue,
	}

	nodesInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: c.handleNode,
		UpdateFunc: func(old, new interface{}) {
			c.handleNode(new)
		},
	})

	return c
}

// Run listens for nodes changes and acts upon.
func (c *Controller) Run(concurrency int, stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()
	defer c.workqueue.ShutDown()

	klog.Infof("Starting %s", controllerName)

	klog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.nodesSynced); !ok {
		return fmt.Errorf("failed waiting for caches to sync")
	}

	klog.Infof("Starting %d workers", concurrency)

	for i := 0; i < concurrency; i++ {
		go wait.Until(c.run, time.Second, stopCh)
	}

	<-stopCh
	klog.Info("Stopping workers")

	return nil
}

func (c *Controller) handleNode(o interface{}) {
	var n *api.Node
	var ok bool

	if n, ok = o.(*api.Node); !ok {
		runtime.HandleError(fmt.Errorf("expected node in handleNode, got %T", o))
		return
	}

	// Label already exists
	if _, ok := n.Labels[c.labelKey]; ok {
		return
	}

	if c.nodeMatcher(n) {
		key, err := cache.MetaNamespaceKeyFunc(o)
		if err == nil {
			c.workqueue.Add(key)
		}
	}
}

func (c *Controller) run() {
	for func() bool {
		o, shutdown := c.workqueue.Get()

		if shutdown {
			return false
		}

		err := func(o interface{}) error {
			defer c.workqueue.Done(o)

			var key string
			var ok bool
			if key, ok = o.(string); !ok {
				c.workqueue.Forget(o)
				runtime.HandleError(fmt.Errorf("expected string in workqueue, got %T", o))
				return nil
			}

			if err := c.syncHandler(key); err != nil {
				c.workqueue.AddRateLimited(key)
				return fmt.Errorf("error syncing %s: %w, requeuing", key, err)
			}

			c.workqueue.Forget(o)
			klog.Infof("Successfully synced %s", key)
			return nil
		}(o)

		if err != nil {
			runtime.HandleError(err)
			return true
		}

		return true
	}() {
	}
}

func (c *Controller) syncHandler(key string) error {
	node, err := c.nodesLister.Get(key)
	if err != nil {
		if errors.IsNotFound(err) {
			runtime.HandleError(fmt.Errorf("node %s not found in workqueue", key))
			return nil
		}

		return err
	}

	err = c.labelNode(node)
	if err != nil {
		return err
	}

	c.recorder.Event(node, api.EventTypeNormal, "Labeled", fmt.Sprintf("Node %s labeled successfully", node.Name))
	return nil
}

func (c *Controller) labelNode(n *api.Node) error {
	n = n.DeepCopy()

	n.Labels[c.labelKey] = c.labelValue

	_, err := c.kubeClient.CoreV1().Nodes().Update(n)

	if err != nil {
		runtime.HandleError(fmt.Errorf("failed labeling node: %v", err))
	}

	return nil
}
