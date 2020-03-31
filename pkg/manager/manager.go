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

	// ErrorChan returns a channel through which errors are propagated during event handling
	ErrorChan() <-chan error
}

// New to create a new manager for a given kubernetes client
func New(client kubernetes.Client) Manager {
	m := managerImpl{
		client:       client,
		eventChan:    make(chan struct{}, 1),
		throttleChan: make(chan struct{}, 1),
		errChan:      make(chan error, 1),
		watchers:     newWatchers(),
		watcherLock:  &sync.Mutex{},
		store:        NewStore(),
	}

	go m.throttle()
	return &m
}

type managerImpl struct {
	client kubernetes.Client

	// channels
	eventChan    chan struct{}
	throttleChan chan struct{}
	errChan      chan error

	// watchers
	watcherLock *sync.Mutex
	watchers    *watchers

	// data store
	store *Store
}

// ErrDataNotReady is returned whenever manager has not fetched all necessary data
// and put it in store.
var ErrDataNotReady = fmt.Errorf("data not ready")

func (m *managerImpl) addWatcher(
	key string,
	watcher watch.Interface,
	eventHandler func(watch.Event) error,
) {
	m.watchers.add(key)

	go func(w watch.Interface) {
		for event := range w.ResultChan() {
			if err := eventHandler(event); err != nil {
				m.errChan <- err
				w.Stop()
				break
			}

			m.throttleChan <- struct{}{}
		}
	}(watcher)
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

// Implementation methods go here

func (m *managerImpl) Endpoints(namespace, name string) (*v1.Endpoints, error) {
	key := fmt.Sprintf("endpoints/%s/%s", namespace, name)

	if !m.watchers.exists(key) {
		watcher, err := m.client.WatchEndpoints(namespace, name)
		if err != nil {
			return nil, fmt.Errorf("unable to start watcher for %s: %w", key, err)
		}

		m.addWatcher(key, watcher, func(event watch.Event) error {
			m.store.Set(key, event.Object.(*v1.Endpoints))
			return nil
		})
	}

	data, present := m.store.Get(key)
	if !present {
		return nil, ErrDataNotReady
	}

	endpoints, ok := data.(*v1.Endpoints)
	if !ok {
		return nil, fmt.Errorf("fetched endpoints data is corrupt")
	}
	return endpoints, nil
}

func (m *managerImpl) PodsWithLabels(namespace string, labels string) (*v1.PodList, error) {
	key := fmt.Sprintf("podsWithLabels/%s/%s", namespace, labels)

	if !m.watchers.exists(key) {
		watcher, err := m.client.WatchPodsWithLabels(namespace, labels)
		if err != nil {
			return nil, fmt.Errorf("unable to start watcher for %s: %w", key, err)
		}

		m.addWatcher(key, watcher, func(event watch.Event) error {
			podList, err := m.client.GetPodsWithLabels(namespace, labels)
			if err != nil {
				return err
			}

			m.store.Set(key, podList)
			return nil
		})
	}

	data, present := m.store.Get(key)
	if !present {
		return nil, ErrDataNotReady
	}

	podList, ok := data.(*v1.PodList)
	if !ok {
		return nil, fmt.Errorf("fetched pods list data is corrupt")
	}

	return podList, nil
}

func (m *managerImpl) EventChan() <-chan struct{} {
	return m.eventChan
}

func (m *managerImpl) ErrorChan() <-chan error {
	return m.errChan
}
