package utils

import "github.com/dave/jennifer/jen"

func ToCode(in []*jen.Statement) []jen.Code {
	result := make([]jen.Code, len(in))
	for i, item := range in {
		result[i] = any(item).(jen.Code)
	}
	return result
}

func IsSubsequence[T comparable](subseq, iterable []T, searchIndex int) (int, bool) {
	if searchIndex+len(subseq) > len(iterable) {
		return 0, false
	}
	for ind, item := range subseq {
		if item != iterable[searchIndex+ind] {
			return ind, false
		}
	}
	return searchIndex + len(subseq), true
}

func SlicesEqual[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	_, ok := IsSubsequence(a, b, 0)
	return ok
}
