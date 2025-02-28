package lang

import (
	"fmt"

	"github.com/bdragon300/go-asyncapi/internal/common"
)

// GoSimple is a simple Go type that does not require any special handling. It can be a built-in type like int, or
// a type imported from library like [time.Time] or [golang.org/x/net/ipv4.Conn].
type GoSimple struct {
	BasePositioned
	// TypeName is the name of the type to be rendered
	TypeName string
	// IsInterface is true if the type is an interface, which means it cannot be rendered as a pointer
	IsInterface bool
	// Import is an optional package name or module to import a type from. E.g. "github.com/your/module" or "time"
	// If set, then while rendering the usage of the type, this import will be added to the file's imports list.
	Import string
	// IsRuntimeImport is true if the Import field contains a subpackage in the tool's runtime subpackage. E.g. "kafka"
	IsRuntimeImport bool
}

func (p *GoSimple) Name() string {
	return p.TypeName
}

func (p *GoSimple) Kind() common.ArtifactKind {
	return common.ArtifactKindOther
}

func (p *GoSimple) Selectable() bool {
	return false
}

func (p *GoSimple) Visible() bool {
	return true
}

func (p *GoSimple) String() string {
	if p.Import != "" {
		return fmt.Sprintf("GoSimple(%s.%s)", p.Import, p.TypeName)
	}
	return "GoSimple(" + p.TypeName + ")"
}

func (p *GoSimple) CanBeAddressed() bool {
	return !p.IsInterface
}

func (p *GoSimple) CanBeDereferenced() bool {
	return false
}

func (p *GoSimple) GoTemplate() string {
	return "code/lang/gosimple"
}
