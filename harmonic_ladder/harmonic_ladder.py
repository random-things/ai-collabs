import math, random, itertools, collections, statistics, hashlib

# ---------------------- Egyptian fraction (Greedy) ----------------------
def egyptian_greedy(a: int, b: int):
    """Return list of denominators for a/b using the Sylvester greedy algorithm."""
    dens = []
    while a != 0:
        d = math.ceil(b / a)
        dens.append(d)
        a, b = a * d - b, b * d
        g = math.gcd(a, b)
        if g > 1:
            a //= g
            b //= g
    return dens

# ---------------------- Harmonic Ladder simulation ----------------------
class HarmonicLadder:
    """
    Simplistic in-memory simulation:
      - global denominator->value map where collisions are reported
      - keys are 32-bit ints mapped to key/2**32 rational
      - greedy Egyptian decomposition
    """
    B = 2**32
    def __init__(self):
        self.cells = {}            # denom -> (key, value)
        self.collisions = 0

    def _encode(self, key: int):
        dens = egyptian_greedy(key, HarmonicLadder.B)
        return dens

    def insert(self, key: int, value):
        dens = self._encode(key)
        for d in dens:
            if d in self.cells and self.cells[d][0] != key:
                self.collisions += 1
            self.cells[d] = (key, value)

    def retrieve(self, key: int):
        dens = self._encode(key)
        vals = [self.cells.get(d, (None, None))[1] for d in dens]
        # If any missing or mismatched owner, retrieval fails
        if None in vals:
            return None
        owners = {self.cells[d][0] for d in dens}
        if len(owners) == 1 and owners.pop() == key:
            return vals[0]
        return None

# ---------------------- Tests ----------------------
N_KEYS = 2000
random.seed(1)
keys = random.sample(range(1, 2**24), N_KEYS)   # pick 2K random 24-bit ints

ladder = HarmonicLadder()
for k in keys:
    ladder.insert(k, f"val{k}")

# collision statistics
total_cells = len(ladder.cells)
collisions = ladder.collisions

# retrieval accuracy
failures = sum(1 for k in keys if ladder.retrieve(k) is None)

# overlap analysis: average shared denominators between different keys
den_map = collections.defaultdict(set)
for k in keys:
    for d in ladder._encode(k):
        den_map[d].add(k)
shared_counts = [len(s) for s in den_map.values()]
avg_overlap = statistics.mean(shared_counts)
max_overlap = max(shared_counts)

print("Harmonic Ladder empirical test (2000 random keys, greedy Egyptian, B=2^32)")
print(f"Total ladder cells used : {total_cells}")
print(f"Collisions encountered  : {collisions}")
print(f"Retrieval failures      : {failures}")
print(f"Avg keys per denominator: {avg_overlap:.2f}   Max: {max_overlap}")
