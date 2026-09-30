package stores

import (
	"errors"
	"time"
)

var (
	ErrKeyNotFound = errors.New("key not found\n")
	ErrWrongType   = errors.New("WRONGTYPE operation performed on a wrong kind of value type\n")
	ErrInvalidDuration = errors.New("expiration must be a valid duration\n")
)

type StoreValue struct {
	Type       DataType
	StrValue   string
	ListValue  []string
	HashValue  map[string]string
	Expiration time.Time
}

type IStore interface {
	// Generic methods
	OGet(key string) (*StoreValue, error)
	Keys() []string
	Len() int
	Clear()
	Close()
	Delete(key string)
	Exists(key string) bool
	Expire(key string, ttl time.Duration) error
	Rename(oldKey, newKey string) error
	RenameNX(oldKey, newKey string) error

	// string methods
	Get(key string) (string, error)
	Set(key string, value string, ttl time.Duration)
	Pop(key string) (string, error)

	// list methods
	RPush(key string, values []string) error
	LRange(key string, start, stop int) ([]string, error)

	// hashmap methods
	HGet(key string) (map[string]string, error)
	HSet(key, name, value string) error
}

type DataType byte

const (
	TypeString DataType = iota
	TypeList
	TypeHash
)
