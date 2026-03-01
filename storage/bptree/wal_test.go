package bptree

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestWALAppendAndReplay(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bptree.wal")

	w, err := Open(path)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() { _ = w.Close() })

	if err := w.AppendSet("k1", []byte("v1")); err != nil {
		t.Fatalf("AppendSet(k1) error = %v", err)
	}
	if err := w.AppendSet("k2", []byte("v2")); err != nil {
		t.Fatalf("AppendSet(k2) error = %v", err)
	}
	if err := w.AppendDel("k1"); err != nil {
		t.Fatalf("AppendDel(k1) error = %v", err)
	}

	var got []Entry
	if err := w.Replay(func(entry Entry) error {
		got = append(got, entry)
		return nil
	}); err != nil {
		t.Fatalf("Replay() error = %v", err)
	}

	want := []Entry{
		{Op: "S", Key: "k1", Value: []byte("v1")},
		{Op: "S", Key: "k2", Value: []byte("v2")},
		{Op: "D", Key: "k1", Value: nil},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Replay() entries = %#v, want %#v", got, want)
	}
}

func TestWALReplayInvalidLine(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.wal")
	if err := os.WriteFile(path, []byte("S\tonly_key\n"), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	w, err := Open(path)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer func() { _ = w.Close() }()

	err = w.Replay(func(entry Entry) error { return nil })
	if err == nil {
		t.Fatalf("Replay() expected error for invalid wal line")
	}
}

func TestWALReplayCallbackError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "callback.wal")
	w, err := Open(path)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer func() { _ = w.Close() }()

	if err := w.AppendSet("k", []byte("v")); err != nil {
		t.Fatalf("AppendSet() error = %v", err)
	}

	expected := errors.New("stop")
	err = w.Replay(func(entry Entry) error { return expected })
	if !errors.Is(err, expected) {
		t.Fatalf("Replay() error = %v, want %v", err, expected)
	}
}
