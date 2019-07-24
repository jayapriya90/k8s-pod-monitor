package v1alpha1

import (
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

// PodMonitors ...
func (c *PodMonitorV1Alpha1Client) PodMonitors(namespace string) PodMonitorInterface {
	return &PodMonitorClient{
		client: c.restClient,
		ns:     namespace,
	}
}

// PodMonitorV1Alpha1Client ...
type PodMonitorV1Alpha1Client struct {
	restClient rest.Interface
}

// PodMonitorInterface ...
type PodMonitorInterface interface {
	Create(obj *PodMonitor) (*PodMonitor, error)
	Update(obj *PodMonitor) (*PodMonitor, error)
	Delete(name string, options *meta_v1.DeleteOptions) error
	Get(name string) (*PodMonitor, error)
}

// PodMonitorClient ...
type PodMonitorClient struct {
	client rest.Interface
	ns     string
}

// Create makes http POST call to API server to create PodMonitor resource
func (c *PodMonitorClient) Create(obj *PodMonitor) (*PodMonitor, error) {
	result := &PodMonitor{}
	err := c.client.Post().
		Namespace(c.ns).Resource("PodMonitors").
		Body(obj).Do().Into(result)
	return result, err
}

// Update makes http PUT call to API server to update PodMonitor resource
func (c *PodMonitorClient) Update(obj *PodMonitor) (*PodMonitor, error) {
	result := &PodMonitor{}
	err := c.client.Put().
		Namespace(c.ns).Resource("PodMonitors").
		Name(obj.ObjectMeta.Name).
		Body(obj).Do().Into(result)
	return result, err
}

// Delete makes http DELETE call to API server to delete PodMonitor resource
func (c *PodMonitorClient) Delete(name string, options *meta_v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).Resource("PodMonitors").
		Name(name).Body(options).Do().
		Error()
}

// Get makes http GET call to API server to get PodMonitor resource
func (c *PodMonitorClient) Get(name string) (*PodMonitor, error) {
	result := &PodMonitor{}
	err := c.client.Get().
		Namespace(c.ns).Resource("PodMonitors").
		Name(name).Do().Into(result)
	return result, err
}
