package csi

import (
	"errors"
	"math"
	"sort"
)

// CSI implements the Constellation Search Index operating on raw bytes.
type CSI struct {
	k     int                  // anchor length in bytes
	gaps  []int                // fixed distances between anchors in bytes
	index map[anchorPair][]int // maps byte-anchor pairs to start offsets
	text  []byte               // original document bytes
}

// anchorPair represents two byte-slice anchors and the gap between them
// a1 and a2 are stored as strings for comparability in map keys.
type anchorPair struct {
	a1  string // first k-byte anchor
	a2  string // second k-byte anchor
	gap int    // distance in bytes between anchors
}

// New constructs a CSI index for the provided text. It chooses k and gaps based on entropy.
func New(text string) *CSI {
	b := []byte(text)
	n := len(b)
	if n == 0 {
		panic("text must be non-empty")
	}

	// Compute byte-level Shannon entropy
	entropy := func(data []byte) float64 {
		freq := make([]int, 256)
		for _, v := range data {
			freq[v]++
		}
		var ent float64
		n := float64(len(data))
		for _, c := range freq {
			if c == 0 {
				continue
			}
			p := float64(c) / n
			ent -= p * math.Log2(p)
		}
		return ent
	}(b)

	// Select k and gaps based on entropy thresholds
	var k int
	var gaps []int
	switch {
	case entropy < 3.5:
		k = 6
		gaps = []int{8, 16, 32, 64}
	case entropy < 4.5:
		k = 5
		gaps = []int{6, 12, 24, 48}
	default:
		k = 4
		gaps = []int{4, 8, 16, 32}
	}
	sort.Ints(gaps)

	csi := &CSI{
		k:     k,
		gaps:  gaps,
		index: make(map[anchorPair][]int),
		text:  b,
	}

	// Build index: for each byte offset, record anchor pairs
	for i := 0; i <= n-k; i++ {
		a1 := string(b[i : i+k])
		for _, d := range gaps {
			j := i + d
			if j+k <= n {
				a2 := string(b[j : j+k])
				key := anchorPair{a1: a1, a2: a2, gap: d}
				csi.index[key] = append(csi.index[key], i)
			}
		}
	}
	return csi
}

// Search returns all byte-offsets where pattern occurs in text
func (c *CSI) Search(pattern string) ([]int, error) {
	pb := []byte(pattern)
	m := len(pb)
	if m == 0 {
		return nil, errors.New("pattern is empty")
	}
	// Patterns shorter than k+minGap+k must fallback
	minLen := c.k + c.gaps[0] + c.k
	if m < minLen {
		return nil, errors.New("pattern too short for CSI; fallback required")
	}

	// Build anchor pairs for pattern
	type pair struct {
		key anchorPair
		pos []int
	}
	var pairs []pair
	for _, d := range c.gaps {
		if d+c.k <= m {
			a1 := string(pb[:c.k])
			a2 := string(pb[d : d+c.k])
			key := anchorPair{a1: a1, a2: a2, gap: d}
			if pos, ok := c.index[key]; ok {
				pairs = append(pairs, pair{key, pos})
			} else {
				return nil, errors.New("no match found")
			}
		}
	}
	if len(pairs) == 0 {
		return nil, errors.New("pattern too short for any anchor pair; fallback required")
	}

	// Sort ascending by posting-list length
	sort.Slice(pairs, func(i, j int) bool {
		return len(pairs[i].pos) < len(pairs[j].pos)
	})

	// Intersect posting lists
	candidates := make(map[int]struct{}, len(pairs[0].pos))
	for _, off := range pairs[0].pos {
		candidates[off] = struct{}{}
	}
	for _, p := range pairs[1:] {
		next := make(map[int]struct{})
		for _, off := range p.pos {
			if _, ok := candidates[off]; ok {
				next[off] = struct{}{}
			}
		}
		if len(next) == 0 {
			return nil, errors.New("no candidates found")
		}
		candidates = next
	}

	// Verify matches with byte-level compare
	dlen := len(pb)
	var result []int
	for off := range candidates {
		if off+dlen <= len(c.text) && string(c.text[off:off+dlen]) == pattern {
			result = append(result, off)
		}
	}
	sort.Ints(result)
	return result, nil
}

// Stats returns basic index statistics
func (c *CSI) Stats() map[string]any {
	tb := len(c.index)
	var total, max int
	for _, pos := range c.index {
		total += len(pos)
		if len(pos) > max {
			max = len(pos)
		}
	}
	return map[string]any{
		"text_length":    len(c.text),
		"anchor_length":  c.k,
		"gaps":           c.gaps,
		"total_buckets":  tb,
		"total_postings": total,
		"max_postings":   max,
		"avg_postings":   float64(total) / float64(tb),
	}
}
