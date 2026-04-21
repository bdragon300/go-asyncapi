package lang

import (
	"strings"

	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/samber/lo"
)

// GoSimple is a simple Go type that does not require any special handling. It can be a built-in type like int, or
// a type imported from library like [time.Time] or [golang.org/x/net/ipv4.Conn].
type GoSimple struct {
	BaseType
	// TypeName is the name of the type to be rendered
	TypeName string
	// IsInterface is true if the type is an interface, which means it cannot be rendered as a pointer
	IsInterface bool

	// OriginalFormat is optional format of the type that is set for a type in document, e.g. "date-time" for string
	OriginalFormat string
	// OriginalType is the original type from the document, e.g. "integer" for int32
	OriginalType string

	StructFieldRenderInfo StructFieldRenderInfo
}

func (p *GoSimple) Name() string {
	if p.HasDefinition {
		return p.OriginalName
	}
	return p.TypeName
}

func (p *GoSimple) Kind() common.ArtifactKind {
	return lo.CoalesceOrEmpty(p.ArtifactKind, common.ArtifactKindOther)
}

func (p *GoSimple) Selectable() bool {
	return p.HasDefinition
}

func (p *GoSimple) String() string {
	b := strings.Builder{}
	b.WriteString("GoSimple(")
	if p.HasDefinition {
		b.WriteString(p.OriginalName)
		b.WriteString(" -> ")
	}
	if p.Import != "" {
		b.WriteString(p.Import)
		b.WriteString(".")
	}
	b.WriteString(p.TypeName)
	if p.OriginalFormat != "" {
		b.WriteString(":")
		b.WriteString(p.OriginalFormat)
	}
	b.WriteString(")")
	return b.String()
}

func (p *GoSimple) CanBeAddressed() bool {
	return !p.IsInterface
}

func (p *GoSimple) GoTemplate() string {
	return "code/lang/gosimple"
}

func (p *GoSimple) StructRenderInfo() StructFieldRenderInfo {
	return p.StructFieldRenderInfo
}
