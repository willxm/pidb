package engines

import (
	"os"
	"path/filepath"
	"sync"

	storagelsm "github.com/willxm/pidb/storage/lsm"
)

const defaultLSMMemtableThreshold = 1024
const defaultLSMCompactThreshold = 4

type LSMTree struct {
	Dir               string
	MemtableThreshold int
	CompactThreshold  int

	mu       sync.RWMutex
	memtable map[string]storagelsm.Record
	wal      *storagelsm.WAL
	sstables []string
	nextID   int
}

func (l *LSMTree) New() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.memtable = make(map[string]storagelsm.Record)
	l.sstables = nil
	l.nextID = 1

	if l.wal != nil {
		_ = l.wal.Close()
		l.wal = nil
	}

	l.applyDefaultsLocked()
	if l.Dir == "" {
		return nil
	}
	if err := os.MkdirAll(l.Dir, 0755); err != nil {
		return err
	}

	if err := l.loadSSTablesLocked(); err != nil {
		return err
	}

	walPath := filepath.Join(l.Dir, "wal.log")
	wal, err := storagelsm.OpenWAL(walPath)
	if err != nil {
		return err
	}
	if err := wal.Replay(func(entry storagelsm.WALEntry) error {
		switch entry.Op {
		case "S":
			l.memtable[entry.Key] = storagelsm.Record{Value: append([]byte(nil), entry.Value...)}
		case "D":
			l.memtable[entry.Key] = storagelsm.Record{Tombstone: true}
		}
		return nil
	}); err != nil {
		_ = wal.Close()
		return err
	}

	l.wal = wal
	return nil
}

func (l *LSMTree) Set(key string, value []byte) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if err := l.ensureInitializedLocked(); err != nil {
		return err
	}

	l.memtable[key] = storagelsm.Record{Value: append([]byte(nil), value...)}
	if err := l.appendWALSetLocked(key, value); err != nil {
		return err
	}

	if len(l.memtable) >= l.MemtableThreshold {
		if err := l.flushMemtableLocked(); err != nil {
			return err
		}
	}
	return nil
}

func (l *LSMTree) Get(key string) ([]byte, error) {
	l.mu.RLock()
	if rec, ok := l.memtable[key]; ok {
		l.mu.RUnlock()
		if rec.Tombstone {
			return nil, nil
		}
		return append([]byte(nil), rec.Value...), nil
	}

	sstables := append([]string(nil), l.sstables...)
	l.mu.RUnlock()

	for _, path := range sstables {
		rec, ok, err := storagelsm.GetRecord(path, key)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		if rec.Tombstone {
			return nil, nil
		}
		return append([]byte(nil), rec.Value...), nil
	}

	return nil, nil
}

func (l *LSMTree) Del(key string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if err := l.ensureInitializedLocked(); err != nil {
		return err
	}

	l.memtable[key] = storagelsm.Record{Tombstone: true}
	if err := l.appendWALDelLocked(key); err != nil {
		return err
	}

	if len(l.memtable) >= l.MemtableThreshold {
		if err := l.flushMemtableLocked(); err != nil {
			return err
		}
	}
	return nil
}

func (l *LSMTree) Exists(key string) (bool, error) {
	v, err := l.Get(key)
	if err != nil {
		return false, err
	}
	return v != nil, nil
}

func (l *LSMTree) Shutdown() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if len(l.memtable) > 0 {
		if err := l.flushMemtableLocked(); err != nil {
			return err
		}
	}

	if l.wal == nil {
		return nil
	}
	err := l.wal.Close()
	l.wal = nil
	return err
}

func (l *LSMTree) ensureInitializedLocked() error {
	if l.memtable == nil {
		l.memtable = make(map[string]storagelsm.Record)
	}
	l.applyDefaultsLocked()
	return nil
}

func (l *LSMTree) applyDefaultsLocked() {
	if l.MemtableThreshold <= 0 {
		l.MemtableThreshold = defaultLSMMemtableThreshold
	}
	if l.CompactThreshold <= 0 {
		l.CompactThreshold = defaultLSMCompactThreshold
	}
}

func (l *LSMTree) loadSSTablesLocked() error {
	sstables, nextID, err := storagelsm.ListSSTables(l.Dir)
	if err != nil {
		return err
	}
	l.sstables = sstables
	if nextID <= 0 {
		nextID = 1
	}
	l.nextID = nextID
	return nil
}

func (l *LSMTree) appendWALSetLocked(key string, value []byte) error {
	if l.wal == nil {
		return nil
	}
	return l.wal.AppendSet(key, value)
}

func (l *LSMTree) appendWALDelLocked(key string) error {
	if l.wal == nil {
		return nil
	}
	return l.wal.AppendDel(key)
}

func (l *LSMTree) flushMemtableLocked() error {
	if len(l.memtable) == 0 {
		return nil
	}
	if l.Dir == "" {
		l.memtable = make(map[string]storagelsm.Record)
		return nil
	}

	target := storagelsm.BuildSSTablePath(l.Dir, l.nextID)
	if err := storagelsm.WriteSSTable(target, l.memtable, false); err != nil {
		return err
	}

	l.sstables = append([]string{target}, l.sstables...)
	l.nextID++
	l.memtable = make(map[string]storagelsm.Record)

	if l.wal != nil {
		if err := l.wal.Reset(); err != nil {
			return err
		}
	}

	if len(l.sstables) >= l.CompactThreshold {
		if err := l.compactSSTablesLocked(); err != nil {
			return err
		}
	}

	return nil
}

func (l *LSMTree) compactSSTablesLocked() error {
	if len(l.sstables) <= 1 || l.Dir == "" {
		return nil
	}

	merged := make(map[string]storagelsm.Record)
	for _, path := range l.sstables {
		records, err := storagelsm.ReadAllRecords(path)
		if err != nil {
			return err
		}
		for k, rec := range records {
			if _, exists := merged[k]; exists {
				continue
			}
			merged[k] = rec
		}
	}

	target := storagelsm.BuildSSTablePath(l.Dir, l.nextID)
	if err := storagelsm.WriteSSTable(target, merged, true); err != nil {
		return err
	}

	old := append([]string(nil), l.sstables...)
	l.sstables = []string{target}
	l.nextID++
	for _, p := range old {
		_ = os.Remove(p)
	}

	return nil
}
