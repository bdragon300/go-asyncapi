package lang

import (
	"github.com/bdragon300/go-asyncapi/internal/common"
)

type GoSimple struct {
	Name        string // type name
	IsInterface bool   // true if type is interface, which means it cannot be rendered as pointer
	Import      string // optional generated package name or module to import a type from
}

func (p GoSimple) Kind() common.ObjectKind {
	return common.ObjectKindLang
}

func (p GoSimple) Selectable() bool {
	return false
}

func (p GoSimple) IsPointer() bool {
	return false
}

func (p GoSimple) D() string {
	//ctx.LogStartRender("GoSimple", p.Import, p.Name, "definition", p.Selectable())
	//defer ctx.LogFinishRender()
	//
	//stmt := jen.Id(p.Name)
	//return []*jen.Statement{stmt}
	return renderTemplate("lang/gosimple/definition", &p)
}

func (p GoSimple) U() string {
	//ctx.LogStartRender("GoSimple", p.Import, p.Name, "usage", p.Selectable())
	//defer ctx.LogFinishRender()
	//
	//stmt := &jen.Statement{}
	//switch {
	//case p.Import != "" && p.Import != context.Context.CurrentPackage:
	//	stmt = stmt.Qual(p.Import, p.Name)
	//default:
	//	stmt = stmt.Id(p.Name)
	//}
	//
	//return []*jen.Statement{stmt}
	return renderTemplate("lang/gosimple/usage", &p)
}

func (p GoSimple) TypeName() string {
	return p.Name
}

func (p GoSimple) String() string {
	if p.Import != "" {
		return "GoSimple /" + p.Import + "." + p.Name
	}
	return "GoSimple " + p.Name
}

func (p GoSimple) DefinitionInfo() (*common.GolangTypeDefinitionInfo, error) {
	return nil, nil
}