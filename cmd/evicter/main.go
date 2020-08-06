package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cruise-automation/k-rail/resource"
	"github.com/golang/glog"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	policy "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/kubernetes/typed/policy/v1beta1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/client-go/tools/record"
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
		instanceID                     = flag.String("instance", uuid.New().String(), "the unique holder identity. used for leader lock")
		leaseLockName                  = flag.String("lease-lock-name", "k-rail-evicter-lock", "the lease lock resource name")
		leaseLockNamespace             = flag.String("lease-lock-namespace", "k-rail", "the lease lock resource namespace")
		probeServerAddress             = flag.String("probe-listen-address", ":8080", "server address for healthz/readiness server")
	)
	flag.Parse()
	flag.Set("logtostderr", "true") // glog: no disk log

	defer func() {
		if err := recover(); err != nil {
			klog.Fatal(err)
		}
	}()

	config, err := clientcmd.BuildConfigFromFlags(*master, *kubeconfig)
	if err != nil {
		klog.Fatal(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	watchSigTerm(cancel)

	probeServer := newProbeServer(*probeServerAddress)
	go func() {
		klog.Fatal(probeServer.ListenAndServe())
	}()

	br := record.NewBroadcaster()
	br.StartRecordingToSink(&corev1.EventSinkImpl{Interface: clientset.CoreV1().Events(metav1.NamespaceAll)})
	br.StartLogging(glog.Infof)
	defer br.Shutdown()
	eventRecorder := br.NewRecorder(scheme.Scheme, v1.EventSource{Component: "k-rail-evicter"})

	lock := &resourcelock.LeaseLock{
		LeaseMeta: metav1.ObjectMeta{
			Name:      *leaseLockName,
			Namespace: *leaseLockNamespace,
		},
		Client: clientset.CoordinationV1(),
		LockConfig: resourcelock.ResourceLockConfig{
			Identity:      *instanceID,
			EventRecorder: eventRecorder,
		},
	}

	// start the leader election code loop
	leaderelection.RunOrDie(ctx, leaderelection.LeaderElectionConfig{
		Lock:            lock,
		ReleaseOnCancel: true,
		LeaseDuration:   60 * time.Second,
		RenewDeadline:   15 * time.Second,
		RetryPeriod:     5 * time.Second,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(ctx context.Context) {
				klog.Infof("Starting leader: %s", *instanceID)
				podListWatcher := cache.NewFilteredListWatchFromClient(clientset.CoreV1().RESTClient(), "pods", metav1.NamespaceAll,
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

				evicter := newPodEvicter(clientset.PolicyV1beta1(), eventRecorder, *terminationGracePeriodSeconds)
				controller := NewController(queue, indexer, informer, evicter, *taintedIncubationPeriodSeconds)
				controller.Run(1, ctx.Done())
			},
			OnStoppedLeading: func() {
				// we can do cleanup here
				cancel()
				ctx, _ := context.WithTimeout(ctx, 100*time.Millisecond)
				_ = probeServer.Shutdown(ctx)
				klog.Infof("Leader lost: %s", *instanceID)
				os.Exit(0)
			},
			OnNewLeader: func(identity string) {
				if identity == *instanceID {
					return
				}
				klog.Infof("New leader elected: %s", identity)
			},
		},
	})
}

func newProbeServer(listenAddr string) http.Server {
	okHandler := func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
	http.HandleFunc("/healthz", okHandler)
	http.HandleFunc("/readyness", okHandler)
	return http.Server{Addr: listenAddr, Handler: nil}
}

func watchSigTerm(cancel context.CancelFunc) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ch
		klog.Info("Received termination, signaling shutdown")
		cancel()
	}()
}

type podEvicter struct {
	client               v1beta1.PolicyV1beta1Interface
	eventRecorder        record.EventRecorder
	defaultDeleteOptions *metav1.DeleteOptions
}

func newPodEvicter(client v1beta1.PolicyV1beta1Interface, recorder record.EventRecorder, gracePeriodSeconds int64) *podEvicter {
	return &podEvicter{
		client:               client,
		eventRecorder:        recorder,
		defaultDeleteOptions: &metav1.DeleteOptions{GracePeriodSeconds: &gracePeriodSeconds},
	}
}

// Evict calls the k8s api to evict the given pod. Reason and notes are stored with the audit event.
func (p *podEvicter) Evict(pod *v1.Pod, reason, msg string) error {
	err := p.client.Evictions(pod.Namespace).Evict(newEviction(pod, p.defaultDeleteOptions))
	klog.Infof("Evicted pod %q (UID: %s)", resource.GetResourceName(pod.ObjectMeta), pod.UID)
	if err != nil {
		return errors.Wrap(err, "eviction")
	}
	p.eventRecorder.Eventf(pod, v1.EventTypeNormal, reason, msg)
	return nil
}

func newEviction(pod *v1.Pod, deleteOption *metav1.DeleteOptions) *policy.Eviction {
	return &policy.Eviction{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "Policy/v1beta1",
			Kind:       "Eviction",
		},
		ObjectMeta:    pod.ObjectMeta,
		DeleteOptions: deleteOption,
	}
}
