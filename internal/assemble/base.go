package assemble

import (
	"path"

	"github.com/bdragon300/asyncapi-codegen/internal/common"
	"github.com/bdragon300/asyncapi-codegen/internal/utils"
	"github.com/dave/jennifer/jen"
)

type BaseType struct {
	Name        string
	Description string

	// Render denotes if this type must be rendered separately. Otherwise, it will only be inlined in a parent definition
	// Such as inlined `field struct{...}` and separate `field StructName`, or `field []type` and `field ArrayName`
	Render  bool
	Package common.PackageKind // optional import path from any generated package
}

func (b *BaseType) AllowRender() bool {
	return b.Render
}

func (b *BaseType) TypeName() string {
	return b.Name
}

type Array struct {
	BaseType
	ItemsType common.GolangType
	Size      int
}

func (a *Array) AssembleDefinition(ctx *common.AssembleContext) []*jen.Statement {
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
	items := utils.ToCode(a.ItemsType.AssembleUsage(ctx))
	res = append(res, stmt.Add(items...))

	return res
}

func (a *Array) AssembleUsage(ctx *common.AssembleContext) []*jen.Statement {
	if a.Render {
		if a.Package != "" && a.Package != ctx.CurrentPackage {
			return []*jen.Statement{jen.Qual(path.Join(ctx.ImportBase, string(a.Package)), a.Name)}
		}
		return []*jen.Statement{jen.Id(a.Name)}
	}

	items := utils.ToCode(a.ItemsType.AssembleUsage(ctx))
	return []*jen.Statement{jen.Index().Add(items...)}
}

type Map struct {
	BaseType
	KeyType   common.GolangType
	ValueType common.GolangType
}

func (m *Map) AssembleDefinition(ctx *common.AssembleContext) []*jen.Statement {
	var res []*jen.Statement
	if m.Description != "" {
		res = append(res, jen.Comment(m.Name+" -- "+utils.ToLowerFirstLetter(m.Description)))
	}

	stmt := jen.Type().Id(m.Name)
	keyType := utils.ToCode(m.KeyType.AssembleUsage(ctx))
	valueType := utils.ToCode(m.ValueType.AssembleUsage(ctx))
	res = append(res, stmt.Map((&jen.Statement{}).Add(keyType...)).Add(valueType...))

	return res
}

func (m *Map) AssembleUsage(ctx *common.AssembleContext) []*jen.Statement {
	if m.Render {
		if m.Package != "" && m.Package != ctx.CurrentPackage {
			return []*jen.Statement{jen.Qual(path.Join(ctx.ImportBase, string(m.Package)), m.Name)}
		}
		return []*jen.Statement{jen.Id(m.Name)}
	}

	keyType := utils.ToCode(m.KeyType.AssembleUsage(ctx))
	valueType := utils.ToCode(m.ValueType.AssembleUsage(ctx))
	return []*jen.Statement{jen.Map((&jen.Statement{}).Add(keyType...)).Add(valueType...)}
}

type TypeAlias struct {
	BaseType
	AliasedType common.GolangType
}

func (p *TypeAlias) AssembleDefinition(ctx *common.AssembleContext) []*jen.Statement {
	var res []*jen.Statement
	if p.Description != "" {
		res = append(res, jen.Comment(p.Name+" -- "+utils.ToLowerFirstLetter(p.Description)))
	}

	aliasedStmt := utils.ToCode(p.AliasedType.AssembleDefinition(ctx))
	res = append(res, jen.Type().Id(p.Name).Add(aliasedStmt...))
	return res
}

func (p *TypeAlias) AssembleUsage(ctx *common.AssembleContext) []*jen.Statement {
	if p.Render {
		if p.Package != "" && p.Package != ctx.CurrentPackage {
			return []*jen.Statement{jen.Qual(path.Join(ctx.ImportBase, string(p.Package)), p.Name)}
		}
		return []*jen.Statement{jen.Id(p.Name)}
	}

	aliasedStmt := utils.ToCode(p.AliasedType.AssembleUsage(ctx))
	return []*jen.Statement{jen.Add(aliasedStmt...)}
}

type Simple struct {
	Type            string // type name with or without package name, such as "json.Marshal" or "string"
	IsIface         bool
	ExternalPackage string             // optional import path, such as "encoding/json"
	Package         common.PackageKind // optional import path from any generated package
}

func (p Simple) AllowRender() bool {
	return false
}

func (p Simple) AssembleDefinition(*common.AssembleContext) []*jen.Statement {
	return []*jen.Statement{jen.Id(p.Type)}
}

func (p Simple) AssembleUsage(ctx *common.AssembleContext) []*jen.Statement {
	if p.ExternalPackage != "" {
		return []*jen.Statement{jen.Qual(p.ExternalPackage, p.Type)}
	}
	if p.Package != "" && p.Package != ctx.CurrentPackage {
		return []*jen.Statement{jen.Qual(path.Join(ctx.ImportBase, string(p.Package)), p.Type)}
	}
	return []*jen.Statement{jen.Id(p.Type)}
}

func (p Simple) TypeName() string {
	return ""
}

type NullableType struct {
	Type   common.GolangType
	Render bool
}

func (n NullableType) AllowRender() bool {
	return n.Render
}

func (n NullableType) AssembleDefinition(ctx *common.AssembleContext) []*jen.Statement {
	return n.Type.AssembleDefinition(ctx)
}

func (n NullableType) AssembleUsage(ctx *common.AssembleContext) []*jen.Statement {
	isPtr := true
	switch v := n.Type.(type) {
	case *Interface:
		isPtr = false
	case *Simple:
		isPtr = !v.IsIface
	}
	if isPtr {
		return []*jen.Statement{jen.Op("*").Add(utils.ToCode(n.Type.AssembleUsage(ctx))...)}
	}
	return n.Type.AssembleUsage(ctx)
}

func (n NullableType) TypeName() string {
	return n.Type.TypeName()
}
