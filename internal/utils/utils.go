package utils

func CastSliceItems[F any, T any](in []F) []T {
	result := make([]T, len(in))
	for i, item := range in {
		result[i] = any(item).(T)
	}
	return result
}
