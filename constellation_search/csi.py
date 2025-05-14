import random, string, time
from collections import defaultdict, Counter

def build_csi_index(text, k=4, gaps=(4, 8, 16, 32)):
    """Return CSI index: dict key -> list of start positions."""
    index = defaultdict(list)
    n = len(text)
    for i in range(n - k + 1):
        a1 = text[i:i + k]
        for d in gaps:
            j = i + d
            if j + k <= n:
                a2 = text[j:j + k]
                key = (a1, a2, d)
                index[key].append(i)
    return index

def csi_search(text, index, pattern, k=4, gaps=(4, 8, 16, 32)):
    """Return sorted list of start positions of exact matches using CSI."""
    # build anchor pairs for pattern
    m = len(pattern)
    pairs = []
    for d in gaps:
        if d + k <= m:
            a1 = pattern[:k]
            a2 = pattern[d:d + k]
            key = (a1, a2, d)
            if key in index:
                pairs.append((key, index[key]))
            else:
                # one anchor pair missing => no match
                return []
    if not pairs:
        # Pattern too short for any pair
        return []
    # choose pair with smallest posting list to start
    pairs.sort(key=lambda x: len(x[1]))
    candidate_set = set(pairs[0][1])
    for _, plist in pairs[1:]:
        candidate_set &= set(plist)
        if not candidate_set:
            break
    # verification
    verified = [pos for pos in candidate_set if text.startswith(pattern, pos)]
    verified.sort()
    return verified

def brute_force_search(text, pattern):
    """Return all occurrences using Python's find."""
    res = []
    start = text.find(pattern)
    while start != -1:
        res.append(start)
        start = text.find(pattern, start + 1)
    return res

def simulate(n=100_000, k=4, gaps=(4,8,16,32), num_queries=100, pattern_len=50, seed=42):
    random.seed(seed)
    alphabet = string.ascii_lowercase
    sigma = len(alphabet)
    text = ''.join(random.choices(alphabet, k=n))
    index = build_csi_index(text, k, gaps)
    
    # statistics
    candidate_sizes = []
    t_csi = 0.0
    t_brute = 0.0
    for _ in range(num_queries):
        start = random.randrange(0, n - pattern_len)
        pattern = text[start:start + pattern_len]
        # brute
        t0 = time.perf_counter()
        brute_res = brute_force_search(text, pattern)
        t_brute += time.perf_counter() - t0
        # csi
        t0 = time.perf_counter()
        csi_res = csi_search(text, index, pattern, k, gaps)
        t_csi += time.perf_counter() - t0
        assert csi_res == brute_res
        candidate_sizes.append(len(csi_res))
    exp_candidates = n / (sigma ** (2*k))
    return {
        "n": n,
        "k": k,
        "gaps": gaps,
        "sigma": sigma,
        "index_buckets": len(index),
        "avg_candidates": sum(candidate_sizes) / len(candidate_sizes),
        "max_candidates": max(candidate_sizes),
        "exp_candidates": exp_candidates,
        "csi_time_ms": t_csi * 1000,
        "brute_time_ms": t_brute * 1000
    }

def simulate_worst_case(n=20_000, k=4, gaps=(4,8,16,32), num_queries=10, pattern_len=50):
    text = 'a' * n
    index = build_csi_index(text, k, gaps)
    candidate_sizes = []
    for i in range(num_queries):
        start = 0  # all patterns identical
        pattern = text[start:start+pattern_len]
        csi_res = csi_search(text, index, pattern, k, gaps)
        candidate_sizes.append(len(csi_res))
    return {
        "n": n,
        "all_A_index_buckets": len(index),
        "all_A_avg_candidates": sum(candidate_sizes)/len(candidate_sizes),
        "all_A_max_candidates": max(candidate_sizes)
    }

results_random = simulate()
results_worst = simulate_worst_case()

print("=== Random uniform text ===")
for k,v in results_random.items():
    print(f"{k}: {v}")
print("\n=== Worst‑case repetitive text (all 'a's) ===")
for k,v in results_worst.items():
    print(f"{k}: {v}")
