package pidb

import "sync"

type Mem struct {
	sync.RWMutex
	data map[string]interface{}
}

// New is initialize memory engine
func (m *Mem) New() error {
	m.data = make(map[string]interface{})
	return nil
}

// Set is set key and value to memory engine
func (m *Mem) Set(key string, value []byte) error {
	m.Lock()
	defer m.Unlock()
	m.data[key] = value
	return nil
}

// Get is get value from memory engine
func (m *Mem) Get(key string) ([]byte, error) {
	m.RLock()
	defer m.RUnlock()
	if v, ok := m.data[key]; ok {
		return v.([]byte), nil
	}
	return nil, nil
}

// Del is delete key from memory engine
func (m *Mem) Del(key string) error {
	m.Lock()
	defer m.Unlock()
	delete(m.data, key)
	return nil
}

// Exists is check key exists in memory engine
func (m *Mem) Exists(key string) (bool, error) {
	m.RLock()
	defer m.RUnlock()
	_, ok := m.data[key]
	return ok, nil
}
