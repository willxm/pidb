package bitcask

import (
	"os"
	"sync"
	"sync/atomic"
)

type Engine struct {
	data   os.File
	index  map[string]int
	mu     sync.RWMutex
	mgFlag atomic.Value
}

func (e *Engine) New() error {
	// create or read file ,then init index
	panic("implement me")
}

func (e *Engine) Set(key string, value []byte) error {
	// append file and maintain index map
	panic("implement me")
}

func (e *Engine) Get(key string) ([]byte, error) {
	// search index
	panic("implement me")
}

func (e *Engine) Del(key string) error {
	// append file and delete index
	panic("implement me")
}

func (e *Engine) Exists(key string) (bool, error) {
	// just use Get
	panic("implement me")
}

func (e *Engine) Merge(key string) (bool, error) {
	// merge file
	panic("implement me")
}
