package bitcask

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
	"sync/atomic"
)

type Engine struct {
	data   *os.File
	index  map[string]int64
	offset int64
	mu     sync.RWMutex
	mgFlag atomic.Value
}

var filename = "/tmp/bitcask/data.csv"

func (e *Engine) New() error {

	e.index = make(map[string]int64, 0)
	e.mgFlag.Store(0)

	_, err := os.Stat(filename)
	if os.IsNotExist(err) {

		err := os.Mkdir("/tmp/bitcask", 0755)
		if err != nil {
			return err
		}

		_, err = os.Create(filename)
		if err != nil {
			return err
		}
	}
	e.data, err = os.Open(filename)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(e.data)

	for scanner.Scan() {

		line := scanner.Text()
		if lineArr := strings.Split(line, ","); len(lineArr) == 2 {
			k := lineArr[0]
			v := lineArr[1]
			if v != "" {
				e.index[k] = e.offset + 1
			} else {
				delete(e.index, k)
			}
		} else {
			return fmt.Errorf("illegal data on %d", e.offset)
		}
		e.offset += int64(len(line) + 1)
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read data file err %f", err)
	}
	return nil
}

func (e *Engine) Set(key string, value []byte) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	n, err := e.data.WriteString(fmt.Sprintf("%s,%s\n", key, string(value)))
	if err != nil {
		return err
	}

	e.index[key] = e.offset
	e.offset += int64(n)

	return nil
}

func (e *Engine) Get(key string) ([]byte, error) {
	e.mu.RLock()
	offset, ok := e.index[key]
	if !ok {
		e.mu.RUnlock()
		return []byte{}, nil
	}
	e.mu.RUnlock()

	b, err := func() ([]byte, error) {
		e.mu.Lock()
		defer e.mu.Unlock()
		_, err := e.data.Seek(offset, 0)
		if err != nil {
			return nil, err
		}
		reader := bufio.NewReader(e.data)

		return reader.ReadBytes('\n')
	}()
	if err != nil {
		return nil, err
	}
	lineArr := strings.Split(string(b), ",")

	if len(lineArr) != 2 {
		return nil, fmt.Errorf("illegal data on %d,content : %s", e.offset, string(b))
	}
	return []byte(lineArr[1]), nil
}

func (e *Engine) Del(key string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	n, err := e.data.WriteString(fmt.Sprintf("%s,%s\n", key, ""))
	if err != nil {
		return err
	}
	delete(e.index, key)

	e.offset += int64(n)
	panic("implement me")
}

func (e *Engine) Exists(key string) (bool, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	_, ok := e.index[key]
	return ok, nil
}

func (e *Engine) Merge(key string) (bool, error) {
	// merge file
	panic("implement me")
}

func (e *Engine) Shutdown() error {
	// merge file
	panic("implement me")
}
