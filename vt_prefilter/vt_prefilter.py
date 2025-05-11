import random, itertools

# --- constants ----------------------------------------------------------
ALPHABET = "abcdefghijklmnopqrstuvwxyz"
sigma    = len(ALPHABET)
M        = 67                               # ≥ sigma+1, prime
to_int   = {c: i+1 for i, c in enumerate(ALPHABET)}

# --- VT key -------------------------------------------------------------
def vt_key(word: str):
    n = len(word)
    if n == 0: return (0, 0, 0)
    S = sum((i+1)*to_int[ch] for i, ch in enumerate(word)) % (n+1)
    P = sum(to_int[ch] for ch in word) % M
    return (n, S, P)

# --- enumerate *every* distance‑1 string variant ------------------------
def edits1(w: str):
    splits  = [(w[:i], w[i:]) for i in range(len(w)+1)]
    deletes = [L+R[1:]                 for L, R in splits if R]
    repls   = [L+c+R[1:] for L, R in splits if R for c in ALPHABET if c!=R[0]]
    inserts = [L+c+R     for L, R in splits for c in ALPHABET]
    return set(deletes + repls + inserts)

def edits1_damerau(w: str):
    """all distance‑1 words, incl. adjacent transpositions"""
    base = edits1(w)                         # ins / del / sub (old)
    swaps = {w[:i] + w[i+1] + w[i] + w[i+2:]
             for i in range(len(w)-1)}
    return base | swaps

# --- algebraic variant generator (k=1) ----------------------------------
# --- distance‑1 variants (now with transpositions) ----------------------
def vt_variants(word: str):
    """
    Keys of all words at Damerau‑Levenshtein distance ≤ 1
    (substitution · insertion · deletion · adjacent‑swap).
    """
    n = len(word)
    if n == 0:
        return {(0, 0, 0)}

    c_vals = [to_int[ch] for ch in word]
    total  = sum(c_vals)

    S0 = sum((i + 1) * v for i, v in enumerate(c_vals)) % (n + 1)
    P0 = total % M
    K  = {(n, S0, P0)}                       # distance‑0 key

    # ---------- substitutions ------------------------------------------
    for j, v_old in enumerate(c_vals):
        for v_new in range(1, sigma + 1):
            if v_new == v_old:
                continue
            diff = v_new - v_old
            K.add((n,
                   (S0 + diff * (j + 1)) % (n + 1),
                   (P0 + diff) % M))

    # ---------- deletions ----------------------------------------------
    for j in range(n):
        K.add(vt_key(word[:j] + word[j + 1:]))

    # ---------- insertions ---------------------------------------------
    for j in range(n + 1):
        for x in range(1, sigma + 1):
            K.add(vt_key(word[:j] + ALPHABET[x - 1] + word[j:]))

    # ---------- adjacent transpositions --------------------------------
    # swap indices j and j+1
    if n >= 2:
        for j in range(n - 1):
            swapped = (word[:j] + word[j + 1] + word[j] + word[j + 2:])
            K.add(vt_key(swapped))

    return K

# --- variants within *k* edits (k ≥ 1) ----------------------------------
def vt_variants_k(word: str, k: int):
    """
    Return the set of keys reachable from *word* by ≤ k
    Damerau‑Levenshtein edits.

    Warning: grows exponentially; practical for k ≤ 2.
    """
    seen_words = {word}
    frontier   = {word}
    keys       = set()

    for _ in range(k):
        next_frontier = set()
        for w in frontier:
            for key in vt_variants(w):     # distance‑1 keys of w
                keys.add(key)
            # enumerate actual strings at dist‑1 to feed next layer
            splits  = [(w[:i], w[i:]) for i in range(len(w) + 1)]
            deletes = [L + R[1:] for L, R in splits if R]
            repls   = [L + c + R[1:] for L, R in splits if R
                       for c in ALPHABET if c != R[0]]
            inserts = [L + c + R for L, R in splits for c in ALPHABET]
            transps = [w[:i] + w[i + 1] + w[i] + w[i + 2:]
                       for i in range(len(w) - 1)]
            for v in itertools.chain(deletes, repls, inserts, transps):
                if v not in seen_words:
                    seen_words.add(v)
                    next_frontier.add(v)
        frontier = next_frontier
        if not frontier:
            break

    # finally, include the original word’s key, too
    keys.add(vt_key(word))
    return keys

# --- empirical test -----------------------------------------------------
def bound(n): return sigma*(2*n+1) + n + 1

# OK = True
# for _ in range(200):                        # 200 random words, len 1–10
#     w     = ''.join(random.choice(ALPHABET) for _ in range(random.randint(1,10)))
#     k_alg = vt_variants(w)
#     k_true = {vt_key(v) for v in edits1_damerau(w)} | {vt_key(w)}

#     if k_true != k_alg:
#         print("❌ mismatch on", w)
#         OK = False
#         break

#     assert len(k_alg) <= bound(len(w)), "size bound violated"

# print("All tests passed ✅" if OK else "Test failed")

# # test k = 1  (should still pass)
# for _ in range(200):
#     w = ''.join(random.choice(ALPHABET) for _ in range(random.randint(1,10)))
#     assert {vt_key(v) for v in edits1(w)} | {vt_key(w)} | \
#            {vt_key(w[:i]+w[i+1]+w[i]+w[i+2:])         # swaps
#             for i in range(len(w)-1)} <= vt_variants(w)

# # test k = 2
# for _ in range(100):
#     w = ''.join(random.choice(ALPHABET) for _ in range(random.randint(1,8)))
#     keys_alg = vt_variants_k(w, 2)
#     # brute‑force all distance‑2 strings:
#     true_set = {vt_key(x)
#                 for v1 in edits1(w)
#                 for x  in edits1(v1)} | {vt_key(w)}
#     assert true_set <= keys_alg

# print("All expanded‑k tests passed ✅")