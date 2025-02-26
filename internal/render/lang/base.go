package lang

import (
	"github.com/bdragon300/go-asyncapi/internal/common"
)

// BaseType is a base for the types representing the basic Go type.
type BaseType struct {
	// OriginalName is the name of the type as it appears in the AsyncAPI document.
	OriginalName string
	// Description is optional description. Renders as Go doc comment.
	Description string

	// HasDefinition is true if this type renders as definition in the generated code. Otherwise, it renders as
	// inline type. Such as inlined `field struct{...}` and separate `field StructName`, or `field []type`
	// and `field ArrayName`
	HasDefinition bool
	// Import is an optional external (or runtime) module to import a type from. E.g. "github.com/your/module"
	Import string
	// ObjectKind for types in [lang] package has values ObjectKindOther or ObjectKindSchema
	ObjectKind common.ObjectKind
}

func (b *BaseType) Name() string {
	return b.OriginalName
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

func (b *BaseType) CanBeAddressed() bool {
	return true
}

func (b *BaseType) CanBeDereferenced() bool {
	return false
}

func (b *BaseType) ObjectHasDefinition() bool {
	return b.HasDefinition
}

// GolangTypeExtractor is an interface for [common.GolangType] types (such as pointers, aliases) that are able to
// wrap another [common.GolangType] type.
type GolangTypeExtractor interface {
	InnerGolangType() common.GolangType
}

// GolangTypeWrapper is an interface for *non* [common.GolangType] types (such as promises, refs) that contain a
// [common.GolangType] type inside.
// Used primarily in templates.
type GolangTypeWrapper interface {
	UnwrapGolangType() common.GolangType
}

type golangStructType interface {
	IsStruct() bool
}
