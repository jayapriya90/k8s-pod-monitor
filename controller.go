package main

import (
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

// Controller struct encapsulates logging, client set, informer,
// worker queue, and handlers
type Controller struct {
	logger    *log.Entry
	clientset kubernetes.Interface
	queue     workqueue.RateLimitingInterface
	informer  cache.SharedIndexInformer
	handler   *PodHandler
}

// Run begins processing items, and will continue looping until a value is sent down stopCh
// Once stopCh is closed, it'll shutdown the workqueue and wait for workers to finish
// processing their current work items
func (c *Controller) Run(stopCh <-chan struct{}) {
	// log and exit during crash
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	c.logger.Info("Initiating controller")
	// run the informer in the background to start watching pod resources
	go c.informer.Run(stopCh)

	// perform initial synchronization to populate resources
	// wait for the cache to be synced before starting workers
	if !cache.WaitForCacheSync(stopCh, c.HasSynced) {
		utilruntime.HandleError(fmt.Errorf("Error syncing cache"))
		return
	}
	c.logger.Info("Cache sync completed successfully")

	// run the runWorker method every second with a stop channel
	wait.Until(c.runWorker, time.Second, stopCh)
}

// HasSynced returns true if the informer's store has been
// informed by at least one full LIST of the authoritative state (API Server)
// of the informer's object collection.
func (c *Controller) HasSynced() bool {
	return c.informer.HasSynced()
}

// runWorker executes the loop to process new items added to the queue
func (c *Controller) runWorker() {
	log.Info("Starting worker")

	// invoke processNextItem to fetch and consume the next change
	// to a watched resource
	for c.processNextItem() {
		log.Info("Processing next item from the queue")
	}

	log.Info("Worker completed processing items")
}

// processNextItem retrieves each queued item and takes the
// necessary handler action based off of if the item was
// created or deleted
func (c *Controller) processNextItem() bool {
	// fetch the next item from the queue to process.
	// if a shutdown is requested, return out of this to stop processing
	key, quit := c.queue.Get()
	if quit {
		return false
	}

	defer c.queue.Done(key)

	// cast key interface to string (format `namespace/name`)
	keyRaw := key.(string)

	// take the string key and get the pod object out of the indexer.
	// item will contain the object for the resource.
	// exists is a bool that'll indicate whether or not the
	// resource was created (true) or deleted (false)
	item, exists, err := c.informer.GetIndexer().GetByKey(keyRaw)
	if err != nil {
		utilruntime.HandleError(err)
	}

	// if the item doesn't exist, invoke ObjectDeleted Handler
	if !exists {
		c.logger.Infof("Object deletion detected: %s", keyRaw)
		c.handler.ObjectDeleted(keyRaw, item)
		// after processing, remove the key from the queue
		c.queue.Forget(key)
	} else {
		// if the object does exist, invoke ObjectCreated Handler
		c.logger.Infof("Object creation detected: %s", keyRaw)
		c.handler.ObjectCreated(keyRaw, item)
		// after processing, remove the key from the queue
		c.queue.Forget(key)
	}

	// keep the worker loop running by returning true
	return true
}
