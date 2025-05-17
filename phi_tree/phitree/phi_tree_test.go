package phitree

import (
	"math"
	"testing"
)

// TestBasicOperations verifies basic tree operations
func TestBasicOperations(t *testing.T) {
	tree := New[string]()

	// Test empty tree
	if _, err := tree.Get(1); err != ErrKeyNotFound {
		t.Errorf("Expected ErrKeyNotFound for empty tree, got %v", err)
	}

	// Test insertion and retrieval
	testCases := []struct {
		key   int
		value string
		valid bool
	}{
		{1, "one", true},
		{2, "two", true},
		{3, "three", true},
		{5, "five", true},
		{8, "eight", true},
		{13, "thirteen", true},
		{-1, "negative", false},
	}

	for _, tc := range testCases {
		t.Logf("TEST: Insert(%d, %s)", tc.key, tc.value)
		if ok := tree.Insert(tc.key, tc.value); ok != tc.valid {
			t.Errorf("Insert(%d) returned %v, want %v", tc.key, ok, tc.valid)
		}
		if tc.valid {
			t.Logf("TEST: Get(%d) after Insert", tc.key)
			got, err := tree.Get(tc.key)
			if err != nil {
				t.Errorf("Get(%d) failed: %v", tc.key, err)
			}
			if got != tc.value {
				t.Errorf("Get(%d) = %v, want %v", tc.key, got, tc.value)
			}
			t.Logf("Successfully inserted and retrieved key %d with value %s", tc.key, tc.value)
		} else {
			t.Logf("TEST: Get(%d) after failed Insert", tc.key)
			if _, err := tree.Get(tc.key); err != ErrKeyNotFound {
				t.Errorf("Expected ErrKeyNotFound for invalid key %d, got %v", tc.key, err)
			}
		}
	}

	t.Logf("TEST: Get(4) for non-existent key")
	if _, err := tree.Get(4); err != ErrKeyNotFound {
		t.Errorf("Expected ErrKeyNotFound for non-existent key, got %v", err)
	}

	t.Logf("TEST: Get(5) before MustGet")
	if got, err := tree.Get(5); err != nil {
		t.Logf("WARNING: Key 5 is not in tree before MustGet! Error: %v", err)
	} else {
		t.Logf("Key 5 is in tree before MustGet with value: %s", got)
	}

	t.Logf("TEST: MustGet(5)")
	if got := tree.MustGet(5); got != "five" {
		t.Errorf("MustGet(5) = %v, want five", got)
	}

	t.Logf("TEST: MustGet(4) expecting panic")
	defer func() {
		if r := recover(); r != ErrKeyNotFound {
			t.Errorf("MustGet(4) panicked with %v, want %v", r, ErrKeyNotFound)
		} else {
			t.Logf("Recovered expected panic: %v", r)
		}
	}()
	tree.MustGet(4)
}

// TestConcurrentOperations verifies thread safety
func TestConcurrentOperations(t *testing.T) {
	tree := New[int]()
	done := make(chan bool)

	// Concurrent writers
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				key := id*1000 + j
				tree.Insert(key, key*2)
			}
			done <- true
		}(i)
	}

	// Concurrent readers
	for i := 0; i < 5; i++ {
		go func() {
			for j := 0; j < 1000; j++ {
				tree.Get(j)
				tree.MaxDepth()
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 15; i++ {
		<-done
	}

	// Verify all values were inserted correctly
	for i := 0; i < 10; i++ {
		for j := 0; j < 100; j++ {
			key := i*1000 + j
			if value, err := tree.Get(key); err != nil || value != key*2 {
				t.Errorf("Concurrent operation failed: Get(%d) = %v, %v, want %d, nil",
					key, value, err, key*2)
			}
		}
	}
}

// TestClear verifies the Clear operation
func TestClear(t *testing.T) {
	tree := New[string]()

	// Insert some values
	tree.Insert(1, "one")
	tree.Insert(2, "two")

	// Clear the tree
	tree.Clear()

	// Verify tree is empty
	if _, err := tree.Get(1); err != ErrKeyNotFound {
		t.Error("Expected ErrKeyNotFound after Clear")
	}
	if tree.MaxDepth() != 0 {
		t.Errorf("Expected max depth 0 after Clear, got %d", tree.MaxDepth())
	}
}

// TestDepthProperties verifies the mathematical properties of the tree depth
func TestDepthProperties(t *testing.T) {
	tree := New[int]()

	// Insert keys in a way that exercises different depths
	keys := []int{1, 2, 3, 5, 8, 13, 21, 34, 55, 89, 144}

	for _, key := range keys {
		tree.Insert(key, key)
		// The depth should be at most log_phi(key) + 1
		maxAllowedDepth := int(math.Ceil(math.Log(float64(key))/math.Log(Phi))) + 1
		if tree.MaxDepth() > maxAllowedDepth {
			t.Errorf("Key %d: depth %d exceeds theoretical maximum %d",
				key, tree.MaxDepth(), maxAllowedDepth)
		}
	}
}

// TestGenericTypes verifies the tree works with different types
func TestGenericTypes(t *testing.T) {
	// Test with int
	intTree := New[int]()
	intTree.Insert(1, 42)
	if got, _ := intTree.Get(1); got != 42 {
		t.Errorf("int tree: got %v, want 42", got)
	}

	// Test with float64
	floatTree := New[float64]()
	floatTree.Insert(1, 3.14)
	if got, _ := floatTree.Get(1); got != 3.14 {
		t.Errorf("float tree: got %v, want 3.14", got)
	}

	// Test with custom type
	type Point struct{ x, y int }
	pointTree := New[Point]()
	pointTree.Insert(1, Point{1, 2})
	if got, _ := pointTree.Get(1); got != (Point{1, 2}) {
		t.Errorf("point tree: got %v, want {1 2}", got)
	}
}

// TestZeckendorfProperties verifies the Zeckendorf representation properties
func TestZeckendorfProperties(t *testing.T) {
	tree := New[int]() // Create a tree to access the cached Fibonacci list

	// Test that no number has consecutive 1s in its representation
	for n := 1; n <= 1000; n++ {
		// Ensure we have enough Fibonacci numbers
		tree.ensureFibonacciCapacity(n)

		// Get a read lock to access the Fibonacci list
		tree.fibMu.RLock()
		fibs := tree.fibs
		tree.fibMu.RUnlock()

		// Convert to Zeckendorf representation
		bits := make([]int, 0, len(fibs))
		remaining := n
		for i := len(fibs) - 1; i >= 0; i-- {
			if fibs[i] <= remaining {
				bits = append(bits, 1)
				remaining -= fibs[i]
			} else {
				bits = append(bits, 0)
			}
		}

		// Check for consecutive 1s
		for i := 0; i < len(bits)-1; i++ {
			if bits[i] == 1 && bits[i+1] == 1 {
				t.Errorf("Number %d has consecutive 1s in its representation: %v", n, bits)
			}
		}
	}

	// Test that each number has a unique representation
	seen := make(map[string]bool)
	for n := 1; n <= 1000; n++ {
		// Ensure we have enough Fibonacci numbers
		tree.ensureFibonacciCapacity(n)

		// Get a read lock to access the Fibonacci list
		tree.fibMu.RLock()
		fibs := tree.fibs
		tree.fibMu.RUnlock()

		// Convert to Zeckendorf representation
		bits := make([]int, 0, len(fibs))
		remaining := n
		for i := len(fibs) - 1; i >= 0; i-- {
			if fibs[i] <= remaining {
				bits = append(bits, 1)
				remaining -= fibs[i]
			} else {
				bits = append(bits, 0)
			}
		}

		// Create a string key for the representation
		key := ""
		for _, b := range bits {
			key += string('0' + byte(b))
		}
		if seen[key] {
			t.Errorf("Number %d has a duplicate representation: %s", n, key)
		}
		seen[key] = true
	}
}

// TestLargeNumbers verifies the tree works with large numbers
func TestLargeNumbers(t *testing.T) {
	tree := New[int]()
	largeKeys := []int{
		1_000_000,
		2_000_000,
		5_000_000,
		10_000_000,
	}

	for _, key := range largeKeys {
		tree.Insert(key, key*2)
		if got, err := tree.Get(key); err != nil || got != key*2 {
			t.Errorf("Large key %d: got %v, %v, want %d, nil",
				key, got, err, key*2)
		}
	}
}

// TestEdgeCases verifies behavior with edge cases and invalid inputs
func TestEdgeCases(t *testing.T) {
	tree := New[int]()

	// Test zero
	if !tree.Insert(0, 42) {
		t.Error("Insert(0) failed")
	}
	if got, err := tree.Get(0); err != nil || got != 42 {
		t.Errorf("Zero key: got %v, %v, want 42, nil", got, err)
	}

	// Test negative numbers
	if tree.Insert(-1, 43) {
		t.Error("Insert(-1) succeeded when it should have failed")
	}
	if _, err := tree.Get(-1); err != ErrKeyNotFound {
		t.Errorf("Expected ErrKeyNotFound for negative key, got %v", err)
	}

	// Test very large numbers
	largeKey := math.MaxInt32
	if !tree.Insert(largeKey, largeKey*2) {
		t.Error("Insert(MaxInt32) failed")
	}
	if got, err := tree.Get(largeKey); err != nil || got != largeKey*2 {
		t.Errorf("Large key: got %v, %v, want %d, nil", got, err, largeKey*2)
	}

	// Test overflow case
	overflowKey := math.MaxInt64
	if !tree.Insert(overflowKey, 1) {
		t.Error("Insert(MaxInt64) failed")
	}
	if got, err := tree.Get(overflowKey); err != nil || got != 1 {
		t.Errorf("Overflow key: got %v, %v, want 1, nil", got, err)
	}

	// Test concurrent operations with edge cases
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			// Mix of valid and invalid operations
			tree.Insert(0, 0)
			tree.Insert(-1, -1)
			tree.Get(0)
			tree.Get(-1)
			done <- true
		}()
	}
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify final state
	if got, err := tree.Get(0); err != nil || got != 0 {
		t.Errorf("After concurrent operations: Get(0) = %v, %v, want 0, nil", got, err)
	}
	if _, err := tree.Get(-1); err != ErrKeyNotFound {
		t.Errorf("After concurrent operations: expected ErrKeyNotFound for -1, got %v", err)
	}
}
