package lsm

type Record struct {
	Value     []byte
	Tombstone bool
}
