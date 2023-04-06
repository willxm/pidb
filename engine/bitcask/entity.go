package bitcask

import (
	"encoding/binary"
	"hash/crc32"
	"time"
)

const entityHeadLen = 4 + 8 + 4 + 4

type Entity struct {
	crc       uint32
	timestamp uint64
	kLen      uint32
	vLen      uint32
	key       []byte
	value     []byte
}

func NewEntity(key, value []byte) *Entity {
	e := &Entity{
		timestamp: uint64(time.Now().UnixMicro()), // irrational, should be updated with write,but I have no idea.
		kLen:      uint32(len(key)),
		vLen:      uint32(len(value)),
		key:       key,
		value:     value,
	}
	e.sign()
	return e
}

func (e *Entity) Encode() []byte {
	b := make([]byte, e.len(), e.len())
	binary.BigEndian.PutUint32(b[:4], e.crc)
	binary.BigEndian.PutUint64(b[4:12], e.timestamp)
	binary.BigEndian.PutUint32(b[12:16], e.kLen)
	binary.BigEndian.PutUint32(b[16:20], e.vLen)
	copy(b[20:20+e.kLen], e.key)
	copy(b[20+e.kLen:20+e.kLen+e.vLen], e.value)
	return b
}

func DecodeHead(b []byte) *Entity {
	e := &Entity{}
	e.crc = binary.BigEndian.Uint32(b[0:4])
	e.timestamp = binary.BigEndian.Uint64(b[4:12])
	e.kLen = binary.BigEndian.Uint32(b[12:16])
	e.vLen = binary.BigEndian.Uint32(b[16:20])
	e.key = make([]byte, e.kLen, e.kLen)
	e.value = make([]byte, e.vLen, e.vLen)
	return e
}

func (e *Entity) DecodeBody(b []byte) *Entity {
	copy(e.key, b[:e.kLen])
	copy(e.value, b[e.kLen:e.kLen+e.vLen])
	return e
}

func (e *Entity) sign() {
	crc32q := crc32.MakeTable(0xedb88320)
	e.crc = crc32.Checksum(e.Encode()[4:], crc32q)
}

func (e *Entity) len() int {
	return int(entityHeadLen + e.kLen + e.vLen)
}
