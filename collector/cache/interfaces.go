package cache

import "sync"

type Cache interface {
	Add(u interface{}, value string) (err error)
	Get(u interface{}) (i string, err error)
}

type Caches struct {
	data map[interface{}]string
	sync.RWMutex
}
