package pidb

import "fmt"

const (
	order = 5
)

// Tree is bp tree
type BpTree struct {
	Root *Node
}

// Node is node of bp tree
type Node struct {
	Pointers []interface{}
	Keys     [][]byte
	Parent   *Node
	IsLeaf   bool
	NumKeys  int
	Next     *Node
}

func (t *BpTree) findLeaf(key string) *Node {
	var i int
	node := t.Root
	for !node.IsLeaf {
		i = 0
		for i < node.NumKeys {
			if string(key) < string(node.Keys[i]) {
				break
			}
			i++
		}
		node = node.Pointers[i].(*Node)
	}
	return node
}

func (t *BpTree) insertIntoLeaf(node *Node, key string, value []byte) {
	var i, insertionPoint int
	insertionPoint = 0
	for insertionPoint < node.NumKeys && string(node.Keys[insertionPoint]) < string(key) {
		insertionPoint++
	}
	for i = node.NumKeys; i > insertionPoint; i-- {
		node.Keys[i] = node.Keys[i-1]
		node.Pointers[i] = node.Pointers[i-1]
	}
	node.Keys[insertionPoint] = []byte(key)
	node.Pointers[insertionPoint] = value
	node.NumKeys++
}

func (t *BpTree) insertIntoLeafAfterSplitting(node *Node, key string, value []byte) {
	var newLeaf, child *Node
	var insertionIndex, i, j int
	var newKey []byte
	newLeaf = &Node{
		Pointers: make([]interface{}, order),
		Keys:     make([][]byte, order-1),
		Parent:   nil,
		IsLeaf:   true,
		NumKeys:  0,
		Next:     nil,
	}
	insertionIndex = 0
	for insertionIndex < order-1 && string(node.Keys[insertionIndex]) < string(key) {
		insertionIndex++
	}
	for i, j = 0, 0; i < node.NumKeys; i, j = i+1, j+1 {
		if j == insertionIndex {
			j++
		}
		newLeaf.Keys[j] = node.Keys[i]
		newLeaf.Pointers[j] = node.Pointers[i]
	}
	newLeaf.Keys[insertionIndex] = []byte(key)
	newLeaf.Pointers[insertionIndex] = value
	newLeaf.NumKeys = order - 1
	node.NumKeys = order - 1
	for i = node.NumKeys; i < order-1; i++ {
		node.Keys[i] = nil
		node.Pointers[i] = nil
	}
	for i = newLeaf.NumKeys; i < order-1; i++ {
		newLeaf.Keys[i] = nil
		newLeaf.Pointers[i] = nil
	}
	newLeaf.Next = node.Next
	node.Next = newLeaf
	newKey = newLeaf.Keys[0]
	child = newLeaf
	for node.Parent != nil {
		node = node.Parent
		t.insertIntoNode(node, child, newKey)
		child = node
	}
	t.insertIntoNewRoot(child, newKey)
}

func (t *BpTree) insertIntoNode(node *Node, right *Node, key []byte) {
	var i, insertionPoint int
	insertionPoint = 0
	for insertionPoint < node.NumKeys && string(node.Keys[insertionPoint]) < string(key) {
		insertionPoint++
	}
	for i = node.NumKeys; i > insertionPoint; i-- {
		node.Keys[i] = node.Keys[i-1]
		node.Pointers[i+1] = node.Pointers[i]
	}
	node.Keys[insertionPoint] = key
	node.Pointers[insertionPoint+1] = right
	node.NumKeys++
}

func (t *BpTree) insertIntoNewRoot(left *Node, key []byte) {
	root := &Node{
		Pointers: make([]interface{}, order),
		Keys:     make([][]byte, order-1),
		Parent:   nil,
		IsLeaf:   false,
		NumKeys:  1,
		Next:     nil,
	}
	root.Keys[0] = key
	root.Pointers[0] = left
	root.Pointers[1] = left.Next
	left.Parent = root
	left.Next.Parent = root
	t.Root = root
}

// New is initialize bp tree
func (t *BpTree) New() error {
	t.Root = &Node{
		Pointers: make([]interface{}, order),
		Keys:     make([][]byte, order-1),
		Parent:   nil,
		IsLeaf:   true,
		NumKeys:  0,
		Next:     nil,
	}
	return nil
}

// Set is set key and value to bp tree
func (t *BpTree) Set(key string, value []byte) error {
	var left *Node
	if ok, _ := t.Exists(key); ok {
		return fmt.Errorf("key %s already exists", key)
	}
	left = t.findLeaf(key)
	if left.NumKeys < order-1 {
		t.insertIntoLeaf(left, key, value)
	} else {
		t.insertIntoLeafAfterSplitting(left, key, value)
	}
	return nil
}

// Get is get value from bp tree
func (t *BpTree) Get(key string) ([]byte, error) {
	var i int
	node := t.findLeaf(key)
	for i = 0; i < node.NumKeys; i++ {
		if string(node.Keys[i]) == key {
			return node.Pointers[i].([]byte), nil
		}
	}
	return nil, nil
}

// Del is delete key from bp tree
func (t *BpTree) Del(key string) error {
	//TODO:
	return nil
}

// Exists is check key exists in bp tree
func (t *BpTree) Exists(key string) (bool, error) {
	var i int
	node := t.findLeaf(key)
	for i = 0; i < node.NumKeys; i++ {
		if string(node.Keys[i]) == key {
			return true, nil
		}
	}
	return false, nil
}
