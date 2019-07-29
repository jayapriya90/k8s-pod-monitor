package main

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/jayapriya90/k8s-pod-monitor/v1alpha1"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Test the Kubernetes resource config using Terratest.
func TestKubernetesBasic(t *testing.T) {
	t.Parallel()

	// Path to the Kubernetes resource config we will test
	kubeResourcePath, err := filepath.Abs("./test_k8s_manifests/nginx-deployment.yaml")
	require.NoError(t, err)

	// Setup the kubectl config and context. Here we choose to use the defaults, which is:
	// - HOME/.kube/config for the kubectl config file
	// - Current context of the kubectl config file
	options := k8s.NewKubectlOptions("", "")

	k8sClient, config := GetKubernetesClient()
	require.NotNil(t, k8sClient)
	require.NotNil(t, config)

	// At the end of the test, run `kubectl delete -f RESOURCE_CONFIG` to clean up any resources that were created.
	defer k8sClient.AppsV1().Deployments(metav1.NamespaceDefault).Delete("nginx-deployment-test", &metav1.DeleteOptions{})

	podList, err := k8sClient.CoreV1().Pods(metav1.NamespaceAll).List(metav1.ListOptions{})
	require.NoError(t, err)
	currentRunningPods := 0
	for _, pod := range podList.Items {
		if pod.Status.Phase == "Running" {
			currentRunningPods++
		}
	}
	// This will run `kubectl apply -f RESOURCE_CONFIG` and fail the test if there are any errors
	k8s.KubectlApply(t, options, kubeResourcePath)

	// wait for the pod to be available
	time.Sleep(30 * time.Second)

	podMonitorClient, err := v1alpha1.NewClient(config)
	require.NoError(t, err)
	podMonitor, err := podMonitorClient.PodMonitors("default").Get("pod-monitor")
	require.NoError(t, err)
	newPodCount := currentRunningPods + 2

	require.Equal(t, int32(newPodCount), podMonitor.Status.PodRunningCount)
}

