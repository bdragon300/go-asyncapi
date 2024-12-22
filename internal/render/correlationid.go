package render

import (
	"fmt"
	"github.com/bdragon300/go-asyncapi/internal/render/lang"
	"github.com/bdragon300/go-asyncapi/internal/tmpl"
	"github.com/bdragon300/go-asyncapi/internal/utils"
	"net/url"
	"strconv"
	"strings"

	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/samber/lo"
)

const structReceiverName = "m"

type CorrelationIDStructField string

const (
	CorrelationIDStructFieldPayload CorrelationIDStructField = "payload"
	CorrelationIDStructFieldHeaders CorrelationIDStructField = "headers"
)

// CorrelationID never renders itself, only as a part of message struct
type CorrelationID struct {
	OriginalName string
	Description  string
	StructField  CorrelationIDStructField // Type field name to store the value to or to load the value from
	LocationPath []string                 // JSONPointer path to the field in the message, should be non-empty
}

func (c *CorrelationID) Kind() common.ObjectKind {
	return common.ObjectKindCorrelationID
}

func (c *CorrelationID) Selectable() bool {
	return false
}

func (c *CorrelationID) Visible() bool {
	return true
}

func (c *CorrelationID) Name() string {
	return utils.CapitalizeUnchanged(c.OriginalName)
}

//func (c CorrelationID) D(_ *common.RenderContext) []*j.Statement {
//	panic("not implemented")
//}
//
//func (c CorrelationID) U(_ *common.RenderContext) []*j.Statement {
//	panic("not implemented")
//}
//
//func (c CorrelationID) ID() string {
//	return c.GetOriginalName
//}

func (c *CorrelationID) String() string {
	return "CorrelationID " + c.OriginalName
}

func (c *CorrelationID) RenderSetterBody(inVar string, inVarType *lang.GoStruct) string {
	//ctx.LogStartRender("CorrelationID.RenderSetterBody", "", c.GetOriginalName, "definition", false)
	//defer ctx.LogFinishRender()

	f, ok := lo.Find(inVarType.Fields, func(item lang.GoStructField) bool { return item.Name == string(c.StructField) })
	if !ok {
		panic(fmt.Errorf("field %s not found in inVarType", c.StructField))
	}

	// Define the first anchor with initial value
	body := []string{
		fmt.Sprintf("v0 := %s.%s", inVar, c.StructField),
	}

	// Extract a value from types chain
	bodySteps, err := c.renderValueExtractionCode(c.LocationPath, f.Type, false)
	if err != nil {
		panic(fmt.Errorf(
			"cannot generate CorrelationID value setter code for types chain at location %q: %s",
			strings.Join(c.LocationPath, "/"),
			err.Error(),
		))
	}

	// Exclude the last step
	body = append(body, lo.FlatMap(bodySteps[:len(bodySteps)-1], func(item correlationIDExpansionStep, _ int) []string {
		return item.codeLines
	})...)

	// Collapse the value back
	exprVal := "value"
	for i := len(bodySteps) - 1; i >= 0; i-- {
		body = append(body, fmt.Sprintf("%s = %s", bodySteps[i].varValue, exprVal))
		exprVal = bodySteps[i].varValueVarName
	}
	body = append(body, fmt.Sprintf("%s.%s = v0", inVar, c.StructField))

	return strings.Join(body, "\n")
}

func (c *CorrelationID) TargetVarType(varType *lang.GoStruct) common.GolangType {
	f, ok := lo.Find(varType.Fields, func(item lang.GoStructField) bool { return item.Name == string(c.StructField) })
	if !ok {
		panic(fmt.Errorf("field %s not found in varType", c.StructField))
	}
	bodySteps, err := c.renderValueExtractionCode(c.LocationPath, f.Type, false)
	if err != nil {
		panic(fmt.Errorf(
			"cannot generate CorrelationID value setter code for types chain at location %q: %s",
			strings.Join(c.LocationPath, "/"),
			err.Error(),
		))
	}
	return bodySteps[len(bodySteps)-1].varType
}

//func (c CorrelationID) RenderSetterDefinition(ctx *common.RenderContext, message *Message) []*j.Statement {
//	ctx.LogStartRender("CorrelationID.RenderSetterDefinition", "", c.GetOriginalName, "definition", false)
//	defer ctx.LogFinishRender()
//
//	f, ok := lo.Find(message.OutType.Fields, func(item GoStructField) bool { return item.GetOriginalName == c.StructField })
//	if !ok {
//		panic(fmt.Errorf("field %s not found in OutType", c.StructField))
//	}
//
//	// Define the first anchor with initial value
//	codeLines := []*j.Statement{
//		j.Op("v0 :=").Id(message.OutType.ReceiverName() + "." + c.StructField),
//	}
//
//	// Extract a value from types chain
//	bodySteps, err := c.renderValueExtractionCode(ctx, c.LocationPath, f.Type, false)
//	if err != nil {
//		panic(fmt.Errorf(
//			"cannot generate CorrelationID value setter code for types chain at location %q: %s",
//			strings.Join(c.LocationPath, "/"),
//			err.Error(),
//		))
//	}
//
//	// Exclude the last step
//	codeLines = append(codeLines, lo.FlatMap(bodySteps[:len(bodySteps)-1], func(item correlationIDExpansionStep, _ int) []*j.Statement {
//		return item.codeLines
//	})...)
//
//	// Collapse the value back
//	exprVal := j.Id("value")
//	for i := len(bodySteps) - 1; i >= 0; i-- {
//		codeLines = append(codeLines, j.Add(bodySteps[i].varValue).Op("=").Add(exprVal.Clone()))
//		exprVal = j.Id(bodySteps[i].varValueVarName)
//	}
//	codeLines = append(codeLines, j.Id(message.OutType.ReceiverName()+"."+c.StructField).Op("= v0"))
//
//	receiver := j.Id(message.OutType.ReceiverName()).Id(message.OutType.GetOriginalName)
//
//	// Method SetCorrelationID(value any)
//	// TODO: comment from description
//	return []*j.Statement{
//		j.Func().Params(receiver.Clone()).Id("SetCorrelationID").
//			Params(j.Id("value").Add(utils.ToCode(bodySteps[len(bodySteps)-1].varType.U(ctx))...)).
//			Block(utils.ToCode(codeLines)...),
//	}
//}

func (c *CorrelationID) RenderGetterBody(outVar string, outVarType *lang.GoStruct) string {
	//ctx.LogStartRender("CorrelationID.RenderGetterDefinition", "", c.GetOriginalName, "definition", false)
	//defer ctx.LogFinishRender()

	f, ok := lo.Find(outVarType.Fields, func(item lang.GoStructField) bool { return item.Name == string(c.StructField) })
	if !ok {
		panic(fmt.Errorf("field %s not found in outVarType", c.StructField))
	}

	// Define the first anchor with initial value
	body := []string {
		fmt.Sprintf("v0 := %s.%s", structReceiverName, c.StructField),
	}

	// Extract a value from types chain
	bodySteps, err := c.renderValueExtractionCode(c.LocationPath, f.Type, true)
	if err != nil {
		panic(fmt.Errorf(
			"cannot generate CorrelationID value getter code for types chain at location %s: %s",
			strings.Join(c.LocationPath, "/"),
			err.Error(),
		))
	}
	body = append(body, lo.FlatMap(bodySteps, func(item correlationIDExpansionStep, _ int) []string {
		return item.codeLines
	})...)
	body = append(body, fmt.Sprintf("%s = %s", outVar, bodySteps[len(bodySteps)-1].varName))

	return strings.Join(body, "\n")
}

//func (c CorrelationID) RenderGetterDefinition(ctx *common.RenderContext, message *Message) []*j.Statement {
//	ctx.LogStartRender("CorrelationID.RenderGetterDefinition", "", c.GetOriginalName, "definition", false)
//	defer ctx.LogFinishRender()
//
//	f, ok := lo.Find(message.InType.Fields, func(item GoStructField) bool { return item.GetOriginalName == c.StructField })
//	if !ok {
//		panic(fmt.Errorf("field %s not found in InType", c.StructField))
//	}
//
//	// Define the first anchor with initial value
//	codeLines := []*j.Statement{
//		j.Id("v0").Op(":=").Id(message.InType.ReceiverName() + "." + c.StructField),
//	}
//
//	// Extract a value from types chain
//	bodySteps, err := c.renderValueExtractionCode(ctx, c.LocationPath, f.Type, true)
//	if err != nil {
//		panic(fmt.Errorf(
//			"cannot generate CorrelationID value getter code for types chain at location %s: %s",
//			strings.Join(c.LocationPath, "/"),
//			err.Error(),
//		))
//	}
//	codeLines = append(codeLines, lo.FlatMap(bodySteps, func(item correlationIDExpansionStep, _ int) []*j.Statement {
//		return item.codeLines
//	})...)
//	receiver := j.Id(message.InType.ReceiverName()).Id(message.InType.GetOriginalName)
//
//	codeLines = append(codeLines,
//		j.Id("value").Op("=").Id(bodySteps[len(bodySteps)-1].varName),
//		j.Return(),
//	)
//
//	// Method CorrelationID() (any, error)
//	// TODO: comment from description
//	return []*j.Statement{
//		j.Func().Params(receiver.Clone()).Id("CorrelationID").
//			Params().
//			Params(j.Id("value").Add(utils.ToCode(bodySteps[len(bodySteps)-1].varType.U(ctx))...), j.Err().Error()).
//			Block(utils.ToCode(codeLines)...),
//	}
//}

type correlationIDExpansionStep struct {
	codeLines []string
	varName   string
	varValue        string
	varValueVarName string
	varType         common.GolangType
}

func (c *CorrelationID) renderValueExtractionCode(
	path []string,
	initialType common.GolangType,
	validationCode bool,
) (items []correlationIDExpansionStep, err error) {
	// TODO: consider also AdditionalProperties in object
	//ctx.Logger.Trace("Render correlationId extraction code", "path", path, "initialType", initialType.String())
	pathIdx := 0

	baseType := initialType
	for pathIdx < len(path) {
		var body []string
		var varValueStmts string

		// Anchor is a variable that holds the current value of the path item
		anchor := fmt.Sprintf("v%d", pathIdx)
		nextAnchor := fmt.Sprintf("v%d", pathIdx+1)

		memberName, err2 := unescapeCorrelationIDPathItem(path[pathIdx])
		if err2 != nil {
			err = fmt.Errorf("cannot unescape CorrelationID path %q, item %q: %w", path, path[pathIdx], err)
			return
		}

		switch typ := baseType.(type) {
		case *lang.GoStruct:
			//ctx.Logger.Trace("In GoStruct", "path", path[:pathIdx], "name", typ.ID(), "member", memberName)
			fld, ok := lo.Find(typ.Fields, func(item lang.GoStructField) bool { return item.MarshalName == memberName })
			if !ok {
				err = fmt.Errorf(
					"field %q not found in struct %s, path: /%s",
					memberName, typ.OriginalName, strings.Join(path[:pathIdx], "/"),
				)
				return
			}
			varValueStmts = fmt.Sprintf("%s.%s", anchor, fld.Name)
			baseType = fld.Type
			body = []string{fmt.Sprintf("%s := %s", nextAnchor, varValueStmts)}
		case *lang.GoMap:
			// TODO: x-parameter in correlationIDs spec section to set numbers as "0" for string keys or 0 for int keys
			//ctx.Logger.Trace("In GoMap", "path", path[:pathIdx], "name", typ.ID(), "member", memberName)
			varValueStmts = fmt.Sprintf("%s[%s]", anchor, utils.ToGoLiteral(memberName))
			baseType = typ.ValueType
			varExpr := fmt.Sprintf("var %s %s", nextAnchor, lo.Must(tmpl.TemplateGoUsage(typ.ValueType)))
			if typ.ValueType.IsPointer() {
				// Append ` = new(TYPE)` to initialize a pointer
				varExpr += fmt.Sprintf(" = new(%s)", lo.Must(tmpl.TemplateGoUsage(typ.ValueType)))
			}

			ifExpr := fmt.Sprintf(`if v, ok := %s; ok {
				%s = v
			}`, varValueStmts, nextAnchor)
			if validationCode {
				fmtErrorf := common.GetContext().QualifiedName("fmt.Errorf")
				ifExpr += fmt.Sprintf(` else {
					err = %s("key %%q not found in map on path /%s", %s)
					return
				}`, fmtErrorf, strings.Join(path[:pathIdx], "/"), utils.ToGoLiteral(memberName))
			}
			body = []string{
				fmt.Sprintf(`if %s == nil { 
					%s = make(%s) 
				}`, anchor, anchor, lo.Must(tmpl.TemplateGoUsage(typ))),
				varExpr,
				ifExpr,
			}
		case *lang.GoArray:
			//ctx.Logger.Trace("In GoArray", "path", path[:pathIdx], "name", typ.ID(), "member", memberName)
			if _, ok := memberName.(string); ok {
				err = fmt.Errorf(
					"index %q is not a number, array %s, path: /%s",
					memberName,
					typ.OriginalName,
					strings.Join(path[:pathIdx], "/"),
				)
				return
			}
			if validationCode {
				fmtErrorf := common.GetContext().QualifiedName("fmt.Errorf")
				body = append(body, fmt.Sprintf(`if len(%s) <= %s {
					err = %s("index %%q is out of range in array of length %%d on path /%s", %s, len(%s))
					return
				}`, anchor, utils.ToGoLiteral(memberName), fmtErrorf, strings.Join(path[:pathIdx], "/"), utils.ToGoLiteral(memberName), anchor))
			}
			varValueStmts = fmt.Sprintf("%s[%s]", anchor, utils.ToGoLiteral(memberName))
			baseType = typ.ItemsType
			body = append(body, fmt.Sprintf("%s := %s", nextAnchor, varValueStmts))
		case *lang.GoSimple: // Should be a terminal type in chain, raise error otherwise (if any path parts left to resolve)
			//ctx.Logger.Trace("In GoSimple", "path", path[:pathIdx], "name", typ.ID(), "member", memberName)
			if pathIdx >= len(path)-1 { // Primitive types should get addressed by the last path item
				err = fmt.Errorf(
					"type %q cannot be resolved further, path: /%s",
					typ.Name(), // TODO: check if this is correct
					strings.Join(path[:pathIdx], "/"),
				)
				return
			}
			baseType = typ
		case lang.GolangTypeWrapperType:
			//ctx.Logger.Trace(
			//	"In wrapper type",
			//	"path", path[:pathIdx], "name", typ.String(), "type", fmt.Sprintf("%T", typ), "member", memberName,
			//)
			t, ok := typ.UnwrapGolangType()
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
			//ctx.Logger.Trace("Unknown type", "path", path[:pathIdx], "name", typ.String(), "type", fmt.Sprintf("%T", typ))
			err = fmt.Errorf(
				"type %s is not addressable, path: /%s",
				typ.Name(), // TODO: check if this is correct
				strings.Join(path[:pathIdx], "/"),
			)
			return
		}

		item := correlationIDExpansionStep{
			codeLines:       body,
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
