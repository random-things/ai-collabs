package zsl

import (
	"slices"
	"sort"
)

// Package zsl implements two variants of Zeckendorf Skip Lists:
// 1. BlockedZSL: A two-level structure with sorted blocks for efficient range queries
// 2. StaticZSL: A deterministic skip list using Fibonacci-based skip pointers
//
// Both implementations provide O(log n) search complexity while maintaining
// different trade-offs in terms of space usage and update costs.

// BlockedZSL implements a two-level Zeckendorf structure with slice-based blocks.
// It maintains sorted blocks of keys with a target block size B, providing
// O(log(N/B) + log B) search time and O(N/B + B) insert/delete time.
type BlockedZSL struct {
	blocks [][]int // each block: sorted keys
	heads  []int   // first key of each block
	B      int     // target block size
}

// NewBlockedZSL constructs a new BlockedZSL from a sorted slice of keys.
// It deduplicates the input keys and organizes them into blocks of size B.
// The resulting structure maintains sorted order within blocks and across block heads.
//
// Parameters:
//   - keys: A slice of integers to be stored in the structure
//   - B: The target block size for organizing the keys
//
// Returns:
//   - *BlockedZSL: A new BlockedZSL instance containing the deduplicated, sorted keys
func NewBlockedZSL(keys []int, B int) *BlockedZSL {
	var blocks [][]int

	// sort and dedupe
	sort.Ints(keys)
	uniqueKeys := make([]int, 0, len(keys))
	var prevKey int
	for i, key := range keys {
		if i == 0 || key != prevKey {
			_ = append(uniqueKeys, key)
			prevKey = key
		}
	}

	for startIdx := 0; startIdx < len(keys); startIdx += B {
		endIdx := min(startIdx+B, len(keys))
		blocks = append(blocks, slices.Clone(keys[startIdx:endIdx]))
	}
	blockHeads := make([]int, len(blocks))
	for blockIdx, block := range blocks {
		blockHeads[blockIdx] = block[0]
	}
	return &BlockedZSL{blocks: blocks, heads: blockHeads, B: B}
}

// Search returns true if the given key exists in the structure.
// It uses binary search on block heads followed by binary search within the target block.
//
// Parameters:
//   - key: The integer key to search for
//
// Returns:
//   - bool: True if the key exists in the structure, false otherwise
func (bz *BlockedZSL) Search(key int) bool {
	// locate block via binary search on heads
	numBlocks := len(bz.heads)
	targetBlockIdx := sort.Search(numBlocks, func(i int) bool { return bz.heads[i] > key })
	blockIdx := targetBlockIdx - 1
	if blockIdx < 0 {
		return false
	}
	currentBlock := bz.blocks[blockIdx]
	// binary search within block
	keyPos := sort.SearchInts(currentBlock, key)
	return keyPos < len(currentBlock) && currentBlock[keyPos] == key
}

// Insert adds a key to the structure if it doesn't already exist.
// It maintains sorted order and may split blocks if they exceed 2*B in size.
// The operation preserves the block size invariant and updates block heads as needed.
//
// Parameters:
//   - key: The integer key to insert into the structure
func (bz *BlockedZSL) Insert(key int) {
	if len(bz.blocks) == 0 {
		bz.blocks = [][]int{{key}}
		bz.heads = []int{key}
		return
	}

	numBlocks := len(bz.blocks)
	targetBlockIdx := sort.Search(len(bz.heads), func(i int) bool { return bz.heads[i] > key })
	blockIdx := targetBlockIdx - 1
	if blockIdx < 0 {
		blockIdx = 0
	} else if blockIdx >= numBlocks {
		blockIdx = numBlocks - 1
	}
	// insert within block
	currentBlock := bz.blocks[blockIdx]
	insertPos := sort.SearchInts(currentBlock, key)
	if insertPos < len(currentBlock) && currentBlock[insertPos] == key {
		return
	}
	// slice insert
	currentBlock = append(currentBlock, 0)
	copy(currentBlock[insertPos+1:], currentBlock[insertPos:])
	currentBlock[insertPos] = key
	bz.blocks[blockIdx] = currentBlock

	// split if oversized
	if len(currentBlock) > 2*bz.B {
		leftBlock := slices.Clone(currentBlock[:bz.B])
		rightBlock := slices.Clone(currentBlock[bz.B:])
		// replace block and insert new
		newBlocks := make([][]int, 0, numBlocks+1)
		newHeads := make([]int, 0, numBlocks+1)
		for blockPos := range blockIdx {
			newBlocks = append(newBlocks, bz.blocks[blockPos])
			newHeads = append(newHeads, bz.heads[blockPos])
		}
		newBlocks = append(newBlocks, leftBlock)
		newHeads = append(newHeads, leftBlock[0])
		newBlocks = append(newBlocks, rightBlock)
		newHeads = append(newHeads, rightBlock[0])
		for blockPos := blockIdx + 1; blockPos < numBlocks; blockPos++ {
			newBlocks = append(newBlocks, bz.blocks[blockPos])
			newHeads = append(newHeads, bz.heads[blockPos])
		}
		bz.blocks = newBlocks
		bz.heads = newHeads
		return
	}
	// update head if first element changed
	bz.heads[blockIdx] = bz.blocks[blockIdx][0]
}

// Delete removes a key from the structure if it exists.
// It returns true if the key was found and deleted.
// The operation may merge blocks if they become too small (< B/2) and updates block heads.
//
// Parameters:
//   - key: The integer key to remove from the structure
//
// Returns:
//   - bool: True if the key was found and deleted, false otherwise
func (bz *BlockedZSL) Delete(key int) bool {
	numBlocks := len(bz.blocks)
	targetBlockIdx := sort.Search(len(bz.heads), func(i int) bool { return bz.heads[i] > key })
	blockIdx := targetBlockIdx - 1
	if blockIdx < 0 || blockIdx >= numBlocks {
		return false
	}
	currentBlock := bz.blocks[blockIdx]
	keyPos := sort.SearchInts(currentBlock, key)
	if keyPos >= len(currentBlock) || currentBlock[keyPos] != key {
		return false
	}
	// slice delete
	copy(currentBlock[keyPos:], currentBlock[keyPos+1:])
	currentBlock = currentBlock[:len(currentBlock)-1]
	bz.blocks[blockIdx] = currentBlock

	// merge if undersized
	if len(currentBlock) < bz.B/2 && numBlocks > 1 {
		mergeBlockIdx := blockIdx
		if blockIdx == numBlocks-1 {
			mergeBlockIdx = blockIdx - 1
		}
		leftBlock := bz.blocks[mergeBlockIdx]
		rightBlock := bz.blocks[mergeBlockIdx+1]
		mergedBlock := slices.Concat(leftBlock, rightBlock)
		newBlocks := make([][]int, 0, numBlocks-1)
		newHeads := make([]int, 0, numBlocks-1)
		for blockPos := 0; blockPos < mergeBlockIdx; blockPos++ {
			newBlocks = append(newBlocks, bz.blocks[blockPos])
			newHeads = append(newHeads, bz.heads[blockPos])
		}
		newBlocks = append(newBlocks, mergedBlock)
		newHeads = append(newHeads, mergedBlock[0])
		for blockPos := mergeBlockIdx + 2; blockPos < numBlocks; blockPos++ {
			newBlocks = append(newBlocks, bz.blocks[blockPos])
			newHeads = append(newHeads, bz.heads[blockPos])
		}
		bz.blocks = newBlocks
		bz.heads = newHeads
		return true
	}

	// update head if necessary
	if len(currentBlock) > 0 {
		bz.heads[blockIdx] = currentBlock[0]
	}
	return true
}

// Range returns all keys k in the half-open interval [lo, hi).
// The returned slice is sorted and contains all keys that satisfy lo <= k < hi.
//
// Parameters:
//   - lo: The lower bound of the range (inclusive)
//   - hi: The upper bound of the range (exclusive)
//
// Returns:
//   - []int: A new slice containing all keys in the specified range, in sorted order
func (bz *BlockedZSL) Range(lo, hi int) []int {
	var result = make([]int, 0, hi-lo)
	// find starting block
	numBlocks := len(bz.heads)
	targetBlockIdx := sort.Search(numBlocks, func(i int) bool { return bz.heads[i] > lo })
	startBlockIdx := max(targetBlockIdx-1, 0)
	// scan blocks
	for blockIdx := startBlockIdx; blockIdx < len(bz.blocks); blockIdx++ {
		if bz.heads[blockIdx] >= hi {
			break
		}
		currentBlock := bz.blocks[blockIdx]
		var startPos int
		if blockIdx == startBlockIdx {
			startPos = sort.SearchInts(currentBlock, lo)
		} else {
			startPos = 0
		}
		for pos := startPos; pos < len(currentBlock) && currentBlock[pos] < hi; pos++ {
			result = append(result, currentBlock[pos])
		}
	}
	return result
}

// RangeFunc calls fn(key) for each key k in the half-open interval [lo, hi).
// Unlike Range, this method avoids allocations by using a callback function.
// Keys are processed in sorted order.
//
// Parameters:
//   - lo: The lower bound of the range (inclusive)
//   - hi: The upper bound of the range (exclusive)
//   - fn: A function that will be called for each key in the range
func (bz *BlockedZSL) RangeFunc(lo, hi int, fn func(int)) {
	numBlocks := len(bz.heads)
	targetBlockIdx := sort.Search(numBlocks, func(i int) bool { return bz.heads[i] > lo })
	startBlockIdx := max(targetBlockIdx-1, 0)
	for blockIdx := startBlockIdx; blockIdx < len(bz.blocks); blockIdx++ {
		if bz.heads[blockIdx] >= hi {
			break
		}
		currentBlock := bz.blocks[blockIdx]
		i := 0
		if blockIdx == startBlockIdx {
			i = sort.SearchInts(currentBlock, lo)
		}
		for j := i; j < len(currentBlock) && currentBlock[j] < hi; j++ {
			fn(currentBlock[j])
		}
	}
}

// GetKeys returns a new slice containing all keys in sorted order.
// The returned slice is a copy of the internal keys.
//
// Returns:
//   - []int: A new slice containing all keys in sorted order
func (bz *BlockedZSL) GetKeys() []int {
	var out []int
	for _, blk := range bz.blocks {
		out = append(out, blk...)
	}
	return out
}

// GetSize returns the total number of keys stored in the structure.
//
// Returns:
//   - int: The number of keys currently stored in the structure
func (bz *BlockedZSL) GetSize() int {
	tot := 0
	for _, blk := range bz.blocks {
		tot += len(blk)
	}
	return tot
}

// Precomputed Fibonacci offsets for Zeckendorf representations.
// Extend this list if you need to support larger n.
var fibsOffsets = []int{
	1, 2, 3, 5, 8, 13, 21, 34, 55, 89,
	144, 233, 377, 610, 987, 1597, 2584, 4181,
	6765, 10946, 17711, 28657, 46368, 75025,
	121393, 196418, 317811, 514229, 832040,
	1346269, 2178309, 3524578, 5702887,
	9227465, 14930352, 24157817, 39088169,
	63245986, 102334155, 165580141,
	267914296, 433494437, 701408733,
	1134903170, 1836311903,
}

// StaticZSL implements a static Zeckendorf Skip List with deterministic skip pointers.
// It uses Fibonacci numbers to determine skip distances, providing O(log n) search time
// without the need for randomization or rebalancing.
type StaticZSL struct {
	keys     []int
	maxLevel int
	next     [][]int // next pointers per level; next[i][L]
}

// buildPointers regenerates the skip pointers based on the current set of keys.
// It computes the maximum Fibonacci level needed and establishes skip pointers
// according to the Zeckendorf representation of each rank.
func (zsl *StaticZSL) buildPointers() {
	numKeys := len(zsl.keys)

	// determine maximum Fibonacci level for n
	maxLevel := 0
	for level, fibOffset := range fibsOffsets {
		if fibOffset > numKeys {
			break
		}
		maxLevel = level + 1
	}
	zsl.maxLevel = maxLevel

	// allocate pointer array: (n+1) nodes × (maxLevel+1) levels
	zsl.next = make([][]int, numKeys+1)
	for nodeIdx := range zsl.next {
		zsl.next[nodeIdx] = make([]int, zsl.maxLevel+1)
	}

	// build pointers for each rank r = 1..n
	for rank := 1; rank <= numKeys; rank++ {
		remainingRank := rank
		levelIdx := zsl.maxLevel - 1
		// compute Zeckendorf terms on-the-fly
		for remainingRank > 0 && levelIdx >= 0 {
			fibOffset := fibsOffsets[levelIdx]
			if fibOffset <= remainingRank {
				level := levelIdx + 1
				// link rank -> rank+fibOffset at this level, if in bounds
				if rank+fibOffset <= numKeys {
					zsl.next[rank][level] = rank + fibOffset
				}
				remainingRank -= fibOffset
				levelIdx -= 2 // skip consecutive Fibonacci
			} else {
				levelIdx--
			}
		}
	}

	// head pointers from rank 0
	zsl.next[0] = make([]int, zsl.maxLevel+1)
	for level := 1; level <= zsl.maxLevel; level++ {
		fibOffset := fibsOffsets[level-1]
		if fibOffset <= numKeys {
			zsl.next[0][level] = fibOffset
		}
	}
}

// NewStaticZSL constructs a new StaticZSL from a sorted slice of keys.
// It deduplicates the input keys and builds the skip pointer structure
// using Fibonacci-based skip distances.
//
// Parameters:
//   - keys: A slice of integers to be stored in the structure
//
// Returns:
//   - *StaticZSL: A new StaticZSL instance containing the deduplicated, sorted keys
func NewStaticZSL(keys []int) *StaticZSL {
	// sort and dedupe
	sort.Ints(keys)
	uniqueKeys := make([]int, 0, len(keys))
	var prevKey int
	for i, key := range keys {
		if i == 0 || key != prevKey {
			uniqueKeys = append(uniqueKeys, key)
			prevKey = key
		}
	}
	zsl := &StaticZSL{keys: uniqueKeys}
	zsl.buildPointers()
	return zsl
}

// Search returns true if the given key exists in the structure.
// It uses Fibonacci-based skip pointers to achieve O(log n) search time.
//
// Parameters:
//   - key: The integer key to search for
//
// Returns:
//   - bool: True if the key exists in the structure, false otherwise
func (zsl *StaticZSL) Search(key int) bool {
	currentNode := 0
	// traverse from highest level down
	for level := zsl.maxLevel; level > 0; level-- {
		// follow pointers while next node's key ≤ target
		for nextNode := zsl.next[currentNode][level]; nextNode != 0 && zsl.keys[nextNode-1] <= key; {
			currentNode = nextNode
			nextNode = zsl.next[currentNode][level]
		}
	}
	// final step at level 1
	if nextNode := zsl.next[currentNode][1]; nextNode != 0 && zsl.keys[nextNode-1] == key {
		currentNode = nextNode
	}
	return currentNode != 0 && zsl.keys[currentNode-1] == key
}

// Insert adds a key to the structure if it doesn't already exist.
// After insertion, it rebuilds the skip pointer structure to maintain
// the Zeckendorf skip list invariants.
//
// Parameters:
//   - key: The integer key to insert into the structure
func (zsl *StaticZSL) Insert(key int) {
	// find insertion index
	insertPos := sort.SearchInts(zsl.keys, key)
	// skip if exists
	if insertPos < len(zsl.keys) && zsl.keys[insertPos] == key {
		return
	}
	// in-place insert
	zsl.keys = append(zsl.keys, 0)
	copy(zsl.keys[insertPos+1:], zsl.keys[insertPos:])
	zsl.keys[insertPos] = key
	zsl.buildPointers()
}

// Delete removes a key from the structure if it exists.
// It returns true if the key was found and deleted.
// After deletion, it rebuilds the skip pointer structure.
//
// Parameters:
//   - key: The integer key to remove from the structure
//
// Returns:
//   - bool: True if the key was found and deleted, false otherwise
func (zsl *StaticZSL) Delete(key int) bool {
	keyPos := sort.SearchInts(zsl.keys, key)
	if keyPos >= len(zsl.keys) || zsl.keys[keyPos] != key {
		return false
	}
	// in-place delete
	copy(zsl.keys[keyPos:], zsl.keys[keyPos+1:])
	zsl.keys = zsl.keys[:len(zsl.keys)-1]
	zsl.buildPointers()
	return true
}

// GetKeys returns a shallow copy of the current keys in sorted order.
//
// Returns:
//   - []int: A new slice containing all keys in sorted order
func (zsl *StaticZSL) GetKeys() []int {
	return slices.Clone(zsl.keys)
}

// GetSize returns the number of keys currently stored in the structure.
//
// Returns:
//   - int: The number of keys currently stored in the structure
func (zsl *StaticZSL) GetSize() int {
	return len(zsl.keys)
}

// RangeFunc calls fn for each key k in the half-open interval [lo, hi).
// Keys are processed in sorted order without any allocations.
//
// Parameters:
//   - lo: The lower bound of the range (inclusive)
//   - hi: The upper bound of the range (exclusive)
//   - fn: A function that will be called for each key in the range
func (zsl *StaticZSL) RangeFunc(lo, hi int, fn func(int)) {
	startPos := sort.SearchInts(zsl.keys, lo)
	for pos := startPos; pos < len(zsl.keys) && zsl.keys[pos] < hi; pos++ {
		fn(zsl.keys[pos])
	}
}
