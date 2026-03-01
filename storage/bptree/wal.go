package bptree

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Entry struct {
	Op    string
	Key   string
	Value []byte
}

type WAL struct {
	file *os.File
}

func Open(path string) (*WAL, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	return &WAL{file: f}, nil
}

func (w *WAL) Replay(fn func(entry Entry) error) error {
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
				return fmt.Errorf("invalid bptree wal line: %s", line)
			}
			key, err := base64.StdEncoding.DecodeString(parts[1])
			if err != nil {
				return err
			}
			value, err := base64.StdEncoding.DecodeString(parts[2])
			if err != nil {
				return err
			}
			if err := fn(Entry{Op: "S", Key: string(key), Value: value}); err != nil {
				return err
			}
		case "D":
			if len(parts) != 2 {
				return fmt.Errorf("invalid bptree wal line: %s", line)
			}
			key, err := base64.StdEncoding.DecodeString(parts[1])
			if err != nil {
				return err
			}
			if err := fn(Entry{Op: "D", Key: string(key)}); err != nil {
				return err
			}
		default:
			return fmt.Errorf("invalid bptree wal op: %s", parts[0])
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
	keyEncoded := base64.StdEncoding.EncodeToString([]byte(key))
	valueEncoded := base64.StdEncoding.EncodeToString(value)
	if _, err := w.file.WriteString("S\t" + keyEncoded + "\t" + valueEncoded + "\n"); err != nil {
		return err
	}
	return w.file.Sync()
}

func (w *WAL) AppendDel(key string) error {
	if w == nil || w.file == nil {
		return nil
	}
	keyEncoded := base64.StdEncoding.EncodeToString([]byte(key))
	if _, err := w.file.WriteString("D\t" + keyEncoded + "\n"); err != nil {
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
