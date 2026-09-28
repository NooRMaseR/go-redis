package stores

import "time"

type IStore interface {
	Keys() []string
	Get(key string) (string, error)
	Set(key string, value string, ttl time.Duration)
	Delete(key string)
	Exists(key string) bool
	Pop(key string) (string, error)
	Rename(oldKey, newKey string) error
	RenameNX(oldKey, newKey string) error
	Len() int
	Clear()
	Close()
}


