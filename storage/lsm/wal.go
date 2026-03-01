package lsm

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type WALEntry struct {
	Op    string
	Key   string
	Value []byte
}

type WAL struct {
	file *os.File
}

func OpenWAL(path string) (*WAL, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	return &WAL{file: f}, nil
}

func (w *WAL) Replay(fn func(entry WALEntry) error) error {
	if w == nil || w.file == nil {
		return nil
	}
	if _, err := w.file.Seek(0, 0); err != nil {
		return err
	}

	scanner := bufio.NewScanner(w.file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		switch parts[0] {
		case "S":
			if len(parts) != 3 {
				return fmt.Errorf("invalid lsm wal line: %s", line)
			}
			k, err := base64.StdEncoding.DecodeString(parts[1])
			if err != nil {
				return err
			}
			v, err := base64.StdEncoding.DecodeString(parts[2])
			if err != nil {
				return err
			}
			if err := fn(WALEntry{Op: "S", Key: string(k), Value: v}); err != nil {
				return err
			}
		case "D":
			if len(parts) != 2 {
				return fmt.Errorf("invalid lsm wal line: %s", line)
			}
			k, err := base64.StdEncoding.DecodeString(parts[1])
			if err != nil {
				return err
			}
			if err := fn(WALEntry{Op: "D", Key: string(k)}); err != nil {
				return err
			}
		default:
			return fmt.Errorf("invalid lsm wal op: %s", parts[0])
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	_, err := w.file.Seek(0, 2)
	return err
}

func (w *WAL) AppendSet(key string, value []byte) error {
	if w == nil || w.file == nil {
		return nil
	}
	k := base64.StdEncoding.EncodeToString([]byte(key))
	v := base64.StdEncoding.EncodeToString(value)
	if _, err := w.file.WriteString("S\t" + k + "\t" + v + "\n"); err != nil {
		return err
	}
	return w.file.Sync()
}

func (w *WAL) AppendDel(key string) error {
	if w == nil || w.file == nil {
		return nil
	}
	k := base64.StdEncoding.EncodeToString([]byte(key))
	if _, err := w.file.WriteString("D\t" + k + "\n"); err != nil {
		return err
	}
	return w.file.Sync()
}

func (w *WAL) Reset() error {
	if w == nil || w.file == nil {
		return nil
	}
	if err := w.file.Truncate(0); err != nil {
		return err
	}
	if _, err := w.file.Seek(0, 2); err != nil {
		return err
	}
	return w.file.Sync()
}

func (w *WAL) Close() error {
	if w == nil || w.file == nil {
		return nil
	}
	if err := w.file.Sync(); err != nil {
		return err
	}
	err := w.file.Close()
	w.file = nil
	return err
}
