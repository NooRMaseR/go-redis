package stores

import (
	"context"
	"fmt"
	"maps"
	"sync"
	"time"
)

type TTLStore struct {
	mut    sync.RWMutex
	wg     sync.WaitGroup
	data   map[string]*StoreValue
	cancel context.CancelFunc
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

func (s *TTLStore) OGet(key string) (*StoreValue, error) {
	s.mut.RLock()
	defer s.mut.RUnlock()

	obj, ok := s.data[key]
	if !ok || obj.IsExpired() {
		return &StoreValue{}, ErrKeyNotFound
	}
	return obj, nil
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
	if val.Type != TypeString {
		return "", ErrWrongType
	}
	return val.StrValue, nil
}

func (s *TTLStore) Set(key, value string, ttl time.Duration) {
	s.mut.Lock()
	defer s.mut.Unlock()

	var exp time.Time

	if ttl > 0 {
		exp = time.Now().Add(ttl)
	}

	s.data[key] = &StoreValue{
		Type:       TypeString,
		StrValue:   value,
		Expiration: exp,
	}
}

func (s *TTLStore) Expire(key string, ttl time.Duration) error {
	s.mut.Lock()
	defer s.mut.Unlock()
	if val, ok := s.data[key]; ok && !val.IsExpired() {
		val.Expiration = time.Now().Add(ttl)
		return nil
	}
	return ErrKeyNotFound
}

func (s *TTLStore) RPush(key string, values []string) error {
	s.mut.Lock()
	defer s.mut.Unlock()

	obj, ok := s.data[key]
	if !ok || obj.IsExpired() {
		s.data[key] = &StoreValue{
			Type:       TypeList,
			ListValue:  append([]string{}, values...),
			Expiration: time.Time{},
		}
		return nil
	}

	if obj.Type != TypeList {
		return ErrWrongType
	}

	obj.ListValue = append(obj.ListValue, values...)
	return nil
}

func (s *TTLStore) LRange(key string, start, stop int) ([]string, error) {
	s.mut.RLock()
	defer s.mut.RUnlock()

	obj, ok := s.data[key]
	if !ok || obj.IsExpired() {
		return []string{}, ErrKeyNotFound
	}

	if obj.Type != TypeList {
		return []string{}, ErrWrongType
	}

	length := len(obj.ListValue)
	if length == 0 {
		return []string{}, nil
	}

	if start < 0 {
		start += length
	}

	if stop < 0 {
		stop += length
	}

	if start < 0 {
		start = 0
	}

	if stop >= length {
		stop = length - 1
	}

	if start > stop {
		return []string{}, nil
	}

	result := make([]string, stop-start+1)
	copy(result, obj.ListValue[start : stop+1])
	return result, nil
}

func (s *TTLStore) HGet(key string) (map[string]string, error) {
	s.mut.RLock()
	defer s.mut.RUnlock()

	obj, ok := s.data[key]
	if !ok || obj.IsExpired() {
		return map[string]string{}, ErrKeyNotFound
	}

	if obj.Type != TypeHash {
		return map[string]string{}, ErrWrongType
	}

	data := make(map[string]string, len(obj.HashValue))
	maps.Copy(data, obj.HashValue)
	return data, nil
}

func (s *TTLStore) HSet(key, name, value string) error {
	s.mut.Lock()
	defer s.mut.Unlock()

	obj, ok := s.data[key]
	if !ok || obj.IsExpired() {
		s.data[key] = &StoreValue{
			Type:       TypeHash,
			HashValue:  map[string]string{name: value},
			Expiration: time.Time{},
		}
		return nil
	}

	if obj.Type != TypeHash {
		return ErrWrongType
	}

	obj.HashValue[name] = value
	return nil
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
		return val.StrValue, nil
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
	if !ok {
		return ErrKeyNotFound
	}
	
	if val.IsExpired() {
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
	s.mut.RLock()
	defer s.mut.RUnlock()

	val, ok := s.data[key]
	return ok && !val.IsExpired()
}

func (s *TTLStore) Clone() *TTLStore {
	s.mut.RLock()
	defer s.mut.RUnlock()

	return &TTLStore{
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
		data:   make(map[string]*StoreValue),
		cancel: cancel,
	}

	if interval > 0 {
		s.wg.Go(func() { s.runJanitor(ctx, interval) })
	}

	return s
}
