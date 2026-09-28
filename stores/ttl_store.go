package stores

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type TTL struct {
	Value      string
	Expiration time.Time
}

type TTLStore struct {
	mut    sync.RWMutex
	data   map[string]TTL
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func (t *TTL) IsExpired() bool {
	if t.Expiration.IsZero() {
		return false
	}
	return time.Now().After(t.Expiration)
}

func (s *TTLStore) Keys() []string {
	s.mut.RLock()
	defer s.mut.RUnlock()

	keys := make([]string, 0, len(s.data))
	for key, val := range s.data {
		if !val.IsExpired() {
			keys = append(keys, key)
		}
	}
	return keys
}

func (s *TTLStore) Get(key string) (string, error) {
	s.mut.RLock()
	val, ok := s.data[key]
	s.mut.RUnlock()

	if !ok {
		return "", ErrKeyNotFound
	}
	if val.IsExpired() {
		s.mut.Lock()
		if cur, exists := s.data[key]; exists && cur.IsExpired() {
			delete(s.data, key)
		}
		s.mut.Unlock()
		return "", ErrKeyNotFound
	}
	return val.Value, nil
}

func (s *TTLStore) Set(key string, value string, ttl time.Duration) {
	s.mut.Lock()
	defer s.mut.Unlock()
	var exp time.Time

	if ttl > 0 {
		exp = time.Now().Add(ttl)
	}
	
	s.data[key] = TTL{
		Value:      value,
		Expiration: exp,
	}
}

func (s *TTLStore) Delete(key string) {
	s.mut.Lock()
	defer s.mut.Unlock()
	delete(s.data, key)
}

func (s *TTLStore) Pop(key string) (string, error) {
	s.mut.Lock()
	defer s.mut.Unlock()

	val, ok := s.data[key]
	if ok {
		delete(s.data, key)
		if val.IsExpired() {
			return "", ErrKeyNotFound
		}
		return val.Value, nil
	}
	return "", ErrKeyNotFound
}

/*
checks for key existence and renames the key if it exists
*/
func (s *TTLStore) Rename(oldKey, newKey string) error {
	return s.rename(oldKey, newKey, false)
}

/*
checks for key existence and renames the key if it exists and check if the New Key Exists
*/
func (s *TTLStore) RenameNX(oldKey, newKey string) error {
	return s.rename(oldKey, newKey, true)
}

func (s *TTLStore) rename(oldKey, newKey string, checkNewKeyExists bool) error {
	s.mut.Lock()
	defer s.mut.Unlock()

	// 1. Verify oldKey is present and alive
	val, ok := s.data[oldKey]
	if !ok || val.IsExpired() {
		delete(s.data, oldKey)
		return ErrKeyNotFound
	}

	// 2. Identity check
	if oldKey == newKey {
		return nil
	}

	// 3. RenameNX check
	if checkNewKeyExists {
		if newVal, newOk := s.data[newKey]; newOk {
			if newVal.IsExpired() {
				delete(s.data, newKey)
			} else {
				return fmt.Errorf("the new key %s already exists, cannot rename\n", newKey)
			}
		}
	}

	// 4. Overwrite & move
	s.data[newKey] = val
	delete(s.data, oldKey)
	return nil
}

func (s *TTLStore) Len() int {
	return len(s.Keys())
}

func (s *TTLStore) Exists(key string) bool {
	if _, err := s.Get(key); err != nil {
		return false
	}
	return true
}

func (s *TTLStore) Clone() TTLStore {
	return TTLStore{
		data: s.data,
	}
}

func (s *TTLStore) Clear() {
	s.mut.Lock()
	defer s.mut.Unlock()
	clear(s.data)
}

func (s *TTLStore) cleanExpired() {
	s.mut.Lock()
	defer s.mut.Unlock()
	for key, val := range s.data {
		if val.IsExpired() {
			delete(s.data, key)
		}
	}
}

func (s *TTLStore) Close() {
	if s.cancel != nil {
		s.cancel()
		s.wg.Wait()
	}
}

func (s *TTLStore) runJanitor(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.cleanExpired()
		}
	}
}

func NewTTLStore(interval time.Duration) *TTLStore {
	ctx, cancel := context.WithCancel(context.Background())

	s := &TTLStore{
		data:   make(map[string]TTL),
		cancel: cancel,
	}

	if interval > 0 {
		s.wg.Go(func() { s.runJanitor(ctx, interval) })
	}

	return s
}
