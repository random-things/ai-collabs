import random, math, hashlib, collections, statistics, time

# -------------- SparseHashWheel ----------------
class SparseHashWheel:
    def __init__(self, primes):
        self.primes = primes
        self.wheels = [[0]*p for p in primes]
    def _h(self, item:int):
        return int.from_bytes(hashlib.blake2b(item.to_bytes(8,'little'),digest_size=8).digest(),'little')
    def insert(self, item, count=1):
        h = self._h(item)
        for arr,p in zip(self.wheels,self.primes):
            idx=h%p
            arr[idx]=min(arr[idx]+count,2**16-1)
    def query(self,item):
        h=self._h(item)
        return min(arr[h%p] for arr,p in zip(self.wheels,self.primes))
    def estimate_cardinality(self):
        densities=[]
        total=0
        for arr in self.wheels:
            nz=sum(1 for x in arr if x)
            densities.append(nz/len(arr))
            total+=len(arr)
        d=[x for x in densities if x>0]
        if not d: return 0
        harmonic=len(d)/sum(1/x for x in d)
        return -total*math.log(max(1e-12,1-harmonic))

# Parameters
PRIMES=[1009,1013,1019]
INSERTS=50000
UNIQUE=10000
ZIPF_S=1.1
random.seed(0)
population=list(range(UNIQUE))
weights=[1/(i+1)**ZIPF_S for i in range(UNIQUE)]
total=sum(weights)
cum=[]
s=0
for w in weights:
    s+=w
    cum.append(s/total)
stream=[]
for _ in range(INSERTS):
    r=random.random()
    idx=next(i for i,x in enumerate(cum) if x>=r)
    stream.append(idx)

true_freq=collections.Counter(stream)
wheel=SparseHashWheel(PRIMES)

for itm in stream:
    wheel.insert(itm)

sample_keys=random.sample(list(true_freq.keys()),2000)
errors=[abs(wheel.query(k)-true_freq[k]) for k in sample_keys]
avg_err=statistics.mean(errors)
max_err=max(errors)

est_card=wheel.estimate_cardinality()
true_card=len(true_freq)

first=wheel.wheels[0]
expected=INSERTS/len(first)
chi_sq=sum((c-expected)**2/expected for c in first)
print("SparseHashWheel summary")
print("Mean abs freq error:", avg_err, "Max:", max_err)
print("Cardinality estimate:", round(est_card), "True:", true_card)
print("Chi-square load:", round(chi_sq,2))
