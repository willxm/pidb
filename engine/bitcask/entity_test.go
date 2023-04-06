package bitcask

import (
	"fmt"
	"testing"
)

func TestNewEntity(t *testing.T) {

	key, value := []byte("hi"), []byte("nico")
	entity := NewEntity(key, value)
	b := entity.Encode()

	fmt.Printf("%x\n", b)
	// e0 ed b2 c8
	// 00 05 f8 ad 52 6a fa 38
	// 00 00 00 02
	// 00 00 00 04
	// 68 69
	// 6e 69 63 6f

	e := DecodeHead(b)
	// read twice

	e = e.DecodeBody(b[20:])

	fmt.Printf("%v\n", e)
	// &{925776056 1680798700936603 2 4 [104 105] [110 105 99 111]}

}
