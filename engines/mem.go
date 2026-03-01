package engines

import "sync"

type Mem struct {
	sync.RWMutex
	data map[string]interface{}
}

func (m *Mem) New() error {
	m.data = make(map[string]interface{})
	return nil
}

func (m *Mem) Set(key string, value []byte) error {
	m.Lock()
	defer m.Unlock()
	m.data[key] = value
	return nil
}

func (m *Mem) Get(key string) ([]byte, error) {
	m.RLock()
	defer m.RUnlock()
	if v, ok := m.data[key]; ok {
		return v.([]byte), nil
	}
	return nil, nil
}

func (m *Mem) Del(key string) error {
	m.Lock()
	defer m.Unlock()
	delete(m.data, key)
	return nil
}

func (m *Mem) Exists(key string) (bool, error) {
	m.RLock()
	defer m.RUnlock()
	_, ok := m.data[key]
	return ok, nil
}
