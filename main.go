package main

import (
	"k8s.io/client-go/rest"
	"os"
	"os/signal"
	"syscall"

	log "github.com/Sirupsen/logrus"
	api_v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"
)

// GetKubernetesClient retrieves the Kubernetes cluster client from outside the
// cluster. If cannot be connected from out of cluster, fallback is to try
// from within the cluster
func GetKubernetesClient() (kubernetes.Interface, *rest.Config) {
	// construct the path to resolve to `~/.kube/config`
	kubeConfigPath := os.Getenv("HOME") + "/.kube/config"

	// create the config from the path
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		// fallback logic
		if config, err = rest.InClusterConfig(); err != nil {
			log.Fatalf("error creating client configuration: %v", err)
		}
	}

	// generate the client based off of the config
	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	log.Info("Successfully constructed k8s client")
	return kubeClient, config
}

func main() {
	// get kubernetes client
	client, config := GetKubernetesClient()

	// create the informer to watch all the pods
	informer := cache.NewSharedIndexInformer(
		// the ListWatch contains two functions that informer requires
		// ListFunc - to list resources and
		// WatchFunc - to watch resources
		&cache.ListWatch{
			ListFunc: func(options meta_v1.ListOptions) (runtime.Object, error) {
				// list pods from all namespaces
				return client.CoreV1().Pods(meta_v1.NamespaceAll).List(options)
			},
			WatchFunc: func(options meta_v1.ListOptions) (watch.Interface, error) {
				// watch pods from all namespaces
				return client.CoreV1().Pods(meta_v1.NamespaceAll).Watch(options)
			},
		},
		&api_v1.Pod{}, // the target type (Pod)
		0,             // no resync (period of 0)
		cache.Indexers{},
	)

	// create a queue to process the resources received by the informer
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	// add event handlers to add/update/delete resources
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		// Add new resources
		AddFunc: func(obj interface{}) {
			// create key for resource object in the format of 'namespace/name'
			key, err := cache.MetaNamespaceKeyFunc(obj)
			log.Infof("Add pod: %s", key)
			if err == nil {
				// add key to the queue for the handler to process
				queue.Add(key)
			}
		},
		// Update existing resources
		UpdateFunc: func(oldObj, newObj interface{}) {
			// create key for resource object in the format of 'namespace/name'
			key, err := cache.MetaNamespaceKeyFunc(newObj)
			log.Infof("Update pod: %s", key)
			if err == nil {
				// add key to the queue for the handler to process
				queue.Add(key)
			}
		},
		// Delete resources
		DeleteFunc: func(obj interface{}) {
			// check if the resource is already in DeletedFinalStateUnknown state where
			// a resource was deleted but it is still contained in the namespace/name index
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			log.Infof("Delete pod: %s", key)
			if err == nil {
				// add key to the queue for the handler to process
				queue.Add(key)
			}
		},
	})

	// construct the Controller object
	controller := Controller{
		logger:    log.NewEntry(log.New()),
		clientset: client,
		informer:  informer,
		queue:     queue,
		handler:   NewPodHandler(config),
	}

	// stopCh channel is to synchronize graceful shutdown
	stopCh := make(chan struct{})
	defer close(stopCh)

	// run the controller loop to process items
	go controller.Run(stopCh)

	// sigTerm channel is to handle OS signals for graceful shutdown/termination
	sigTerm := make(chan os.Signal, 1)
	signal.Notify(sigTerm, syscall.SIGTERM)
	signal.Notify(sigTerm, syscall.SIGINT)
	<-sigTerm
}
