package lang

import (
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/utils"
)

type BaseType struct {
	OriginalName string
	Description  string

	// HasDefinition is true if this type should have a definition in the generated code. Otherwise, it renders as
	// inline type. Such as inlined `field struct{...}` and separate `field StructName`, or `field []type`
	// and `field ArrayName`
	HasDefinition bool
	Import        string // optional external (or runtime) module to import a type from
	// Typically it's ObjectKindOther or ObjectKindSchema
	ObjectKind common.ObjectKind
}

func (b *BaseType) Kind() common.ObjectKind {
	return b.ObjectKind
}

func (b *BaseType) Selectable() bool {
	return b.HasDefinition
}

func (b *BaseType) Visible() bool {
	return true
}

func (b *BaseType) IsPointer() bool {
	return false
}

func (b *BaseType) Name() string {
	return utils.CapitalizeUnchanged(b.OriginalName)
}

func (b *BaseType) ObjectHasDefinition() bool {
	return b.HasDefinition
}

type GolangTypeWrapperType interface {
	UnwrapGolangType() (common.GolangType, bool)
}

type golangStructType interface {
	IsStruct() bool
}
