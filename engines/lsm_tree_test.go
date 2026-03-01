package engines

import (
	"path/filepath"
	"testing"
)

func TestLSMTreeBasicCRUD(t *testing.T) {
	lsm := &LSMTree{}
	if err := lsm.New(); err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if err := lsm.Set("k", []byte("v1")); err != nil {
		t.Fatalf("Set() error = %v", err)
	}
	v, err := lsm.Get("k")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if string(v) != "v1" {
		t.Fatalf("Get() = %q, want %q", string(v), "v1")
	}

	if err := lsm.Set("k", []byte("v2")); err != nil {
		t.Fatalf("Set overwrite error = %v", err)
	}
	v, err = lsm.Get("k")
	if err != nil {
		t.Fatalf("Get() after overwrite error = %v", err)
	}
	if string(v) != "v2" {
		t.Fatalf("Get() after overwrite = %q, want %q", string(v), "v2")
	}

	if err := lsm.Del("k"); err != nil {
		t.Fatalf("Del() error = %v", err)
	}
	exists, err := lsm.Exists("k")
	if err != nil {
		t.Fatalf("Exists() error = %v", err)
	}
	if exists {
		t.Fatalf("Exists() = true, want false")
	}
}

func TestLSMTreePersistence(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "lsm")

	lsm := &LSMTree{Dir: dir, MemtableThreshold: 2, CompactThreshold: 10}
	if err := lsm.New(); err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if err := lsm.Set("a", []byte("1")); err != nil {
		t.Fatalf("Set(a) error = %v", err)
	}
	if err := lsm.Set("b", []byte("2")); err != nil {
		t.Fatalf("Set(b) error = %v", err)
	}
	if err := lsm.Set("c", []byte("3")); err != nil {
		t.Fatalf("Set(c) error = %v", err)
	}
	if err := lsm.Del("b"); err != nil {
		t.Fatalf("Del(b) error = %v", err)
	}

	if err := lsm.Shutdown(); err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}

	recovered := &LSMTree{Dir: dir, MemtableThreshold: 2, CompactThreshold: 10}
	if err := recovered.New(); err != nil {
		t.Fatalf("recovered New() error = %v", err)
	}
	defer func() { _ = recovered.Shutdown() }()

	v, err := recovered.Get("a")
	if err != nil {
		t.Fatalf("Get(a) error = %v", err)
	}
	if string(v) != "1" {
		t.Fatalf("Get(a) = %q, want %q", string(v), "1")
	}

	v, err = recovered.Get("c")
	if err != nil {
		t.Fatalf("Get(c) error = %v", err)
	}
	if string(v) != "3" {
		t.Fatalf("Get(c) = %q, want %q", string(v), "3")
	}

	exists, err := recovered.Exists("b")
	if err != nil {
		t.Fatalf("Exists(b) error = %v", err)
	}
	if exists {
		t.Fatalf("Exists(b) = true, want false")
	}
}

func TestLSMTreeCompaction(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "lsm-compact")
	lsm := &LSMTree{Dir: dir, MemtableThreshold: 1, CompactThreshold: 2}
	if err := lsm.New(); err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer func() { _ = lsm.Shutdown() }()

	if err := lsm.Set("k1", []byte("v1")); err != nil {
		t.Fatalf("Set(k1) error = %v", err)
	}
	if err := lsm.Set("k2", []byte("v2")); err != nil {
		t.Fatalf("Set(k2) error = %v", err)
	}
	if err := lsm.Set("k1", []byte("v1-new")); err != nil {
		t.Fatalf("Set(k1,new) error = %v", err)
	}
	if err := lsm.Del("k2"); err != nil {
		t.Fatalf("Del(k2) error = %v", err)
	}

	if got := len(lsm.sstables); got != 1 {
		t.Fatalf("sstable count = %d, want 1 after compaction", got)
	}

	v, err := lsm.Get("k1")
	if err != nil {
		t.Fatalf("Get(k1) error = %v", err)
	}
	if string(v) != "v1-new" {
		t.Fatalf("Get(k1) = %q, want %q", string(v), "v1-new")
	}

	exists, err := lsm.Exists("k2")
	if err != nil {
		t.Fatalf("Exists(k2) error = %v", err)
	}
	if exists {
		t.Fatalf("Exists(k2) = true, want false")
	}
}
