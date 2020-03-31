package manager

import (
	"sync"
)

// NewStore is a constructor to create store.
func NewStore() *Store {
	return &Store{
		data:  make(map[string]interface{}),
		mutex: &sync.Mutex{},
	}
}

// Store is an thread safe KV store.
type Store struct {
	data  map[string]interface{}
	mutex *sync.Mutex
}

// Get a value from the store for the given key.
// False is returned if the key is not present
// The operation is thread safe.
func (s *Store) Get(key string) (interface{}, bool) {
	s.mutex.Lock()
	data, exists := s.data[key]
	s.mutex.Unlock()

	if !exists {
		return nil, false
	}

	return data, true
}

// Set the given key, value inside the store.
// If the key already exists, it is overridden.
// The operation is thread safe.
func (s *Store) Set(key string, value interface{}) {
	s.mutex.Lock()
	s.data[key] = value
	s.mutex.Unlock()
}
