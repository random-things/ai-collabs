package phitree

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"testing"

	"github.com/emirpasic/gods/trees/redblacktree"
	"github.com/google/btree"
)

// BenchmarkItem implements btree.Item for Google's BTree
type BenchmarkItem struct {
	key   int
	value int
}

func (a BenchmarkItem) Less(b btree.Item) bool {
	return a.key < b.(BenchmarkItem).key
}

// generateKeys creates a slice of n random keys in the range [0, max)
func generateKeys(n, max int) []int {
	keys := make([]int, n)
	for i := range keys {
		keys[i] = rand.Intn(max)
	}
	return keys
}

// benchmarkInsert measures insertion performance
func benchmarkInsert(b *testing.B, size int) {
	// Generate keys once for consistent benchmarking
	keys := generateKeys(size, size*10)
	values := make([]int, size)
	for i := range values {
		values[i] = i
	}

	b.Run("PhiTree", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			tree := New[int]()
			for j := 0; j < size; j++ {
				tree.Insert(keys[j], values[j])
			}
		}
	})

	b.Run("RedBlackTree", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			tree := redblacktree.NewWithIntComparator()
			for j := 0; j < size; j++ {
				tree.Put(keys[j], values[j])
			}
		}
	})

	b.Run("BTree", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			tree := btree.New(32) // 32 is a common degree for B-trees
			for j := 0; j < size; j++ {
				tree.ReplaceOrInsert(BenchmarkItem{keys[j], values[j]})
			}
		}
	})
}

// benchmarkGet measures lookup performance
func benchmarkGet(b *testing.B, size int) {
	// Generate keys and create populated trees
	keys := generateKeys(size, size*10)
	values := make([]int, size)
	for i := range values {
		values[i] = i
	}

	// Create and populate trees once
	phiTree := New[int]()
	rbTree := redblacktree.NewWithIntComparator()
	btree := btree.New(32)
	for i := 0; i < size; i++ {
		phiTree.Insert(keys[i], values[i])
		rbTree.Put(keys[i], values[i])
		btree.ReplaceOrInsert(BenchmarkItem{keys[i], values[i]})
	}

	// Generate random keys for lookups
	lookupKeys := generateKeys(b.N, size*10)

	b.Run("PhiTree", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			phiTree.Get(lookupKeys[i%len(lookupKeys)])
		}
	})

	b.Run("RedBlackTree", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			rbTree.Get(lookupKeys[i%len(lookupKeys)])
		}
	})

	b.Run("BTree", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			btree.Get(BenchmarkItem{lookupKeys[i%len(lookupKeys)], 0})
		}
	})
}

// benchmarkConcurrent measures concurrent operation performance
func benchmarkConcurrent(b *testing.B, size int) {
	// Generate keys once
	keys := generateKeys(size, size*10)
	values := make([]int, size)
	for i := range values {
		values[i] = i
	}

	b.Run("PhiTree", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			tree := New[int]()
			done := make(chan bool)

			// Start 4 concurrent writers
			for w := 0; w < 4; w++ {
				go func(wid int) {
					for j := 0; j < size/4; j++ {
						idx := wid*(size/4) + j
						tree.Insert(keys[idx], values[idx])
					}
					done <- true
				}(w)
			}

			// Start 4 concurrent readers
			for r := 0; r < 4; r++ {
				go func() {
					for j := 0; j < size/4; j++ {
						tree.Get(keys[j])
					}
					done <- true
				}()
			}

			// Wait for all goroutines
			for j := 0; j < 8; j++ {
				<-done
			}
		}
	})

	b.Run("RedBlackTree", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			tree := redblacktree.NewWithIntComparator()
			var mu sync.Mutex
			done := make(chan bool)

			// Start 4 concurrent writers
			for w := 0; w < 4; w++ {
				go func(wid int) {
					for j := 0; j < size/4; j++ {
						idx := wid*(size/4) + j
						mu.Lock()
						tree.Put(keys[idx], values[idx])
						mu.Unlock()
					}
					done <- true
				}(w)
			}

			// Start 4 concurrent readers
			for r := 0; r < 4; r++ {
				go func() {
					for j := 0; j < size/4; j++ {
						mu.Lock()
						tree.Get(keys[j])
						mu.Unlock()
					}
					done <- true
				}()
			}

			// Wait for all goroutines
			for j := 0; j < 8; j++ {
				<-done
			}
		}
	})

	b.Run("BTree", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			tree := btree.New(32)
			var mu sync.Mutex
			done := make(chan bool)

			// Start 4 concurrent writers
			for w := 0; w < 4; w++ {
				go func(wid int) {
					for j := 0; j < size/4; j++ {
						idx := wid*(size/4) + j
						mu.Lock()
						tree.ReplaceOrInsert(BenchmarkItem{keys[idx], values[idx]})
						mu.Unlock()
					}
					done <- true
				}(w)
			}

			// Start 4 concurrent readers
			for r := 0; r < 4; r++ {
				go func() {
					for j := 0; j < size/4; j++ {
						mu.Lock()
						tree.Get(BenchmarkItem{keys[j], 0})
						mu.Unlock()
					}
					done <- true
				}()
			}

			// Wait for all goroutines
			for j := 0; j < 8; j++ {
				<-done
			}
		}
	})
}

// Run benchmarks for different tree sizes
func BenchmarkInsert100(b *testing.B)    { benchmarkInsert(b, 100) }
func BenchmarkInsert1000(b *testing.B)   { benchmarkInsert(b, 1000) }
func BenchmarkInsert10000(b *testing.B)  { benchmarkInsert(b, 10000) }
func BenchmarkInsert100000(b *testing.B) { benchmarkInsert(b, 100000) }

func BenchmarkGet100(b *testing.B)    { benchmarkGet(b, 100) }
func BenchmarkGet1000(b *testing.B)   { benchmarkGet(b, 1000) }
func BenchmarkGet10000(b *testing.B)  { benchmarkGet(b, 10000) }
func BenchmarkGet100000(b *testing.B) { benchmarkGet(b, 100000) }

func BenchmarkConcurrent100(b *testing.B)    { benchmarkConcurrent(b, 100) }
func BenchmarkConcurrent1000(b *testing.B)   { benchmarkConcurrent(b, 1000) }
func BenchmarkConcurrent10000(b *testing.B)  { benchmarkConcurrent(b, 10000) }
func BenchmarkConcurrent100000(b *testing.B) { benchmarkConcurrent(b, 100000) }

// BenchmarkMemoryUsage measures memory usage for different tree sizes
func BenchmarkMemoryUsage(b *testing.B) {
	sizes := []int{100, 1000, 10000, 100000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("PhiTree_%d", size), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				tree := New[int]()
				for j := 0; j < size; j++ {
					tree.Insert(j, j)
				}
				b.StopTimer()
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				b.ReportMetric(float64(m.Alloc), "bytes/op")
				b.StartTimer()
			}
		})

		b.Run(fmt.Sprintf("RedBlackTree_%d", size), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				tree := redblacktree.NewWithIntComparator()
				for j := 0; j < size; j++ {
					tree.Put(j, j)
				}
				b.StopTimer()
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				b.ReportMetric(float64(m.Alloc), "bytes/op")
				b.StartTimer()
			}
		})

		b.Run(fmt.Sprintf("BTree_%d", size), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				tree := btree.New(32)
				for j := 0; j < size; j++ {
					tree.ReplaceOrInsert(BenchmarkItem{j, j})
				}
				b.StopTimer()
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				b.ReportMetric(float64(m.Alloc), "bytes/op")
				b.StartTimer()
			}
		})
	}
}
