package common

type GolangType interface {
	Assembled
	// CanBePointer returns true if a pointer may be applied yet to a type during rendering. E.g. types that are
	// already pointers can't be pointed the second time -- this function returns false
	CanBePointer() bool
	TypeName() string
}
