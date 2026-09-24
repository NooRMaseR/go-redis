package stores

import (
	"errors"
	"fmt"
	"sync"
)

var ErrKeyNotFound = errors.New("key not found")

type Store struct {
	mut  sync.RWMutex
	data map[string]string
}

func (s *Store) Keys() []string {
	s.mut.RLock()
	defer s.mut.RUnlock()
	keys := make([]string, 0, len(s.data))
	for key := range s.data {
		keys = append(keys, key)
	}
	return keys
}

func (s *Store) Get(key string) (string, error) {
	s.mut.RLock()
	defer s.mut.RUnlock()
	val, ok := s.data[key]
	if !ok {
		return "", fmt.Errorf("Key %v not found", key)
	}
	return val, nil
}

func (s *Store) Set(key, value string) {
	s.mut.Lock()
	defer s.mut.Unlock()
	s.data[key] = value
}

func (s *Store) Delete(key string) {
	s.mut.Lock()
	defer s.mut.Unlock()
	delete(s.data, key)
}

func (s *Store) Pop(key string) (string, error) {
	s.mut.Lock()
	defer s.mut.Unlock()
	val, ok := s.data[key]
	if ok {
		delete(s.data, key)
		return val, nil
	}
	return "", ErrKeyNotFound
}

func (s *Store) Rename(oldKey, newKey string) error {
	if oldKey == newKey {
		return nil
	}

	s.mut.Lock()
	defer s.mut.Unlock()

	if _, ok := s.data[oldKey]; !ok {
		return fmt.Errorf("Key %v not found to rename", oldKey)
	}
	s.data[newKey] = s.data[oldKey]
	delete(s.data, oldKey)
	return nil
}

func (s *Store) Len() int {
	s.mut.RLock()
	defer s.mut.RUnlock()
	return len(s.data)
}

func (s *Store) Clear() {
	s.mut.Lock()
	defer s.mut.Unlock()
	clear(s.data)
}

func NewStore() *Store {
	return &Store{
		data: make(map[string]string),
	}
}
