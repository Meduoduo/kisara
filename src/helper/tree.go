package helper

import "errors"

type Compareable interface {
	int | int8 | int16 | int32 | int64 |
		uint | uint8 | uint16 | uint32 | uint64 |
		string
}

// Not thread safety
type Tree[K Compareable, V any] struct {
	root     bool
	key      K
	parent   *Tree[K, V]
	value    *V
	children map[K]*Tree[K, V]
}

var (
	ErrTreeNotFoundParent = errors.New("not found parent node")
)

func NewTree[K, V Compareable]() *Tree[K, V] {
	return &Tree[K, V]{
		root: true,
	}
}

func (t *Tree[K, V]) GetAllChildren() map[K]*V {
	m := make(map[K]*V)
	if len(t.children) == 0 {
		return m
	}

	for k, v := range t.children {
		m[k] = v.value
		children := v.GetAllChildren()
		for ck, cv := range children {
			m[ck] = cv
		}
	}

	return m
}

func (t *Tree[K, V]) GetNode(id K) *Tree[K, V] {
	root := t
	if root.key == id {
		return t
	}

	if len(root.children) == 0 {
		return nil
	}

	for _, v := range root.children {
		target := v.GetNode(id)
		if target != nil {
			return target
		}
	}

	return nil
}

// check if the node holds parent_id as key is parent of node holds client_id as key
func (t *Tree[K, V]) CheckIfParent(parent_id K, client_id K) bool {
	// walk through all nodes to find parent
	parent := t.GetNode(parent_id)
	if parent == nil {
		return false
	}

	client := parent.GetNode(client_id)

	return client != nil
}

// add node
func (t *Tree[K, V]) AddNode(key K, value *V) {
	node := &Tree[K, V]{
		parent: t,
		key:    key,
		value:  value,
	}

	if t.children == nil {
		t.children = make(map[K]*Tree[K, V])
	}

	t.children[key] = node
}

// add parent
// if you want to add parent to a relative root node
func (t *Tree[K, V]) AddParent(key K, value *V) {
	node := &Tree[K, V]{
		key:      key,
		value:    value,
		children: make(map[K]*Tree[K, V]),
	}

	current_parent := t.parent
	// if t is root node
	if current_parent == nil {
		t.parent = node
		node.children[t.key] = t
	} else {
		// if t is not root node
		// parent -> t
		// remove t from parent
		delete(current_parent.children, t.key)
		// parent t
		// set node to parent's children field
		current_parent.children[key] = node
		// parent -> node t
		// set t to node's children field
		node.children[t.key] = t
		// parent -> node -> t
	}
}

// add to parent
func (t *Tree[K, V]) AddToParent(parent_id K, key K, value *V) error {
	parent := t.GetNode(parent_id)
	if parent != nil {
		parent.AddNode(key, value)
	} else {
		return ErrTreeNotFoundParent
	}
	return nil
}

// remove key from children
func (t *Tree[K, V]) Remove(key K) {
	if len(t.children) == 0 {
		return
	}

	for id, v := range t.children {
		if id == key {
			delete(t.children, id)
		} else {
			v.Remove(key)
		}
	}
}

// Walk tree from bottom leaf to root
func (t *Tree[K, V]) WalkReverse(cb func(K, *V)) {
	if len(t.children) != 0 {
		for _, v := range t.children {
			v.WalkReverse(cb)
		}
	}

	if !t.root {
		cb(t.key, t.value)
	}
}
