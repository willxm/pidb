package engines

import (
	"fmt"
	"sort"
	"sync"

	storagebptree "github.com/willxm/pidb/storage/bptree"
)

const bpTreeMaxKeys = 4

type BpTree struct {
	File string

	mu   sync.RWMutex
	root *bpNode
	wal  *storagebptree.WAL
}

type bpNode struct {
	isLeaf   bool
	keys     []string
	values   [][]byte
	children []*bpNode
	next     *bpNode
}

func (t *BpTree) New() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.wal != nil {
		_ = t.wal.Close()
		t.wal = nil
	}
	t.root = &bpNode{isLeaf: true}

	if t.File == "" {
		return nil
	}

	wal, err := storagebptree.Open(t.File)
	if err != nil {
		return err
	}
	if err := wal.Replay(func(entry storagebptree.Entry) error {
		switch entry.Op {
		case "S":
			return t.setLocked(entry.Key, entry.Value, false, false)
		case "D":
			return t.delLocked(entry.Key, false)
		default:
			return fmt.Errorf("invalid bptree wal op: %s", entry.Op)
		}
	}); err != nil {
		_ = wal.Close()
		return err
	}
	t.wal = wal
	return nil
}

func (t *BpTree) Set(key string, value []byte) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.setLocked(key, value, true, true)
}

func (t *BpTree) setLocked(key string, value []byte, rejectDuplicate bool, appendLog bool) error {

	if t.root == nil {
		t.root = &bpNode{isLeaf: true}
	}

	promote, right, split, err := t.insertRecursive(t.root, key, append([]byte(nil), value...), rejectDuplicate)
	if err != nil {
		return err
	}
	if split {
		t.root = &bpNode{
			isLeaf:   false,
			keys:     []string{promote},
			children: []*bpNode{t.root, right},
		}
	}
	if appendLog && t.wal != nil {
		if err := t.wal.AppendSet(key, value); err != nil {
			return err
		}
	}
	return nil
}

func (t *BpTree) Get(key string) ([]byte, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	leaf := t.findLeaf(key)
	if leaf == nil {
		return nil, nil
	}
	index := sort.SearchStrings(leaf.keys, key)
	if index < len(leaf.keys) && leaf.keys[index] == key {
		return append([]byte(nil), leaf.values[index]...), nil
	}
	return nil, nil
}

func (t *BpTree) Del(key string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.delLocked(key, true)
}

func (t *BpTree) delLocked(key string, appendLog bool) error {

	leaf := t.findLeaf(key)
	if leaf == nil {
		return nil
	}
	index := sort.SearchStrings(leaf.keys, key)
	if index >= len(leaf.keys) || leaf.keys[index] != key {
		return nil
	}

	leaf.keys = append(leaf.keys[:index], leaf.keys[index+1:]...)
	leaf.values = append(leaf.values[:index], leaf.values[index+1:]...)
	if appendLog && t.wal != nil {
		if err := t.wal.AppendDel(key); err != nil {
			return err
		}
	}
	return nil
}

func (t *BpTree) Exists(key string) (bool, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	leaf := t.findLeaf(key)
	if leaf == nil {
		return false, nil
	}
	index := sort.SearchStrings(leaf.keys, key)
	return index < len(leaf.keys) && leaf.keys[index] == key, nil
}

func (t *BpTree) findLeaf(key string) *bpNode {
	node := t.root
	if node == nil {
		return nil
	}
	for !node.isLeaf {
		nextIndex := sort.Search(len(node.keys), func(i int) bool {
			return key < node.keys[i]
		})
		node = node.children[nextIndex]
	}
	return node
}

func (t *BpTree) insertRecursive(node *bpNode, key string, value []byte, rejectDuplicate bool) (string, *bpNode, bool, error) {
	if node.isLeaf {
		index := sort.SearchStrings(node.keys, key)
		if index < len(node.keys) && node.keys[index] == key {
			if rejectDuplicate {
				return "", nil, false, fmt.Errorf("key %s already exists", key)
			}
			node.values[index] = append([]byte(nil), value...)
			return "", nil, false, nil
		}
		node.keys = append(node.keys, "")
		copy(node.keys[index+1:], node.keys[index:])
		node.keys[index] = key

		node.values = append(node.values, nil)
		copy(node.values[index+1:], node.values[index:])
		node.values[index] = value

		if len(node.keys) <= bpTreeMaxKeys {
			return "", nil, false, nil
		}

		mid := len(node.keys) / 2
		right := &bpNode{isLeaf: true}
		right.keys = append(right.keys, node.keys[mid:]...)
		right.values = append(right.values, node.values[mid:]...)
		node.keys = node.keys[:mid]
		node.values = node.values[:mid]

		right.next = node.next
		node.next = right

		return right.keys[0], right, true, nil
	}

	childIndex := sort.Search(len(node.keys), func(i int) bool {
		return key < node.keys[i]
	})
	promote, right, split, err := t.insertRecursive(node.children[childIndex], key, value, rejectDuplicate)
	if err != nil {
		return "", nil, false, err
	}
	if !split {
		return "", nil, false, nil
	}

	node.keys = append(node.keys, "")
	copy(node.keys[childIndex+1:], node.keys[childIndex:])
	node.keys[childIndex] = promote

	node.children = append(node.children, nil)
	copy(node.children[childIndex+2:], node.children[childIndex+1:])
	node.children[childIndex+1] = right

	if len(node.keys) <= bpTreeMaxKeys {
		return "", nil, false, nil
	}

	mid := len(node.keys) / 2
	newPromote := node.keys[mid]
	rightInternal := &bpNode{isLeaf: false}
	rightInternal.keys = append(rightInternal.keys, node.keys[mid+1:]...)
	rightInternal.children = append(rightInternal.children, node.children[mid+1:]...)

	node.keys = node.keys[:mid]
	node.children = node.children[:mid+1]

	return newPromote, rightInternal, true, nil
}

func (t *BpTree) Shutdown() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.wal == nil {
		return nil
	}
	err := t.wal.Close()
	t.wal = nil
	return err
}
