package pidb

import (
	"path/filepath"
	"testing"

	"github.com/willxm/pidb/storage/bitcask"
)

func TestEngineIntegration(t *testing.T) {
	tests := []struct {
		name     string
		build    func(t *testing.T) Engine
		teardown func(Engine)
	}{
		{
			name: "mem",
			build: func(t *testing.T) Engine {
				t.Helper()
				return &Mem{}
			},
			teardown: func(Engine) {},
		},
		{
			name: "lsm_tree",
			build: func(t *testing.T) Engine {
				t.Helper()
				dir := filepath.Join(t.TempDir(), "lsm")
				return &LSMTree{Dir: dir, MemtableThreshold: 2}
			},
			teardown: func(e Engine) {
				if lsm, ok := e.(*LSMTree); ok {
					_ = lsm.Shutdown()
				}
			},
		},
		{
			name: "bptree",
			build: func(t *testing.T) Engine {
				t.Helper()
				return &BpTree{}
			},
			teardown: func(Engine) {},
		},
		{
			name: "bitcask",
			build: func(t *testing.T) Engine {
				t.Helper()
				dir := t.TempDir()
				return &Bitcask{Engine: bitcask.Engine{File: filepath.Join(dir, "data.csv")}}
			},
			teardown: func(e Engine) {
				if b, ok := e.(*Bitcask); ok {
					_ = b.Shutdown()
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := tt.build(t)
			t.Cleanup(func() { tt.teardown(engine) })

			if err := engine.New(); err != nil {
				t.Fatalf("New() error = %v", err)
			}

			exists, err := engine.Exists("missing")
			if err != nil {
				t.Fatalf("Exists(missing) error = %v", err)
			}
			if exists {
				t.Fatalf("Exists(missing) = true, want false")
			}

			value, err := engine.Get("missing")
			if err != nil {
				t.Fatalf("Get(missing) error = %v", err)
			}
			if len(value) != 0 {
				t.Fatalf("Get(missing) len = %d, want 0", len(value))
			}

			if err := engine.Set("k1", []byte("v1")); err != nil {
				t.Fatalf("Set() error = %v", err)
			}

			exists, err = engine.Exists("k1")
			if err != nil {
				t.Fatalf("Exists(k1) error = %v", err)
			}
			if !exists {
				t.Fatalf("Exists(k1) = false, want true")
			}

			value, err = engine.Get("k1")
			if err != nil {
				t.Fatalf("Get(k1) error = %v", err)
			}
			if string(value) != "v1" {
				t.Fatalf("Get(k1) = %q, want %q", string(value), "v1")
			}

			if err := engine.Del("k1"); err != nil {
				t.Fatalf("Del(k1) error = %v", err)
			}

			exists, err = engine.Exists("k1")
			if err != nil {
				t.Fatalf("Exists(k1) after delete error = %v", err)
			}
			if exists {
				t.Fatalf("Exists(k1) after delete = true, want false")
			}
		})
	}
}
