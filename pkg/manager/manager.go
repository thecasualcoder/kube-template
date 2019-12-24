package manager

import (
	"fmt"
	"github.com/thecasualcoder/kube-template/pkg/kubernetes"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
	"sync"
	"time"
)

// Manager is an interface through which kubernetes objects
// can be queried as template functions.
type Manager interface {
	// Endpoints to list endpoints given namespace and name
	Endpoints(namespace, name string) (*v1.Endpoints, error)
	// PodsWithLabels to list pods given namespace and labels
	PodsWithLabels(namespace string, labels string) (*v1.PodList, error)
	// EventChan will send events whenever there are changes to resources used by the render function
	EventChan() <-chan struct{}
}

// New to create a new manager for a given kubernetes client
func New(client kubernetes.Client) Manager {
	impl := managerImpl{
		client:       client,
		eventChan:    make(chan struct{}, 1),
		throttleChan: make(chan struct{}, 1),
		watchers:     make(map[string]watch.Interface),
		watcherLock:  &sync.Mutex{},
	}
	impl.eventChan <- struct{}{}
	go impl.throttle()
	return &impl
}

type managerImpl struct {
	client       kubernetes.Client
	eventChan    chan struct{}
	throttleChan chan struct{}
	watcherLock  *sync.Mutex
	watchers     map[string]watch.Interface
}

func (m *managerImpl) Endpoints(namespace, name string) (*v1.Endpoints, error) {
	endpoints, err := m.client.GetEndpoints(namespace, name)
	if err != nil {
		return nil, err
	}

	watcherKey := fmt.Sprintf("endpoints/%s/%s", namespace, name)
	if _, exists := m.watchers[watcherKey]; !exists {
		watcher, err := m.client.WatchEndpoints(namespace, name)
		if err != nil {
			return nil, fmt.Errorf("unable to start watcher for %s: %w", watcherKey, err)
		}

		m.addWatcher(watcherKey, watcher)
	}

	return endpoints, nil
}

func (m *managerImpl) addWatcher(key string, watcher watch.Interface) {
	m.watcherLock.Lock()
	m.watchers[key] = watcher
	m.watcherLock.Unlock()

	go func() {
		for range watcher.ResultChan() {
			m.throttleChan <- struct{}{}
		}
	}()
}

func (m *managerImpl) PodsWithLabels(namespace string, labels string) (*v1.PodList, error) {
	podList, err := m.client.GetPodsWithLabels(namespace, labels)
	if err != nil {
		return podList, err
	}

	watcherKey := fmt.Sprintf("podsWithLabels/%s/%s", namespace, labels)
	if _, exists := m.watchers[watcherKey]; !exists {
		watcher, err := m.client.WatchPodsWithLabels(namespace, labels)
		if err != nil {
			return podList, fmt.Errorf("unable to start watcher for %s: %w", watcherKey, err)
		}

		m.addWatcher(watcherKey, watcher)
	}

	return podList, nil
}

func (m *managerImpl) EventChan() <-chan struct{} {
	return m.eventChan
}

func (m *managerImpl) throttle() {
	timerFired := false

	for range m.throttleChan {
		if !timerFired {
			timerFired = true
			time.AfterFunc(2*time.Second, func() {
				timerFired = false
				m.eventChan <- struct{}{}
			})
		}
	}
}
