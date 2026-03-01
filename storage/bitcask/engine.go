package bitcask

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type Engine struct {
	data   *DataFile
	index  map[string]int64
	offset int64
	mu     sync.RWMutex
	File   string
}

var filename = "/tmp/bitcask/data.csv"

func (e *Engine) New() error {
	e.index = make(map[string]int64)

	dataFile := e.File
	if dataFile == "" {
		dataFile = filename
	}

	if err := os.MkdirAll(filepath.Dir(dataFile), 0755); err != nil {
		return err
	}

	file, err := OpenDataFile(dataFile)
	if err != nil {
		return err
	}
	e.data = file

	e.offset = 0
	err = e.data.Scan(func(offset int64, rec Record) error {
		if rec.Tombstone {
			delete(e.index, rec.Key)
		} else {
			e.index[rec.Key] = offset
		}
		e.offset = offset + rec.Size
		return nil
	})
	if err != nil {
		return fmt.Errorf("read data file err %v", err)
	}
	return nil
}

func (e *Engine) Set(key string, value []byte) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	rec := Record{Key: key, Value: append([]byte(nil), value...)}
	offset, n, err := e.data.Append(rec)
	if err != nil {
		return err
	}
	e.index[key] = offset
	e.offset = offset + int64(n)
	return nil
}

func (e *Engine) Get(key string) ([]byte, error) {
	e.mu.RLock()
	offset, ok := e.index[key]
	e.mu.RUnlock()
	if !ok {
		return []byte{}, nil
	}

	rec, err := e.data.ReadAt(offset)
	if err != nil {
		return nil, err
	}
	if rec.Tombstone {
		return []byte{}, nil
	}
	return rec.Value, nil
}

func (e *Engine) Del(key string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	rec := Record{Key: key, Tombstone: true}
	offset, n, err := e.data.Append(rec)
	if err != nil {
		return err
	}
	delete(e.index, key)
	e.offset = offset + int64(n)
	return nil
}

func (e *Engine) Exists(key string) (bool, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	_, ok := e.index[key]
	return ok, nil
}

func (e *Engine) Merge(key string) (bool, error) {
	return false, nil
}

func (e *Engine) Shutdown() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.data == nil {
		return nil
	}
	return e.data.Close()
}
