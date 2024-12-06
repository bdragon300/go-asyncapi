package lang

import (
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render/context"
	"github.com/bdragon300/go-asyncapi/internal/tpl"
	"strings"
)

type BaseType struct {
	Name        string
	Description string

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

func (b *BaseType) TypeName() string {
	return b.Name
}

func (b *BaseType) IsPointer() bool {
	return false
}

func (b *BaseType) DefinitionInfo() (*common.GolangTypeDefinitionInfo, error) {
	if b.definitionInfo == nil {
		return nil, common.ErrObjectDefinitionUnknownYet
	}
	return b.definitionInfo, nil
}

type GolangTypeWrapperType interface {
	UnwrapGolangType() (common.GolangType, bool)
	//String() string
}

type GolangPointerType interface {  // TODO: replace with common.GolangType? or move where it is used
	IsPointer() bool
}

type golangStructType interface {
	IsStruct() bool
}

func renderTemplate[T common.Renderable](name string, obj T) string {
	var b strings.Builder
	tmpl := tpl.LoadTemplate(name)
	if tmpl == nil {
		panic("template not found: " + name)
	}

	tmpl = tmpl.Funcs(tpl.GetTemplateFunctions(context.Context))
	if err := tmpl.Execute(&b, obj); err != nil {
		panic(err)
	}
	return b.String()
}