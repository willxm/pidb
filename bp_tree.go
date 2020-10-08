package pidb

import (
	"bytes"
	"errors"
)

const (
	order = 5
)

var (
	ErrKeyNotFound = errors.New("key not found")
)

type Tree struct {
	Root *Node
}

type Node struct {
	Pointers []interface{}
	Keys     [][]byte
	Parent   *Node
	IsLeaf   bool
	NumKeys  int
	Next     *Node
}

func New() *Tree {
	return &Tree{}
}

var queue *Node

func enqueue(node *Node) {
	var c *Node

	if queue == nil {
		queue = node
		queue.Next = nil
	} else {
		c = queue
		for c.Next != nil {
			c = c.Next
		}
		c.Next = node
		node.Next = nil
	}
}

func dequeue() *Node {
	n := queue
	queue = queue.Next

	return n
}

func (t *Tree) hight() int {
	h := 0
	p := t.Root
	for !p.IsLeaf {
		p = p.Pointers[0].(*Node)
		h++
	}
	return h
}

func (t *Tree) findRange(start, end []byte) (numFound int, keys [][]byte, pointers []interface{}) {
	var (
		n        *Node
		i, j     int
		scanFlag bool
	)

	if n = t.FindLeaf(start); n == nil {
		return 0, nil, nil
	}

	for j = 0; j < n.NumKeys && bytes.Compare(n.Keys[j], start) < 0; {
		j++
	}

	scanFlag = true
	for n != nil && scanFlag {
		for i = j; i < n.NumKeys; i++ {
			if bytes.Compare(n.Keys[i], end) > 0 {
				scanFlag = false
				break
			}
			keys = append(keys, n.Keys[i])
			pointers = append(pointers, n.Pointers[i])
			numFound++
		}

		n, _ = n.Pointers[order-1].(*Node)

		j = 0
	}

	return
}

func (t *Tree) FindLeaf(key []byte) *Node {
	var i int
	var curr *Node

	if curr = t.Root; curr == nil {
		return nil
	}

	for !curr.IsLeaf {
		i = 0
		for i < curr.NumKeys {
			if bytes.Compare(key, curr.Keys[i]) >= 0 {
				i++
			} else {
				break
			}
		}
		curr = curr.Pointers[i].(*Node)
	}

	return curr
}

func (t *Tree) Find(key []byte) (*Record, error) {
	var (
		leaf *Node
		i    int
	)

	// Find leaf by key.
	leaf = t.FindLeaf(key)

	if leaf == nil {
		return nil, ErrKeyNotFound
	}

	for i = 0; i < leaf.NumKeys; i++ {
		if bytes.Compare(key, leaf.Keys[i]) == 0 {
			break
		}
	}

	if i == leaf.NumKeys {
		return nil, ErrKeyNotFound
	}

	return leaf.Pointers[i].(*Record), nil
}

func (t *Tree) Insert(key string, value []byte) error {
	//TODO:
	return nil
}

func (t *Tree) Delete(key string) error {
	//TODO:
	return nil
}
