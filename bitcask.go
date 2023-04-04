package pidb

import "github.com/willxm/pidb/engine/bitcask"

type Bitcask struct {
	bitcask.Engine
}

func (b *Bitcask) New() error {
	//TODO implement me
	panic("implement me")
}

func (b *Bitcask) Set(key string, value []byte) error {
	//TODO implement me
	panic("implement me")
}

func (b *Bitcask) Get(key string) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (b *Bitcask) Del(key string) error {
	//TODO implement me
	panic("implement me")
}

func (b *Bitcask) Exists(key string) (bool, error) {
	//TODO implement me
	panic("implement me")
}
