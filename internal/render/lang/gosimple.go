package lang

import (
	"github.com/bdragon300/go-asyncapi/internal/common"
)

type GoSimple struct {
	TypeName    string // type name
	IsInterface bool   // true if type is interface, which means it cannot be rendered as pointer  // TODO: use or remove
	Import      string // optional package name or module to import a type from
	RuntimeImport bool // true indicates that Import contains a runtime package
}

func (p *GoSimple) Kind() common.ObjectKind {
	return common.ObjectKindOther
}

func (p *GoSimple) Selectable() bool {
	return false
}

func (p *GoSimple) Visible() bool {
	return true
}

func (p *GoSimple) Addressable() bool {
	return !p.IsInterface
}

func (p *GoSimple) IsPointer() bool {
	return false
}

func (p *GoSimple) GoTemplate() string {
	return "lang/gosimple"
}

func (p *GoSimple) String() string {
	if p.Import != "" {
		return "GoSimple /" + p.Import + "." + p.TypeName
	}
	return "GoSimple " + p.TypeName
}

func (p *GoSimple) Name() string {
	return p.TypeName
}