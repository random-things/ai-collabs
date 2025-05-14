package csi

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"
)

// TestNew ensures that we can construct and close the Rust-backed index.
func TestNewAndClose(t *testing.T) {
	idx := New("hello world")
	defer idx.Close()
	if idx == nil {
		t.Fatal("New returned nil CSI")
	}
}

// TestSearch verifies basic Search behavior against the Rust implementation.
func TestSearch(t *testing.T) {
	text := "hello world, hello there, hello universe"
	idx := New(text)
	defer idx.Close()

	tests := []struct {
		pattern  string
		expected []int
	}{
		// short patterns should simply return no matches (no error)
		{"hello", nil},
		{"xyz", nil},
		{"", nil}, // empty pattern returns error below
		{"universe", nil},
		// longer pattern that does match once
		{"ello world, hello there", []int{1}},
	}

	for _, tt := range tests {
		offs, err := idx.Search(tt.pattern)
		if tt.pattern == "" {
			if err == nil {
				t.Errorf("Search(%q): expected error for empty pattern", tt.pattern)
			}
			continue
		}
		if err != nil {
			t.Errorf("Search(%q) returned unexpected error: %v", tt.pattern, err)
			continue
		}
		if len(offs) != len(tt.expected) {
			t.Errorf("Search(%q) = %v, want %v", tt.pattern, offs, tt.expected)
			continue
		}
		for i := range offs {
			if offs[i] != tt.expected[i] {
				t.Errorf("Search(%q)[%d] = %d, want %d",
					tt.pattern, i, offs[i], tt.expected[i])
			}
		}
	}
}

// TestSearchNoIndexReuse ensures that multiple successive searches work.
func TestMultipleSearches(t *testing.T) {
	idx := New("abc abc abc")
	defer idx.Close()

	for _, pat := range []string{"a", "ab", "abc", "bc a", "c a"} {
		_, err := idx.Search(pat)
		// only empty pattern errors
		if pat == "" {
			if err == nil {
				t.Errorf("Search(%q): expected error for empty pattern", pat)
			}
		} else if err != nil {
			t.Errorf("Search(%q) unexpected error: %v", pat, err)
		}
	}
}

// BenchmarkBuild measures the Rust index build time
func BenchmarkBuild(b *testing.B) {
	text := strings.Repeat("abcdefghijklmnopqrstuvwxyz", 4000) // ~100kB
	for i := 0; i < b.N; i++ {
		idx := New(text)
		idx.Close()
	}
}

// BenchmarkSearch measures the Rust-backed Search performance
func BenchmarkSearch(b *testing.B) {
	text := strings.Repeat("the quick brown fox jumps over the lazy dog ", 1000)
	idx := New(text)
	defer idx.Close()

	pattern := "brown fox jumps"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := idx.Search(pattern)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBuildPowersOfTwo(b *testing.B) {
	// Start with 2^10 (1KB) and go up to 2^30 (1GB)
	// We'll stop if any single build takes more than 10 seconds
	sizes := []int{
		1 << 10, // 1KB
		1 << 12, // 4KB
		1 << 14, // 16KB
		1 << 16, // 64KB
		1 << 18, // 256KB
		1 << 20, // 1MB
		1 << 22, // 4MB
		1 << 24, // 16MB
		1 << 26, // 64MB
		1 << 28, // 256MB
		1 << 30, // 1GB
	}

	// Create a timer that will stop the benchmark after 10 seconds
	timeout := time.After(10 * time.Second)

	for _, size := range sizes {
		// Check if we've exceeded the timeout
		select {
		case <-timeout:
			b.Logf("Stopping benchmarks after 10 seconds")
			return
		default:
		}

		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			// Generate random text of the specified size
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			text := make([]byte, size)
			for i := range text {
				text[i] = byte(32 + r.Intn(95)) // Use printable ASCII chars
			}
			s := string(text)

			// Reset timer and memory stats before the actual benchmark
			b.ResetTimer()
			b.ReportAllocs()

			// Run the benchmark
			for i := 0; i < b.N; i++ {
				_ = New(s)
			}

			// Report memory usage per operation
			b.ReportMetric(float64(size), "bytes/op")
		})

		// Check timeout again after each size
		select {
		case <-timeout:
			b.Logf("Stopping benchmarks after 10 seconds")
			return
		default:
		}
	}
}
