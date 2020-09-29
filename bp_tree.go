package pidb

type Tree struct {
	Root *Node
}

type Node struct {
	Pointers []interface{}
	Keys     []string
	Parent   *Node
	IsLeaf   bool
	NumKeys  int
	Next     *Node
}

func New() *Tree {
	return &Tree{}
}

func (t *Tree) Insert(key string, value []byte) error {
	//TODO:
	return nil
}

func (t *Tree) Find(key string) (*Record, error) {
	//TODO:
	return nil, nil
}

func (t *Tree) Delete(key string) error {
	//TODO:
	return nil
}
