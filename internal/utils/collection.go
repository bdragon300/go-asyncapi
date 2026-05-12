package utils

import (
	"cmp"
	"iter"
	"maps"
	"slices"
)

// OrderedKeysIter returns an iterator over the map's keys and values, in ascending order of the keys.
func OrderedKeysIter[K cmp.Ordered, V any](m map[K]V) iter.Seq2[K, V] {
	keys := slices.Sorted(maps.Keys(m))
	return func(yield func(K, V) bool) {
		for _, k := range keys {
			if !yield(k, m[k]) {
				return
			}
		}
	}
}

// ZipLongest2 returns an iterator that zips two slices together, yielding pairs of elements from both slices.
// If one slice is shorter than the other, it yields zero values for the missing elements.
func ZipLongest2[A, B any](a []A, b []B) iter.Seq2[A, B] {
	return func(yield func(A, B) bool) {
		maxLen := max(len(a), len(b))
		for i := 0; i < maxLen; i++ {
			var av A
			var bv B
			if i < len(a) {
				av = a[i]
			}
			if i < len(b) {
				bv = b[i]
			}
			if !yield(av, bv) {
				return
			}
		}
	}
}
