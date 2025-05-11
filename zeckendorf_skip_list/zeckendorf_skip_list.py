"""A proof-of-concept implementation of a Zeckendorf Skip List.

This module implements a static Zeckendorf Skip List (ZSL) using Fibonacci numbers
to determine skip distances. The implementation provides O(log n) search time
without the need for randomization or rebalancing.

The key idea is to use the Zeckendorf representation of each rank to determine
skip pointers, where each number can be represented as a sum of distinct Fibonacci
numbers. This creates a deterministic skip list structure that maintains good
search performance while being simpler to implement than traditional skip lists.

Example:
    >>> keys = list(range(1000))
    >>> zsl = StaticZSL(keys)
    >>> zsl.search(42)
    True
    >>> zsl.search(1001)
    False
"""

import bisect
import random
import time


def zeckendorf_terms(target_number: int, fibonacci_numbers: list[int]) -> list[int]:
    """Computes the Zeckendorf representation of a number using given Fibonacci numbers.

    The Zeckendorf representation expresses a number as a sum of distinct Fibonacci
    numbers, where no two consecutive Fibonacci numbers are used. This function
    returns the indices of the Fibonacci numbers used in the representation.

    Args:
        target_number: The number to represent in Zeckendorf form.
        fibonacci_numbers: A list of Fibonacci numbers in ascending order.

    Returns:
        A list of indices (1-based) into the fibonacci_numbers list, representing the
        Fibonacci numbers that sum to target_number. The indices are in descending order.

    Example:
        >>> fibonacci_numbers = [1, 2, 3, 5, 8, 13]
        >>> zeckendorf_terms(7, fibonacci_numbers)
        [4, 2]  # 7 = 5 + 2
    """
    zeckendorf_indices = []
    remaining_sum = target_number
    current_fib_index = len(fibonacci_numbers) - 1

    while remaining_sum > 0 and current_fib_index >= 0:
        current_fib = fibonacci_numbers[current_fib_index]
        if current_fib <= remaining_sum:
            # Add 1 to convert to 1-based level indexing
            zeckendorf_indices.append(current_fib_index + 1)
            remaining_sum -= current_fib
            # Skip consecutive Fibonacci numbers
            current_fib_index -= 2
        else:
            current_fib_index -= 1

    return zeckendorf_indices


class StaticZSL:
    """A static Zeckendorf Skip List implementation.

    This class implements a skip list where skip distances are determined by
    Fibonacci numbers, using the Zeckendorf representation of each rank to
    establish skip pointers. The structure is static (no insertions/deletions)
    and provides O(log n) search time.

    The skip list is constructed from a sorted list of keys, and uses a
    dictionary-based representation for skip pointers at each level.

    Attributes:
        keys: The sorted list of keys stored in the skip list.
        num_keys: The number of keys in the skip list.
        fibonacci_numbers: The list of Fibonacci numbers used for skip distances.
        max_level: The maximum level of skip pointers.
        next: A list of dictionaries, where next[i][L] gives the node at
            level L that follows node i.

    Example:
        >>> keys = [1, 3, 5, 7, 9]
        >>> zsl = StaticZSL(keys)
        >>> zsl.search(5)
        True
        >>> zsl.search(6)
        False
    """

    def __init__(self, sorted_keys: list[int]):
        """Initializes a StaticZSL from a sorted list of keys.

        Args:
            sorted_keys: A sorted list of unique integers to store in the skip list.

        Note:
            The input keys should be sorted in ascending order. The constructor
            will not verify this, and incorrect results may occur if the input
            is not properly sorted.
        """
        self.keys = sorted_keys
        self.num_keys = len(sorted_keys)

        # Precompute Fibonacci numbers up to num_keys
        fibonacci_numbers = [1, 2]
        while fibonacci_numbers[-1] <= self.num_keys:
            next_fib = fibonacci_numbers[-1] + fibonacci_numbers[-2]
            fibonacci_numbers.append(next_fib)
        
        # Remove the last Fibonacci number if it exceeds num_keys
        if fibonacci_numbers[-1] > self.num_keys:
            fibonacci_numbers.pop()
        
        self.fibonacci_numbers = fibonacci_numbers
        self.max_level = len(fibonacci_numbers)

        # Initialize skip pointers: next[rank][level] gives the next node at that level
        # Index 0 represents the head node
        self.next = [dict() for _ in range(self.num_keys + 1)]

        # Build skip pointers for each rank
        for current_rank in range(1, self.num_keys + 1):
            skip_levels = zeckendorf_terms(current_rank, self.fibonacci_numbers)
            for level in skip_levels:
                skip_distance = self.fibonacci_numbers[level - 1]
                target_rank = current_rank + skip_distance
                if target_rank <= self.num_keys:
                    self.next[current_rank][level] = target_rank

        # Initialize head pointers (from rank 0)
        # Head has pointers at all levels for initial jumps
        for level in range(1, self.max_level + 1):
            skip_distance = self.fibonacci_numbers[level - 1]
            if skip_distance <= self.num_keys:
                self.next[0][level] = skip_distance

    def search(self, target_key: int) -> bool:
        """Searches for a key in the skip list.

        The search uses the skip pointers to quickly navigate through the list,
        following the highest-level pointers possible at each step, then
        stepping at the base level to find the exact position.

        Args:
            target_key: The integer key to search for.

        Returns:
            True if the key is found in the skip list, False otherwise.

        Example:
            >>> zsl = StaticZSL([1, 3, 5, 7, 9])
            >>> zsl.search(5)
            True
            >>> zsl.search(6)
            False
        """
        current_node = 0  # Start at head node

        # Navigate using highest possible level pointers
        for level in range(self.max_level, 0, -1):
            while (level in self.next[current_node] and 
                   self.keys[self.next[current_node][level] - 1] <= target_key):
                current_node = self.next[current_node][level]

        # Take final step at base level if possible
        if 1 in self.next[current_node]:
            current_node = self.next[current_node][1]

        # Check if we found the target key
        return current_node != 0 and self.keys[current_node - 1] == target_key


# Benchmarking code
def run_benchmarks() -> list[dict]:
    """Runs a comparative benchmark between ZSL and list+bisect search.

    This function benchmarks the search performance of the StaticZSL against
    Python's built-in bisect module on sorted lists. It tests various sizes
    and reports the relative performance.

    The benchmark:
    1. Tests sizes from 1,000 to 20,000 elements
    2. Uses 10,000 random queries for each size
    3. Tests both present and absent keys
    4. Reports timing and speedup ratios

    Returns:
        A list of dictionaries containing benchmark results for each size.
    """
    test_sizes = [1000, 5000, 10000, 20000]
    benchmark_results = []

    for num_keys in test_sizes:
        # Prepare test data
        sorted_keys = list(range(num_keys))
        zsl = StaticZSL(sorted_keys)
        bisect_list = sorted_keys.copy()

        # Generate test queries
        num_queries = 10000
        present_keys = random.choices(sorted_keys, k=num_queries)
        absent_keys = random.choices(range(num_keys, 2 * num_keys), k=num_queries)

        # Benchmark ZSL search
        zsl_start_time = time.perf_counter()
        for key in present_keys:
            zsl.search(key)
        for key in absent_keys:
            zsl.search(key)
        zsl_time = time.perf_counter() - zsl_start_time

        # Benchmark list+bisect search
        bisect_start_time = time.perf_counter()
        for key in present_keys:
            insert_pos = bisect.bisect_left(bisect_list, key)
            _ = (insert_pos < len(bisect_list) and bisect_list[insert_pos] == key)
        for key in absent_keys:
            insert_pos = bisect.bisect_left(bisect_list, key)
            _ = (insert_pos < len(bisect_list) and bisect_list[insert_pos] == key)
        bisect_time = time.perf_counter() - bisect_start_time

        benchmark_results.append({
            'n': num_keys,
            'ZSL_search_time_s': zsl_time,
            'List+bisect_search_time_s': bisect_time
        })

    return benchmark_results


def print_benchmark_results(results: list[dict]) -> None:
    """Prints benchmark results in a formatted table.

    Args:
        results: A list of dictionaries containing benchmark results.
    """
    print("\nBenchmark Results:")
    print("-" * 75)
    print(f"{'Size (n)':>10} {'ZSL Time (s)':>15} {'List+bisect Time (s)':>20} {'Speedup':>10}")
    print("-" * 75)
    
    for result in results:
        speedup = result['List+bisect_search_time_s'] / result['ZSL_search_time_s']
        print(f"{result['n']:10d} {result['ZSL_search_time_s']:15.4f} "
              f"{result['List+bisect_search_time_s']:20.4f} {speedup:10.2f}x")
    
    print("-" * 75)


if __name__ == '__main__':
    benchmark_results = run_benchmarks()
    print_benchmark_results(benchmark_results)
