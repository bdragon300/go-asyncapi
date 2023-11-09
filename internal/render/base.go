package render

import (
	"github.com/samber/lo"

	"github.com/bdragon300/asyncapi-codegen-go/internal/common"
	"github.com/bdragon300/asyncapi-codegen-go/internal/utils"
	"github.com/dave/jennifer/jen"
)

type BaseType struct {
	Name        string
	Description string

	// Render denotes if this type must be rendered separately. Otherwise, it will only be inlined in a parent definition
	// Such as inlined `field struct{...}` and separate `field StructName`, or `field []type` and `field ArrayName`
	Render      bool
	PackageName string // optional import path from any generated package
}

func (b *BaseType) AllowRender() bool {
	return b.Render
}

func (b *BaseType) TypeName() string {
	return b.Name
}

func (b *BaseType) String() string {
	return b.Name
}

type Array struct {
	BaseType
	ItemsType common.GolangType
	Size      int
}

func (a *Array) RenderDefinition(ctx *common.RenderContext) []*jen.Statement {
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
	items := utils.ToCode(a.ItemsType.RenderUsage(ctx))
	res = append(res, stmt.Add(items...))

	return res
}

func (a *Array) RenderUsage(ctx *common.RenderContext) []*jen.Statement {
	if a.Render {
		if a.PackageName != "" && a.PackageName != ctx.CurrentPackage {
			return []*jen.Statement{jen.Qual(ctx.GeneratedPackage(a.PackageName), a.Name)}
		}
		return []*jen.Statement{jen.Id(a.Name)}
	}

	items := utils.ToCode(a.ItemsType.RenderUsage(ctx))
	return []*jen.Statement{jen.Index().Add(items...)}
}

type Map struct {
	BaseType
	KeyType   common.GolangType
	ValueType common.GolangType
}

func (m *Map) RenderDefinition(ctx *common.RenderContext) []*jen.Statement {
	var res []*jen.Statement
	if m.Description != "" {
		res = append(res, jen.Comment(m.Name+" -- "+utils.ToLowerFirstLetter(m.Description)))
	}

	stmt := jen.Type().Id(m.Name)
	keyType := utils.ToCode(m.KeyType.RenderUsage(ctx))
	valueType := utils.ToCode(m.ValueType.RenderUsage(ctx))
	res = append(res, stmt.Map((&jen.Statement{}).Add(keyType...)).Add(valueType...))

	return res
}

func (m *Map) RenderUsage(ctx *common.RenderContext) []*jen.Statement {
	if m.Render {
		if m.PackageName != "" && m.PackageName != ctx.CurrentPackage {
			return []*jen.Statement{jen.Qual(ctx.GeneratedPackage(m.PackageName), m.Name)}
		}
		return []*jen.Statement{jen.Id(m.Name)}
	}

	keyType := utils.ToCode(m.KeyType.RenderUsage(ctx))
	valueType := utils.ToCode(m.ValueType.RenderUsage(ctx))
	return []*jen.Statement{jen.Map((&jen.Statement{}).Add(keyType...)).Add(valueType...)}
}

type TypeAlias struct {
	BaseType
	AliasedType common.GolangType
}

func (p *TypeAlias) RenderDefinition(ctx *common.RenderContext) []*jen.Statement {
	var res []*jen.Statement
	if p.Description != "" {
		res = append(res, jen.Comment(p.Name+" -- "+utils.ToLowerFirstLetter(p.Description)))
	}

	aliasedStmt := utils.ToCode(p.AliasedType.RenderDefinition(ctx))
	res = append(res, jen.Type().Id(p.Name).Add(aliasedStmt...))
	return res
}

func (p *TypeAlias) RenderUsage(ctx *common.RenderContext) []*jen.Statement {
	if p.Render {
		if p.PackageName != "" && p.PackageName != ctx.CurrentPackage {
			return []*jen.Statement{jen.Qual(ctx.GeneratedPackage(p.PackageName), p.Name)}
		}
		return []*jen.Statement{jen.Id(p.Name)}
	}

	aliasedStmt := utils.ToCode(p.AliasedType.RenderUsage(ctx))
	return []*jen.Statement{jen.Add(aliasedStmt...)}
}

type Simple struct {
	Name            string // type name with or without package name, such as "json.Marshal" or "string"
	IsIface         bool
	Package         string            // optional import path from any generated package
	TypeParamValues []common.Renderer // optional type parameter types to be filled in definition and usage
}

func (p Simple) AllowRender() bool {
	return false
}

func (p Simple) RenderDefinition(ctx *common.RenderContext) []*jen.Statement {
	stmt := jen.Id(p.Name)
	if len(p.TypeParamValues) > 0 {
		typeParams := lo.FlatMap(p.TypeParamValues, func(item common.Renderer, index int) []jen.Code {
			return utils.ToCode(item.RenderUsage(ctx))
		})
		stmt = stmt.Types(typeParams...)
	}
	return []*jen.Statement{stmt}
}

func (p Simple) RenderUsage(ctx *common.RenderContext) []*jen.Statement {
	stmt := &jen.Statement{}
	switch {
	case p.Package != "" && p.Package != ctx.CurrentPackage:
		stmt = stmt.Qual(p.Package, p.Name)
	default:
		stmt = stmt.Id(p.Name)
	}

	if len(p.TypeParamValues) > 0 {
		typeParams := lo.FlatMap(p.TypeParamValues, func(item common.Renderer, index int) []jen.Code {
			return utils.ToCode(item.RenderUsage(ctx))
		})
		stmt = stmt.Types(typeParams...)
	}

	return []*jen.Statement{stmt}
}

func (p Simple) TypeName() string {
	return ""
}

func (p Simple) String() string {
	return p.Name
}

type NullableType struct {
	Type   common.GolangType
	Render bool
}

func (n NullableType) AllowRender() bool {
	return n.Render
}

func (n NullableType) RenderDefinition(ctx *common.RenderContext) []*jen.Statement {
	return n.Type.RenderDefinition(ctx)
}

func (n NullableType) RenderUsage(ctx *common.RenderContext) []*jen.Statement {
	isPtr := true
	switch v := n.Type.(type) {
	case *Interface:
		isPtr = false
	case *Simple:
		isPtr = !v.IsIface
	}
	if isPtr {
		return []*jen.Statement{jen.Op("*").Add(utils.ToCode(n.Type.RenderUsage(ctx))...)}
	}
	return n.Type.RenderUsage(ctx)
}

func (n NullableType) TypeName() string {
	return n.Type.TypeName()
}

func (n NullableType) String() string {
	return n.Type.String()
}