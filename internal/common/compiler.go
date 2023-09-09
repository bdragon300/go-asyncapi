package common

type GolangType interface {
	Assembler
	TypeName() string
}
