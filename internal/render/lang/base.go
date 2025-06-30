package lang

import (
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/jsonpointer"
)

// BaseJSONPointed holds a JSON Pointer to a current object position in the AsyncAPI document.
// It is a utility type intended to be embedded in other types, don't use it directly.
type BaseJSONPointed struct {
	pointer jsonpointer.JSONPointer
}

func (b *BaseJSONPointed) Pointer() jsonpointer.JSONPointer {
	return b.pointer
}

func (b *BaseJSONPointed) SetPointer(pointer jsonpointer.JSONPointer) {
	b.pointer = pointer
}

// BaseType is a base for the types representing the basic Go type.
type BaseType struct {
	BaseJSONPointed
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
	// ArtifactKind for types in [lang] package has values ArtifactKindOther or ArtifactKindSchema
	ArtifactKind common.ArtifactKind
}

func (b *BaseType) Name() string {
	return b.OriginalName
}

func (b *BaseType) Kind() common.ArtifactKind {
	return b.ArtifactKind
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

// GolangWrappedType is an interface for [common.GolangType] types (such as pointers, aliases) that are able to
// wrap another [common.GolangType] type.
type GolangWrappedType interface {
	// UnwrapGolangType recursively unwraps the wrapped [common.GolangType] type.
	UnwrapGolangType() common.GolangType
}

// GolangReferenceType is an interface for references to golang type (such as promises, refs).
// Used primarily in templates.
type GolangReferenceType interface {
	// DerefGolangType recursively unwraps the referenced [common.GolangType] type.
	DerefGolangType() common.GolangType
}

type golangStructType interface {
	IsStruct() bool
}
