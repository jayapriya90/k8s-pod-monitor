package main

import (
	apiextension "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/jayapriya90/k8s-pod-monitor/v1alpha1"
	core_v1 "k8s.io/api/core/v1"
)

// PodHandler is a sample implementation of Handler
type PodHandler struct {
	crdClient   *v1alpha1.PodMonitorV1Alpha1Client
	podMonitor *v1alpha1.PodMonitor
	podsPending map[string]bool
	podsRunning map[string]bool
}

func createCRDClient(config *rest.Config) (*v1alpha1.PodMonitorV1Alpha1Client, *v1alpha1.PodMonitor, error) {
	client, err := apiextension.NewForConfig(config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Create CRD
	err = v1alpha1.CreateCRD(client)
	if err != nil {
		log.Fatalf("Failed to create crd: %v", err)
	}

	// Wait for the CRD to be created before using it
	time.Sleep(3 * time.Second)

	// Create a new clientset which includes PodMonitor CRD schema
	crdclient, err := v1alpha1.NewClient(config)
	if err != nil {
		panic(err)
	}

	podMonitor := v1alpha1.PodMonitor{
		TypeMeta: v1.TypeMeta{Kind: "PodMonitor", APIVersion: "v1alpha1"},
		ObjectMeta: v1.ObjectMeta{Name: "pod-monitor"},
		Spec: v1alpha1.PodMonitorSpec{},
		Status: v1alpha1.PodMonitorStatus{PodPendingCount: 0, PodRunningCount: 0},
	}
	// check if pod-monitor resource already exists in default namespace
	// create if it does not exist
	pm, err := crdclient.PodMonitors("default").Get("pod-monitor")
	if err != nil && errors.IsNotFound(err) {
		pm, err = crdclient.PodMonitors("default").Create(&podMonitor)
		if err != nil {
			panic(err)
		}
	}
	return crdclient, pm, nil
}

// Handler initialization
func NewPodHandler(config *rest.Config) *PodHandler {
	crdClient, podMonitor, err := createCRDClient(config)
	if err != nil {
		panic(err)
	}
	return &PodHandler{crdClient: crdClient, podMonitor: podMonitor, podsPending: make(map[string]bool), podsRunning: make(map[string]bool)}
}

// ObjectCreated is called when an object is created
func (t *PodHandler) ObjectCreated(key string, obj interface{}) {
	log.Infof("PodHandler.ObjectCreated -> %s", key)
	// assert the type to a Pod object to pull out relevant data
	pod := obj.(*core_v1.Pod)
	t.updateCounters(key, pod)
}

func (t *PodHandler) updateCounters(key string, pod *core_v1.Pod) {
	qualifiedName := key
	if pod.Status.Phase == "Pending" {
		// current state is Pending, but previous state is Running
		// remove from running state and move to pending state amp
		if _, exists := t.podsRunning[qualifiedName]; exists {
			delete(t.podsRunning, qualifiedName)
		}
		t.podsPending[qualifiedName] = true
	} else if pod.Status.Phase == "Running" {
		// current state is Running, but previous state is Pending
		// remove from running state and move to pending state map
		if _, exists := t.podsPending[qualifiedName]; exists {
			delete(t.podsPending, qualifiedName)
		}
		t.podsRunning[qualifiedName] = true
	}
	log.Infof("    podsRunning: %d", len(t.podsRunning))
	log.Infof("    podsPending: %d", len(t.podsPending))
	t.updateCRD()
}

func (t *PodHandler) updateCRD() {
	t.podMonitor.Status.PodRunningCount = int32(len(t.podsRunning))
	t.podMonitor.Status.PodPendingCount = int32(len(t.podsPending))
	updated, err := t.crdClient.PodMonitors("default").Update(t.podMonitor)
	if err != nil {
		panic(err)
	}
	t.podMonitor = updated
}

// ObjectDeleted is called when an object is deleted
func (t *PodHandler) ObjectDeleted(key string, obj interface{}) {
	log.Infof("PodHandler.ObjectDeleted -> %s", key)
	qualifiedName := key
	delete(t.podsPending, qualifiedName)
	delete(t.podsRunning, qualifiedName)
	log.Infof("    podsRunning: %d", len(t.podsRunning))
	log.Infof("    podsPending: %d", len(t.podsPending))
	t.updateCRD()
}

// ObjectUpdated is called when an object is updated
func (t *PodHandler) ObjectUpdated(key string, objOld, objNew interface{}) {
	log.Infof("PodHandler.ObjectUpdated -> %s", key)
	pod := objOld.(*core_v1.Pod)
	t.updateCounters(key, pod)
}
