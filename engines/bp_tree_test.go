package engines

import (
	"path/filepath"
	"testing"
)

func TestBpTreeBasicCRUD(t *testing.T) {
	tree := &BpTree{}
	if err := tree.New(); err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if err := tree.Set("a", []byte("1")); err != nil {
		t.Fatalf("Set(a) error = %v", err)
	}
	if err := tree.Set("b", []byte("2")); err != nil {
		t.Fatalf("Set(b) error = %v", err)
	}

	v, err := tree.Get("a")
	if err != nil {
		t.Fatalf("Get(a) error = %v", err)
	}
	if string(v) != "1" {
		t.Fatalf("Get(a) = %q, want %q", string(v), "1")
	}

	exists, err := tree.Exists("b")
	if err != nil {
		t.Fatalf("Exists(b) error = %v", err)
	}
	if !exists {
		t.Fatalf("Exists(b) = false, want true")
	}

	if err := tree.Del("a"); err != nil {
		t.Fatalf("Del(a) error = %v", err)
	}

	exists, err = tree.Exists("a")
	if err != nil {
		t.Fatalf("Exists(a) after delete error = %v", err)
	}
	if exists {
		t.Fatalf("Exists(a) after delete = true, want false")
	}
}

func TestBpTreeSplitAndLookup(t *testing.T) {
	tree := &BpTree{}
	if err := tree.New(); err != nil {
		t.Fatalf("New() error = %v", err)
	}

	keys := []string{"k01", "k02", "k03", "k04", "k05", "k06", "k07", "k08", "k09", "k10"}

	for _, k := range keys {
		if err := tree.Set(k, []byte("v"+k[1:])); err != nil {
			t.Fatalf("Set(%s) error = %v", k, err)
		}
	}

	for _, k := range keys {
		want := "v" + k[1:]
		got, err := tree.Get(k)
		if err != nil {
			t.Fatalf("Get(%s) error = %v", k, err)
		}
		if string(got) != want {
			t.Fatalf("Get(%s) = %q, want %q", k, string(got), want)
		}
	}
}

func TestBpTreeDuplicateKey(t *testing.T) {
	tree := &BpTree{}
	if err := tree.New(); err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if err := tree.Set("dup", []byte("1")); err != nil {
		t.Fatalf("first Set() error = %v", err)
	}
	if err := tree.Set("dup", []byte("2")); err == nil {
		t.Fatalf("second Set() expected duplicate key error")
	}
}

func TestBpTreePersistence(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bptree.aof")

	tree := &BpTree{File: path}
	if err := tree.New(); err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if err := tree.Set("k1", []byte("v1")); err != nil {
		t.Fatalf("Set(k1) error = %v", err)
	}
	if err := tree.Set("k2", []byte("v2")); err != nil {
		t.Fatalf("Set(k2) error = %v", err)
	}
	if err := tree.Del("k1"); err != nil {
		t.Fatalf("Del(k1) error = %v", err)
	}
	if err := tree.Shutdown(); err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}

	recovered := &BpTree{File: path}
	if err := recovered.New(); err != nil {
		t.Fatalf("recovered New() error = %v", err)
	}
	defer func() { _ = recovered.Shutdown() }()

	v, err := recovered.Get("k2")
	if err != nil {
		t.Fatalf("Get(k2) error = %v", err)
	}
	if string(v) != "v2" {
		t.Fatalf("Get(k2) = %q, want %q", string(v), "v2")
	}

	exists, err := recovered.Exists("k1")
	if err != nil {
		t.Fatalf("Exists(k1) error = %v", err)
	}
	if exists {
		t.Fatalf("Exists(k1) = true, want false")
	}
}
