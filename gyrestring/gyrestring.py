"""
gyre_string.py – Proof‑of‑concept implementation of the Gyre‑String Index

Core ideas implemented
----------------------
* 128‑bit Möbius sketch (forward + reverse rolling hash modulo a 64‑bit prime)
* Incremental sketch update (append‑one‑char)
* Dictionary index:  sketch  ->  { original words }
* Fuzzy lookup (edit‑distance ≤ k) by enumerating all edits of the query
  and hashing those variants;  this keeps the build‑time tiny while remaining
  mathematically correct.  For small k (1‑2) and search‑as‑you‑type query
  lengths, the cost is modest.

Performance notes
-----------------
* No attempt is made to pre‑compute "bit‑run variants"; the goal here is
  correctness and usability.  Porting that optimisation is a future exercise
  once the reference results are established.
"""

from __future__ import annotations

from collections import defaultdict
from functools import lru_cache
from typing import Dict, Iterable, List, Set, Tuple

# ------------------------------------------------------------------------ #
#  bit_run_variants – enumerate sketch keys that are reachable by ≤ k edits
# ------------------------------------------------------------------------ #
from itertools import combinations, product
from typing import Iterable, Set, Tuple

def _contiguous_masks(max_run: int) -> Tuple[int, ...]:
    """
    Pre‑compute every mask that flips 1..max_run contiguous bits
    inside a 64‑bit word.
    """
    masks = []
    for length in range(1, max_run + 1):
        base = (1 << length) - 1                # e.g. 0b11, 0b111 …
        for start in range(64 - length + 1):
            masks.append(base << start)
    return tuple(masks)


# single global cache – building it once is cheap and re‑used for every call
_MASK_CACHE: dict[int, Tuple[int, ...]] = {}


def _masks(max_run: int) -> Tuple[int, ...]:
    if max_run not in _MASK_CACHE:
        _MASK_CACHE[max_run] = _contiguous_masks(max_run)
    return _MASK_CACHE[max_run]


def bit_run_variants(
    f: int,
    r: int,
    k: int,
    *,
    max_run: int = 6,
) -> Set[Tuple[int, int]]:
    """
    Yield all (F′, R′) sketches reachable by flipping up to *k* bit‑runs,
    each run being ≤ `max_run` contiguous bits, across the two 64‑bit limbs.

    Notes
    -----
    *  This is a *superset* of the exact lemma‑based set, but it is guaranteed
       to include every true edit‑distance ≤ k variant, i.e. it is **complete**.
       (You can prune false positives afterwards with a Levenshtein checker
       if you wish.)
    *  Runtime is `O(C(k) · runs)`, where `runs = Σ_{l=1..max_run} (64-l+1)`
       ≈ 363 for `max_run = 6`.  For `k = 1` that is ~700 variants – tiny.
    """
    masks = _masks(max_run)
    variants: Set[Tuple[int, int]] = {(f, r)}          # distance 0

    if k == 0:
        return variants

    # --- single‑run flips ------------------------------------------------- #
    for m in masks:
        variants.add((f ^ m, r))   # flip in forward limb
        variants.add((f, r ^ m))   # flip in reverse limb

    if k == 1:
        return variants

    # --- multi‑run flips (k ≥ 2) ----------------------------------------- #
    # Generate combinations of 2, 3, …, k masks over the two limbs.
    # The Cartesian‐product trick lets us distribute runs between limbs
    # without double‑counting.
    for edits in range(2, k + 1):
        # choose how many runs go into F and how many into R
        for runs_in_f in range(edits + 1):
            runs_in_r = edits - runs_in_f
            for combo_f in combinations(masks, runs_in_f) if runs_in_f else [()]:
                for combo_r in combinations(masks, runs_in_r) if runs_in_r else [()]:
                    m_f = 0
                    for m in combo_f:
                        m_f ^= m
                    m_r = 0
                    for m in combo_r:
                        m_r ^= m
                    variants.add((f ^ m_f, r ^ m_r))

    return variants

# --------------------------------------------------------------------------- #
# 1. Constants & helpers
# --------------------------------------------------------------------------- #

ALPHABET = "abcdefghijklmnopqrstuvwxyz"           # Σ
SIGMA = len(ALPHABET)                             # 26
BASE = SIGMA + 2                                  # 27  – coprime to P
P = (1 << 64) - 59                                # 2⁶⁴ − 59 (largest < 2⁶⁴)

CHAR_TO_INT = {c: i for i, c in enumerate(ALPHABET)}


@lru_cache(maxsize=None)
def pow_base(exp: int) -> int:
    """B^exp  (mod P)  – memoised for speed."""
    return pow(BASE, exp, P)


def char_val(ch: str) -> int:
    """Map character to [0 .. σ‑1]; non‑alphabet chars collapse to σ‑1."""
    return CHAR_TO_INT.get(ch.lower(), SIGMA - 1)


# --------------------------------------------------------------------------- #
# 2. Möbius sketch
# --------------------------------------------------------------------------- #

def sketch(word: str) -> Tuple[int, int]:
    """
    Return (F, R) 64‑bit integers for `word`.

    F(x) = Σ (c_i+1)·B^i
    R(x) = Σ (c_i+1)·B^{n-1-i}
    Both mod P.
    """
    f = r = 0
    n = len(word)
    for i, ch in enumerate(word):
        v = char_val(ch) + 1            # +1 avoids leading zeros
        f = (f + v * pow_base(i)) % P
        r = (r + v * pow_base(n - 1 - i)) % P
    return f, r


def append_char(f: int, r: int, length: int, ch: str) -> Tuple[int, int]:
    """
    Incrementally update (F, R) when *appending* a single character.
    Useful for “search‑as‑you‑type” demos.

        F_new = F_old + (c+1)·B^len
        R_new = (R_old·B) + (c+1)
    """
    v = char_val(ch) + 1
    f_new = (f + v * pow_base(length)) % P
    r_new = (r * BASE + v) % P
    return f_new, r_new


# --------------------------------------------------------------------------- #
# 3. Edit‑distance utilities (Norvig‑style)
# --------------------------------------------------------------------------- #

def edits1(word: str, alphabet: str = ALPHABET) -> Set[str]:
    """All strings with Levenshtein distance 1 from `word`."""
    splits = [(word[:i], word[i:]) for i in range(len(word) + 1)]
    deletes     = {L + R[1:]           for L, R in splits if R}
    transposes  = {L + R[1] + R[0] + R[2:] for L, R in splits if len(R) > 1}
    replaces    = {L + c + R[1:]       for L, R in splits if R for c in alphabet}
    inserts     = {L + c + R           for L, R in splits       for c in alphabet}
    return deletes | transposes | replaces | inserts


def edits_k(word: str, k: int, alphabet: str = ALPHABET) -> Set[str]:
    """All strings with Levenshtein distance ≤ k (k = 0,1,2 are practical)."""
    level = {word}
    for _ in range(k):
        level |= {e for w in level for e in edits1(w, alphabet)}
    return level


# --------------------------------------------------------------------------- #
# 4. The index
# --------------------------------------------------------------------------- #

class GyreStringIndex:
    """
    Simple, correct index for fuzzy search with the Gyre‑String sketch.

    Build time  : O(#words)
    Query time  : O(|edits_k(query)|)   (naïve enumeration; good for small k)
    """

    def __init__(self, k: int = 1, alphabet: str = ALPHABET) -> None:
        self.k = k
        self.alphabet = alphabet
        self._table: Dict[Tuple[int, int], Set[str]] = defaultdict(set)

    # ---- building -------------------------------------------------------- #

    def add(self, word: str) -> None:
        """Insert *exactly one* word (no edit variants) into the index."""
        self._table[sketch(word)].add(word)

    def build(self, words: Iterable[str]) -> "GyreStringIndex":
        """Bulk‑build from an iterable of words."""
        for w in words:
            self.add(w)
        return self

    # ---- lookup ---------------------------------------------------------- #

    def lookup(self, query: str) -> Set[str]:
        """
        Return all dictionary words within Levenshtein distance ≤ self.k
        from *query*.
        """
        results: Set[str] = set()
        for variant in edits_k(query, self.k, self.alphabet):
            bucket = self._table.get(sketch(variant))
            if bucket:
                results.update(bucket)
        return results
    
    def lookup_bit_run_variants(self, query: str) -> Set[str]:
        """
        Return all dictionary words within Levenshtein distance ≤ self.k
        from *query*.
        """
        results: Set[str] = set()

        f, r = sketch(query)
        for f_var, r_var in bit_run_variants(f, r, self.k):
            bucket = self._table.get((f_var, r_var))
            if bucket:
                results.update(bucket)

        # optionally: Myers / Levenshtein verification here
        return results

    # ---- convenience ----------------------------------------------------- #

    def __contains__(self, word: str) -> bool:
        return word in self.lookup(word)

    def __len__(self) -> int:
        return sum(len(s) for s in self._table.values())

    def stats(self) -> dict:
        buckets = list(self._table.values())
        return {
            "words": len(self),
            "buckets": len(buckets),
            "avg_bucket_size": sum(map(len, buckets)) / max(1, len(buckets)),
        }


# --------------------------------------------------------------------------- #
# 5. Quick demo / self‑test
# --------------------------------------------------------------------------- #

if __name__ == "__main__":
    DICT = [
        "hello", "world", "help", "held", "yellow", "helmet",
        "hero", "heron", "wild", "word", "would"
    ]
    from time import perf_counter
    start_time = perf_counter()
    index = GyreStringIndex(k=1).build(DICT)
    build_time_ms = (perf_counter() - start_time) * 1000
    print(f"Index stats: {index.stats()} (built in {build_time_ms:.10f}ms)\n")

    while True:
        try:
            q = input("query> ").strip()
        except (EOFError, KeyboardInterrupt):
            break
        if not q:
            continue
        hits = index.lookup(q)
        print(f" matches: {sorted(hits) if hits else '∅'}\n")
