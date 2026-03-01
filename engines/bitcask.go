package engines

import "github.com/willxm/pidb/storage/bitcask"

type Bitcask struct {
	bitcask.Engine
}

func (b *Bitcask) New() error {
	return b.Engine.New()
}

func (b *Bitcask) Set(key string, value []byte) error {
	return b.Engine.Set(key, value)
}

func (b *Bitcask) Get(key string) ([]byte, error) {
	value, err := b.Engine.Get(key)
	if err != nil {
		return nil, err
	}
	if len(value) == 0 {
		return nil, nil
	}
	return value, nil
}

func (b *Bitcask) Del(key string) error {
	return b.Engine.Del(key)
}

func (b *Bitcask) Exists(key string) (bool, error) {
	return b.Engine.Exists(key)
}
