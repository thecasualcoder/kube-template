package manager

import (
	v1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Manager is an interface through which kubernetes objects
// can be queried as template functions.
type Manager interface {
	// Endpoints to list endpoints given namespace and name
	Endpoints(namespace, name string) (*v1.Endpoints, error)
}

// New to create a new manager for a given kubernetes client
func New(clientset *kubernetes.Clientset) Manager {
	return &managerImpl{
		clientset: clientset,
	}
}

type managerImpl struct {
	clientset *kubernetes.Clientset
}

func (m *managerImpl) Endpoints(namespace, name string) (*v1.Endpoints, error) {
	return m.clientset.CoreV1().Endpoints(namespace).Get(name, metaV1.GetOptions{})
}
