package phitree

import (
	"errors"
	"math"
	"sync"
)

// Phi is the golden ratio constant
var Phi = (1 + math.Sqrt(5)) / 2

// ErrKeyNotFound is returned when attempting to get a non-existent key
var ErrKeyNotFound = errors.New("key not found")

// Node represents a node in the PhiTree
type Node[T any] struct {
	value T
	left  *Node[T]
	right *Node[T]
}

// reset resets a node for reuse
func (n *Node[T]) reset() {
	var zero T
	n.value = zero
	n.left = nil
	n.right = nil
}

// PhiTree is a tree data structure that uses Zeckendorf representation
// for efficient storage and retrieval.
type PhiTree[T any] struct {
	root     *Node[T]
	maxDepth int
	pool     sync.Pool
	fibs     []int // Cached Fibonacci numbers
	fibMu    sync.RWMutex
}

// New creates a new empty PhiTree
func New[T any]() *PhiTree[T] {
	t := &PhiTree[T]{
		fibs: []int{1, 2}, // Initialize with base Fibonacci numbers
	}
	t.pool = sync.Pool{
		New: func() interface{} {
			return &Node[T]{}
		},
	}
	// Create root node from pool and initialize it
	t.root = t.pool.Get().(*Node[T])
	t.root.reset()
	return t
}

// ensureFibonacciCapacity ensures the Fibonacci list is large enough for n
func (t *PhiTree[T]) ensureFibonacciCapacity(n int) {
	t.fibMu.RLock()
	if len(t.fibs) > 0 && t.fibs[len(t.fibs)-1] > n {
		t.fibMu.RUnlock()
		return
	}
	t.fibMu.RUnlock()

	t.fibMu.Lock()
	defer t.fibMu.Unlock()

	// Double check after acquiring write lock
	if len(t.fibs) > 0 && t.fibs[len(t.fibs)-1] > n {
		return
	}

	// Grow the Fibonacci list until we exceed n
	for t.fibs[len(t.fibs)-1] <= n {
		next := t.fibs[len(t.fibs)-1] + t.fibs[len(t.fibs)-2]
		if next <= t.fibs[len(t.fibs)-1] { // Check for overflow
			break
		}
		t.fibs = append(t.fibs, next)
	}
}

// zeckendorfBits returns the Zeckendorf representation of n as a slice of bits (MSB first)
func zeckendorfBits(fibs []int, n int) []int {
	bits := make([]int, 0, len(fibs))
	remaining := n
	used := false
	for i := len(fibs) - 1; i >= 0; i-- {
		if fibs[i] <= remaining {
			bits = append(bits, 1)
			remaining -= fibs[i]
			used = true
		} else if used {
			bits = append(bits, 0)
		}
	}
	return bits
}

// walk traverses the tree following the Zeckendorf representation of key
// if create is true, it creates missing nodes
func (t *PhiTree[T]) walk(key int, create bool) (*Node[T], int) {
	if key < 0 {
		return nil, 0
	}
	if key == 0 {
		return t.root, 0
	}
	t.ensureFibonacciCapacity(key)
	t.fibMu.RLock()
	fibs := t.fibs
	t.fibMu.RUnlock()

	bits := zeckendorfBits(fibs, key)
	node := t.root
	depth := 0
	for _, bit := range bits {
		if bit == 1 {
			next := node.right
			if next == nil {
				if !create {
					return nil, depth
				}
				next = t.pool.Get().(*Node[T])
				next.reset()
				node.right = next
			}
			node = next
		} else {
			next := node.left
			if next == nil {
				if !create {
					return nil, depth
				}
				next = t.pool.Get().(*Node[T])
				next.reset()
				node.left = next
			}
			node = next
		}
		depth++
	}
	return node, depth
}

// Insert adds a value to the tree at the given key.
// Returns false if the key is invalid (negative or would cause overflow).
func (t *PhiTree[T]) Insert(key int, value T) bool {
	if key < 0 {
		return false
	}
	node, depth := t.walk(key, true)
	if node == nil {
		return false
	}
	node.value = value
	if depth > t.maxDepth {
		t.maxDepth = depth
	}
	return true
}

// Get retrieves the value stored at the given key.
// Returns ErrKeyNotFound if the key doesn't exist or is invalid.
func (t *PhiTree[T]) Get(key int) (T, error) {
	if key < 0 {
		var zero T
		return zero, ErrKeyNotFound
	}
	node, _ := t.walk(key, false)
	if node == nil {
		var zero T
		return zero, ErrKeyNotFound
	}
	return node.value, nil
}

// MustGet retrieves the value stored at the given key.
// Panics if the key doesn't exist.
func (t *PhiTree[T]) MustGet(key int) T {
	value, err := t.Get(key)
	if err != nil {
		panic(err)
	}
	return value
}

// MaxDepth returns the maximum depth of the tree.
func (t *PhiTree[T]) MaxDepth() int {
	return t.maxDepth
}

// Clear removes all elements from the tree and returns nodes to the pool.
func (t *PhiTree[T]) Clear() {
	// Recursively return nodes to the pool
	var clearNode func(*Node[T])
	clearNode = func(n *Node[T]) {
		if n == nil {
			return
		}
		clearNode(n.left)
		clearNode(n.right)
		n.reset()
		t.pool.Put(n)
	}

	// Clear children of root and reset root
	clearNode(t.root.left)
	clearNode(t.root.right)
	t.root.reset()
	t.root.left = nil
	t.root.right = nil
	t.maxDepth = 0
}
