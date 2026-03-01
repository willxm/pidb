package bitcask

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type DataFile struct {
	file *os.File
}

func OpenDataFile(path string) (*DataFile, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	return &DataFile{file: f}, nil
}

func (d *DataFile) Scan(fn func(offset int64, rec Record) error) error {
	if d == nil || d.file == nil {
		return nil
	}
	if _, err := d.file.Seek(0, 0); err != nil {
		return err
	}

	scanner := bufio.NewScanner(d.file)
	offset := int64(0)
	for scanner.Scan() {
		line := scanner.Text()
		rec, err := decodeRecordLine(line)
		if err != nil {
			return err
		}
		rec.Size = int64(len(line) + 1)
		if err := fn(offset, rec); err != nil {
			return err
		}
		offset += rec.Size
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	_, err := d.file.Seek(0, 2)
	return err
}

func (d *DataFile) Append(record Record) (offset int64, n int, err error) {
	if d == nil || d.file == nil {
		return 0, 0, fmt.Errorf("data file is not initialized")
	}
	offset, err = d.file.Seek(0, 2)
	if err != nil {
		return 0, 0, err
	}
	encoded := encodeRecord(record)
	n, err = d.file.WriteString(encoded)
	if err != nil {
		return 0, 0, err
	}
	err = d.file.Sync()
	return offset, n, err
}

func (d *DataFile) ReadAt(offset int64) (Record, error) {
	if d == nil || d.file == nil {
		return Record{}, fmt.Errorf("data file is not initialized")
	}
	if _, err := d.file.Seek(offset, 0); err != nil {
		return Record{}, err
	}
	reader := bufio.NewReader(d.file)
	line, err := reader.ReadString('\n')
	if err != nil {
		return Record{}, err
	}
	line = strings.TrimSuffix(line, "\n")
	rec, err := decodeRecordLine(line)
	if err != nil {
		return Record{}, err
	}
	rec.Size = int64(len(line) + 1)
	return rec, nil
}

func (d *DataFile) Close() error {
	if d == nil || d.file == nil {
		return nil
	}
	if err := d.file.Sync(); err != nil {
		return err
	}
	err := d.file.Close()
	d.file = nil
	return err
}
