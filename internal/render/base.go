package render

import "github.com/bdragon300/asyncapi-codegen-go/internal/common"

type BaseType struct {
	Name        string
	Description string

	// DirectRender is true if the rendering of this type must be invoked directly on rendering phase. Otherwise, the
	// rendering of this type is invoked indirectly by another type.
	// Such as inlined `field struct{...}` and separate `field StructName`, or `field []type` and `field ArrayName`
	DirectRender bool
	PackageName  string // optional generated package name or module to import a type from
}

func (b *BaseType) DirectRendering() bool {
	return b.DirectRender
}

func (b *BaseType) TypeName() string {
	return b.Name
}

func (b *BaseType) ID() string {
	return b.Name
}

func (b *BaseType) String() string {
	if b.PackageName != "" {
		return "GoType ." + b.PackageName + "." + b.Name
	}
	return "GoType " + b.Name
}

type golangTypeWrapperType interface {
	WrappedGolangType() (common.GolangType, bool)
	String() string
}

type golangPointerType interface {
	IsPointer() bool
}

type golangStructType interface {
	IsStruct() bool
}
