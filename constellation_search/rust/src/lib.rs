// src/lib.rs
//! Single‐threaded, zero‐allocation Rust CSI using one‐pass open‐address buckets.

use std::slice;
use std::os::raw::c_uchar;

/// Opaque handle passed over FFI
#[repr(C)]
pub struct CSIHandle {
    inner: Box<CSIIndex>,
}

/// A single gap’s open‐address bucket table
struct FlatIndex {
    table_size: usize,
    keys:       Vec<u64>,   // len = table_size, 0 means empty
    starts:     Vec<usize>, // len = table_size, prefix‐sum start offsets
    lens:       Vec<usize>, // len = table_size, count of entries
    offs:       Vec<usize>, // all offsets, grouped by bucket
}

/// Main index
struct CSIIndex {
    k:    usize,
    gaps: Vec<usize>,
    flat: Vec<FlatIndex>,
    text: Vec<u8>,
    pw:   Vec<u64>, // rolling‐hash powers
}

const BASE_P: u64 = 1315423911;

impl CSIIndex {
    #[inline(always)]
    fn build_from(data: &[u8]) -> Self {
        let n = data.len();

        // 1) entropy → choose k & gaps
        let entropy = compute_entropy(data);
        let (k, mut gaps) = if entropy < 3.5 {
            (6, vec![8,16,32,64])
        } else if entropy < 4.5 {
            (5, vec![6,12,24,48])
        } else {
            (4, vec![4,8,16,32])
        };
        gaps.sort_unstable();

        // 2) prefix‐hash & powers
        let mut ph = Vec::with_capacity(n+1);
        let mut pw = Vec::with_capacity(n+1);
        ph.push(0u64); pw.push(1u64);
        for i in 0..n {
            ph.push(ph[i].wrapping_mul(BASE_P).wrapping_add(data[i] as u64));
            pw.push(pw[i].wrapping_mul(BASE_P));
        }

        // 3) one‐pass open‐address bucket for each gap
        let mut flat = Vec::with_capacity(gaps.len());
        for &d in &gaps {
            flat.push(FlatIndex::new(&ph, &pw, n, k, d));
        }

        CSIIndex {
            k,
            gaps,
            flat,
            text: data.to_vec(),
            pw,
        }
    }

    fn search(&self, pat: &[u8]) -> Vec<usize> {
        let m = pat.len();
        let minlen = self.k + self.gaps[0] + self.k;
        if m < minlen {
            return Vec::new();
        }
        // build pattern hash
        let mut php = Vec::with_capacity(m+1);
        php.push(0u64);
        for i in 0..m {
            php.push(php[i].wrapping_mul(BASE_P).wrapping_add(pat[i] as u64));
        }
        // collect postings
        let mut lists = Vec::with_capacity(self.gaps.len());
        for (idx, &d) in self.gaps.iter().enumerate() {
            if d + self.k <= m {
                let h1 = php[self.k]
                    .wrapping_sub(php[0].wrapping_mul(self.pw[self.k]));
                let h2 = php[d + self.k]
                    .wrapping_sub(php[d].wrapping_mul(self.pw[self.k]));
                let key = combine_hashes(h1, h2, d as u64);

                // open‐address lookup
                let fi = &self.flat[idx];
                let mut slot = (key as usize) & (fi.table_size - 1);
                loop {
                    let k2 = unsafe { *fi.keys.get_unchecked(slot) };
                    if k2 == 0 {
                        return Vec::new();
                    }
                    if k2 == key {
                        let start = unsafe { *fi.starts.get_unchecked(slot) };
                        let len   = unsafe { *fi.lens.get_unchecked(slot) };
                        lists.push(&fi.offs[start..start+len]);
                        break;
                    }
                    slot = (slot + 1) & (fi.table_size - 1);
                }
            }
        }
        if lists.is_empty() {
            return Vec::new();
        }
        // intersect two‐pointer
        let mut acc = lists[0].to_vec();
        for lst in &lists[1..] {
            acc = intersect_sorted(&acc, lst);
            if acc.is_empty() {
                return Vec::new();
            }
        }
        // verify
        let dlen = m;
        acc.into_iter()
            .filter(|&off| off + dlen <= self.text.len() &&
                unsafe { &*self.text.get_unchecked(off..off+dlen) } == pat)
            .collect()
    }
}

impl FlatIndex {
    #[inline(always)]
    fn new(ph: &[u64], pw: &[u64], n: usize, k: usize, d: usize) -> Self {
        // number of entries
        let m = if n >= k + d { n - (k + d) + 1 } else { 0 };
        // table size = next power of two ≥ 2*m, min 16
        let ts = (m * 2).next_power_of_two().max(16);
        // arrays
        let mut keys   = vec![0u64; ts];
        let mut counts = vec![0usize; ts];

        // Pass 1: count per-key
        for i in 0..m {
            let h1 = unsafe {
                ph.get_unchecked(i+k).wrapping_sub(
                    ph.get_unchecked(i).wrapping_mul(*pw.get_unchecked(k))
                )
            };
            let j = i + d;
            let h2 = unsafe {
                ph.get_unchecked(j+k).wrapping_sub(
                    ph.get_unchecked(j).wrapping_mul(*pw.get_unchecked(k))
                )
            };
            let key = combine_hashes(h1, h2, d as u64);
            let mut slot = (key as usize) & (ts - 1);
            loop {
                let k2 = unsafe { *keys.get_unchecked(slot) };
                if k2 == 0 {
                    // claim empty
                    unsafe { *keys.get_unchecked_mut(slot) = key; }
                    unsafe { *counts.get_unchecked_mut(slot) = 1; }
                    break;
                }
                if k2 == key {
                    unsafe { *counts.get_unchecked_mut(slot) += 1; }
                    break;
                }
                slot = (slot + 1) & (ts - 1);
            }
        }

        // prefix-sum to get starts
        let mut starts = vec![0usize; ts];
        let mut sum = 0;
        for idx in 0..ts {
            if unsafe { *keys.get_unchecked(idx) } != 0 {
                unsafe { *starts.get_unchecked_mut(idx) = sum; }
                sum += unsafe { *counts.get_unchecked(idx) };
            }
        }

        // reset counts → use as write‐idx
        for idx in 0..ts {
            if unsafe { *keys.get_unchecked(idx) } != 0 {
                unsafe { *counts.get_unchecked_mut(idx) = 0; }
            }
        }

        let mut offs = vec![0usize; sum];
        // Pass 2: fill offs
        for i in 0..m {
            let h1 = unsafe {
                ph.get_unchecked(i+k).wrapping_sub(
                    ph.get_unchecked(i).wrapping_mul(*pw.get_unchecked(k))
                )
            };
            let j = i + d;
            let h2 = unsafe {
                ph.get_unchecked(j+k).wrapping_sub(
                    ph.get_unchecked(j).wrapping_mul(*pw.get_unchecked(k))
                )
            };
            let key = combine_hashes(h1, h2, d as u64);
            let mut slot = (key as usize) & (ts - 1);
            loop {
                if unsafe { *keys.get_unchecked(slot) } == key {
                    let start = unsafe { *starts.get_unchecked(slot) };
                    let cnt   = unsafe { *counts.get_unchecked(slot) };
                    unsafe { *offs.get_unchecked_mut(start + cnt) = i; }
                    unsafe { *counts.get_unchecked_mut(slot) = cnt + 1; }
                    break;
                }
                slot = (slot + 1) & (ts - 1);
            }
        }

        FlatIndex { table_size: ts, keys, starts, lens: counts, offs }
    }
}

#[inline(always)]
fn intersect_sorted(a: &[usize], b: &[usize]) -> Vec<usize> {
    let mut res = Vec::with_capacity(a.len().min(b.len()));
    let (mut i, mut j) = (0, 0);
    while i < a.len() && j < b.len() {
        match unsafe { *a.get_unchecked(i) }.cmp(&unsafe { *b.get_unchecked(j) }) {
            std::cmp::Ordering::Less    => i += 1,
            std::cmp::Ordering::Greater => j += 1,
            std::cmp::Ordering::Equal   => {
                res.push(unsafe { *a.get_unchecked(i) });
                i += 1; j += 1;
            }
        }
    }
    res
}

#[inline(always)]
fn combine_hashes(h1: u64, h2: u64, d: u64) -> u64 {
    let mut x = h1 ^ (h2 << 1) ^ (d << 2);
    x ^= x >> 33;
    x = x.wrapping_mul(0xff51afd7ed558ccd);
    x ^= x >> 33;
    x = x.wrapping_mul(0xc4ceb9fe1a85ec53);
    x ^= x >> 33;
    x
}

fn compute_entropy(data: &[u8]) -> f64 {
    let mut freq = [0usize; 256];
    for &b in data {
        freq[b as usize] += 1;
    }
    let n = data.len() as f64;
    let mut ent = 0.0;
    for &c in &freq {
        if c > 0 {
            let p = (c as f64) / n;
            ent -= p * p.log2();
        }
    }
    ent
}

// FFI exports

#[unsafe(no_mangle)]
pub extern "C" fn csi_new(data: *const c_uchar, len: usize) -> *mut CSIHandle {
    if data.is_null() || len == 0 { return std::ptr::null_mut() }
    let slice = unsafe { slice::from_raw_parts(data, len) };
    let idx = CSIIndex::build_from(slice);
    let handle = CSIHandle { inner: Box::new(idx) };
    Box::into_raw(Box::new(handle))
}

#[unsafe(no_mangle)]
pub extern "C" fn csi_free(handle: *mut CSIHandle) {
    if !handle.is_null() {
        unsafe { let _ = Box::from_raw(handle); }
    }
}

#[unsafe(no_mangle)]
pub extern "C" fn csi_search(
    handle: *const CSIHandle,
    pat:    *const c_uchar,
    pat_len: usize,
    out:    *mut usize,
    max_out: usize,
) -> usize {
    if handle.is_null() || pat.is_null() { return 0 }
    let idx = unsafe { &*((*handle).inner) };
    let pat_slice = unsafe { slice::from_raw_parts(pat, pat_len) };
    let matches = idx.search(pat_slice);
    let n = matches.len().min(max_out);
    let out_slice = unsafe { slice::from_raw_parts_mut(out, n) };
    for i in 0..n {
        unsafe { *out_slice.get_unchecked_mut(i) = matches[i]; }
    }
    n
}
