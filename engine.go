package pidb

type Record struct {
	Value []byte
}

type Engine interface {
	New() error
	Set(key string, value []byte) error
	Get(key string) ([]byte, error)
	Del(key string) error
	Exists(key string) (bool, error)
}
