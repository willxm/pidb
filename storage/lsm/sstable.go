package lsm

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const FilePrefix = "sst-"
const FileSuffix = ".db"

func ListSSTables(dir string) ([]string, int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, 0, err
	}

	type fileWithID struct {
		id   int
		path string
	}
	files := make([]fileWithID, 0)
	maxID := 0

	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasPrefix(name, FilePrefix) || !strings.HasSuffix(name, FileSuffix) {
			continue
		}
		idPart := strings.TrimSuffix(strings.TrimPrefix(name, FilePrefix), FileSuffix)
		id, err := strconv.Atoi(idPart)
		if err != nil {
			continue
		}
		if id > maxID {
			maxID = id
		}
		files = append(files, fileWithID{id: id, path: filepath.Join(dir, name)})
	}

	sort.Slice(files, func(i, j int) bool { return files[i].id > files[j].id })
	paths := make([]string, 0, len(files))
	for _, f := range files {
		paths = append(paths, f.path)
	}
	return paths, maxID + 1, nil
}

func BuildSSTablePath(dir string, id int) string {
	return filepath.Join(dir, fmt.Sprintf("%s%06d%s", FilePrefix, id, FileSuffix))
}

func WriteSSTable(path string, records map[string]Record, skipTombstone bool) error {
	keys := make([]string, 0, len(records))
	for k, rec := range records {
		if skipTombstone && rec.Tombstone {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	tmp := path + ".tmp"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
	if err != nil {
		return err
	}

	writer := bufio.NewWriter(f)
	for _, key := range keys {
		rec := records[key]
		k := base64.StdEncoding.EncodeToString([]byte(key))
		if rec.Tombstone {
			if _, err = writer.WriteString("D\t" + k + "\n"); err != nil {
				_ = f.Close()
				return err
			}
			continue
		}
		v := base64.StdEncoding.EncodeToString(rec.Value)
		if _, err = writer.WriteString("S\t" + k + "\t" + v + "\n"); err != nil {
			_ = f.Close()
			return err
		}
	}

	if err = writer.Flush(); err != nil {
		_ = f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		_ = f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}

	return os.Rename(tmp, path)
}

func GetRecord(path, key string) (Record, bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return Record{}, false, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		switch parts[0] {
		case "S":
			if len(parts) != 3 {
				return Record{}, false, fmt.Errorf("invalid sstable line: %s", line)
			}
			k, err := base64.StdEncoding.DecodeString(parts[1])
			if err != nil {
				return Record{}, false, err
			}
			if string(k) != key {
				continue
			}
			v, err := base64.StdEncoding.DecodeString(parts[2])
			if err != nil {
				return Record{}, false, err
			}
			return Record{Value: v}, true, nil
		case "D":
			if len(parts) != 2 {
				return Record{}, false, fmt.Errorf("invalid sstable line: %s", line)
			}
			k, err := base64.StdEncoding.DecodeString(parts[1])
			if err != nil {
				return Record{}, false, err
			}
			if string(k) == key {
				return Record{Tombstone: true}, true, nil
			}
		default:
			return Record{}, false, fmt.Errorf("invalid sstable op: %s", parts[0])
		}
	}
	if err := scanner.Err(); err != nil {
		return Record{}, false, err
	}
	return Record{}, false, nil
}

func ReadAllRecords(path string) (map[string]Record, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	records := make(map[string]Record)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		switch parts[0] {
		case "S":
			if len(parts) != 3 {
				return nil, fmt.Errorf("invalid sstable line: %s", line)
			}
			k, err := base64.StdEncoding.DecodeString(parts[1])
			if err != nil {
				return nil, err
			}
			v, err := base64.StdEncoding.DecodeString(parts[2])
			if err != nil {
				return nil, err
			}
			records[string(k)] = Record{Value: v}
		case "D":
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid sstable line: %s", line)
			}
			k, err := base64.StdEncoding.DecodeString(parts[1])
			if err != nil {
				return nil, err
			}
			records[string(k)] = Record{Tombstone: true}
		default:
			return nil, fmt.Errorf("invalid sstable op: %s", parts[0])
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return records, nil
}
