package utils

import (
	"cmp"
	"iter"
	"maps"
	"slices"

	"github.com/samber/lo"
)

// WithoutBy returns slice excluding all given values executing the predicate function.
func WithoutBy[T comparable, Slice ~[]T](collection Slice, exclude Slice, predicate func(a, b T) bool) Slice {
	result := make(Slice, 0, len(collection))
	for i := range collection {
		contains := lo.ContainsBy(exclude, func(item T) bool { return predicate(collection[i], item) })
		if !contains {
			result = append(result, collection[i])
		}
	}
	return result
}

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
