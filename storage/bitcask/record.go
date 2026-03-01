package bitcask

import (
	"fmt"
	"strings"
)

type Record struct {
	Key       string
	Value     []byte
	Tombstone bool
	Size      int64
}

func encodeRecord(record Record) string {
	if record.Tombstone {
		return fmt.Sprintf("%s,\n", record.Key)
	}
	return fmt.Sprintf("%s,%s\n", record.Key, string(record.Value))
}

func decodeRecordLine(line string) (Record, error) {
	parts := strings.SplitN(line, ",", 2)
	if len(parts) != 2 {
		return Record{}, fmt.Errorf("invalid record line: %s", line)
	}
	rec := Record{Key: parts[0]}
	if parts[1] == "" {
		rec.Tombstone = true
		return rec, nil
	}
	rec.Value = []byte(parts[1])
	return rec, nil
}
