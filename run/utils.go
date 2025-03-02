package run

// FromPtrOrZero returns the dereferenced value of the pointer or the zero value of the type if it is nil.
func FromPtrOrZero[T any](x *T) T {
	if x != nil {
		return *x
	}
	return *new(T)
}

// ToPtr is helper function the just returns a pointer to the value. Useful where applying & operator is not allowed.
func ToPtr[T any](x T) *T {
	return &x
}
