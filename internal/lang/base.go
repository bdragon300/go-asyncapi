package lang

import (
	"fmt"
	"strings"

	"github.com/bdragon300/asyncapi-codegen/internal/render"
	"github.com/bdragon300/asyncapi-codegen/internal/utils"
	"github.com/dave/jennifer/jen"
)

type LangType interface {
	render.LangRenderer
	// canBePointer returns true if a pointer may be applied yet to a type during rendering. E.g. types that are
	// already pointers can't be pointed the second time -- this function returns false
	canBePointer() bool
	GetName() string
}

type BaseType struct {
	Name        string
	Description string

	Imports map[string]string

	// Render denotes if this type must be rendered separately. Otherwise, it will only be inlined in a parent definition
	// Such as inlined `field struct{...}` and separate `field StructName`, or `field []type` and `field ArrayName`
	Render bool
}

func (b *BaseType) AllowRender() bool {
	return b.Render
}

func (b *BaseType) GetName() string {
	return b.Name
}

func (b *BaseType) AdditionalImports() map[string]string {
	return b.Imports
	// return lo.Assign(lo.Map(b.MethodSnippets, func(item DefinitionSnippet, index int) map[string]string {
	//	return item.Imports
	// })...)
}

type Array struct {
	BaseType
	ItemsType LangType
	Size      int
}

func (a *Array) canBePointer() bool {
	return false
}

func (a *Array) RenderDefinition(ctx *render.Context) []*jen.Statement {
	var res []*jen.Statement
	if a.Description != "" {
		res = append(res, jen.Comment(a.Name+" -- "+utils.ToLowerFirstLetter(a.Description)))
	}

	stmt := jen.Type().Id(a.Name)
	if a.Size > 0 {
		stmt = stmt.Index(jen.Lit(a.Size))
	} else {
		stmt = stmt.Index()
	}
	items := utils.CastSliceItems[*jen.Statement, jen.Code](a.ItemsType.RenderUsage(ctx))
	res = append(res, stmt.Add(items...))
	// res = append(res, lo.FlatMap(a.MethodSnippets, func(item DefinitionSnippet, index int) []*jen.Statement {
	//	return item.renderDefinitions(a)
	// })...)

	return res
}

func (a *Array) RenderUsage(ctx *render.Context) []*jen.Statement {
	if a.Render {
		if ctx.ForceImportPackage != "" {
			return []*jen.Statement{jen.Qual(ctx.ForceImportPackage, a.Name)}
		}
		return []*jen.Statement{jen.Id(a.Name)}
	}

	items := utils.CastSliceItems[*jen.Statement, jen.Code](a.ItemsType.RenderUsage(ctx))
	return []*jen.Statement{jen.Index().Add(items...)}
}

type Map struct {
	BaseType
	KeyType   LangType
	ValueType LangType
}

func (m *Map) canBePointer() bool {
	return false
}

func (m *Map) RenderDefinition(ctx *render.Context) []*jen.Statement {
	var res []*jen.Statement
	if m.Description != "" {
		res = append(res, jen.Comment(m.Name+" -- "+utils.ToLowerFirstLetter(m.Description)))
	}

	stmt := jen.Type().Id(m.Name)
	keyType := utils.CastSliceItems[*jen.Statement, jen.Code](m.KeyType.RenderUsage(ctx))
	valueType := utils.CastSliceItems[*jen.Statement, jen.Code](m.ValueType.RenderUsage(ctx))
	res = append(res, stmt.Map((&jen.Statement{}).Add(keyType...)).Add(valueType...))
	// res = append(res, lo.FlatMap(m.MethodSnippets, func(item DefinitionSnippet, index int) []*jen.Statement {
	//	return item.renderDefinitions(m)
	// })...)

	return res
}

func (m *Map) RenderUsage(ctx *render.Context) []*jen.Statement {
	if m.Render {
		if ctx.ForceImportPackage != "" {
			return []*jen.Statement{jen.Qual(ctx.ForceImportPackage, m.Name)}
		}
		return []*jen.Statement{jen.Id(m.Name)}
	}

	keyType := utils.CastSliceItems[*jen.Statement, jen.Code](m.KeyType.RenderUsage(ctx))
	valueType := utils.CastSliceItems[*jen.Statement, jen.Code](m.ValueType.RenderUsage(ctx))
	return []*jen.Statement{jen.Map((&jen.Statement{}).Add(keyType...)).Add(valueType...)}
}

type TypeAlias struct {
	BaseType
	AliasedType LangType

	// Render config
	Nullable bool
}

func (p *TypeAlias) canBePointer() bool {
	return !p.Nullable
}

func (p *TypeAlias) RenderDefinition(ctx *render.Context) []*jen.Statement {
	var res []*jen.Statement
	if p.Description != "" {
		res = append(res, jen.Comment(p.Name+" -- "+utils.ToLowerFirstLetter(p.Description)))
	}

	aliasedStmt := utils.CastSliceItems[*jen.Statement, jen.Code](p.AliasedType.RenderDefinition(ctx))
	res = append(res, jen.Type().Id(p.Name).Add(aliasedStmt...))
	// res = append(res, lo.FlatMap(p.MethodSnippets, func(item DefinitionSnippet, index int) []*jen.Statement {
	//	return item.renderDefinitions(p)
	// })...)
	return res
}

func (p *TypeAlias) RenderUsage(ctx *render.Context) []*jen.Statement {
	stmt := &jen.Statement{}
	if p.Nullable {
		stmt = stmt.Op("*")
	}
	if p.Render {
		if ctx.ForceImportPackage != "" {
			return []*jen.Statement{jen.Qual(ctx.ForceImportPackage, p.Name)}
		}
		return []*jen.Statement{stmt.Id(p.Name)}
	}

	aliasedStmt := utils.CastSliceItems[*jen.Statement, jen.Code](p.AliasedType.RenderUsage(ctx))
	return []*jen.Statement{stmt.Add(aliasedStmt...)}
}

type Simple struct {
	TypeName   string // type name with or without package name, such as "json.Marshal" or "string"
	ImportPath string // optional import path, such as "encoding/json"
}

func (p Simple) AllowRender() bool {
	return false
}

func (p Simple) RenderDefinition(*render.Context) []*jen.Statement {
	if p.ImportPath != "" {
		return []*jen.Statement{jen.Qual(p.ImportPath, p.TypeName)}
	}
	return []*jen.Statement{jen.Id(p.TypeName)}
}

func (p Simple) RenderUsage(ctx *render.Context) []*jen.Statement {
	if p.ImportPath != "" {
		if ctx.ForceImportPackage != "" {
			return []*jen.Statement{jen.Qual(ctx.ForceImportPackage, p.TypeName)}
		}
		return []*jen.Statement{jen.Qual(p.ImportPath, p.TypeName)}
	}
	return []*jen.Statement{jen.Id(p.TypeName)}
}

func (p Simple) AdditionalImports() map[string]string {
	if p.ImportPath != "" {
		parts := strings.Split(p.ImportPath, "/")
		if len(parts) < 2 {
			panic(fmt.Sprintf("Wrong import path %q", p.ImportPath))
		}
		return map[string]string{p.ImportPath: parts[len(parts)-1]}
	}
	return nil
}

func (p Simple) canBePointer() bool {
	return false
}

func (p Simple) GetName() string {
	return ""
}
