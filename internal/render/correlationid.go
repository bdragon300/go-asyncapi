package render

import (
	"fmt"
	"strings"

	"github.com/bdragon300/asyncapi-codegen-go/internal/common"
	"github.com/bdragon300/asyncapi-codegen-go/internal/utils"
	j "github.com/dave/jennifer/jen"
	"github.com/samber/lo"
)

// CorrelationID never renders itself, only as a part of message struct
type CorrelationID struct {
	Name        string
	Description string
	StructField string   // Payload or Headers
	Path        []string // Should be non-empty
}

func (c CorrelationID) DirectRendering() bool {
	return false
}

func (c CorrelationID) RenderDefinition(_ *common.RenderContext) []*j.Statement {
	panic("not implemented")
}

func (c CorrelationID) RenderUsage(_ *common.RenderContext) []*j.Statement {
	panic("not implemented")
}

func (c CorrelationID) String() string {
	return c.Name
}

func (c CorrelationID) RenderSetterDefinition(ctx *common.RenderContext, message *Message) []*j.Statement {
	ctx.LogRender("CorrelationID_Setter", "", c.Name, "definition", false)
	defer ctx.LogReturn()

	f, ok := lo.Find(message.OutStruct.Fields, func(item StructField) bool { return item.Name == c.StructField })
	if !ok {
		// FIXME: output error without panic
		panic(fmt.Sprintf("Field %s not found in OutStruct", c.StructField))
	}

	body := []*j.Statement{
		j.Op("v0 :=").Id(message.OutStruct.ReceiverName() + "." + c.StructField),
	}

	bodyItems, err := c.renderMemberExtractionCode(ctx, c.Path, f.Type, true)
	if err != nil {
		// FIXME: output error without panic
		panic(fmt.Sprintf("Cannot render CorrelationID %s: %s", strings.Join(c.Path, "/"), err.Error()))
	}
	// Exclude the last definition statement
	body = append(body, lo.FlatMap(bodyItems[:len(bodyItems)-1], func(item correlationIDBodyItem, index int) []*j.Statement {
		return item.body
	})...)

	exprVal := j.Id("value")
	for i := len(bodyItems) - 1; i >= 0; i-- {
		body = append(body, j.Add(bodyItems[i].varValue).Op("=").Add(exprVal.Clone()))
		exprVal = j.Id(bodyItems[i].varValueVarName)
	}
	body = append(body, j.Id(message.OutStruct.ReceiverName()+"."+c.StructField).Op("= v0"))
	receiver := j.Id(message.OutStruct.ReceiverName()).Id(message.OutStruct.Name)

	// Method SetCorrelationID(value any)
	// TODO: comment from description
	return []*j.Statement{
		j.Func().Params(receiver.Clone()).Id("SetCorrelationID").
			Params(j.Id("value").Add(utils.ToCode(bodyItems[len(bodyItems)-1].varType.RenderUsage(ctx))...)).
			Block(utils.ToCode(body)...),
	}
}

func (c CorrelationID) RenderGetterDefinition(ctx *common.RenderContext, message *Message) []*j.Statement {
	ctx.LogRender("CorrelationID_Getter", "", c.Name, "definition", false)
	defer ctx.LogReturn()

	f, ok := lo.Find(message.InStruct.Fields, func(item StructField) bool { return item.Name == c.StructField })
	if !ok {
		// FIXME: output error without panic
		panic(fmt.Sprintf("Field %s not found in InStruct", c.StructField))
	}

	body := []*j.Statement{
		j.Id("v0").Op(":=").Id(message.InStruct.ReceiverName() + "." + c.StructField),
	}

	bodyItems, err := c.renderMemberExtractionCode(ctx, c.Path, f.Type, true)
	if err != nil {
		// FIXME: output error without panic
		panic(fmt.Sprintf("Cannot render CorrelationID %s: %s", strings.Join(c.Path, "/"), err.Error()))
	}
	body = append(body, lo.FlatMap(bodyItems, func(item correlationIDBodyItem, index int) []*j.Statement {
		return item.body
	})...)
	receiver := j.Id(message.InStruct.ReceiverName()).Id(message.InStruct.Name)

	body = append(body,
		j.Id("value").Op("=").Id(bodyItems[len(bodyItems)-1].varName),
		j.Return(),
	)

	// Method CorrelationID() (any, error)
	// TODO: comment from description
	return []*j.Statement{
		j.Func().Params(receiver.Clone()).Id("CorrelationID").
			Params().
			Params(j.Id("value").Add(utils.ToCode(bodyItems[len(bodyItems)-1].varType.RenderUsage(ctx))...), j.Err().Error()).
			Block(utils.ToCode(body)...),
	}
}

type correlationIDBodyItem struct {
	body            []*j.Statement
	varName         string
	varValue        *j.Statement
	varValueVarName string
	varType         common.GolangType
}

func (c CorrelationID) renderMemberExtractionCode(
	ctx *common.RenderContext,
	path []string,
	initialType common.GolangType,
	validationCode bool,
) (items []correlationIDBodyItem, err error) {
	// TODO: consider also AdditionalProperties in object
	ctx.Logger.Trace("Render correlationId extraction code", "path", path, "initialType", initialType.String())
	pathIdx := 0

	baseType := initialType
	for pathIdx < len(path) {
		var body []*j.Statement
		var varValueStmts *j.Statement
		anchor := fmt.Sprintf("v%d", pathIdx)
		nextAnchor := fmt.Sprintf("v%d", pathIdx+1)
		memberName := path[pathIdx]

		switch typ := baseType.(type) {
		case *Struct:
			ctx.Logger.Trace("In Struct", "path", path[:pathIdx], "name", typ.String(), "member", memberName)
			fld, ok := lo.Find(typ.Fields, func(item StructField) bool { return item.MarshalName == memberName })
			if !ok {
				err = fmt.Errorf("field %q not found, path: /%s", memberName, strings.Join(path[:pathIdx], "/"))
				return
			}
			varValueStmts = j.Id(anchor).Dot(fld.Name)
			baseType = fld.Type
			body = []*j.Statement{j.Id(nextAnchor).Op(":=").Add(varValueStmts)}
		case *Map:
			// TODO: x-parameter in correlationIDs spec section to set numbers as "0" for string keys or 0 for int keys
			ctx.Logger.Trace("In Map", "path", path[:pathIdx], "name", typ.String(), "member", memberName)
			varValueStmts = j.Id(anchor).Index(j.Lit(memberName))
			baseType = typ.ValueType
			varExpr := j.Var().Id(nextAnchor).Add(utils.ToCode(typ.ValueType.RenderUsage(ctx))...)
			if t, ok := typ.ValueType.(golangPointerType); ok && t.IsPointer() {
				// Append ` = new(TYPE)` to initialize a pointer
				varExpr = varExpr.Op("=").New(j.Add(utils.ToCode(typ.ValueType.RenderUsage(ctx))...))
			}

			ifExpr := j.If(j.Op("v, ok :=").Add(varValueStmts), j.Id("ok")).Block(
				j.Id(nextAnchor).Op("=").Id("v"),
			)
			if validationCode {
				ifExpr = ifExpr.Else().Block(
					j.Err().Op("=").Qual("fmt", "Errorf").Call(
						j.Lit(fmt.Sprintf("key %%q not found in map on path /%s", strings.Join(path[:pathIdx], "/"))),
						j.Lit(memberName),
					),
					j.Return(),
				)
			}
			body = []*j.Statement{
				j.If(j.Id(anchor).Op("==").Nil()).Block(
					j.Id(anchor).Op("=").Make(utils.ToCode(typ.RenderUsage(ctx))...),
				),
				varExpr,
				ifExpr,
			}
		case *Array:
			ctx.Logger.Trace("In Array", "path", path[:pathIdx], "name", typ.String(), "member", memberName)
			varValueStmts = j.Id(anchor).Index(j.Lit(memberName))
			baseType = typ.ItemsType
			body = []*j.Statement{j.Id(nextAnchor).Op(":=").Add(varValueStmts)}
			if validationCode {
				body = append(body, j.If(j.Len(j.Id(anchor)).Op("<=").Lit(memberName)).Block(
					j.Err().Op("=").Qual("fmt", "Errorf").Call(
						j.Lit(fmt.Sprintf("index %%q not found in array of length %%d on path /%s", strings.Join(path[:pathIdx], "/"))),
						j.Len(j.Id(anchor)),
						j.Lit(memberName),
					),
					j.Return(),
				))
			}
		case *Simple: // Should be a terminal type in chain, raise error otherwise (if any path parts left to resolve)
			ctx.Logger.Trace("In Simple", "path", path[:pathIdx], "name", typ.String(), "member", memberName)
			if pathIdx >= len(path)-1 { // Primitive types should get addressed by the last path item
				err = fmt.Errorf(
					"type %q doesn't contain addressable elements, path: /%s",
					typ.TypeName(),
					strings.Join(path[:pathIdx], "/"),
				)
				return
			}
			baseType = typ
		case golangTypeWrapperType:
			ctx.Logger.Trace(
				"In wrapper type",
				"path", path[:pathIdx], "name", typ.String(), "type", fmt.Sprintf("%T", typ), "member", memberName,
			)
			t, ok := typ.WrappedGolangType()
			if !ok {
				err = fmt.Errorf(
					"wrapped type %T does not contain a wrapped GolangType, path: /%s",
					typ,
					strings.Join(path[:pathIdx], "/"),
				)
				return
			}
			baseType = t
			continue
		default:
			ctx.Logger.Trace("Unknown type", "path", path[:pathIdx], "name", typ.String(), "type", fmt.Sprintf("%T", typ))
			err = fmt.Errorf(
				"type %s is not addressable, path: /%s",
				typ.TypeName(),
				strings.Join(path[:pathIdx], "/"),
			)
			return
		}

		item := correlationIDBodyItem{
			body:            body,
			varName:         nextAnchor,
			varValue:        varValueStmts,
			varValueVarName: anchor,
			varType:         baseType,
		}
		items = append(items, item)
		pathIdx++
	}

	return
}
