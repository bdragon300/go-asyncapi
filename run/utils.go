package run

func DerefOrZero[T any](x *T) T {
	if x != nil {
		return *x
	}
	return *new(T)
}

func ToPtr[T any](x T) *T {
	return &x
}
