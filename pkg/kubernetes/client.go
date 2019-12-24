package kubernetes

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/clientcmd"
)

// NewClient creates a clientset for given kubeconfig file.
// If kubeconfig is empty, it creates a InClusterClient.
// Errors out if client cannot be created.
func NewClient(kubeconfig string) (Client, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &clientImpl{clientset}, nil
}

// Client represents a Kubernetes client. It abstracts and proxies call to kubernetes API.
// Make sure to return Kubernetes objects always so that it remains as a proxy with few abstractions
type Client interface {
	// GetEndpoints fetches the endpoints for a given namespace and name
	GetEndpoints(namespace, name string) (*v1.Endpoints, error)
	// WatchEndpoints returns a watcher of Endpoints watch API
	WatchEndpoints(namespace, name string) (watch.Interface, error)
	// GetPodsWithLabels fetches pod list for given namespace and label selectors
	// Label selectors need to sent in the format of key=value,key2=value2
	GetPodsWithLabels(namespace, labelSelectors string) (*v1.PodList, error)
	// WatchPodsWithLabels watches pods for given namespace and label selectors
	// Label selectors need to sent in the format of key=value,key2=value2
	WatchPodsWithLabels(namespace, labelSelectors string) (watch.Interface, error)
}

type clientImpl struct {
	*kubernetes.Clientset
}

func (c clientImpl) GetEndpoints(namespace, name string) (*v1.Endpoints, error) {
	return c.CoreV1().Endpoints(namespace).Get(name, metaV1.GetOptions{})
}

func (c clientImpl) WatchEndpoints(namespace, name string) (watch.Interface, error) {
	return c.CoreV1().Endpoints(namespace).Watch(metaV1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.name=%s", name),
	})
}

func (c clientImpl) GetPodsWithLabels(namespace, labelSelectors string) (*v1.PodList, error) {
	return c.CoreV1().Pods(namespace).List(metaV1.ListOptions{
		LabelSelector: labelSelectors,
	})
}

func (c clientImpl) WatchPodsWithLabels(namespace, labelSelectors string) (watch.Interface, error) {
	return c.CoreV1().Pods(namespace).Watch(metaV1.ListOptions{
		LabelSelector: labelSelectors,
	})
}
