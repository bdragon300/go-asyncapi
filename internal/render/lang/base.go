package lang

import "github.com/bdragon300/go-asyncapi/internal/common"

type BaseType struct {
	Name        string
	Description string

	// HasDefinition is true if this type has a definition in the generated code. Otherwise, it renders as inline type.
	// Such as inlined `field struct{...}` and separate `field StructName`, or `field []type` and `field ArrayName`
	HasDefinition bool
	Import        string // optional generated package name or module to import a type from // TODO: replace to "is in runtime module" flag?
}

func (b BaseType) Kind() common.ObjectKind {
	return common.ObjectKindLang
}

func (b *BaseType) Selectable() bool {
	return b.HasDefinition
}

func (b *BaseType) TypeName() string {
	return b.Name
}

func (b *BaseType) String() string {
	if b.Import != "" {
		return "GoType /" + b.Import + "." + b.Name
	}
	return "GoType " + b.Name
}

type GolangTypeWrapperType interface {
	WrappedGolangType() (common.GolangType, bool)
	String() string
}

type GolangPointerType interface {
	IsPointer() bool
}

type golangStructType interface {
	IsStruct() bool
}
