package lang

import (
	"github.com/bdragon300/go-asyncapi/internal/common"
)

type GoSimple struct {
	Name        string // type name
	IsInterface bool   // true if type is interface, which means it cannot be rendered as pointer  // TODO: use or remove
	Import      string // optional generated package name or module to import a type from
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

func (p *GoSimple) IsPointer() bool {
	return false
}

func (p *GoSimple) GoTemplate() string {
	return "lang/gosimple"
}

func (p *GoSimple) String() string {
	if p.Import != "" {
		return "GoSimple /" + p.Import + "." + p.Name
	}
	return "GoSimple " + p.Name
}

func (p *GoSimple) GetOriginalName() string {
	return p.Name
}