package lang

import (
	"github.com/bdragon300/go-asyncapi/internal/common"
)

type BaseType struct {
	OriginalName string
	Description  string

	// HasDefinition is true if this type should have a definition in the generated code. Otherwise, it renders as
	// inline type. Such as inlined `field struct{...}` and separate `field StructName`, or `field []type`
	// and `field ArrayName`
	HasDefinition bool
	Import        string // optional external (or runtime) module to import a type from
	definitionInfo *common.GolangTypeDefinitionInfo
}

func (b *BaseType) Kind() common.ObjectKind {
	return common.ObjectKindOther
}

func (b *BaseType) Selectable() bool {
	return b.HasDefinition
}

func (b *BaseType) IsPointer() bool {
	return false
}

func (b *BaseType) DefinitionInfo() (*common.GolangTypeDefinitionInfo, error) {
	if b.definitionInfo == nil {
		return nil, common.ErrDefinitionLocationUnknown
	}
	return b.definitionInfo, nil
}

func (b *BaseType) SetDefinitionInfo(info *common.GolangTypeDefinitionInfo) {
	b.definitionInfo = info
}

func (b *BaseType) GetOriginalName() string {
	return b.OriginalName
}

type GolangTypeWrapperType interface {
	UnwrapGolangType() (common.GolangType, bool)
}

type golangStructType interface {
	IsStruct() bool
}
