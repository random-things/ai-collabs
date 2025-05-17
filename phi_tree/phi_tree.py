import math, random
from typing import List, Tuple

phi = (1 + 5 ** 0.5) / 2

def fibonacci_up_to(n: int) -> List[int]:
    fibs = [1, 2]
    while fibs[-1] <= n:
        fibs.append(fibs[-1] + fibs[-2])
    return fibs

def zeckendorf_bits(n: int) -> List[int]:
    """
    Return Zeckendorf digits (MSB‑>LSB) using the standard greedy algorithm.
    Guarantees no consecutive 1s.
    """
    # Build Fibonacci numbers >=1,2,3,...
    fibs = [1, 2]  # F1=1, F2=2 in Zeckendorf indexing
    while fibs[-1] <= n:
        fibs.append(fibs[-1] + fibs[-2])
    fibs.pop()  # remove the last that exceeded n
    bits = []
    remaining = n
    for f in reversed(fibs):
        if f <= remaining:
            bits.append(1)
            remaining -= f
        else:
            bits.append(0)
    # assert uniqueness and correctness outside    
    return bits

class Node:
    __slots__ = ("value", "left", "right")
    def __init__(self):
        self.value = None
        self.left = None
        self.right = None

class PhiTree:
    def __init__(self):
        self.root = Node()
        self.max_depth = 0
    def _walk(self, key: int, create=False) -> Tuple[Node, int]:
        node = self.root
        depth = 0
        for bit in zeckendorf_bits(key):
            nxt = node.right if bit else node.left
            if nxt is None:
                if not create:
                    return None, depth
                nxt = Node()
                if bit:
                    node.right = nxt
                else:
                    node.left = nxt
            node = nxt
            depth += 1
        return node, depth
    def insert(self, key: int, value):
        node, depth = self._walk(key, create=True)
        node.value = value
        if depth > self.max_depth:
            self.max_depth = depth
    def get(self, key: int):
        node, _ = self._walk(key, create=False)
        return None if node is None else node.value

def test_zeckendorf_properties(N: int):
    seen = set()
    violations = {'duplicate': 0, 'consecutive_1s': 0, 'length_bound': 0}
    worst_delta = 0
    for n in range(1, N+1):
        bits = zeckendorf_bits(n)
        word = ''.join(map(str, bits))
        if word in seen:
            violations['duplicate'] += 1
        else:
            seen.add(word)
        if '11' in word:
            violations['consecutive_1s'] += 1
        bound = math.ceil(math.log(n, phi))
        if len(bits) > bound:
            violations['length_bound'] += 1
            worst_delta = max(worst_delta, len(bits) - bound)
    return violations, worst_delta

def test_phi_tree(keys: List[int]):
    tree = PhiTree()
    for k in keys:
        tree.insert(k, k*3)
    ok = all(tree.get(k) == k*3 for k in keys)
    predicted = math.ceil(math.log(max(keys), phi))
    return tree.max_depth, predicted, ok

N = 50000
violations, worst_delta = test_zeckendorf_properties(N)
print("Zeckendorf tests for 1..", N)
print("Violations:", violations, "Worst extra length:", worst_delta)

keys = random.sample(range(1, 2_000_000), 40000)
depth, pred, ok = test_phi_tree(keys)
print("\nΦ‑Tree random key test (40k keys up to 2e6)")
print("Observed max depth:", depth, "Predicted bound:", pred)
print("All retrievals correct:", ok)
