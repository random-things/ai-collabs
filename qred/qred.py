from math import ceil, gcd

def qred(a: int, b: int) -> list[int]:
    """
    Perform QRED (Quadratic-Residue Egyptian Decomposition)
    on a rational number a/b (a < b, a and b are coprime).

    Returns a list of denominators of the unit fractions such that:
        a/b = sum(1/d_i for d_i in result)
    """
    if a <= 0 or b <= 0 or a >= b:
        raise ValueError("Require 0 < a < b for proper fraction.")
    if gcd(a, b) != 1:
        raise ValueError("Fraction must be in reduced form (coprime numerator and denominator).")

    terms = []

    while a > 1:
        d0 = ceil(b / a)
        found = False
        for k in range(a):
            d = d0 + k
            r = a * d - b
            if 0 < r < b and b % r == 0:
                terms.append(d)
                b = b // r * d  # new denominator
                a = 1           # the remainder is 1/b now
                terms.append(b)
                found = True
                break
        if not found:
            raise RuntimeError("QRED failed to find valid decomposition.")
    
    if a == 1:
        terms.append(b)

    return terms


if __name__ == "__main__":
    a, b = 47, 112
    decomposition = qred(a, b)
    print(f"{a}/{b} = " + " + ".join(f"1/{d}" for d in decomposition))
