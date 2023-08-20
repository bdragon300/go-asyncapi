package lang

import (
	"github.com/bdragon300/asyncapi-codegen/internal/scanner"
	"github.com/bdragon300/asyncapi-codegen/internal/utils"
	"github.com/dave/jennifer/jen"
	"github.com/samber/lo"
)

// Struct defines the data required to generate a struct in Go.
type Struct struct {
	BaseType
	Fields  []StructField
	Methods []StructMethod

	// Render config
	Nullable bool
}

func (s *Struct) canBePointer() bool {
	return !s.Nullable
}

func (s *Struct) RenderDefinition() []*jen.Statement {
	var res []*jen.Statement
	if s.Description != "" {
		res = append(res, jen.Comment(s.Name+" -- "+utils.ToLowerFirstLetter(s.Description)))
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

// StructField defines the data required to generate a field in Go.
type StructField struct {
	Name          string
	Description   string
	Type          LangType
	RequiredValue bool
	Tags          map[string]string
}

func (f *StructField) renderDefinition() []*jen.Statement {
	var res []*jen.Statement
	if f.Description != "" {
		res = append(res, jen.Comment(f.Name+" -- "+utils.ToLowerFirstLetter(f.Description)))
	}

	stmt := jen.Id(f.Name)
	if f.Type.canBePointer() && f.RequiredValue {
		stmt = stmt.Op("*")
	}
	items := lo.Map(f.Type.(scanner.LangRenderer).RenderUsage(), func(item *jen.Statement, index int) jen.Code { return item })
	res = append(res, stmt.Add(items...))

	return res
}

type StructMethod struct {
	Name            string
	Description     string
	ReceiverName    string
	PointerReceiver bool
	Parameters      map[string]LangType
	ReturnType      map[string]LangType
	Body            MethodBody
}

func (s *StructMethod) renderDefinition(strct *Struct) []*jen.Statement {
	var res []*jen.Statement
	if s.Description != "" {
		res = append(res, jen.Comment(s.Name+" -- "+utils.ToLowerFirstLetter(s.Description)))
	}
	stmt := jen.Func()

	// Receiver
	receiver := jen.Id(s.ReceiverName)
	if s.PointerReceiver {
		receiver = receiver.Op("*")
	}
	stmt = stmt.Params(receiver.Id(strct.Name)).Id(s.Name)

	// Parameters
	var code []jen.Code
	for _, param := range s.Parameters {
		items := lo.Map(param.(scanner.LangRenderer).RenderUsage(), func(item *jen.Statement, index int) jen.Code { return item })
		code = append(code, items...)
	}
	stmt = stmt.Params(code...)
	code = code[:0]

	// Return value
	for _, ret := range s.ReturnType {
		items := lo.Map(ret.(scanner.LangRenderer).RenderUsage(), func(item *jen.Statement, index int) jen.Code { return item })
		code = append(code, items...)
	}
	if len(code) > 1 {
		stmt = stmt.Params(code...) // Several return values
	} else {
		stmt = stmt.Add(code[0]) // Single return value
	}

	// Body
	code = lo.Map(s.Body.renderDefinition(strct), func(item *jen.Statement, index int) jen.Code { return item })
	stmt = stmt.Block(code...)
	res = append(res, stmt)

	return res
}

type MethodBody struct {
	ReturnLiter       string
	ReturnName        string
	ReturnStructField string
}

func (m *MethodBody) renderDefinition(strct *Struct) []*jen.Statement {
	switch {
	case m.ReturnLiter != "":
		return []*jen.Statement{jen.Lit(m.ReturnLiter)}
	case m.ReturnName != "":
		return []*jen.Statement{jen.Id(m.ReturnName)}
	case m.ReturnStructField != "":
		return []*jen.Statement{jen.Id(strct.Name).Dot(m.ReturnName)}
	}
	return []*jen.Statement{jen.Empty()}
}
