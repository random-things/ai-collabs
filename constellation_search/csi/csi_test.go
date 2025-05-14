package csi

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		k       int
		gaps    []int
		wantErr bool
	}{
		{
			name:    "valid parameters",
			text:    "hello world",
			k:       4,
			gaps:    []int{4, 8, 16},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); (r != nil) != tt.wantErr {
					t.Errorf("New() panic = %v, wantErr %v", r, tt.wantErr)
				}
			}()
			_ = New(tt.text)
		})
	}
}

func TestSearch(t *testing.T) {
	text := "hello world, hello there, hello universe"
	t.Logf("Test text: %q", text)
	t.Logf("Text length: %d", len(text))
	t.Logf("'universe' starts at: %d", strings.Index(text, "universe"))

	csi := New(text)

	tests := []struct {
		name      string
		pattern   string
		expected  []int
		expectErr bool
	}{
		{
			name:      "exact match",
			pattern:   "hello",
			expected:  []int{0, 13, 26},
			expectErr: true, // "hello" is 5 chars, less than min 12
		},
		{
			name:      "no match",
			pattern:   "xyz",
			expected:  nil,
			expectErr: true, // "xyz" is 3 chars, less than min 12
		},
		{
			name:      "empty pattern",
			pattern:   "",
			expected:  nil,
			expectErr: true, // empty pattern is less than min 12
		},
		{
			name:      "pattern too short",
			pattern:   "hel",
			expected:  nil,
			expectErr: true, // "hel" is 3 chars, less than min 12
		},
		{
			name:      "pattern at end",
			pattern:   "universe",
			expected:  []int{32},
			expectErr: true, // "universe" is 8 chars, less than min 12
		},
	}

	// Add a test case with a pattern of adequate length
	longPattern := "ello world, hello there"
	longPatternPos := strings.Index(text, longPattern)
	if longPatternPos >= 0 {
		tests = append(tests, struct {
			name      string
			pattern   string
			expected  []int
			expectErr bool
		}{
			name:      "pattern long enough",
			pattern:   longPattern,
			expected:  []int{longPatternPos},
			expectErr: false,
		})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := csi.Search(tt.pattern)
			if tt.expectErr {
				if err == nil {
					t.Errorf("Search() expected an error for pattern %q (len=%d), got nil",
						tt.pattern, len(tt.pattern))
				}
				// If error expected, no need to check the results
				return
			} else {
				if err != nil {
					t.Errorf("Search() unexpected error for pattern %q (len=%d): %v",
						tt.pattern, len(tt.pattern), err)
					return
				}
				if len(got) != len(tt.expected) {
					t.Errorf("Search() got %v positions, want %v", len(got), len(tt.expected))
					return
				}
				for i := range got {
					if got[i] != tt.expected[i] {
						t.Errorf("Search() position %d = %v, want %v", i, got[i], tt.expected[i])
						if tt.name == "pattern at end" {
							t.Logf("Text around position %d: %q", got[i], text[got[i]:got[i]+len(tt.pattern)])
							t.Logf("Text around position %d: %q", tt.expected[i], text[tt.expected[i]:tt.expected[i]+len(tt.pattern)])
						}
					}
				}
			}
		})
	}
}

func TestStats(t *testing.T) {
	text := "hello world, hello there, hello universe"
	csi := New(text)
	stats := csi.Stats()

	expectedFields := []string{
		"text_length", "anchor_length", "gaps", "total_buckets", "total_postings", "max_postings", "avg_postings",
	}
	for _, field := range expectedFields {
		if _, ok := stats[field]; !ok {
			t.Errorf("Stats() missing field %s", field)
		}
	}

	// Check specific values
	if stats["text_length"] != len(text) {
		t.Errorf("Stats() text_length = %v, want %v", stats["text_length"], len(text))
	}
	if stats["anchor_length"] != 6 {
		t.Errorf("Stats() anchor_length = %v, want 6", stats["anchor_length"])
	}
}

func BenchmarkBuild(b *testing.B) {
	// (A) Worst case (random text) – CSI buckets nearly full, build is slower
	b.Run("worst_case", func(b *testing.B) {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		text := make([]byte, 100_000)
		for i := range text {
			text[i] = byte('a' + r.Intn(26))
		}
		s := string(text)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = New(s)
		}
	})

	// (B) "Normal" (repetitive) text – CSI buckets are sparse, build is faster
	b.Run("normal", func(b *testing.B) {
		normal := "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum."
		// Repeat so that the total length is comparable (e.g. 100_000 chars)
		normal = strings.Repeat(normal, 100)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = New(normal)
		}
	})
}

// Find all occurrences of pattern in text using Go's standard library
func allOccurrences(text, pattern string) []int {
	var res []int
	for i := 0; ; {
		idx := strings.Index(text[i:], pattern)
		if idx == -1 {
			break
		}
		res = append(res, i+idx)
		i += idx + 1 // move past this match
	}
	return res
}

// generateRandomUnicodeText creates a string of random Unicode characters of the specified length in bytes
func generateRandomUnicodeText(size int) string {
	// Unicode ranges to sample from (inclusive)
	ranges := []struct {
		start rune
		end   rune
	}{
		{0x4E00, 0x9FFF},   // CJK Unified Ideographs
		{0x3040, 0x309F},   // Hiragana
		{0x30A0, 0x30FF},   // Katakana
		{0xAC00, 0xD7AF},   // Hangul Syllables
		{0x0900, 0x097F},   // Devanagari
		{0x0600, 0x06FF},   // Arabic
		{0x0400, 0x04FF},   // Cyrillic
		{0x1F600, 0x1F64F}, // Emoticons
		{0x1F300, 0x1F5FF}, // Miscellaneous Symbols and Pictographs
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	var builder strings.Builder
	builder.Grow(size) // Pre-allocate space

	// Keep adding characters until we reach or exceed the target size
	for builder.Len() < size {
		// Pick a random range
		rangeIdx := r.Intn(len(ranges))
		rng := ranges[rangeIdx]

		// Generate a random rune from the selected range
		runeRange := rng.end - rng.start
		runeValue := rng.start + rune(r.Int63n(int64(runeRange)))

		builder.WriteRune(runeValue)
	}

	// Trim to exact size if we went over
	result := builder.String()
	if len(result) > size {
		// Convert to runes to avoid cutting in the middle of a character
		runes := []rune(result)
		for len(string(runes)) > size {
			runes = runes[:len(runes)-1]
		}
		result = string(runes)
	}

	return result
}

func BenchmarkSearchImplementations(b *testing.B) {
	// Natural language text
	naturalText := `Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. 
	Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. 
	Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. 
	Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.`
	naturalText = strings.Repeat(naturalText, 100)

	// Unicode text
	unicodeSample := "Hello 世界! 🌍 こんにちは! 안녕하세요! Привет мир! مرحبا بالعالم! Bonjour le monde! 😀😁😂🤣😃😄😅😆😉😊😋😎😍😘🥰😗😙😚🙂🤗🤩🤔🤨"
	unicodeText := strings.Repeat(unicodeSample, 50000) // ~5M chars, much larger and more varied

	// Generate Unicode patterns of required lengths
	unicodePatternLens := []int{16, 32, 64, 128}
	unicodePatterns := make([]string, len(unicodePatternLens))

	// Convert to runes for proper Unicode handling
	unicodeRunes := []rune(unicodeSample)
	// Ensure we have enough runes for our patterns
	if len(unicodeRunes) < unicodePatternLens[len(unicodePatternLens)-1] {
		// If not enough runes, repeat the sample until we have enough
		neededLen := unicodePatternLens[len(unicodePatternLens)-1] * 2 // Double what we need to ensure enough space
		repeats := neededLen/len(unicodeRunes) + 1
		unicodeRunes = []rune(strings.Repeat(unicodeSample, repeats))
	}

	for i, plen := range unicodePatternLens {
		// Calculate a safe start position that ensures we have enough space for the pattern
		maxStart := len(unicodeRunes) - plen
		if maxStart <= 0 {
			panic(fmt.Sprintf("not enough runes for pattern length %d (have %d runes)", plen, len(unicodeRunes)))
		}
		// Use a fixed offset from the start to ensure we get different patterns
		start := (i * 10) % maxStart
		unicodePatterns[i] = string(unicodeRunes[start : start+plen])
	}

	// Repetitive text
	repetitiveText := strings.Repeat("ABCDEFGHIJKLMNOPQRSTUVWXYZ", 1000)

	// Random text of different sizes
	sizes := []int{10_000, 100_000, 1_000_000, 10_000_000}
	randomTexts := make(map[int]string)
	for _, size := range sizes {
		randomTexts[size] = generateRandomUnicodeText(size)
	}

	// Test cases
	testCases := []struct {
		name     string
		text     string
		patterns []string
	}{
		{
			name: "natural_language",
			text: naturalText,
			patterns: []string{
				"Lorem ipsum dolor sit amet",
				"consectetur adipiscing elit. Sed do eiusmod tempor",
				"Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.",
			},
		},
		{
			name:     "unicode_text",
			text:     unicodeText,
			patterns: unicodePatterns,
		},
		{
			name: "repetitive_text",
			text: repetitiveText,
			patterns: []string{
				"ABCDEFGHIJKL",
				"MNOPQRSTUVWXYZABCDE",
				"UVWXYZABCDEFGHIJKLMNOPQRSTUVWXYZABCDEFGHIJK",
			},
		},
	}

	// Add random text test cases
	for _, size := range sizes {
		patterns := make([]string, 1000)
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		text := randomTexts[size]
		for i := range patterns {
			start := r.Intn(len(text) - 32)
			patterns[i] = text[start : start+32]
		}
		testCases = append(testCases, struct {
			name     string
			text     string
			patterns []string
		}{
			name:     fmt.Sprintf("random_text_%d", size),
			text:     text,
			patterns: patterns,
		})
	}

	// Run benchmarks
	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			// CSI implementation
			b.Run("csi", func(b *testing.B) {
				csi := New(tc.text)
				var totalMatches int
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					pattern := tc.patterns[i%len(tc.patterns)]
					matches, err := csi.Search(pattern)
					if err != nil {
						b.Fatalf("Error searching for pattern %q: %v", pattern, err)
					}
					totalMatches += len(matches)
				}
				b.ReportMetric(float64(totalMatches)/float64(b.N), "matches/op")
			})

			// Built-in implementation
			b.Run("builtin", func(b *testing.B) {
				var totalMatches int
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					pattern := tc.patterns[i%len(tc.patterns)]
					matches := allOccurrences(tc.text, pattern)
					totalMatches += len(matches)
				}
				b.ReportMetric(float64(totalMatches)/float64(b.N), "matches/op")
			})

			// Verify match counts are identical
			csi := New(tc.text)
			for _, pattern := range tc.patterns {
				csiMatches, _ := csi.Search(pattern)
				builtinMatches := allOccurrences(tc.text, pattern)
				if len(csiMatches) != len(builtinMatches) {
					b.Errorf("Mismatch in number of matches for pattern %q: CSI found %d, built-in found %d",
						pattern, len(csiMatches), len(builtinMatches))
				}
			}
		})
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
