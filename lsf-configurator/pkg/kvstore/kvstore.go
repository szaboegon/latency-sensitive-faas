package kvstore

import (
	"fmt"
	"sync"
)

type kvStore[T any] struct {
	mu    sync.RWMutex
	store map[string]T
}

func NewStore[T any]() Client[T] {
	return &kvStore[T]{
		store: make(map[string]T),
	}
}

func (s *kvStore[T]) Set(key string, value T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.store[key] = value
}

func (s *kvStore[T]) Get(key string) (T, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, exists := s.store[key]
	if !exists {
		var zero T
		return zero, fmt.Errorf("entity with key %s not found", key)
	}
	return val, nil
}

func (s *kvStore[T]) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.store, key)
}
