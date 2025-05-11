// Package zsl_test contains benchmarks for the zsl package.
// These benchmarks compare the performance of BlockedZSL and StaticZSL implementations
// against each other and against a standard B-tree implementation.
// The benchmarks measure various operations (insert, delete, search, range) across
// different data sizes and access patterns.
package zsl

import (
	"fmt"
	"math/rand"
	"runtime"
	"testing"
	"time"

	"github.com/google/btree"
)

// IntWrapper implements the btree.Item interface for integers.
// It allows us to use integers in the B-tree implementation.
type IntWrapper int

// Less implements the btree.Item interface.
func (a IntWrapper) Less(b btree.Item) bool {
	return a < b.(IntWrapper)
}

func init() { rand.New(rand.NewSource(time.Now().UnixNano())) }

// --- Benchmark Helpers --------------------------------------------------------

// benchStaticZSLInsert benchmarks insertion operations for StaticZSL.
// It pre-populates the structure with n keys and then measures the time
// to insert keys in a modulo pattern (i % n).
//
// Parameters:
//   - b: The benchmark context
//   - n: The number of keys to pre-populate
func benchStaticZSLInsert(b *testing.B, n int) {
	keys := make([]int, n)
	for i := range keys {
		keys[i] = i
	}
	z := NewStaticZSL(keys)
	b.ResetTimer()
	for i := range b.N {
		z.Insert(i % n)
	}
}

// benchStaticZSLDelete benchmarks deletion operations for StaticZSL.
// It pre-populates the structure with 2n keys and then measures the time
// to delete keys in a modulo pattern (i % n).
//
// Parameters:
//   - b: The benchmark context
//   - n: The number of keys to delete
func benchStaticZSLDelete(b *testing.B, n int) {
	keys := make([]int, 2*n)
	for i := range keys {
		keys[i] = i
	}
	z := NewStaticZSL(keys)
	b.ResetTimer()
	for i := range b.N {
		z.Delete(i % n)
	}
}

// benchStaticZSLSearch benchmarks search operations for StaticZSL.
// It pre-populates the structure with n keys and then measures the time
// to search for random keys within the range [0, n).
//
// Parameters:
//   - b: The benchmark context
//   - n: The upper bound for random key generation
func benchStaticZSLSearch(b *testing.B, n int) {
	keys := make([]int, n)
	for i := range keys {
		keys[i] = i
	}
	z := NewStaticZSL(keys)
	b.ResetTimer()
	for range b.N {
		z.Search(rand.Intn(n))
	}
}

// benchBlockedZSLInsert benchmarks insertion operations for BlockedZSL.
// It pre-populates the structure with n keys and then measures the time
// to insert keys in a modulo pattern (i % n).
//
// Parameters:
//   - b: The benchmark context
//   - n: The number of keys to pre-populate
//   - B: The block size for the BlockedZSL
func benchBlockedZSLInsert(b *testing.B, n int, B int) {
	keys := make([]int, n)
	for i := range keys {
		keys[i] = i
	}
	bz := NewBlockedZSL(keys, B)
	b.ResetTimer()
	for i := range b.N {
		bz.Insert(i % n)
	}
}

// benchBlockedZSLDelete benchmarks deletion operations for BlockedZSL.
// It pre-populates the structure with 2n keys and then measures the time
// to delete keys in a modulo pattern (i % n).
//
// Parameters:
//   - b: The benchmark context
//   - n: The number of keys to delete
//   - B: The block size for the BlockedZSL
func benchBlockedZSLDelete(b *testing.B, n int, B int) {
	keys := make([]int, 2*n)
	for i := range keys {
		keys[i] = i
	}
	bz := NewBlockedZSL(keys, B)
	b.ResetTimer()
	for i := range b.N {
		bz.Delete(i % n)
	}
}

// benchBlockedZSLSearch benchmarks search operations for BlockedZSL.
// It pre-populates the structure with n keys and then measures the time
// to search for random keys within the range [0, n).
//
// Parameters:
//   - b: The benchmark context
//   - n: The upper bound for random key generation
//   - B: The block size for the BlockedZSL
func benchBlockedZSLSearch(b *testing.B, n int, B int) {
	keys := make([]int, n)
	for i := range keys {
		keys[i] = i
	}
	bz := NewBlockedZSL(keys, B)
	b.ResetTimer()
	for range b.N {
		bz.Search(rand.Intn(n))
	}
}

// benchBlockedZSLRangeFunc benchmarks range iteration for BlockedZSL.
// It pre-populates the structure with n keys and then measures the time
// to iterate over random ranges, counting the number of elements.
//
// Parameters:
//   - b: The benchmark context
//   - n: The number of keys to pre-populate
//   - B: The block size for the BlockedZSL
func benchBlockedZSLRangeFunc(b *testing.B, n, B int) {
	keys := make([]int, n)
	for i := range keys {
		keys[i] = i
	}
	bz := NewBlockedZSL(keys, B)
	b.ResetTimer()
	for range b.N {
		lo := rand.Intn(n)
		hi := rand.Intn(n)
		if lo > hi {
			lo, hi = hi, lo
		}
		var count int
		bz.RangeFunc(lo, hi, func(_ int) {
			count++
		})
	}
}

// benchBTreeInsert benchmarks insertion operations for B-tree.
// It pre-populates the tree with n keys and then measures the time
// to insert keys in a modulo pattern (i % n).
//
// Parameters:
//   - b: The benchmark context
//   - n: The number of keys to pre-populate
func benchBTreeInsert(b *testing.B, n int) {
	tree := btree.New(32)
	for i := range n {
		tree.ReplaceOrInsert(IntWrapper(i))
	}
	b.ResetTimer()
	for i := range b.N {
		tree.ReplaceOrInsert(IntWrapper(i % n))
	}
}

// benchBTreeDelete benchmarks deletion operations for B-tree.
// It pre-populates the tree with 2n keys and then measures the time
// to delete keys in a modulo pattern (i % n).
//
// Parameters:
//   - b: The benchmark context
//   - n: The number of keys to delete
func benchBTreeDelete(b *testing.B, n int) {
	tree := btree.New(32)
	for i := range 2 * n {
		tree.ReplaceOrInsert(IntWrapper(i))
	}
	b.ResetTimer()
	for i := range b.N {
		tree.Delete(IntWrapper(i % n))
	}
}

// benchBTreeSearch benchmarks search operations for B-tree.
// It pre-populates the tree with n keys and then measures the time
// to search for random keys within the range [0, n).
//
// Parameters:
//   - b: The benchmark context
//   - n: The upper bound for random key generation
func benchBTreeSearch(b *testing.B, n int) {
	tree := btree.New(32)
	for i := range n {
		tree.ReplaceOrInsert(IntWrapper(i))
	}
	b.ResetTimer()
	for range b.N {
		tree.Get(IntWrapper(rand.Intn(n)))
	}
}

// --- Non-random Access Patterns ----------------------------------------------

// benchSequentialInsert benchmarks sequential insertion patterns.
// It measures the time to insert keys in ascending order.
//
// Parameters:
//   - b: The benchmark context
//   - ds: A data structure that implements Insert(int)
func benchSequentialInsert(b *testing.B, ds interface{ Insert(int) }) {
	b.ResetTimer()
	for i := range b.N {
		ds.Insert(i)
	}
}

// benchReverseInsert benchmarks reverse sequential insertion patterns.
// It measures the time to insert keys in descending order.
//
// Parameters:
//   - b: The benchmark context
//   - ds: A data structure that implements Insert(int)
//   - n: The number of keys to insert
func benchReverseInsert(b *testing.B, ds interface{ Insert(int) }, n int) {
	b.ResetTimer()
	for i := range b.N {
		ds.Insert(n - 1 - (i % n))
	}
}

// benchSlashInsertDelete benchmarks a pattern of alternating insert and delete
// operations, where each delete operation targets a key k positions ahead
// of the inserted key.
//
// Parameters:
//   - b: The benchmark context
//   - ds: A data structure that implements Insert(int) and Delete(int) bool
//   - n: The modulo for key generation
//   - k: The offset for delete operations
func benchSlashInsertDelete(b *testing.B, ds interface {
	Insert(int)
	Delete(int) bool
}, n, k int) {
	b.ResetTimer()
	for i := range b.N {
		x := i % n
		ds.Insert(x)
		ds.Delete((x + k) % n)
	}
}

// benchZipfSearch benchmarks search operations using a Zipf distribution,
// which models real-world access patterns where some keys are accessed
// more frequently than others.
//
// Parameters:
//   - b: The benchmark context
//   - ds: A data structure that implements Search(int) bool
//   - n: The upper bound for key generation
func benchZipfSearch(b *testing.B, ds interface{ Search(int) bool }, n int) {
	zipf := rand.NewZipf(rand.New(rand.NewSource(42)), 1.2, 1, uint64(n-1))
	b.ResetTimer()
	for range b.N {
		ds.Search(int(zipf.Uint64()))
	}
}

// RangeIterable defines the interface for range iteration operations.
// It is used to benchmark range operations across different implementations.
type RangeIterable interface {
	RangeFunc(lo, hi int, fn func(int))
}

// BTreeAdapter adapts btree.Tree to our benchmark interfaces.
// It allows us to benchmark the B-tree implementation using the same
// interface as our custom implementations.
type BTreeAdapter struct {
	tree *btree.BTree
}

// Insert adds a key to the B-tree.
func (a *BTreeAdapter) Insert(x int) {
	a.tree.ReplaceOrInsert(IntWrapper(x))
}

// Delete removes a key from the B-tree.
func (a *BTreeAdapter) Delete(x int) bool {
	return a.tree.Delete(IntWrapper(x)) != nil
}

// Search checks if a key exists in the B-tree.
func (a *BTreeAdapter) Search(x int) bool {
	_, ok := a.tree.Get(IntWrapper(x)).(IntWrapper)
	return ok
}

// RangeFunc iterates over keys in the range [lo, hi).
func (a *BTreeAdapter) RangeFunc(lo, hi int, fn func(int)) {
	a.tree.AscendRange(IntWrapper(lo), IntWrapper(hi), func(item btree.Item) bool {
		fn(int(item.(IntWrapper)))
		return true
	})
}

// benchSlidingWindowRange benchmarks range iteration with a sliding window.
// It measures the time to iterate over ranges that slide through the key space.
//
// Parameters:
//   - b: The benchmark context
//   - ds: A data structure that implements RangeIterable
//   - n: The total number of keys
//   - w: The window size for range iteration
func benchSlidingWindowRange(b *testing.B, ds RangeIterable, n, w int) {
	b.ResetTimer()
	for i := range b.N {
		lo := (i * w / 2) % (n - w)
		hi := lo + w
		var count int
		ds.RangeFunc(lo, hi, func(_ int) {
			count++
		})
	}
}

// --- Benchmark Parameters ----------------------------------------------------
var (
	// sizes defines the data sizes to benchmark against.
	// These represent the number of keys in the data structure.
	sizes = []int{1_000, 100_000, 1_000_000, 10_000_000}

	// blockSizes defines the block sizes to test for BlockedZSL.
	// These represent different trade-offs between search and update costs.
	blockSizes = []int{64, 128, 256, 512, 1024, 2048}
)

// --- Grouped Benchmarks ------------------------------------------------------

// BenchmarkInsert benchmarks insertion operations across different
// implementations and data sizes. It measures both time and memory
// allocation for each operation.
func BenchmarkInsert(b *testing.B) {
	for _, n := range sizes {
		b.Run(fmt.Sprintf("Insert_N%d", n), func(b *testing.B) {
			// StaticZSL
			b.Run("StaticZSL", func(b *testing.B) {
				runtime.GC()
				var m0, m1 runtime.MemStats
				runtime.ReadMemStats(&m0)
				benchStaticZSLInsert(b, n)
				b.StopTimer()
				runtime.GC()
				runtime.ReadMemStats(&m1)
				b.ReportMetric(float64(m1.Alloc-m0.Alloc)/float64(b.N), "heapBytes/op")
				b.ReportMetric(float64(m1.HeapObjects-m0.HeapObjects)/float64(b.N), "heapObjects/op")
			})

			// BlockedZSL for each B
			for _, B := range blockSizes {
				name := fmt.Sprintf("BlockedZSL_B%d", B)
				b.Run(name, func(b *testing.B) {
					runtime.GC()
					var m0, m1 runtime.MemStats
					runtime.ReadMemStats(&m0)
					benchBlockedZSLInsert(b, n, B)
					b.StopTimer()
					runtime.GC()
					runtime.ReadMemStats(&m1)
					b.ReportMetric(float64(m1.Alloc-m0.Alloc)/float64(b.N), "heapBytes/op")
					b.ReportMetric(float64(m1.HeapObjects-m0.HeapObjects)/float64(b.N), "heapObjects/op")
				})
			}

			// B-Tree
			b.Run("BTree", func(b *testing.B) {
				runtime.GC()
				var m0, m1 runtime.MemStats
				runtime.ReadMemStats(&m0)
				benchBTreeInsert(b, n)
				b.StopTimer()
				runtime.GC()
				runtime.ReadMemStats(&m1)
				b.ReportMetric(float64(m1.Alloc-m0.Alloc)/float64(b.N), "heapBytes/op")
				b.ReportMetric(float64(m1.HeapObjects-m0.HeapObjects)/float64(b.N), "heapObjects/op")
			})
		})
	}
}

// BenchmarkDelete benchmarks deletion operations across different
// implementations and data sizes. It measures both time and memory
// allocation for each operation.
func BenchmarkDelete(b *testing.B) {
	for _, n := range sizes {
		b.Run(fmt.Sprintf("Delete_N%d", n), func(b *testing.B) {
			// StaticZSL
			b.Run("StaticZSL", func(b *testing.B) {
				runtime.GC()
				var m0, m1 runtime.MemStats
				runtime.ReadMemStats(&m0)
				benchStaticZSLDelete(b, n)
				b.StopTimer()
				runtime.GC()
				runtime.ReadMemStats(&m1)
				b.ReportMetric(float64(m1.Alloc-m0.Alloc)/float64(b.N), "heapBytes/op")
				b.ReportMetric(float64(m1.HeapObjects-m0.HeapObjects)/float64(b.N), "heapObjects/op")
			})

			// BlockedZSL
			for _, B := range blockSizes {
				b.Run(fmt.Sprintf("BlockedZSL_B%d", B), func(b *testing.B) {
					runtime.GC()
					var m0, m1 runtime.MemStats
					runtime.ReadMemStats(&m0)
					benchBlockedZSLDelete(b, n, B)
					b.StopTimer()
					runtime.GC()
					runtime.ReadMemStats(&m1)
					b.ReportMetric(float64(m1.Alloc-m0.Alloc)/float64(b.N), "heapBytes/op")
					b.ReportMetric(float64(m1.HeapObjects-m0.HeapObjects)/float64(b.N), "heapObjects/op")
				})
			}

			// B-Tree
			b.Run("BTree", func(b *testing.B) {
				runtime.GC()
				var m0, m1 runtime.MemStats
				runtime.ReadMemStats(&m0)
				benchBTreeDelete(b, n)
				b.StopTimer()
				runtime.GC()
				runtime.ReadMemStats(&m1)
				b.ReportMetric(float64(m1.Alloc-m0.Alloc)/float64(b.N), "heapBytes/op")
				b.ReportMetric(float64(m1.HeapObjects-m0.HeapObjects)/float64(b.N), "heapObjects/op")
			})
		})
	}
}

// BenchmarkSearch benchmarks search operations across different
// implementations and data sizes. It measures both time and memory
// allocation for each operation.
func BenchmarkSearch(b *testing.B) {
	for _, n := range sizes {
		b.Run(fmt.Sprintf("Search_N%d", n), func(b *testing.B) {
			// StaticZSL
			b.Run("StaticZSL", func(b *testing.B) {
				runtime.GC()
				var m0, m1 runtime.MemStats
				runtime.ReadMemStats(&m0)
				benchStaticZSLSearch(b, n)
				b.StopTimer()
				runtime.GC()
				runtime.ReadMemStats(&m1)
				b.ReportMetric(float64(m1.Alloc-m0.Alloc)/float64(b.N), "heapBytes/op")
				b.ReportMetric(float64(m1.HeapObjects-m0.HeapObjects)/float64(b.N), "heapObjects/op")
			})

			// BlockedZSL
			for _, B := range blockSizes {
				b.Run(fmt.Sprintf("BlockedZSL_B%d", B), func(b *testing.B) {
					runtime.GC()
					var m0, m1 runtime.MemStats
					runtime.ReadMemStats(&m0)
					benchBlockedZSLSearch(b, n, B)
					b.StopTimer()
					runtime.GC()
					runtime.ReadMemStats(&m1)
					b.ReportMetric(float64(m1.Alloc-m0.Alloc)/float64(b.N), "heapBytes/op")
					b.ReportMetric(float64(m1.HeapObjects-m0.HeapObjects)/float64(b.N), "heapObjects/op")
				})
			}

			// B-Tree
			b.Run("BTree", func(b *testing.B) {
				runtime.GC()
				var m0, m1 runtime.MemStats
				runtime.ReadMemStats(&m0)
				benchBTreeSearch(b, n)
				b.StopTimer()
				runtime.GC()
				runtime.ReadMemStats(&m1)
				b.ReportMetric(float64(m1.Alloc-m0.Alloc)/float64(b.N), "heapBytes/op")
				b.ReportMetric(float64(m1.HeapObjects-m0.HeapObjects)/float64(b.N), "heapObjects/op")
			})
		})
	}
}

// BenchmarkRange benchmarks range iteration operations across different
// implementations and data sizes. It measures the time to iterate over
// random ranges within the key space.
func BenchmarkRange(b *testing.B) {
	for _, n := range sizes {
		b.Run(fmt.Sprintf("Range_N%d", n), func(b *testing.B) {
			// BlockedZSL Range
			for _, B := range blockSizes {
				b.Run(fmt.Sprintf("BlockedZSL_B%d", B), func(b *testing.B) {
					benchBlockedZSLRangeFunc(b, n, B)
				})
			}

			// B-Tree Range via AscendRange
			b.Run("BTree", func(b *testing.B) {
				// prepare
				tree := btree.New(32)
				for i := range n {
					tree.ReplaceOrInsert(IntWrapper(i))
				}
				b.ResetTimer()
				for range b.N {
					lo := IntWrapper(rand.Intn(n))
					hi := IntWrapper(rand.Intn(n))
					if lo > hi {
						lo, hi = hi, lo
					}
					var count int
					tree.AscendRange(lo, hi, func(item btree.Item) bool {
						count++
						return true
					})
				}
			})
		})
	}
}

// BenchmarkPatterns benchmarks various non-random access patterns
// across different implementations. These patterns include:
// - Sequential insertion
// - Reverse sequential insertion
// - Alternating insert/delete operations
// - Zipf-distributed searches
// - Sliding window range iteration
//
// All patterns are benchmarked with n=100,000 keys to provide
// a realistic workload size.
func BenchmarkPatterns(b *testing.B) {
	n := 100_000
	window := n / 10

	b.Run("SeqInsert_StaticZSL", func(b *testing.B) {
		z := NewStaticZSL(make([]int, n))
		benchSequentialInsert(b, z)
	})
	b.Run("SeqInsert_BlockedZSL", func(b *testing.B) {
		bz := NewBlockedZSL(make([]int, n), 256)
		benchSequentialInsert(b, bz)
	})
	b.Run("SeqInsert_BTree", func(b *testing.B) {
		t := btree.New(32)
		benchSequentialInsert(b, &BTreeAdapter{tree: t})
	})

	b.Run("ReverseInsert_StaticZSL", func(b *testing.B) {
		z := NewStaticZSL(make([]int, n))
		benchReverseInsert(b, z, n)
	})
	b.Run("ReverseInsert_BlockedZSL", func(b *testing.B) {
		bz := NewBlockedZSL(make([]int, n), 256)
		benchReverseInsert(b, bz, n)
	})
	b.Run("ReverseInsert_BTree", func(b *testing.B) {
		t := btree.New(32)
		benchReverseInsert(b, &BTreeAdapter{tree: t}, n)
	})
	b.Run("Slash_InsertDelete_StaticZSL", func(b *testing.B) {
		z := NewStaticZSL(make([]int, n))
		benchSlashInsertDelete(b, z, n, window)
	})
	b.Run("Slash_InsertDelete_BlockedZSL", func(b *testing.B) {
		bz := NewBlockedZSL(make([]int, n), 256)
		benchSlashInsertDelete(b, bz, n, window)
	})
	b.Run("Slash_InsertDelete_BTree", func(b *testing.B) {
		t := btree.New(32)
		benchSlashInsertDelete(b, &BTreeAdapter{tree: t}, n, window)
	})
	b.Run("ZipfSearch_StaticZSL", func(b *testing.B) {
		z := NewStaticZSL(make([]int, n))
		benchZipfSearch(b, z, n)
	})
	b.Run("ZipfSearch_BlockedZSL", func(b *testing.B) {
		bz := NewBlockedZSL(make([]int, n), 256)
		benchZipfSearch(b, bz, n)
	})
	b.Run("ZipfSearch_BTree", func(b *testing.B) {
		t := btree.New(32)
		for i := 0; i < n; i++ {
			t.ReplaceOrInsert(IntWrapper(i))
		}
		benchZipfSearch(b, &BTreeAdapter{tree: t}, n)
	})
	b.Run("SlidingRange_BlockedZSL", func(b *testing.B) {
		bz := NewBlockedZSL(make([]int, n), 256)
		for i := 0; i < n; i++ {
			bz.Insert(i)
		}
		benchSlidingWindowRange(b, bz, n, window)
	})
	b.Run("SlidingRange_BTree", func(b *testing.B) {
		tree := btree.New(32)
		for i := 0; i < n; i++ {
			tree.ReplaceOrInsert(IntWrapper(i))
		}
		benchSlidingWindowRange(b, &BTreeAdapter{tree: tree}, n, window)
	})
}
