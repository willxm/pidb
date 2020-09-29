package pidb

type Record struct {
	Value []byte
}

type Engine interface {
	New()
	Insert(key string, value []byte)
	Find(key string)
}
