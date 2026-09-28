package stores

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

var ErrKeyNotFound = errors.New("key not found\n")

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
		return "", ErrKeyNotFound
	}
	return val, nil
}

func (s *Store) Set(key, value string, ttl time.Duration) {
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
	return s.rename(oldKey, newKey, false)
}

func (s *Store) RenameNX(oldKey, newKey string) error {
	return s.rename(oldKey, newKey, true)
}

func (s *Store) rename(oldKey, newKey string, checkNewKeyExists bool) error {
	s.mut.Lock()
	defer s.mut.Unlock()

	// 1. Verify oldKey is present and alive
	val, ok := s.data[oldKey]
	if !ok {
		delete(s.data, oldKey)
		return ErrKeyNotFound
	}

	// 2. Identity check
	if oldKey == newKey {
		return nil
	}

	// 3. RenameNX check
	if checkNewKeyExists {
		if _, newOk := s.data[newKey]; newOk {
			return fmt.Errorf("the new key %s already exists, cannot rename\n", newKey)
		}
	}

	// 4. Overwrite & move
	s.data[newKey] = val
	delete(s.data, oldKey)
	return nil
}

func (s *Store) Len() int {
	return len(s.Keys())
}

func (s *Store) Exists(key string) bool {
	if _, err := s.Get(key); err != nil {
		return false
	}
	return true
}

func (s *Store) Clear() {
	s.mut.Lock()
	defer s.mut.Unlock()
	clear(s.data)
}

func (s *Store) Clone() Store {
	return Store{
		data: s.data,
	}
}

func (s *Store) Close() {}

func NewStore() *Store {
	return &Store{
		data: make(map[string]string),
	}
}
