# Varshamov-Tenengolts (VT) Syndrome Prefilter

This package implements a novel approach to search-as-you-type using Varshamov-Tenengolts (VT) syndromes as keys for efficient fuzzy matching. The implementation is particularly optimized for short-length strings (typically ≤ 32 characters) and provides significant performance improvements over existing algorithms.

## Why VT Syndromes?

### 1. Constant-Space Key Representation
- Each string is mapped to a compact VT key (n, S, P) where:
  - n: string length
  - S: weighted sum of character positions
  - P: sum of character values
- This provides a constant-space representation regardless of string length
- Keys can be efficiently stored and compared

### 2. Efficient Edit Distance Filtering
- VT syndromes have a unique property: similar strings (within edit distance k) map to similar keys
- For distance-1 edits:
  - Substitutions: Change P by character difference, S by weighted difference
  - Deletions/Insertions: Change n and recalculate S, P
  - Transpositions: Change S based on position differences
- This allows us to quickly filter out strings that cannot be within edit distance k

### 3. Performance Advantages

#### Memory Efficiency
- Traditional approaches (e.g., BK-trees, Levenshtein automata) require O(n) space per string
- VT keys require only O(1) space per string
- This enables efficient in-memory indexing of large datasets

#### Query Performance
- Traditional approaches require O(n²) time for edit distance calculations
- VT syndrome filtering:
  - O(1) key calculation
  - O(1) key comparison
  - Only compute actual edit distance for candidates that pass the filter
- For short strings, this provides orders of magnitude better performance

#### Search-as-You-Type Optimization
- As users type, we can incrementally update VT keys
- Each keystroke only requires O(1) operations to update the key
- This enables real-time filtering even with large datasets

### 4. Theoretical Guarantees
- VT syndromes provide a theoretical upper bound on the number of possible variants
- This allows for predictable memory allocation and performance characteristics

## Where This Fits

The VT keys aren't an actual index. They serve as a prefiltering step to limit the
search space. To use in practice with maximum performance benefits:

### Build an index of VT keys from the prefixes of all words in your dictionary

```
for word in dictionary:
  for prefix in get_prefixes(word):
    index[vtkey(prefix)] = word
```

Note that you don't have to index all variants, just prefixes.

### Naive search of the index

```
query_keys = vtvariants(query)

for key in query_keys:
  if key in index:
    results.add(index[key])

# Collisions are expected!
# You *will* get false positives, but no false negatives.

# Standard Levenshtein distance is a fast and viable post-filter
for result in results:
  if edit_distance(query, result) > k:
    results.remove(result)
```

### Search-as-you-type

This optimization applies in search-as-you-type cases. The following code lets
you update your search key without having to recalculate S and P (an O(n) operation)
if you need maximum performance.

```
# to_int is a 1-based array of your alphabet
# M is your chosen prime > sigma
# sigma is len(alphabet)
v = to_int[c]
n += 1
S = (S + n * v) % (n + 1)
P = (P + v) % M
total += v
```

Then n, S, and P can be passed to your variants-generating function, skipping
the calculation. Without generating variants, this is only a shortcut for exact
match.

## Implementation Details

The package provides:
- `CalculateVTKey`: Maps strings to VT keys
- `VTVariants`: Generates all possible VT keys within edit distance 1
- `VTVariantsK`: Generates all possible VT keys within edit distance k

## Use Cases

This implementation is particularly effective for:
- Search-as-you-type interfaces
- Autocomplete systems
- Spell checkers
- Short string fuzzy matching
- Real-time filtering of large datasets

## Limitations

- Preferable for short strings (≤ 32 characters)
- Best performance with small edit distances (k ≤ 2)
- Requires character set of known size (currently implemented for lowercase a-z)

## Potential Future Work

- Extend to larger character sets
- Support substring matching
- Optimize for larger edit distances
- Implement parallel processing for large datasets
- Add support for weighted edit distances
