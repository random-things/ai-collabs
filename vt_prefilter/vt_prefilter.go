package vt_prefilter

import (
	"math"
)

// VTConstants for Varshamov-Tenengolts encoding
const (
	VTAlphabet = "abcdefghijklmnopqrstuvwxyz"
	VTSigma    = len(VTAlphabet)
	VTM        = 67 // ≥ sigma+1, prime
)

// vtToInt maps characters to their 1-based index
var vtToInt = make(map[rune]int)

func init() {
	for i, c := range VTAlphabet {
		vtToInt[c] = i + 1
	}
}

// posMod returns a positive modulo result (like Python's % operator)
func posMod(a, m int) int {
	return ((a % m) + m) % m
}

// VTKey represents a Varshamov-Tenengolts key (n, S, P)
type VTKey struct {
	N int // length
	S int // sum of (i+1)*char_value
	P int // sum of char_values mod M
}

// CalculateVTKey calculates the VT key for a word
func CalculateVTKey(word string) VTKey {
	n := len(word)
	if n == 0 {
		return VTKey{0, 0, 0}
	}

	var s, p int
	for i, ch := range word {
		val := vtToInt[ch]
		s += (i + 1) * val // (i+1) because we want 1-based indexing
		p += val
	}
	return VTKey{
		N: n,
		S: posMod(s, n+1), // mod (n+1) because we want values in [0,n]
		P: posMod(p, VTM), // mod M to keep values in [0,M-1]
	}
}

// VTEdits1 generates all distance-1 string variants
func VTEdits1(w string) map[string]bool {
	result := make(map[string]bool)

	// Deletions
	for i := 0; i < len(w); i++ {
		result[w[:i]+w[i+1:]] = true
	}

	// Substitutions
	for i := 0; i < len(w); i++ {
		for _, c := range VTAlphabet {
			if rune(w[i]) != c {
				result[w[:i]+string(c)+w[i+1:]] = true
			}
		}
	}

	// Insertions
	for i := 0; i <= len(w); i++ {
		for _, c := range VTAlphabet {
			result[w[:i]+string(c)+w[i:]] = true
		}
	}

	return result
}

// VTEdits1Damerau includes adjacent transpositions
func VTEdits1Damerau(w string) map[string]bool {
	result := VTEdits1(w)

	// Add adjacent transpositions
	for i := 0; i < len(w)-1; i++ {
		swapped := w[:i] + string(w[i+1]) + string(w[i]) + w[i+2:]
		result[swapped] = true
	}

	return result
}

// VTVariants returns keys of all words at Damerau-Levenshtein distance ≤ 1
func VTVariants(word string) map[VTKey]bool {
	n := len(word)
	if n == 0 {
		// For empty string, consider all possible single-character insertions
		result := map[VTKey]bool{{0, 0, 0}: true} // Include empty key
		for _, c := range VTAlphabet {
			result[CalculateVTKey(string(c))] = true
		}
		return result
	}

	// Calculate original key
	key := CalculateVTKey(word)
	result := map[VTKey]bool{key: true}

	// Get all variants using existing function
	variants := VTEdits1Damerau(word)

	// Calculate VT keys for each variant
	for variant := range variants {
		result[CalculateVTKey(variant)] = true
	}

	return result
}

// VTVariantsK returns keys reachable by ≤ k Damerau-Levenshtein edits
func VTVariantsK(word string, k int) map[VTKey]bool {
	seenWords := map[string]bool{word: true}
	frontier := map[string]bool{word: true}
	keys := make(map[VTKey]bool)

	// Include original word's key
	keys[CalculateVTKey(word)] = true

	for i := 0; i < k; i++ {
		nextFrontier := make(map[string]bool)
		for w := range frontier {
			// Generate next layer variants
			variants := VTEdits1Damerau(w)
			for v := range variants {
				if !seenWords[v] {
					seenWords[v] = true
					nextFrontier[v] = true
					// Calculate VT key for each new variant
					keys[CalculateVTKey(v)] = true
				}
			}
		}
		frontier = nextFrontier
		if len(frontier) == 0 {
			break
		}
	}

	return keys
}

// VTBound returns the theoretical upper bound on variant count for k edits
func VTBound(n, k int) int {
	if n == 0 {
		if k == 0 {
			return 1 // Only empty string
		}
		return VTSigma + 1 // Empty string + all single character insertions
	}

	if k == 0 {
		return 1 // Only the original word
	}

	// For k=1, calculate bound mathematically
	if k == 1 {
		// For a word of length n:
		// - Original word: 1
		// - Deletions: Each deletion position can result in different VT keys
		//   because the remaining characters' positions change.
		//   For each position i (0 to n-1):
		//   - Characters after i shift left
		//   - Affects both S and P values differently based on position
		//   So we need to consider each position and each possible word:
		//   n positions × VTSigma possible words = n×VTSigma
		// - Substitutions: Each position and character can result in different VT keys
		//   For each position i (0 to n-1) and each character c:
		//   - Changes both S and P values differently based on position
		//   So we need to consider each position and character:
		//   n positions × (VTSigma-1) characters × VTSigma possible words = n×(VTSigma-1)×VTSigma
		// - Insertions: Each position can result in different VT keys
		//   For each position i (0 to n), inserting a character c:
		//   - Changes the position of all characters after i
		//   - Affects both S and P values differently based on position
		//   So we need to consider each position separately:
		//   (n+1) positions × VTSigma characters × VTSigma possible words = (n+1)×VTSigma×VTSigma
		// - Transpositions: Each transposition can result in different VT keys
		//   For each position i (0 to n-2):
		//   - Changes positions of two characters
		//   - Affects both S and P values differently based on position
		//   So we need to consider each position and each possible word:
		//   (n-1) positions × VTSigma possible words = (n-1)×VTSigma
		deletions := n * VTSigma
		substitutions := n * (VTSigma - 1) * VTSigma
		insertions := (n + 1) * VTSigma * VTSigma
		transpositions := max(0, n-1) * VTSigma
		return 1 + deletions + substitutions + insertions + transpositions
	}

	// For k>1, calculate bound by considering all possible intermediate lengths
	// and the maximum number of variants possible at each length
	maxLen := n + k       // Maximum possible length after k insertions
	minLen := max(0, n-k) // Minimum possible length after k deletions

	// Calculate maximum variants at each possible length
	total := 0
	for l := minLen; l <= maxLen; l++ {
		// At each length l, we can have:
		// - Original word: 1
		// - l×VTSigma deletions
		// - l×(VTSigma-1)×VTSigma substitutions
		// - (l+1)×VTSigma×VTSigma insertions
		// - (l-1)×VTSigma transpositions (if l >= 2)
		variantsAtLength := 1 + // original word
			l*VTSigma + // deletions
			l*(VTSigma-1)*VTSigma + // substitutions
			(l+1)*VTSigma*VTSigma + // insertions
			max(0, l-1)*VTSigma // transpositions

		// For k>1, we need to consider that each edit can be applied to any position
		// in any intermediate string. The number of possible intermediate strings
		// grows exponentially with k, so we multiply by k^k to account for this.
		// This is a very loose bound, but it's guaranteed to be larger than the
		// actual number of variants.
		total += variantsAtLength * int(math.Pow(float64(k), float64(k)))
	}

	return total
}
