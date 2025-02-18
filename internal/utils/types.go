package utils

import "github.com/samber/lo"

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
