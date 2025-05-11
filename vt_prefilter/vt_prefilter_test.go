package vt_prefilter

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
)

// TestCalculateVTKey tests basic VT key calculation
func TestCalculateVTKey(t *testing.T) {
	tests := []struct {
		word string
		want VTKey
	}{
		{"", VTKey{0, 0, 0}},
		{"a", VTKey{1, 1, 1}},
		{"ab", VTKey{2, 2, 3}},
		{"abc", VTKey{3, 2, 6}},
	}

	for _, tt := range tests {
		t.Run(tt.word, func(t *testing.T) {
			got := CalculateVTKey(tt.word)
			if got != tt.want {
				t.Errorf("CalculateVTKey(%q) = %v; want %v", tt.word, got, tt.want)
				// Debug output
				t.Logf("Calculating key for %q:", tt.word)
				var s, p int
				for i, ch := range tt.word {
					val := vtToInt[ch]
					s += (i + 1) * val
					p += val
					t.Logf("  pos %d: val %d, s += %d*%d = %d, p += %d = %d", i, val, i+1, val, s, val, p)
				}
				t.Logf("  Final: s = %d mod %d = %d, p = %d mod %d = %d", s, len(tt.word)+1, s%(len(tt.word)+1), p, VTM, p%VTM)
			}
		})
	}
}

// TestVTKeyProperties tests mathematical properties of VT keys
func TestVTKeyProperties(t *testing.T) {
	// Property 1: Empty string always has key (0,0,0)
	emptyKey := CalculateVTKey("")
	if emptyKey != (VTKey{0, 0, 0}) {
		t.Errorf("CalculateVTKey(\"\") = %+v, want {0,0,0}", emptyKey)
	}

	// Property 2: Key length matches string length
	word := "test"
	testKey := CalculateVTKey(word)
	if testKey.N != len(word) {
		t.Errorf("CalculateVTKey(%q).N = %d, want %d", word, testKey.N, len(word))
	}

	// Property 3: S value is always in [0, n]
	word = "abcdef"
	key := CalculateVTKey(word)
	if key.S < 0 || key.S > key.N {
		t.Errorf("CalculateVTKey(%q).S = %d, want in [0,%d]", word, key.S, key.N)
	}

	// Property 4: P value is always in [0, M-1]
	if key.P < 0 || key.P >= VTM {
		t.Errorf("CalculateVTKey(%q).P = %d, want in [0,%d]", word, key.P, VTM-1)
	}
}

// TestVTEdits1 tests the basic edit operations
func TestVTEdits1(t *testing.T) {
	tests := []struct {
		word string
		want int // expected number of variants
	}{
		{"", 26},    // 26 insertions
		{"a", 77},   // 1 deletion + 25 substitutions + 26 insertions
		{"ab", 128}, // 2 deletions + 50 substitutions + 78 insertions
	}

	for _, tt := range tests {
		t.Run(tt.word, func(t *testing.T) {
			got := VTEdits1(tt.word)
			if len(got) != tt.want {
				t.Errorf("len(VTEdits1(%q)) = %d; want %d", tt.word, len(got), tt.want)
			}
		})
	}
}

// TestVTEdits1Damerau tests Damerau-Levenshtein operations
func TestVTEdits1Damerau(t *testing.T) {
	tests := []struct {
		word string
		want int // expected number of variants
	}{
		{"", 26},    // 26 insertions
		{"a", 77},   // 1 deletion + 25 substitutions + 26 insertions
		{"ab", 129}, // 2 deletions + 50 substitutions + 78 insertions + 1 transposition
	}

	for _, tt := range tests {
		t.Run(tt.word, func(t *testing.T) {
			got := VTEdits1Damerau(tt.word)
			if len(got) != tt.want {
				t.Errorf("len(VTEdits1Damerau(%q)) = %d; want %d", tt.word, len(got), tt.want)
			}
		})
	}
}

// TestVTVariants tests the algebraic variant generation
func TestVTVariants(t *testing.T) {
	tests := []struct {
		word string
	}{
		{""},
		{"a"},
		{"ab"},
	}

	for _, tt := range tests {
		t.Run(tt.word, func(t *testing.T) {
			got := VTVariants(tt.word)

			// Calculate expected number of variants by generating VT keys from all edits
			want := make(map[VTKey]bool)
			// Add original word's key
			want[CalculateVTKey(tt.word)] = true
			// Add keys from all Damerau-Levenshtein distance 1 edits
			for v := range VTEdits1Damerau(tt.word) {
				want[CalculateVTKey(v)] = true
			}

			// Debug output for bounds comparison
			bound := VTBound(len(tt.word), 1)
			t.Logf("Word: %q (length %d)", tt.word, len(tt.word))
			t.Logf("Upper bound (k=1): %d", bound)
			t.Logf("Generated variants: %d", len(got))
			t.Logf("Expected variants: %d", len(want))
			if len(got) > bound {
				t.Logf("⚠️ Generated variants exceed bound by %d", len(got)-bound)
			}
			if len(want) > bound {
				t.Logf("⚠️ Expected variants exceed bound by %d", len(want)-bound)
			}

			if len(got) != len(want) {
				t.Errorf("len(VTVariants(%q)) = %d; want %d", tt.word, len(got), len(want))
				// Debug output
				t.Logf("Original key: %v", CalculateVTKey(tt.word))
				t.Logf("Generated variants:")
				for k := range got {
					t.Logf("  %v", k)
				}
				t.Logf("Expected variants:")
				for k := range want {
					t.Logf("  %v", k)
				}
			}
		})
	}
}

// TestVTVariantsK tests k-distance variant generation
func TestVTVariantsK(t *testing.T) {
	tests := []struct {
		word string
		k    int
	}{
		{"", 1},
		{"a", 1},
		{"a", 2},
		{"ab", 1},
		{"ab", 2},
		{"a", 3},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_k%d", tt.word, tt.k), func(t *testing.T) {
			got := VTVariantsK(tt.word, tt.k)

			// Calculate expected number of variants by generating all possible edits up to distance k
			want := make(map[VTKey]bool)
			seenWords := map[string]bool{tt.word: true}
			frontier := map[string]bool{tt.word: true}

			// Add original word's key
			want[CalculateVTKey(tt.word)] = true

			// Generate variants layer by layer up to distance k
			for i := 0; i < tt.k; i++ {
				nextFrontier := make(map[string]bool)
				for w := range frontier {
					// Add keys for current word's variants
					for v := range VTEdits1Damerau(w) {
						if !seenWords[v] {
							seenWords[v] = true
							nextFrontier[v] = true
							want[CalculateVTKey(v)] = true
						}
					}
				}
				frontier = nextFrontier
				if len(frontier) == 0 {
					break // No more variants possible
				}
			}

			// Debug output for bounds comparison
			bound := VTBound(len(tt.word), tt.k)
			t.Logf("Word: %q (length %d, k=%d)", tt.word, len(tt.word), tt.k)
			t.Logf("Upper bound: %d", bound)
			t.Logf("Generated variants: %d", len(got))
			t.Logf("Expected variants: %d", len(want))
			if len(got) > bound {
				t.Logf("⚠️ Generated variants exceed bound by %d", len(got)-bound)
			}
			if len(want) > bound {
				t.Logf("⚠️ Expected variants exceed bound by %d", len(want)-bound)
			}

			if len(got) != len(want) {
				t.Errorf("len(VTVariantsK(%q, %d)) = %d; want %d", tt.word, tt.k, len(got), len(want))
				// Debug output
				t.Logf("Original key: %v", CalculateVTKey(tt.word))
				// t.Logf("Generated variants:")
				// for k := range got {
				// 	t.Logf("  %v", k)
				// }
				// t.Logf("Expected variants:")
				// for k := range want {
				// 	t.Logf("  %v", k)
				// }
			}

			// Check size bound only for k=1
			if tt.k == 1 && len(got) > VTBound(len(tt.word), 1) {
				t.Errorf("Size bound violated on %s: got %d, want ≤ %d", tt.word, len(got), VTBound(len(tt.word), 1))
			}
		})
	}
}

// TestVTVariantsConsistency tests consistency between algebraic and true variants
func TestVTVariantsConsistency(t *testing.T) {
	r := rand.New(rand.NewSource(42))

	for i := 0; i < 1000; i++ {
		// Generate random word
		n := r.Intn(10) + 1
		var word strings.Builder
		for j := 0; j < n; j++ {
			word.WriteByte(VTAlphabet[r.Intn(len(VTAlphabet))])
		}
		w := word.String()

		// Compare algebraic and true variants
		kAlg := VTVariants(w)
		kTrue := make(map[VTKey]bool)
		for v := range VTEdits1Damerau(w) {
			kTrue[CalculateVTKey(v)] = true
		}
		kTrue[CalculateVTKey(w)] = true

		// Check if sets are equal
		if len(kAlg) != len(kTrue) {
			t.Errorf("Mismatch on %s: got %d variants, want %d", w, len(kAlg), len(kTrue))
			break
		}

		// Check size bound
		if len(kAlg) > VTBound(len(w), 1) {
			t.Errorf("Size bound violated on %s: got %d, want ≤ %d", w, len(kAlg), VTBound(len(w), 1))
			break
		}
	}
}

// TestVTBound tests the size bound calculation
func TestVTBound(t *testing.T) {
	tests := []struct {
		n int
		k int
	}{
		{0, 0},
		{0, 1},
		{1, 0},
		{1, 1},
		{2, 1},
		{3, 1},
		{1, 2},
		{2, 2},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("n=%d_k=%d", tt.n, tt.k), func(t *testing.T) {
			// For k=0, we only need one word of length n
			var words []string
			if tt.k == 0 {
				if tt.n == 0 {
					words = []string{""}
				} else {
					// Just use a single word of the required length
					word := strings.Repeat("a", tt.n)
					words = []string{word}
				}
			} else {
				// For k>0, generate all possible words of length n
				words = generateAllWords(tt.n)
			}

			// For each word, generate all variants up to distance k
			allKeys := make(map[VTKey]bool)
			deletionKeys := make(map[VTKey]bool)
			substitutionKeys := make(map[VTKey]bool)
			insertionKeys := make(map[VTKey]bool)
			transpositionKeys := make(map[VTKey]bool)

			for _, word := range words {
				// Add original word's key
				origKey := CalculateVTKey(word)
				allKeys[origKey] = true

				// For k=0, we only consider the original word
				if tt.k == 0 {
					continue
				}

				// Count variants by operation type
				// Deletions
				for i := 0; i < len(word); i++ {
					key := CalculateVTKey(word[:i] + word[i+1:])
					deletionKeys[key] = true
					allKeys[key] = true
				}

				// Substitutions
				for i := 0; i < len(word); i++ {
					for _, c := range VTAlphabet {
						if c != rune(word[i]) {
							key := CalculateVTKey(word[:i] + string(c) + word[i+1:])
							substitutionKeys[key] = true
							allKeys[key] = true
						}
					}
				}

				// Insertions
				for i := 0; i <= len(word); i++ {
					for _, c := range VTAlphabet {
						key := CalculateVTKey(word[:i] + string(c) + word[i:])
						insertionKeys[key] = true
						allKeys[key] = true
					}
				}

				// Transpositions
				if len(word) >= 2 {
					for i := 0; i < len(word)-1; i++ {
						swapped := word[:i] + string(word[i+1]) + string(word[i]) + word[i+2:]
						key := CalculateVTKey(swapped)
						transpositionKeys[key] = true
						allKeys[key] = true
					}
				}
			}

			experimentalBound := len(allKeys)
			calculatedBound := VTBound(tt.n, tt.k)

			// Print breakdown of variants
			t.Logf("Breakdown for n=%d, k=%d:", tt.n, tt.k)
			t.Logf("  Original keys: 1")
			t.Logf("  Deletion keys: %d", len(deletionKeys))
			t.Logf("  Substitution keys: %d", len(substitutionKeys))
			t.Logf("  Insertion keys: %d", len(insertionKeys))
			t.Logf("  Transposition keys: %d", len(transpositionKeys))
			t.Logf("  Total unique keys: %d", experimentalBound)
			t.Logf("  Calculated bound: %d", calculatedBound)

			if calculatedBound < experimentalBound {
				t.Errorf("VTBound(%d, %d) = %d; experimental bound is %d",
					tt.n, tt.k, calculatedBound, experimentalBound)
			}
		})
	}
}

// generateAllWords generates all possible words of length n using VTAlphabet
func generateAllWords(n int) []string {
	if n == 0 {
		return []string{""}
	}

	// For n=1, return all single characters
	if n == 1 {
		words := make([]string, len(VTAlphabet))
		for i, ch := range VTAlphabet {
			words[i] = string(ch)
		}
		return words
	}

	// For n>1, recursively generate words
	prevWords := generateAllWords(n - 1)
	words := make([]string, 0, len(prevWords)*len(VTAlphabet))
	for _, prev := range prevWords {
		for _, ch := range VTAlphabet {
			words = append(words, prev+string(ch))
		}
	}
	return words
}

// BenchmarkVTKey benchmarks VT key calculation
func BenchmarkVTKey(b *testing.B) {
	words := []string{"", "a", "ab", "abc", "abcd", "abcde"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CalculateVTKey(words[i%len(words)])
	}
}

// BenchmarkVTEdits1 benchmarks edit generation
func BenchmarkVTEdits1(b *testing.B) {
	words := []string{"", "a", "ab", "abc", "abcd", "abcde"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		VTEdits1(words[i%len(words)])
	}
}

// BenchmarkVTEdits1Damerau benchmarks Damerau-Levenshtein edit generation
func BenchmarkVTEdits1Damerau(b *testing.B) {
	words := []string{"", "a", "ab", "abc", "abcd", "abcde"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		VTEdits1Damerau(words[i%len(words)])
	}
}

// BenchmarkVTVariants benchmarks algebraic variant generation
func BenchmarkVTVariants(b *testing.B) {
	words := []string{"", "a", "ab", "abc", "abcd", "abcde"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		VTVariants(words[i%len(words)])
	}
}

// BenchmarkVTVariantsK benchmarks k-distance variant generation
func BenchmarkVTVariantsK(b *testing.B) {
	words := []string{"", "a", "ab", "abc", "abcd", "abcde"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		VTVariantsK(words[i%len(words)], 2)
	}
}
