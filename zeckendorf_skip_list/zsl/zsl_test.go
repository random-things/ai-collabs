// Package zsl_test contains tests for the zsl package.
// These tests verify the correctness of both BlockedZSL and StaticZSL implementations,
// including their core operations and edge cases.
package zsl

import (
	"sort"
	"testing"
)

// verifySorted is a test helper that verifies a slice of integers is sorted.
// It fails the test if the slice is not in ascending order.
func verifySorted(t *testing.T, keys []int) {
	t.Helper()
	if !sort.IntsAreSorted(keys) {
		t.Errorf("keys not sorted: %v", keys)
	}
}

// verifyKeys is a test helper that compares two slices of integers for equality.
// It fails the test if the slices have different lengths or contain different values.
func verifyKeys(t *testing.T, got, want []int) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("got %d keys, want %d", len(got), len(want))
		return
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("keys[%d] = %d, want %d", i, got[i], want[i])
		}
	}
}

// TestStaticZSL tests the StaticZSL implementation, which uses Fibonacci-based skip pointers.
// It verifies the correctness of all operations including construction, search,
// insertion, deletion, and range queries.
func TestStaticZSL(t *testing.T) {
	// TestNewStaticZSL verifies that NewStaticZSL correctly constructs a new skip list
	// from various input sequences, handling sorting and deduplication.
	t.Run("NewStaticZSL", func(t *testing.T) {
		tests := []struct {
			name     string
			keys     []int
			wantSize int
		}{
			{"empty", []int{}, 0},                   // Test empty input
			{"single", []int{1}, 1},                 // Test single element
			{"sorted", []int{1, 2, 3, 4, 5}, 5},     // Test already sorted input
			{"duplicates", []int{1, 1, 2, 2, 3}, 3}, // Test duplicate removal
			{"unsorted", []int{5, 3, 1, 4, 2}, 5},   // Test sorting of input
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				zsl := NewStaticZSL(tt.keys)
				if got := zsl.GetSize(); got != tt.wantSize {
					t.Errorf("GetSize() = %d, want %d", got, tt.wantSize)
				}
				verifySorted(t, zsl.GetKeys())
			})
		}
	})

	// TestSearch verifies that Search correctly identifies the presence or absence
	// of keys in the skip list, including edge cases at the boundaries.
	t.Run("Search", func(t *testing.T) {
		keys := []int{1, 3, 5, 7, 9}
		zsl := NewStaticZSL(keys)

		tests := []struct {
			key  int
			want bool
		}{
			{0, false},  // Test below range
			{1, true},   // Test first element
			{3, true},   // Test middle element
			{9, true},   // Test last element
			{10, false}, // Test above range
			{2, false},  // Test between elements
			{4, false},  // Test between elements
		}

		for _, tt := range tests {
			t.Run("", func(t *testing.T) {
				if got := zsl.Search(tt.key); got != tt.want {
					t.Errorf("Search(%d) = %v, want %v", tt.key, got, tt.want)
				}
			})
		}
	})

	// TestInsert verifies that Insert correctly adds new keys while maintaining
	// sorted order and handling duplicates appropriately.
	t.Run("Insert", func(t *testing.T) {
		tests := []struct {
			name     string
			initial  []int
			insert   int
			wantKeys []int
		}{
			{"empty", []int{}, 1, []int{1}},                  // Test empty list
			{"beginning", []int{2, 3}, 1, []int{1, 2, 3}},    // Test insert at start
			{"middle", []int{1, 3}, 2, []int{1, 2, 3}},       // Test insert in middle
			{"end", []int{1, 2}, 3, []int{1, 2, 3}},          // Test insert at end
			{"duplicate", []int{1, 2, 3}, 2, []int{1, 2, 3}}, // Test duplicate handling
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				zsl := NewStaticZSL(tt.initial)
				zsl.Insert(tt.insert)
				verifyKeys(t, zsl.GetKeys(), tt.wantKeys)
			})
		}
	})

	// TestDelete verifies that Delete correctly removes keys and returns
	// appropriate success/failure status.
	t.Run("Delete", func(t *testing.T) {
		tests := []struct {
			name     string
			initial  []int
			delete   int
			wantKeys []int
			wantOk   bool
		}{
			{"empty", []int{}, 1, []int{}, false},                   // Test empty list
			{"single", []int{1}, 1, []int{}, true},                  // Test single element
			{"first", []int{1, 2, 3}, 1, []int{2, 3}, true},         // Test delete first
			{"middle", []int{1, 2, 3}, 2, []int{1, 3}, true},        // Test delete middle
			{"last", []int{1, 2, 3}, 3, []int{1, 2}, true},          // Test delete last
			{"not found", []int{1, 2, 3}, 4, []int{1, 2, 3}, false}, // Test non-existent
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				zsl := NewStaticZSL(tt.initial)
				if got := zsl.Delete(tt.delete); got != tt.wantOk {
					t.Errorf("Delete() = %v, want %v", got, tt.wantOk)
				}
				verifyKeys(t, zsl.GetKeys(), tt.wantKeys)
			})
		}
	})

	// TestRangeFunc verifies that RangeFunc correctly iterates over keys
	// in the specified range, handling various boundary conditions.
	t.Run("RangeFunc", func(t *testing.T) {
		keys := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
		zsl := NewStaticZSL(keys)

		tests := []struct {
			name     string
			lo, hi   int
			wantKeys []int
		}{
			{"empty", 10, 20, []int{}},           // Test empty range
			{"full", 1, 10, keys},                // Test full range
			{"partial", 3, 7, []int{3, 4, 5, 6}}, // Test partial range
			{"single", 5, 6, []int{5}},           // Test single element
			{"nonexistent", 2, 2, []int{}},       // Test empty range
			{"below", 0, 1, []int{}},             // Test below range
			{"above", 9, 10, []int{9}},           // Test above range
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var got []int
				zsl.RangeFunc(tt.lo, tt.hi, func(k int) {
					got = append(got, k)
				})
				verifyKeys(t, got, tt.wantKeys)
			})
		}
	})
}

// TestBlockedZSL tests the BlockedZSL implementation, which uses a two-level
// structure with sorted blocks. It verifies the correctness of all operations
// including block management, search, insertion, deletion, and range queries.
func TestBlockedZSL(t *testing.T) {
	// TestNewBlockedZSL verifies that NewBlockedZSL correctly constructs a new
	// blocked structure from various input sequences, handling sorting and
	// block organization according to the specified block size.
	t.Run("NewBlockedZSL", func(t *testing.T) {
		tests := []struct {
			name     string
			keys     []int
			B        int
			wantSize int
		}{
			{"empty", []int{}, 2, 0},                     // Test empty input
			{"single", []int{1}, 2, 1},                   // Test single element
			{"exact block", []int{1, 2}, 2, 2},           // Test exact block size
			{"multiple blocks", []int{1, 2, 3, 4}, 2, 4}, // Test multiple blocks
			{"unsorted", []int{4, 2, 1, 3}, 2, 4},        // Test sorting
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				bz := NewBlockedZSL(tt.keys, tt.B)
				if got := bz.GetSize(); got != tt.wantSize {
					t.Errorf("GetSize() = %d, want %d", got, tt.wantSize)
				}
				verifySorted(t, bz.GetKeys())
			})
		}
	})

	// TestSearch verifies that Search correctly identifies the presence or absence
	// of keys in the blocked structure, including edge cases at the boundaries
	// and across block boundaries.
	t.Run("Search", func(t *testing.T) {
		keys := []int{1, 3, 5, 7, 9}
		bz := NewBlockedZSL(keys, 2)

		tests := []struct {
			key  int
			want bool
		}{
			{0, false},  // Test below range
			{1, true},   // Test first element
			{3, true},   // Test middle element
			{9, true},   // Test last element
			{10, false}, // Test above range
			{2, false},  // Test between elements
			{4, false},  // Test between elements
		}

		for _, tt := range tests {
			t.Run("", func(t *testing.T) {
				if got := bz.Search(tt.key); got != tt.want {
					t.Errorf("Search(%d) = %v, want %v", tt.key, got, tt.want)
				}
			})
		}
	})

	// TestInsert verifies that Insert correctly adds new keys while maintaining
	// sorted order and block size invariants, including block splitting when
	// necessary.
	t.Run("Insert", func(t *testing.T) {
		tests := []struct {
			name     string
			initial  []int
			B        int
			insert   int
			wantKeys []int
		}{
			{"empty", []int{}, 2, 1, []int{1}},                             // Test empty list
			{"beginning", []int{2, 3}, 2, 1, []int{1, 2, 3}},               // Test insert at start
			{"middle", []int{1, 3}, 2, 2, []int{1, 2, 3}},                  // Test insert in middle
			{"end", []int{1, 2}, 2, 3, []int{1, 2, 3}},                     // Test insert at end
			{"duplicate", []int{1, 2, 3}, 2, 2, []int{1, 2, 3}},            // Test duplicate
			{"split block", []int{1, 2, 3, 4}, 2, 5, []int{1, 2, 3, 4, 5}}, // Test block split
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				bz := NewBlockedZSL(tt.initial, tt.B)
				bz.Insert(tt.insert)
				verifyKeys(t, bz.GetKeys(), tt.wantKeys)
			})
		}
	})

	// TestDelete verifies that Delete correctly removes keys and maintains
	// block size invariants, including block merging when necessary.
	t.Run("Delete", func(t *testing.T) {
		tests := []struct {
			name     string
			initial  []int
			B        int
			delete   int
			wantKeys []int
			wantOk   bool
		}{
			{"empty", []int{}, 2, 1, []int{}, false},                        // Test empty list
			{"single", []int{1}, 2, 1, []int{}, true},                       // Test single element
			{"first", []int{1, 2, 3}, 2, 1, []int{2, 3}, true},              // Test delete first
			{"middle", []int{1, 2, 3}, 2, 2, []int{1, 3}, true},             // Test delete middle
			{"last", []int{1, 2, 3}, 2, 3, []int{1, 2}, true},               // Test delete last
			{"not found", []int{1, 2, 3}, 2, 4, []int{1, 2, 3}, false},      // Test non-existent
			{"merge blocks", []int{1, 2, 3, 4}, 2, 2, []int{1, 3, 4}, true}, // Test block merge
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				bz := NewBlockedZSL(tt.initial, tt.B)
				if got := bz.Delete(tt.delete); got != tt.wantOk {
					t.Errorf("Delete() = %v, want %v", got, tt.wantOk)
				}
				verifyKeys(t, bz.GetKeys(), tt.wantKeys)
			})
		}
	})

	// TestRange verifies that Range correctly returns all keys in the specified
	// range, handling various boundary conditions and cross-block ranges.
	t.Run("Range", func(t *testing.T) {
		keys := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
		bz := NewBlockedZSL(keys, 3)

		tests := []struct {
			name     string
			lo, hi   int
			wantKeys []int
		}{
			{"empty", 10, 20, []int{}},           // Test empty range
			{"full", 1, 10, keys},                // Test full range
			{"partial", 3, 7, []int{3, 4, 5, 6}}, // Test partial range
			{"single", 5, 6, []int{5}},           // Test single element
			{"nonexistent", 2, 2, []int{}},       // Test empty range
			{"below", 0, 1, []int{}},             // Test below range
			{"above", 9, 10, []int{9}},           // Test above range
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := bz.Range(tt.lo, tt.hi)
				verifyKeys(t, got, tt.wantKeys)
			})
		}
	})

	// TestRangeFunc verifies that RangeFunc correctly iterates over keys
	// in the specified range without allocations, handling various boundary
	// conditions and cross-block ranges.
	t.Run("RangeFunc", func(t *testing.T) {
		keys := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
		bz := NewBlockedZSL(keys, 3)

		tests := []struct {
			name     string
			lo, hi   int
			wantKeys []int
		}{
			{"empty", 10, 20, []int{}},           // Test empty range
			{"full", 1, 10, keys},                // Test full range
			{"partial", 3, 7, []int{3, 4, 5, 6}}, // Test partial range
			{"single", 5, 6, []int{5}},           // Test single element
			{"nonexistent", 2, 2, []int{}},       // Test empty range
			{"below", 0, 1, []int{}},             // Test below range
			{"above", 9, 10, []int{9}},           // Test above range
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var got []int
				bz.RangeFunc(tt.lo, tt.hi, func(k int) {
					got = append(got, k)
				})
				verifyKeys(t, got, tt.wantKeys)
			})
		}
	})

	// TestBlockOperations verifies the block management operations of BlockedZSL,
	// including splitting oversized blocks and merging undersized blocks.
	t.Run("Block Operations", func(t *testing.T) {
		// TestSplitBlock verifies that blocks are correctly split when they
		// exceed the maximum size threshold (2B).
		t.Run("split block", func(t *testing.T) {
			// Test that blocks split correctly when they exceed 2B
			keys := []int{1, 2, 3, 4, 5}
			bz := NewBlockedZSL(keys, 2)
			bz.Insert(6) // Should cause a block split
			verifyKeys(t, bz.GetKeys(), []int{1, 2, 3, 4, 5, 6})
		})

		// TestMergeBlocks verifies that blocks are correctly merged when they
		// fall below the minimum size threshold (B/2).
		t.Run("merge blocks", func(t *testing.T) {
			// Test that blocks merge when they become too small
			keys := []int{1, 2, 3, 4, 5}
			bz := NewBlockedZSL(keys, 3)
			bz.Delete(3) // Should cause blocks to merge
			verifyKeys(t, bz.GetKeys(), []int{1, 2, 4, 5})
		})
	})
}
