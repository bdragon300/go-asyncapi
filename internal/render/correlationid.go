package render

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/utils"
	j "github.com/dave/jennifer/jen"
	"github.com/samber/lo"
)

// CorrelationID never renders itself, only as a part of message struct
type CorrelationID struct {
	Name         string
	Description  string
	StructField  string   // Payload or Headers
	LocationPath []string // Should be non-empty
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

func (c CorrelationID) ID() string {
	return c.Name
}

func (c CorrelationID) String() string {
	return "CorrelationID " + c.Name
}

func (c CorrelationID) RenderSetterDefinition(ctx *common.RenderContext, message *Message) []*j.Statement {
	ctx.LogStartRender("CorrelationID.RenderSetterDefinition", "", c.Name, "definition", false)
	defer ctx.LogFinishRender()

	f, ok := lo.Find(message.OutStruct.Fields, func(item GoStructField) bool { return item.Name == c.StructField })
	if !ok {
		panic(fmt.Errorf("field %s not found in OutStruct", c.StructField))
	}

	// Define the first anchor with initial value
	body := []*j.Statement{
		j.Op("v0 :=").Id(message.OutStruct.ReceiverName() + "." + c.StructField),
	}

	// Extract a value from types chain
	bodySteps, err := c.renderValueExtractionCode(ctx, c.LocationPath, f.Type, false)
	if err != nil {
		panic(fmt.Errorf(
			"cannot generate CorrelationID value setter code for types chain at location %q: %s",
			strings.Join(c.LocationPath, "/"),
			err.Error(),
		))
	}

	// Exclude the last step
	body = append(body, lo.FlatMap(bodySteps[:len(bodySteps)-1], func(item correlationIDExpansionStep, _ int) []*j.Statement {
		return item.body
	})...)

	// Collapse the value back
	exprVal := j.Id("value")
	for i := len(bodySteps) - 1; i >= 0; i-- {
		body = append(body, j.Add(bodySteps[i].varValue).Op("=").Add(exprVal.Clone()))
		exprVal = j.Id(bodySteps[i].varValueVarName)
	}
	body = append(body, j.Id(message.OutStruct.ReceiverName()+"."+c.StructField).Op("= v0"))

	receiver := j.Id(message.OutStruct.ReceiverName()).Id(message.OutStruct.Name)

	// Method SetCorrelationID(value any)
	// TODO: comment from description
	return []*j.Statement{
		j.Func().Params(receiver.Clone()).Id("SetCorrelationID").
			Params(j.Id("value").Add(utils.ToCode(bodySteps[len(bodySteps)-1].varType.RenderUsage(ctx))...)).
			Block(utils.ToCode(body)...),
	}
}

func (c CorrelationID) RenderGetterDefinition(ctx *common.RenderContext, message *Message) []*j.Statement {
	ctx.LogStartRender("CorrelationID.RenderGetterDefinition", "", c.Name, "definition", false)
	defer ctx.LogFinishRender()

	f, ok := lo.Find(message.InStruct.Fields, func(item GoStructField) bool { return item.Name == c.StructField })
	if !ok {
		panic(fmt.Errorf("field %s not found in InStruct", c.StructField))
	}

	// Define the first anchor with initial value
	body := []*j.Statement{
		j.Id("v0").Op(":=").Id(message.InStruct.ReceiverName() + "." + c.StructField),
	}

	// Extract a value from types chain
	bodySteps, err := c.renderValueExtractionCode(ctx, c.LocationPath, f.Type, true)
	if err != nil {
		panic(fmt.Errorf(
			"cannot generate CorrelationID value getter code for types chain at location %s: %s",
			strings.Join(c.LocationPath, "/"),
			err.Error(),
		))
	}
	body = append(body, lo.FlatMap(bodySteps, func(item correlationIDExpansionStep, _ int) []*j.Statement {
		return item.body
	})...)
	receiver := j.Id(message.InStruct.ReceiverName()).Id(message.InStruct.Name)

	body = append(body,
		j.Id("value").Op("=").Id(bodySteps[len(bodySteps)-1].varName),
		j.Return(),
	)

	// Method CorrelationID() (any, error)
	// TODO: comment from description
	return []*j.Statement{
		j.Func().Params(receiver.Clone()).Id("CorrelationID").
			Params().
			Params(j.Id("value").Add(utils.ToCode(bodySteps[len(bodySteps)-1].varType.RenderUsage(ctx))...), j.Err().Error()).
			Block(utils.ToCode(body)...),
	}
}

type correlationIDExpansionStep struct {
	body            []*j.Statement
	varName         string
	varValue        *j.Statement
	varValueVarName string
	varType         common.GolangType
}

func (c CorrelationID) renderValueExtractionCode(
	ctx *common.RenderContext,
	path []string,
	initialType common.GolangType,
	validationCode bool,
) (items []correlationIDExpansionStep, err error) {
	// TODO: consider also AdditionalProperties in object
	ctx.Logger.Trace("Render correlationId extraction code", "path", path, "initialType", initialType.String())
	pathIdx := 0

	baseType := initialType
	for pathIdx < len(path) {
		var body []*j.Statement
		var varValueStmts *j.Statement

		// Anchor is a variable that holds the current value of the path item
		anchor := fmt.Sprintf("v%d", pathIdx)
		nextAnchor := fmt.Sprintf("v%d", pathIdx+1)

		memberName, err2 := unescapeCorrelationIDPathItem(path[pathIdx])
		if err2 != nil {
			err = fmt.Errorf("cannot unescape CorrelationID path %q, item %q: %w", path, path[pathIdx], err)
			return
		}

		switch typ := baseType.(type) {
		case *GoStruct:
			ctx.Logger.Trace("In GoStruct", "path", path[:pathIdx], "name", typ.ID(), "member", memberName)
			fld, ok := lo.Find(typ.Fields, func(item GoStructField) bool { return item.MarshalName == memberName })
			if !ok {
				err = fmt.Errorf(
					"field %q not found in struct %s, path: /%s",
					memberName, typ.Name, strings.Join(path[:pathIdx], "/"),
				)
				return
			}
			varValueStmts = j.Id(anchor).Dot(fld.Name)
			baseType = fld.Type
			body = []*j.Statement{j.Id(nextAnchor).Op(":=").Add(varValueStmts)}
		case *GoMap:
			// TODO: x-parameter in correlationIDs spec section to set numbers as "0" for string keys or 0 for int keys
			ctx.Logger.Trace("In GoMap", "path", path[:pathIdx], "name", typ.ID(), "member", memberName)
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
		case *GoArray:
			ctx.Logger.Trace("In GoArray", "path", path[:pathIdx], "name", typ.ID(), "member", memberName)
			if _, ok := memberName.(string); ok {
				err = fmt.Errorf(
					"index %q is not a number, array %s, path: /%s",
					memberName,
					typ.Name,
					strings.Join(path[:pathIdx], "/"),
				)
				return
			}
			if validationCode {
				body = append(body, j.If(j.Len(j.Id(anchor)).Op("<=").Lit(memberName)).Block(
					j.Err().Op("=").Qual("fmt", "Errorf").Call(
						j.Lit(fmt.Sprintf(
							"index %%q is out of range in array of length %%d on path /%s",
							strings.Join(path[:pathIdx], "/"),
						)),
						j.Len(j.Id(anchor)),
						j.Lit(memberName),
					),
					j.Return(),
				))
			}
			varValueStmts = j.Id(anchor).Index(j.Lit(memberName))
			baseType = typ.ItemsType
			body = append(body, j.Id(nextAnchor).Op(":=").Add(varValueStmts))
		case *GoSimple: // Should be a terminal type in chain, raise error otherwise (if any path parts left to resolve)
			ctx.Logger.Trace("In GoSimple", "path", path[:pathIdx], "name", typ.ID(), "member", memberName)
			if pathIdx >= len(path)-1 { // Primitive types should get addressed by the last path item
				err = fmt.Errorf(
					"type %q cannot be resolved further, path: /%s",
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

		item := correlationIDExpansionStep{
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

func unescapeCorrelationIDPathItem(value string) (any, error) {
	if v, err := strconv.Atoi(value); err == nil {
		return v, nil // Number path items are treated as integers
	}

	// Unquote path item if it is quoted. Quoted forces a path item to be treated as a string, not number.
	quoted := strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") ||
		strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")
	if quoted {
		value = value[1 : len(value)-1] // Unquote
	}

	// RFC3986 URL unescape
	value, err := url.PathUnescape(value)
	if err != nil {
		return nil, err
	}

	// RFC6901 JSON Pointer unescape: replace `~1` to `/` and `~0` to `~`
	return strings.ReplaceAll(strings.ReplaceAll(value, "~1", "/"), "~0", "~"), nil
}
