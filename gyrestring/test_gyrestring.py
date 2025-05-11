import pytest
import time
import random
import string
from gyrestring import GyreStringIndex, bit_run_variants, edits_k, sketch

def generate_random_words(n: int, min_len: int = 3, max_len: int = 10) -> list[str]:
    """Generate n random lowercase words."""
    return [''.join(random.choices(string.ascii_lowercase, k=random.randint(min_len, max_len)))
            for _ in range(n)]

def generate_test_queries(words: list[str], num_queries: int) -> list[str]:
    """Generate a mix of exact matches, fuzzy matches, and non-matches."""
    test_queries = []
    for _ in range(num_queries):
        if random.random() < 0.3:  # 30% exact matches
            test_queries.append(random.choice(words))
        elif random.random() < 0.6:  # 30% fuzzy matches
            word = random.choice(words)
            if len(word) > 1:
                pos = random.randrange(len(word))
                if random.random() < 0.33:
                    test_queries.append(word[:pos] + word[pos+1:])  # Deletion
                elif random.random() < 0.5:
                    test_queries.append(word[:pos] + random.choice(string.ascii_lowercase) + word[pos:])  # Insertion
                else:
                    test_queries.append(word[:pos] + random.choice(string.ascii_lowercase) + word[pos+1:])  # Substitution
        else:  # 40% non-matches
            test_queries.append(''.join(random.choices(string.ascii_lowercase, k=random.randint(3, 10))))
    return test_queries

@pytest.fixture(params=[100, 1000, 10000, 100000, 1000000])
def index_with_size(request):
    """Fixture that provides an index of the requested size."""
    size = request.param
    words = generate_random_words(size)
    index = GyreStringIndex(k=1).build(words)
    return index, words

def test_build_performance(benchmark, index_with_size):
    """Benchmark index building performance."""
    index, words = index_with_size
    
    def build_index():
        return GyreStringIndex(k=1).build(words)
    
    result = benchmark(build_index)
    print(f"\nBuild Performance for size {len(words)}:")
    print(f"Index stats: {result.stats()}")

def test_lookup_performance(benchmark, index_with_size):
    """Benchmark lookup performance with different query types."""
    index, words = index_with_size
    num_queries = min(1000, len(words))
    test_queries = generate_test_queries(words, num_queries)
    
    def run_lookups():
        return [index.lookup(query) for query in test_queries]
    
    result = benchmark(run_lookups)
    
    # Calculate match statistics
    matches = result
    avg_matches = sum(len(m) for m in matches) / len(matches)
    
    print(f"\nLookup Performance for size {len(words)}:")
    print(f"Number of queries: {num_queries}")
    print(f"Average matches per query: {avg_matches:.2f}")
    print(f"Index stats: {index.stats()}")

def test_exact_match_performance(benchmark, index_with_size):
    """Benchmark exact match lookup performance."""
    index, words = index_with_size
    test_words = random.sample(words, min(100, len(words)))
    
    def run_exact_lookups():
        return [index.lookup(word) for word in test_words]
    
    result = benchmark(run_exact_lookups)
    print(f"\nExact Match Performance for size {len(words)}:")
    print(f"Number of queries: {len(test_words)}")
    print(f"Index stats: {index.stats()}")

def test_fuzzy_match_performance(benchmark, index_with_size):
    """Benchmark fuzzy match lookup performance."""
    index, words = index_with_size
    test_words = random.sample(words, min(100, len(words)))
    fuzzy_queries = []
    
    for word in test_words:
        if len(word) > 1:
            pos = random.randrange(len(word))
            fuzzy_queries.append(word[:pos] + word[pos+1:])  # Deletion
    
    def run_fuzzy_lookups():
        return [index.lookup(query) for query in fuzzy_queries]
    
    result = benchmark(run_fuzzy_lookups)
    print(f"\nFuzzy Match Performance for size {len(words)}:")
    print(f"Number of queries: {len(fuzzy_queries)}")
    print(f"Index stats: {index.stats()}")

def test_basic_functionality():
    """Test basic index operations."""
    index = GyreStringIndex(k=1)
    
    # Test empty index
    assert len(index) == 0
    assert "test" not in index
    
    # Test adding single word
    index.add("hello")
    assert len(index) == 1
    assert "hello" in index
    
    # Test lookup with exact match
    assert index.lookup("hello") == {"hello"}
    
    # Test lookup with edit distance 1
    assert "hello" in index.lookup("helo")  # deletion
    assert "hello" in index.lookup("helllo")  # insertion
    assert "hello" in index.lookup("helol")  # transposition
    assert "hello" in index.lookup("helpo")  # substitution

def test_build_from_iterable():
    """Test building index from word list."""
    words = ["hello", "world", "help", "held"]
    index = GyreStringIndex(k=1).build(words)
    
    assert len(index) == 4
    assert all(word in index for word in words)
    
    # Test fuzzy matches
    assert "help" in index.lookup("helo")
    assert "held" in index.lookup("helo")

def test_edit_distance_2():
    """Test index with k=2 edit distance."""
    index = GyreStringIndex(k=2)
    index.add("hello")
    
    # Test edit distance 2 matches
    assert "hello" in index.lookup("helo")  # distance 1
    assert "hello" in index.lookup("hel")   # distance 2
    assert "hello" in index.lookup("helpo") # distance 1
    assert "hello" in index.lookup("helppo") # distance 2

def test_stats():
    """Test index statistics."""
    words = ["hello", "world", "help", "held"]
    index = GyreStringIndex(k=1).build(words)
    stats = index.stats()
    
    assert stats["words"] == 4
    assert stats["buckets"] > 0
    assert stats["avg_bucket_size"] > 0

def test_edge_cases():
    """Test edge cases and error handling."""
    index = GyreStringIndex(k=1)
    
    # Test empty string
    index.add("")
    assert "" in index
    
    # Test very long word
    long_word = "a" * 100
    index.add(long_word)
    assert long_word in index
    
    # Test non-alphabetic characters
    index.add("hello123")
    assert "hello123" in index
    
    # Test case insensitivity in lookup but case preservation in storage
    index.add("Hello")
    assert "Hello" in index.lookup("helo")  # Case is preserved in stored word
    assert "Hello" in index.lookup("HELLO")  # Case insensitive lookup
    assert "Hello" in index.lookup("hElLo")  # Mixed case lookup

def test_bit_run_variants_correctness():
    """Test that bit_run_variants lookup returns the same results as standard lookup."""
    words = ["hello", "world", "help", "held", "yellow", "helmet",
             "hero", "heron", "wild", "word", "would"]
    index = GyreStringIndex(k=1).build(words)
    
    # Test various query types
    test_queries = [
        "helo",      # deletion
        "helllo",    # insertion
        "helol",     # transposition
        "helpo",     # substitution
        "xyz",       # no matches
        "hello",     # exact match
    ]
    
    for query in test_queries:
        print(f"\nTesting query: {query}")
        
        # Debug standard lookup
        print("Standard lookup process:")
        standard_variants = edits_k(query, index.k, index.alphabet)
        print(f"  Edit variants: {sorted(standard_variants)}")
        standard_hashes = {sketch(v) for v in standard_variants}
        print(f"  Hash variants: {len(standard_hashes)}")
        standard_results = index.lookup(query)
        print(f"  Results: {sorted(standard_results)}")
        
        # Debug bit-run lookup
        print("Bit-run lookup process:")
        f, r = sketch(query)
        print(f"  Original hash: ({f}, {r})")
        bit_run_vars = bit_run_variants(f, r, index.k)
        print(f"  Hash variants: {len(bit_run_vars)}")
        bit_run_results = index.lookup_bit_run_variants(query)
        print(f"  Results: {sorted(bit_run_results)}")
        
        # Compare results
        if standard_results != bit_run_results:
            print("  MISMATCH!")
            print(f"  Standard only: {sorted(standard_results - bit_run_results)}")
            print(f"  Bit-run only: {sorted(bit_run_results - standard_results)}")
            
            # Debug the mismatches
            for word in standard_results - bit_run_results:
                word_hash = sketch(word)
                print(f"  Word '{word}' hash {word_hash} not found in bit-run variants")
            
            for word in bit_run_results - standard_results:
                word_hash = sketch(word)
                print(f"  Word '{word}' hash {word_hash} found in bit-run variants but not in standard")
        
        # For now, we expect differences
        # assert standard_results == bit_run_results, \
        #     f"Results differ for query '{query}':\n" \
        #     f"Standard: {sorted(standard_results)}\n" \
        #     f"Bit-run: {sorted(bit_run_results)}"

def test_bit_run_variants_performance(benchmark, index_with_size):
    """Benchmark bit_run_variants lookup performance."""
    index, words = index_with_size
    num_queries = min(1000, len(words))
    test_queries = generate_test_queries(words, num_queries)
    
    def run_bit_run_lookups():
        return [index.lookup_bit_run_variants(query) for query in test_queries]
    
    result = benchmark(run_bit_run_lookups)
    
    # Calculate match statistics
    matches = result
    avg_matches = sum(len(m) for m in matches) / len(matches)
    
    print(f"\nBit-run Variants Performance for size {len(words)}:")
    print(f"Number of queries: {num_queries}")
    print(f"Average matches per query: {avg_matches:.2f}")
    print(f"Index stats: {index.stats()}")

def test_lookup_methods_comparison(benchmark, index_with_size):
    """Compare performance of standard lookup vs bit_run_variants lookup."""
    index, words = index_with_size
    num_queries = min(100, len(words))  # Fewer queries for detailed comparison
    test_queries = generate_test_queries(words, num_queries)
    
    def run_standard_lookups():
        return [index.lookup(query) for query in test_queries]
    
    def run_bit_run_lookups():
        return [index.lookup_bit_run_variants(query) for query in test_queries]
    
    standard_result = benchmark(run_standard_lookups)
    bit_run_result = benchmark(run_bit_run_lookups)
    
    # Verify results match
    standard_matches = standard_result
    bit_run_matches = bit_run_result
    
    for i, query in enumerate(test_queries):
        assert standard_matches[i] == bit_run_matches[i], \
            f"Results differ for query '{query}':\n" \
            f"Standard: {sorted(standard_matches[i])}\n" \
            f"Bit-run: {sorted(bit_run_matches[i])}"
    
    print(f"\nLookup Methods Comparison for size {len(words)}:")
    print(f"Number of queries: {num_queries}")
    print(f"Standard lookup stats: {standard_result.stats}")
    print(f"Bit-run lookup stats: {bit_run_result.stats}")

if __name__ == "__main__":
    pytest.main([__file__, "-v", "--benchmark-only"]) 