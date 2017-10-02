package cache

import "sync"

type Cache interface {
	Add(u interface{}, value interface{}) (err error)
	Get(u interface{}) (i interface{}, err error)
}

type Caches struct {
	data map[interface{}]record
	sync.RWMutex
}

type record struct {
	value interface{}
}
