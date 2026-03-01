package bitcask

import (
	"path/filepath"
	"testing"
)

func TestRecordEncodeDecode(t *testing.T) {
	line := encodeRecord(Record{Key: "k1", Value: []byte("v1")})
	if line != "k1,v1\n" {
		t.Fatalf("encode set line = %q", line)
	}

	rec, err := decodeRecordLine("k1,v1")
	if err != nil {
		t.Fatalf("decode set line error = %v", err)
	}
	if rec.Key != "k1" || string(rec.Value) != "v1" || rec.Tombstone {
		t.Fatalf("decoded set record unexpected: %+v", rec)
	}

	tomb, err := decodeRecordLine("k1,")
	if err != nil {
		t.Fatalf("decode tombstone line error = %v", err)
	}
	if tomb.Key != "k1" || !tomb.Tombstone {
		t.Fatalf("decoded tombstone unexpected: %+v", tomb)
	}
}

func TestDataFileScanAndReadAt(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.csv")
	df, err := OpenDataFile(path)
	if err != nil {
		t.Fatalf("OpenDataFile error = %v", err)
	}
	defer func() { _ = df.Close() }()

	offset1, _, err := df.Append(Record{Key: "a", Value: []byte("1")})
	if err != nil {
		t.Fatalf("append 1 error = %v", err)
	}
	_, _, err = df.Append(Record{Key: "a", Tombstone: true})
	if err != nil {
		t.Fatalf("append tombstone error = %v", err)
	}

	rec, err := df.ReadAt(offset1)
	if err != nil {
		t.Fatalf("ReadAt error = %v", err)
	}
	if rec.Key != "a" || string(rec.Value) != "1" || rec.Tombstone {
		t.Fatalf("ReadAt record unexpected: %+v", rec)
	}

	count := 0
	err = df.Scan(func(offset int64, rec Record) error {
		count++
		return nil
	})
	if err != nil {
		t.Fatalf("Scan error = %v", err)
	}
	if count != 2 {
		t.Fatalf("scan count = %d, want 2", count)
	}
}
