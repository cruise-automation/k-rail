package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
)

type podProvisioner interface {
	Evict(pod *v1.Pod, reason string) error
}

type Controller struct {
	podStore                cache.Indexer
	queue                   workqueue.RateLimitingInterface
	informer                cache.Controller
	podProvisioner          podProvisioner
	incubationPeriodSeconds time.Duration
	started                 time.Time
}

func NewController(queue workqueue.RateLimitingInterface, indexer cache.Indexer, informer cache.Controller, podProvisioner podProvisioner, incubationPeriodSeconds int64) *Controller {
	return &Controller{
		informer:                informer,
		podStore:                indexer,
		queue:                   queue,
		podProvisioner:          podProvisioner,
		incubationPeriodSeconds: time.Duration(incubationPeriodSeconds) * time.Second,
	}
}

func (c *Controller) processNextItem() bool {
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	defer c.queue.Done(key)

	err := c.evictPod(key.(string))
	c.handleErr(err, key)
	return true
}

const (
	annotationPreventEviction = "k-rails/tainted-prevent-eviction"
	annotationTimestamp       = "k-rails/tainted-timestamp"
	annotationReason          = "k-rails/tainted-reason"
)
const defaultEvictionReason = "exec"

// evictPod is the business logic of the controller. it checks the the eviction rules and conditions before calling the pod provisioner.
func (c *Controller) evictPod(key string) error {
	obj, exists, err := c.podStore.GetByKey(key)
	switch {
	case err != nil:
		return err
	case !exists:
		return nil
	}
	pod, ok := obj.(*v1.Pod)
	if !ok {
		return fmt.Errorf("unsupported type: %T", obj)
	}
	if !canEvict(pod, c.incubationPeriodSeconds) {
		return nil
	}

	reason, ok := pod.Annotations[annotationReason]
	if !ok || reason == "" {
		reason = defaultEvictionReason
	}

	return c.podProvisioner.Evict(pod, reason)
}

func canEvict(pod *v1.Pod, incubationPeriod time.Duration) bool {
	if pod == nil {
		return false
	}
	val, ok := pod.Annotations[annotationPreventEviction]
	if ok {
		if val == "yes" || val == "true" {
			return false
		}
	}

	val, ok = pod.Annotations[annotationTimestamp]
	if ok {
		i, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			// todo: log
			return true
		}
		timestamp := time.Unix(i, 0)
		if time.Since(timestamp) < incubationPeriod {
			return false
		}
	}
	return true
}

const maxWorkerRetries = 5

// handleErr checks if an error happened and makes sure we will retry later.
func (c *Controller) handleErr(err error, key interface{}) {
	if err == nil {
		// Forget about the #AddRateLimited history of the key on every successful synchronization.
		// This ensures that future processing of updates for this key is not delayed because of
		// an outdated error history.
		c.queue.Forget(key)
		return
	}

	// This controller retries 5 times if something goes wrong. After that, it stops trying.
	if c.queue.NumRequeues(key) < maxWorkerRetries {
		klog.Infof("Error syncing pod %v: %v", key, err)

		// Re-enqueue the key rate limited. Based on the rate limiter on the
		// queue and the re-enqueue history, the key will be processed later again.
		c.queue.AddRateLimited(key)
		return
	}

	c.queue.Forget(key)
	// Report to an external entity that, even after several retries, we could not successfully process this key
	runtime.HandleError(err)
	klog.Infof("Dropping pod %q out of the queue: %v", key, err)
}

const reconciliationTick = 30 * time.Second
const startupGracePeriod = 90 * time.Second

func (c *Controller) Run(threadiness int, stopCh chan struct{}) {
	defer runtime.HandleCrash()

	// Let the workers stop when we are done
	defer c.queue.ShutDown()
	klog.Info("Starting Pod controller")

	go c.informer.Run(stopCh)

	// Wait for all involved caches to be synced, before processing items from the queue is started
	if !cache.WaitForCacheSync(stopCh, c.informer.HasSynced) {
		runtime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
		return
	}

	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	wait.Until(func() {
		if time.Since(c.started) < startupGracePeriod {
			return
		}
		if err := c.doReconciliation(); err != nil {
			klog.Errorf("Reconciliation failed: %s", err)
		}
	}, reconciliationTick, stopCh)

	klog.Info("Stopping Pod controller")
}

func (c *Controller) runWorker() {
	for c.processNextItem() {
	}
}

func (c *Controller) doReconciliation() error {
	klog.Info("Reconciliation started")
	for _, key := range c.podStore.ListKeys() {
		if err := c.evictPod(key); err != nil {
			return errors.Wrapf(err, "pod %q", key)
		}
	}
	klog.Info("Reconciliation completed")
	return nil
}
