package render

import (
	"fmt"
	"github.com/bdragon300/go-asyncapi/internal/log"
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
	CorrelationIDStructFieldPayload CorrelationIDStructField = "Payload"
	CorrelationIDStructFieldHeaders CorrelationIDStructField = "Headers"
)

// CorrelationID never renders itself, only as a part of message struct
type CorrelationID struct {
	OriginalName string
	Description  string
	StructField  CorrelationIDStructField // Type field name to store the value to or to load the value from
	LocationPath []string                 // JSONPointer path to the field in the message, should be non-empty
	Dummy bool
}

func (c *CorrelationID) Kind() common.ObjectKind {
	return common.ObjectKindOther
}

func (c *CorrelationID) Selectable() bool {
	return false
}

func (c *CorrelationID) Visible() bool {
	return !c.Dummy
}

func (c *CorrelationID) Name() string {
	return c.OriginalName
}

func (c *CorrelationID) String() string {
	return "CorrelationID " + c.OriginalName
}

func (c *CorrelationID) RenderSetterBody(inVar string, inVarType *lang.GoStruct) string {
	logger := log.GetLogger(log.LoggerPrefixRendering)
	logger.Trace(
		"--> Render CorrelationID setter body", "object", c.String(), "inVar", inVar, "inVarType",inVarType.String(),
	)

	f, ok := lo.Find(inVarType.Fields, func(item lang.GoStructField) bool { return item.Name == string(c.StructField) })
	if !ok {
		panic(fmt.Errorf("field %s not found in %s", c.StructField, inVarType))
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

func (c *CorrelationID) RenderGetterBody(outVar string, outVarType *lang.GoStruct) string {
	logger := log.GetLogger(log.LoggerPrefixRendering)
	logger.Trace(
		"--> Render CorrelationID getter body", "object", c.String(), "outVar", outVar, "outVarType",outVarType.String(),
	)

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

func (c *CorrelationID) TargetVarType(varType *lang.GoStruct) common.GolangType {
	f, ok := lo.Find(varType.Fields, func(item lang.GoStructField) bool {
		return item.Name == string(c.StructField)
	})
	if !ok {
		panic(fmt.Errorf("struct field %q not found in %s", c.StructField, varType))
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
	logger := log.GetLogger(log.LoggerPrefixRendering)
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
			logger.Trace("---> GoStruct", "path", path[:pathIdx], "member", memberName, "object", typ.String())
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
			logger.Trace("---> GoMap", "path", path[:pathIdx], "member", memberName, "object", typ.String())
			varValueStmts = fmt.Sprintf("%s[%s]", anchor, utils.ToGoLiteral(memberName))
			baseType = typ.ValueType
			// TODO: replace TemplateGoUsage calls to smth another to remove import from impl, this is a potential circular import
			varExpr := fmt.Sprintf("var %s %s", nextAnchor, lo.Must(tmpl.TemplateGoUsage(typ.ValueType)))
			if typ.ValueType.Addressable() {
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
			logger.Trace("---> GoArray", "path", path[:pathIdx], "member", memberName, "object", typ.String())
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
			logger.Trace("---> GoSimple", "path", path[:pathIdx], "member", memberName, "object", typ.String())
			if pathIdx >= len(path)-1 { // Primitive types should get addressed by the last path item
				err = fmt.Errorf(
					"type %q cannot be resolved further, path: /%s",
					typ.Name(), // TODO: check if this is correct
					strings.Join(path[:pathIdx], "/"),
				)
				return
			}
			baseType = typ
		case lang.GolangTypeExtractor:
			logger.Trace(
				"---> GolangTypeExtractor",
				"path", path[:pathIdx], "member", memberName, "object", baseType.String(), "type", fmt.Sprintf("%T", typ),
			)
			t := typ.InnerGolangType()
			if lo.IsNil(t) {
				err = fmt.Errorf(
					"type %T does not contain a wrapped GolangType, path: /%s",
					typ,
					strings.Join(path[:pathIdx], "/"),
				)
				return
			}
			baseType = t
			continue
		case lang.GolangTypeWrapper:
			logger.Trace(
				"---> GolangTypeWrapper",
				"path", path[:pathIdx], "member", memberName, "object", baseType.String(), "type", fmt.Sprintf("%T", typ),
			)
			t := typ.UnwrapGolangType()
			if lo.IsNil(t) {
				err = fmt.Errorf(
					"type %T does not contain a wrapped GolangType, path: /%s",
					typ,
					strings.Join(path[:pathIdx], "/"),
				)
				return
			}
			baseType = t
			continue
		default:
			logger.Trace("---> Unknown type", "path", path[:pathIdx], "object", typ.String(), "type", fmt.Sprintf("%T", typ))
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
		logger.Trace("---> Add step", "lines", body, "varName", nextAnchor, "varValue", varValueStmts, "varType", baseType.String())

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
