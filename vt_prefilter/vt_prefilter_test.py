import pytest
from vt_prefilter import vt_variants, vt_key, ALPHABET

def test_empty_string():
    """Test that empty string returns only the zero key"""
    result = vt_variants("")
    assert result == {(0, 0, 0)}

def test_single_character():
    """Test variants of a single character word"""
    word = "a"
    result = vt_variants(word)
    
    # Calculate expected size by generating all possible variants
    expected_keys = set()
    
    # Original key
    expected_keys.add(vt_key(word))
    
    # Substitutions
    for c in ALPHABET:
        if c != word[0]:
            expected_keys.add(vt_key(c))
    
    # Insertions
    for pos in range(2):  # before or after
        for c in ALPHABET:
            expected_keys.add(vt_key(word[:pos] + c + word[pos:]))
    
    # Deletion
    expected_keys.add(vt_key(""))
    
    assert len(result) == len(expected_keys)
    
    # Verify original key is present
    assert vt_key(word) in result
    
    # Verify empty string key is present
    assert (0, 0, 0) in result

def test_two_characters():
    """Test variants of a two character word"""
    word = "ab"
    result = vt_variants(word)
    
    # Calculate expected size by generating all possible variants
    expected_keys = set()
    
    # Original key
    expected_keys.add(vt_key(word))
    
    # Substitutions
    for pos in range(2):
        for c in ALPHABET:
            if c != word[pos]:
                new_str = word[:pos] + c + word[pos+1:]
                expected_keys.add(vt_key(new_str))
    
    # Insertions
    for pos in range(3):  # before, between, or after
        for c in ALPHABET:
            expected_keys.add(vt_key(word[:pos] + c + word[pos:]))
    
    # Deletions
    expected_keys.add(vt_key(word[0]))  # delete second char
    expected_keys.add(vt_key(word[1]))  # delete first char
    
    # Transposition
    expected_keys.add(vt_key(word[1] + word[0]))
    
    assert len(result) == len(expected_keys)
    
    # Verify original key is present
    assert vt_key(word) in result
    
    # Verify transposition key is present
    assert vt_key("ba") in result

@pytest.mark.parametrize("word", ["a", "ab", "abc", "hello", "world"])
def test_all_variants_valid(word):
    """Test that all generated variants are valid keys"""
    result = vt_variants(word)
    for key in result:
        n, S, P = key
        # Check key constraints
        assert isinstance(n, int)
        assert isinstance(S, int)
        assert isinstance(P, int)
        assert n >= 0
        assert S >= 0
        assert P >= 0
        assert S < n + 1
        assert P < 67  # M = 67

@pytest.mark.parametrize("word", ["a", "ab", "abc", "hello", "world", ""])
def test_variants_include_original(word):
    """Test that variants always include the original word's key"""
    result = vt_variants(word)
    assert vt_key(word) in result

@pytest.mark.parametrize("word", ["a", "ab", "abc", "hello", "world"])
def test_variants_size_bound(word):
    """Test that the number of variants is within theoretical bounds"""
    def bound(n): return len(ALPHABET)*(2*n+1) + n + 1
    
    result = vt_variants(word)
    assert len(result) <= bound(len(word))

def test_specific_variants():
    """Test specific known variants for a simple word"""
    word = "ab"
    result = vt_variants(word)
    
    # Test some specific variants we know should exist
    assert vt_key("a") in result    # deletion
    assert vt_key("b") in result    # deletion
    assert vt_key("ba") in result   # transposition
    assert vt_key("cab") in result  # insertion
    assert vt_key("acb") in result  # insertion
    assert vt_key("abc") in result  # insertion

@pytest.mark.parametrize("word", ["a", "ab", "abc", "hello", "world"])
def test_variants_consistency(word):
    """Test that generating variants multiple times gives same results"""
    result1 = vt_variants(word)
    result2 = vt_variants(word)
    assert result1 == result2 

def test_actual_search():
    M = 67
    words = ["hello", "world", "python", "programming", "language", "help", "held", "helm"]
    query = "helf"
    expected_results = ["hello", "world", "programming"]

    def all_prefixes(word):
        return [word[:i] for i in range(len(word) + 1)]

    # Build the index
    index = {}
    for word in words:
        for prefix in all_prefixes(word):
            index[vt_key(prefix)] = word

    # Get all variants of the query
    query_keys = vt_variants(query)

    # Find matches in the index
    matches = set()
    for key in query_keys:
        if key in index:
            matches.add(index[key])

    # Verify the results
    print(matches)
