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
