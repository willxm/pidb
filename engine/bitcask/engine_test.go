package bitcask

import (
	"fmt"
	"testing"
)

func TestNew(t *testing.T) {
	engine := Engine{}
	err := engine.New()
	if err != nil {
		t.Error(err)
	}

	b, err := engine.Get("key")
	if err != nil {
		t.Error(err)
	}
	fmt.Println(string(b))

	b2, err := engine.Get("key2")
	if err != nil {
		t.Error(err)
	}
	fmt.Println(string(b2))

}
