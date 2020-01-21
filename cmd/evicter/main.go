package main

import (
	"flag"

	v1 "k8s.io/api/core/v1"
	policy "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/typed/policy/v1beta1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
)

func main() {
	var (
		kubeconfig                     = flag.String("kubeconfig", "", "absolute path to the kubeconfig file: `<home>/.kube/config`")
		master                         = flag.String("master", "", "master url")
		labelSelector                  = flag.String("label-selector", "tainted=true", "label selector to discover tainted pods")
		terminationGracePeriodSeconds  = flag.Int64("termination-grace-period", 30, "pod termination grace period in seconds")
		taintedIncubationPeriodSeconds = flag.Int64("incubation-period", 24*60*60, "time in seconds a tainted pod can run before eviction")
	)
	flag.Parse()
	flag.Set("logtostderr", "true") // glog: no disk log

	config, err := clientcmd.BuildConfigFromFlags(*master, *kubeconfig)
	if err != nil {
		klog.Fatal(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Fatal(err)
	}
	podListWatcher := cache.NewFilteredListWatchFromClient(clientset.CoreV1().RESTClient(), "pods", metav1.NamespaceDefault,
		func(options *metav1.ListOptions) {
			options.LabelSelector = *labelSelector
		})

	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	// Bind the workqueue to a cache with the help of an informer. This way we make sure that
	// whenever the cache is updated, the pod key is added to the workqueue.
	// Note that when we finally process the item from the workqueue, we might see a newer version
	// of the Pod than the version which was responsible for triggering the update.
	indexer, informer := cache.NewIndexerInformer(podListWatcher, &v1.Pod{}, 0, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if key, err := cache.MetaNamespaceKeyFunc(obj); err == nil {
				queue.Add(key)
			}
		},
		UpdateFunc: func(old interface{}, new interface{}) {
			if key, err := cache.MetaNamespaceKeyFunc(new); err == nil {
				queue.Add(key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			// IndexerInformer uses a delta queue, therefore for deletes we have to use this key function.
			if key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj); err == nil {
				queue.Add(key)
			}
		},
	}, cache.Indexers{})
	evicter := newPodEvicter(clientset.PolicyV1beta1(), *terminationGracePeriodSeconds)
	controller := NewController(queue, indexer, informer, evicter, *taintedIncubationPeriodSeconds)

	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(1, stop)

	// todo: watch sigterm
	// todo: recover panic to log
	select {}
}

type podEvicter struct {
	client               v1beta1.PolicyV1beta1Interface
	defaultDeleteOptions *metav1.DeleteOptions
}

func newPodEvicter(client v1beta1.PolicyV1beta1Interface, gracePeriodSeconds int64) *podEvicter {
	return &podEvicter{client: client, defaultDeleteOptions: &metav1.DeleteOptions{GracePeriodSeconds: &gracePeriodSeconds}}
}

func (p *podEvicter) Evict(namespace, podName string) error {
	return p.client.Evictions(namespace).Evict(newEviction(namespace, podName, p.defaultDeleteOptions))
}

func newEviction(ns, evictionName string, deleteOption *metav1.DeleteOptions) *policy.Eviction {
	return &policy.Eviction{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "Policy/v1beta1",
			Kind:       "Eviction",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      evictionName,
			Namespace: ns,
		},
		DeleteOptions: deleteOption,
	}
}
