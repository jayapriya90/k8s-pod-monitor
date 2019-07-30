package main

import (
	apiextension "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/retry"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/jayapriya90/k8s-pod-monitor/v1alpha1"
	core_v1 "k8s.io/api/core/v1"
)

// PodHandler is a sample implementation of Handler
type PodHandler struct {
	crdClient   *v1alpha1.PodMonitorV1Alpha1Client
	startedTimestamp time.Time
	podsCreated map[string]bool
	podsRunning map[string]bool
}

func createCRDClient(config *rest.Config) (*v1alpha1.PodMonitorV1Alpha1Client, time.Time, error) {
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
		Status: v1alpha1.PodMonitorStatus{PodCreatedCount: 0, PodRunningCount: 0},
	}
	// check if pod-monitor resource already exists in default namespace
	// create if it does not exist
	_, err = crdclient.PodMonitors("default").Get("pod-monitor")
	var startedTs time.Time
	if err != nil && errors.IsNotFound(err) {
		_, err := crdclient.PodMonitors("default").Create(&podMonitor)
		if err != nil {
			panic(err)
		}
		startedTs = time.Now()
	}
	return crdclient, startedTs, nil
}

// Handler initialization
func NewPodHandler(config *rest.Config) *PodHandler {
	crdClient, startedTs, err := createCRDClient(config)
	if err != nil {
		panic(err)
	}
	return &PodHandler{crdClient: crdClient, startedTimestamp: startedTs, podsCreated: make(map[string]bool), podsRunning: make(map[string]bool)}
}

// ObjectCreated is called when an object is created
func (t *PodHandler) ObjectCreated(key string, obj interface{}) {
	log.Infof("PodHandler.ObjectCreated -> %s", key)
	// assert the type to a Pod object to pull out relevant data
	pod := obj.(*core_v1.Pod)
	if pod.Status.Phase == "Pending" {
		if pod.CreationTimestamp.Time.Before(t.startedTimestamp) {
			log.Infof("%s pod created before k8s pod monitor service start..ignoring", key)
			return
		}
		if _, exists := t.podsCreated[key]; !exists {
			t.podsCreated[key] = true
		}
	}
	if pod.Status.Phase == "Running" {
		if _, exists := t.podsRunning[key]; !exists {
			t.podsRunning[key] = true
		}
	}
	log.Infof("    podsCreated: %d", len(t.podsCreated))
	log.Infof("    podsRunning: %d", len(t.podsRunning))
	t.updateCRD()
}

// ObjectDeleted is called when an object is deleted
func (t *PodHandler) ObjectDeleted(key string, obj interface{}) {
	log.Infof("PodHandler.ObjectDeleted -> %s", key)
	delete(t.podsRunning, key)
	log.Infof("    podsCreated: %d", len(t.podsCreated))
	log.Infof("    podsRunning: %d", len(t.podsRunning))
	t.updateCRD()
}

func (t *PodHandler) updateCRD() {
	current, err := t.crdClient.PodMonitors("default").Get("pod-monitor")
	current.Status.PodRunningCount = int32(len(t.podsRunning))
	current.Status.PodCreatedCount = int32(len(t.podsCreated))
	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		_, err := t.crdClient.PodMonitors("default").Update(current)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		log.Errorf("%v", err)
	}
}

