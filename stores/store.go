package stores

import (
	"fmt"
	"maps"
	"sync"
	"time"
)

type Store struct {
	mut  sync.RWMutex
	data map[string]*StoreValue
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

func (s *Store) OGet(key string) (*StoreValue, error) {
	s.mut.RLock()
	defer s.mut.RUnlock()

	obj, ok := s.data[key]
	if !ok {
		return &StoreValue{}, ErrKeyNotFound
	}
	return obj, nil
}

func (s *Store) Get(key string) (string, error) {
	s.mut.RLock()
	defer s.mut.RUnlock()
	val, ok := s.data[key]

	if !ok {
		return "", ErrKeyNotFound
	}

	if val.Type != TypeString {
		return "", ErrWrongType
	}

	return val.StrValue, nil
}

func (s *Store) Set(key, value string, ttl time.Duration) {
	s.mut.Lock()
	defer s.mut.Unlock()

	s.data[key] = &StoreValue{
		Type:       TypeString,
		StrValue:   value,
		Expiration: time.Time{},
	}
}

func (s *Store) Expire(key string, ttl time.Duration) error {return nil}

func (s *Store) RPush(key string, values []string) error {
	s.mut.Lock()
	defer s.mut.Unlock()

	obj, ok := s.data[key]
	if !ok {
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

func (s *Store) LRange(key string, start, stop int) ([]string, error) {
	s.mut.RLock()
	defer s.mut.RUnlock()

	obj, ok := s.data[key]
	if !ok {
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
	copy(result, obj.ListValue[start:stop+1])
	return result, nil
}

func (s *Store) HGet(key string) (map[string]string, error) {
	s.mut.RLock()
	defer s.mut.RUnlock()

	obj, ok := s.data[key]
	if !ok {
		return map[string]string{}, ErrKeyNotFound
	}

	if obj.Type != TypeHash {
		return map[string]string{}, ErrWrongType
	}

	data := make(map[string]string, len(obj.HashValue))
	maps.Copy(data, obj.HashValue)
	return data, nil
}

func (s *Store) HSet(key, name, value string) error {
	s.mut.Lock()
	defer s.mut.Unlock()

	obj, ok := s.data[key]
	if !ok {
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
		return val.StrValue, nil
	}
	return "", ErrKeyNotFound
}

/*
checks for key existence and renames the key if it exists
*/
func (s *Store) Rename(oldKey, newKey string) error {
	return s.rename(oldKey, newKey, false)
}

/*
checks for key existence and renames the key if it exists and check if the New Key Exists
*/
func (s *Store) RenameNX(oldKey, newKey string) error {
	return s.rename(oldKey, newKey, true)
}

func (s *Store) rename(oldKey, newKey string, checkNewKeyExists bool) error {
	s.mut.Lock()
	defer s.mut.Unlock()

	// 1. Verify oldKey is present and alive
	val, ok := s.data[oldKey]
	if !ok {
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
	s.mut.RLock()
	defer s.mut.RUnlock()

	_, ok := s.data[key]
	return ok
}

func (s *Store) Clone() *Store {
	s.mut.RLock()
	defer s.mut.RUnlock()

	return &Store{
		data: s.data,
	}
}

func (s *Store) Clear() {
	s.mut.Lock()
	defer s.mut.Unlock()
	clear(s.data)
}

func (s *Store) Close() {}

func NewStore() *Store {
	return &Store{
		data: make(map[string]*StoreValue),
	}
}
