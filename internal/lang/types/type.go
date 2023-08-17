package types

import (
	"github.com/bdragon300/asyncapi-codegen/internal/scancontext"
	"github.com/dave/jennifer/jen"
	"github.com/samber/lo"
)

type LangType interface {
	// canBePointer returns true if a pointer may be applied yet to a type during rendering. E.g. types that are
	// already pointers can't be pointed the second time -- this function returns false
	canBePointer() bool
}

type BaseType struct {
	Name        string
	DefaultName string
	Description string

	// Inline type will be inlined on usage rendering, otherwise it will be declared as a separate type.
	// Such as inlined `field struct{...}` and separate `field StructName`, or `field []type` and `field ArrayName`
	Inline bool
}

func (b *BaseType) PrepareRender(name string) {
	b.Name = name
}

func (b *BaseType) GetDefaultName() string {
	return b.DefaultName
}

func (b *BaseType) SkipRender() bool {
	return b.Inline
}

// Struct defines the data required to generate a struct in Go.
type Struct struct {
	BaseType
	Fields map[string]Field

	// Render config
	Nullable bool // When struct is built from type: ["object", "null"]
}

func (s *Struct) canBePointer() bool {
	return !s.Nullable
}

func (s *Struct) RenderDefinition() []*jen.Statement {
	var res []*jen.Statement
	if s.Description != "" {
		res = append(res, jen.Comment(s.Name+" "+s.Description))
	}
	var structFields []jen.Code
	for _, f := range s.Fields {
		items := lo.Map(f.renderDefinition(), func(item *jen.Statement, index int) jen.Code { return item })
		structFields = append(structFields, items...)
	}

	stmt := jen.Type().Id(s.Name).Struct(structFields...)
	res = append(res, stmt)

	return res
}

func (s *Struct) RenderUsage() []*jen.Statement {
	stmt := &jen.Statement{}
	if s.Nullable {
		stmt = stmt.Op("*")
	}
	if !s.Inline {
		return []*jen.Statement{stmt.Id(s.Name)}
	}

	var structFields []jen.Code
	for _, f := range s.Fields {
		items := lo.Map(f.renderDefinition(), func(item *jen.Statement, index int) jen.Code { return item })
		structFields = append(structFields, items...)
	}
	return []*jen.Statement{stmt.Struct(structFields...)}
}

// Field defines the data required to generate a field in Go.
type Field struct {
	Name          string
	Description   string
	Type          LangType
	RequiredValue bool
	Tags          map[string]string
}

func (f *Field) renderDefinition() []*jen.Statement {
	var res []*jen.Statement
	if f.Description != "" {
		res = append(res, jen.Comment(f.Name+" "+f.Description))
	}

	stmt := jen.Id(f.Name)
	if f.Type.canBePointer() && f.RequiredValue {
		stmt = stmt.Op("*")
	}
	items := lo.Map(f.Type.(scancontext.LangRenderer).RenderUsage(), func(item *jen.Statement, index int) jen.Code { return item })
	res = append(res, stmt.Add(items...))

	return res
}

type Array struct {
	BaseType
	ItemsType LangType
}

func (a *Array) canBePointer() bool {
	return false
}

func (a *Array) RenderDefinition() []*jen.Statement {
	var res []*jen.Statement
	if a.Description != "" {
		res = append(res, jen.Comment(a.Name+" "+a.Description))
	}

	stmt := jen.Type().Id(a.Name).Index()
	items := lo.Map(a.ItemsType.(scancontext.LangRenderer).RenderUsage(), func(item *jen.Statement, index int) jen.Code { return item })
	res = append(res, stmt.Add(items...))

	return res
}

func (a *Array) RenderUsage() []*jen.Statement {
	if !a.Inline {
		return []*jen.Statement{jen.Id(a.Name)}
	}

	items := lo.Map(a.ItemsType.(scancontext.LangRenderer).RenderUsage(), func(item *jen.Statement, index int) jen.Code { return item })
	return []*jen.Statement{jen.Index().Add(items...)}
}

type Map struct {
	BaseType
	KeyType   string
	ValueType LangType
}

func (m *Map) canBePointer() bool {
	return false
}

func (m *Map) RenderDefinition() []*jen.Statement {
	var res []*jen.Statement
	if m.Description != "" {
		res = append(res, jen.Comment(m.Name+" "+m.Description))
	}

	stmt := jen.Type().Id(m.Name).Map(jen.Id(m.KeyType))
	items := lo.Map(m.ValueType.(scancontext.LangRenderer).RenderUsage(), func(item *jen.Statement, index int) jen.Code { return item })
	res = append(res, stmt.Add(items...))

	return res
}

func (m *Map) RenderUsage() []*jen.Statement {
	if !m.Inline {
		return []*jen.Statement{jen.Id(m.Name)}
	}

	items := lo.Map(m.ValueType.(scancontext.LangRenderer).RenderUsage(), func(item *jen.Statement, index int) jen.Code { return item })
	return []*jen.Statement{jen.Map(jen.Id(m.KeyType)).Add(items...)}
}

type PrimitiveType struct {
	BaseType
	LangType string

	// Render config
	Nullable bool
}

func (p *PrimitiveType) canBePointer() bool {
	return !p.Nullable
}

func (p *PrimitiveType) RenderDefinition() []*jen.Statement {
	var res []*jen.Statement
	if p.Description != "" {
		res = append(res, jen.Comment(p.Name+" "+p.Description))
	}

	res = append(res, jen.Type().Id(p.Name).Id(p.LangType))
	return res
}

func (p *PrimitiveType) RenderUsage() []*jen.Statement {
	stmt := &jen.Statement{}
	if p.Nullable {
		stmt = stmt.Op("*")
	}
	if !p.Inline {
		return []*jen.Statement{stmt.Id(p.Name)}
	}

	return []*jen.Statement{stmt.Id(p.LangType)}
}

type Any struct {
	BaseType
}

func (a *Any) canBePointer() bool {
	return false
}

func (a *Any) RenderDefinition() []*jen.Statement {
	var res []*jen.Statement
	if a.Description != "" {
		res = append(res, jen.Comment(a.Name+" "+a.Description))
	}

	res = append(res, jen.Type().Id(a.Name).Any())
	return res
}

func (a *Any) RenderUsage() []*jen.Statement {
	if !a.Inline {
		return []*jen.Statement{jen.Id(a.Name)}
	}

	return []*jen.Statement{jen.Any()}
}

type TypeBindWrapper struct {
	BaseType
	RefQuery *scancontext.RefQuery[LangType]
}

func (r TypeBindWrapper) RenderDefinition() []*jen.Statement {
	return r.RefQuery.Link.(scancontext.LangRenderer).RenderDefinition()
}

func (r TypeBindWrapper) RenderUsage() []*jen.Statement {
	return r.RefQuery.Link.(scancontext.LangRenderer).RenderUsage()
}

func (r TypeBindWrapper) canBePointer() bool {
	return r.RefQuery.Link.canBePointer()
}
