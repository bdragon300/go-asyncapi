package tmpl

import (
	"errors"
	"fmt"
	"go/token"
	"path"
	"strconv"
	"strings"
	"text/template"

	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/log"
	"github.com/bdragon300/go-asyncapi/internal/render/lang"
	"github.com/bdragon300/go-asyncapi/internal/tmpl/manager"
	"github.com/bdragon300/go-asyncapi/internal/utils"
	"github.com/samber/lo"
)

type pinnable interface {
	// Pinnable is true if the object may be pinned to a specific file in the generated code and then be referenced
	// from other files.
	Pinnable() bool
}

// GetTemplateFunctions returns a map of functions to use in templates. These functions include all
// [github.com/go-sprout/sprout] functions and go-asyncapi specific functions.
func GetTemplateFunctions(renderManager *manager.TemplateRenderManager) template.FuncMap {
	type golangWrapperType interface {
		UnwrapGolangType() common.GolangType
	}

	logger := log.GetLogger(log.LoggerPrefixRendering)
	traceCall := func(funcName string, args ...any) {
		if logger.GetLevel() > log.TraceLevel {
			return
		}
		argsStr := strings.Join(lo.Map(args, func(arg any, _ int) string {
			return fmt.Sprintf("%[1]v", arg)
		}), "; ")
		logger.Debugf("--> %s(%s)", funcName, argsStr)
	}

	extraFuncs := template.FuncMap{
		// go* functions return Go code snippets
		"goLit":     func(val any) (string, error) { traceCall("goLit", val); return templateGoLit(renderManager, val) },
		"goIDUpper": func(val any) string { traceCall("goIDUpper", val); return templateGoID(renderManager, val, true) },
		"goID":      func(val any) string { traceCall("goID", val); return templateGoID(renderManager, val, false) },
		"goComment": func(text string) (string, error) { traceCall("goComment", text); return templateGoComment(text) },
		"goQual":    func(parts ...string) string { traceCall("goQual", parts); return templateGoQual(renderManager, parts...) },
		"goQualR": func(parts ...string) string {
			traceCall("goQualR", parts)
			return templateGoQualRuntime(renderManager, parts...)
		},
		"goDef": func(r common.GolangType) (string, error) {
			traceCall("goDef", r)
			tplName := path.Join(r.GoTemplate(), "definition")
			if _, ok := r.(pinnable); ok {
				renderManager.NamespaceManager.DeclareArtifact(r, renderManager, true)
			} else if logger.GetLevel() <= log.TraceLevel {
				logger.Debug("---> goDef: skip pinning due to object is not pinnable")
			}
			if v, ok := r.(golangReferenceType); ok {
				r = v.DerefGolangType()
			}
			res, err := templateExecTemplate(renderManager, tplName, r)
			if err != nil {
				return "", fmt.Errorf("%s: %w", r, err)
			}
			return res, nil
		},
		"goPkg": func(obj any) (pkg string, err error) {
			traceCall("goPkg", obj)
			switch v := obj.(type) {
			case common.Artifact:
				pkg, err = qualifiedTypeGeneratedPackage(renderManager, v)
			case *common.ImplementationObject:
				if lo.IsNil(v) {
					return "", errors.New("argument is nil")
				}
				pkg, err = qualifiedImplementationGeneratedPackage(renderManager, *v)
			default:
				return "", fmt.Errorf("type is not supported %[1]T: %[1]v", obj)
			}

			if err != nil {
				return "", fmt.Errorf("%s: %w", obj, err)
			}
			return lo.Ternary(pkg != "", pkg+".", ""), nil
		},
		"goUsage": func(r common.GolangType) (string, error) {
			traceCall("goUsage", r)
			return templateGoUsage(renderManager, r)
		},

		// Artifact helpers
		"innerType": func(val common.GolangType) common.GolangType {
			traceCall("innerType", val)
			if v, ok := any(val).(golangWrapperType); ok {
				return v.UnwrapGolangType()
			}
			return nil
		},
		"isVisible": func(a common.Artifact) common.Artifact {
			traceCall("isVisible", a)
			return lo.Ternary(!lo.IsNil(a) && a.Visible(), a, nil)
		},

		// Call templates dynamically
		"tmpl": func(templateName string, ctx any) (string, error) {
			traceCall("tmpl", templateName, ctx)
			return templateExecTemplate(renderManager, templateName, ctx)
		},
		"tryTmpl": func(templateName string, ctx any) (string, error) {
			traceCall("tryTmpl", templateName, ctx)
			res, err := templateExecTemplate(renderManager, templateName, ctx)
			switch {
			case errors.Is(err, ErrTemplateNotFound):
				return "", nil
			case err != nil:
				return "", err
			}

			return res, nil
		},

		// Working with template namespace
		"pin": func(object common.Artifact) (string, error) {
			traceCall("pin", object)
			if lo.IsNil(object) {
				return "", fmt.Errorf("cannot pin nil")
			}
			if _, ok := object.(pinnable); !ok {
				return "", fmt.Errorf("type %T is not pinnable", object)
			}
			renderManager.NamespaceManager.DeclareArtifact(object, renderManager, false)
			return "", nil
		},
		"once": func(r any) any {
			traceCall("once", r)
			return templateOnce(renderManager, r)
		},

		// Other
		"impl": func(protocol string) *common.ImplementationObject {
			traceCall("impl", protocol)
			impl, found := lo.Find(renderManager.Implementations, func(def manager.ImplementationItem) bool {
				return def.Object.Manifest.Protocol == protocol
			})
			if !found {
				return nil
			}
			return &impl.Object
		},
		"toQuotable": func(s string) string {
			traceCall("toQuotable", s)
			return strings.TrimSuffix(strings.TrimPrefix(strconv.Quote(s), "\""), "\"")
		},
		"ellipsisStart": func(maxlen int, s string) string {
			traceCall("ellipsisStart", maxlen, s)
			if len(s) <= maxlen {
				return s
			}
			if maxlen <= 3 {
				return strings.Repeat(".", maxlen)
			}
			return "..." + s[len(s)-(maxlen-3):]
		},
		"debug": func(args ...any) string {
			for _, arg := range args {
				logger.Debugf("debug: [%[1]p][%[1]T] %[1]v", arg)
			}
			return ""
		},
		"runtimeExpressionCode": func(c lang.BaseRuntimeExpression, varStruct *lang.GoStruct, addValidationCode bool) (items []correlationIDExtractionStep, err error) {
			traceCall("runtimeExpressionCode", c, varStruct, addValidationCode)
			return templateCorrelationIDExtractionCode(renderManager, c, varStruct, addValidationCode)
		},
	}

	return lo.Assign(sproutFunctions, extraFuncs)
}

// templateOnce adds the given o to the namespace if it is not already added. Return the object if it was not added before,
// or nil if it was already added.
//
// The purpose of this function is similar to [sync.Once], but for template rendering -- to ensure that certain code snippets
// (e.g. type definitions) are rendered only once even if the template is included multiple times.
func templateOnce(mng *manager.TemplateRenderManager, o any) any {
	if lo.IsNil(o) {
		return nil
	}
	if mng.NamespaceManager.ContainsObject(o) {
		return nil
	}
	mng.NamespaceManager.AddObject(o)
	return o
}

type golangReferenceType interface {
	DerefGolangType() common.GolangType
}

// templateGoUsage returns a Go code snippet that represents the usage of the given Go type. If this type is defined
// in other module, the necessary import is also added to the current file and the returned value contains the package
// name as well. If the type is not defined yet, it returns ErrNotPinned.
//
// Type usage snippet uses for example in function parameters of this type, variable definitions of this type, etc.
//
// For example, consider for the [lang.GoStruct] object representing this struct
//
//	type MyStruct struct {
//	    Field1 string
//	    Field2 int
//	}
//
// the function returns "MyStruct". Or "pkg.MyStruct" if the struct is defined in “github.com/path/to/pkg” module.
func templateGoUsage(mng *manager.TemplateRenderManager, r common.GolangType) (string, error) {
	tplName := path.Join(r.GoTemplate(), "usage")
	if v, ok := r.(golangReferenceType); ok {
		r = v.DerefGolangType()
	}
	res, err := templateExecTemplate(mng, tplName, r)
	if err != nil {
		return "", fmt.Errorf("%s: %w", r, err)
	}
	return res, nil
}

// templateExecTemplate executes the template with the given name and context from other template. This differs from
// “template” directive in that it can receive dynamic template name.
func templateExecTemplate(mng *manager.TemplateRenderManager, templateName string, ctx any) (string, error) {
	var bld strings.Builder

	tpl, err := mng.TemplateLoader.LoadTemplate(templateName)
	if err != nil {
		return "", fmt.Errorf("template %q: %w", templateName, err)
	}
	if err := tpl.Execute(&bld, ctx); err != nil {
		return "", fmt.Errorf("template %q: %w", templateName, err)
	}
	return bld.String(), nil
}

// templateGoLit returns a Go literal for the given value.
func templateGoLit(mng *manager.TemplateRenderManager, val any) (string, error) {
	if v, ok := val.(common.GolangType); ok {
		return templateGoUsage(mng, v)
	}
	return toGoLiteral(val), nil
}

// toGoLiteral returns a Go code string with literal for the given value.
func toGoLiteral(val any) string {
	var res string
	switch val.(type) {
	case bool, string, int, complex128:
		// default constant types can be left bare
		return fmt.Sprintf("%#v", val)
	case float64:
		res = fmt.Sprintf("%#v", val)
		if !strings.Contains(res, ".") && !strings.Contains(res, "e") {
			// If the formatted value is not in scientific notation, and does not have a dot, then
			// we add ".0". Otherwise, it will be interpreted as an int.
			// See: https://github.com/golang/go/issues/26363
			res += ".0"
		}
		return res
	case float32, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr:
		// other built-in types need specific type info
		return fmt.Sprintf("%T(%#v)", val, val)
	case complex64:
		// fmt package already renders parenthesis for complex64
		return fmt.Sprintf("%T%#v", val, val)
	}

	panic(fmt.Sprintf("unsupported type for literal: %T", val))
}

// templateGoID returns a Go identifier for the given value. The exportedName argument controls if the returned
// identifier should be exported (start with an uppercase letter) or not.
//
// If value is an object (common.GolangType or lang.Ref or any common.Artifact object like channel or parameter),
// the function returns its name as Go identifier.
// Considers the document references names, aliases, x-go-name tags and so on.
// For that, it calls object's Name method and make a Go identifier from it.
//
// If value is a string, the function just makes a Go identifier.
func templateGoID(mng *manager.TemplateRenderManager, val any, exportedName bool) string {
	var res string

	switch v := val.(type) {
	case common.Artifact:
		// Prefer the name of the topObject over the name of the val if the topObject is a Ref and points to val.
		// Otherwise, use the name of the val.
		//
		// For example, the topObject is a lang.Ref defined in `servers.myServer`. val contains the render.Server
		// defined in `components.servers.reusableServer` that this Ref is points to. Then we'll use the "myServer"
		// as the server name in generated code: functions, structs, etc.
		topObject := mng.CurrentObject
		if lo.IsNil(topObject) || !common.CheckSameArtifacts(topObject, v) {
			res = v.Name() // nil could appear when we render the app template
		} else {
			res = topObject.Name()
		}
	case string:
		res = v
	default:
		panic(fmt.Sprintf("type is not supported %[1]T: %[1]v", val))
	}

	if res == "" {
		return ""
	}
	exported := true
	if !exportedName {
		exported = token.IsExported(res)
	}
	return utils.ToGolangName(res, exported)
}

// templateGoComment returns a Go comment for the given text. If the text contains newlines, it is formatted as a block
// comment. Otherwise, it is formatted as a line comment.
func templateGoComment(text string) (string, error) {
	if strings.HasPrefix(text, "//") || strings.HasPrefix(text, "/*") {
		// automatic formatting disabled.
		return text, nil
	}

	var b strings.Builder
	if strings.Contains(text, "\n") {
		if _, err := b.WriteString("/*\n"); err != nil {
			return "", err
		}
	} else {
		if _, err := b.WriteString("// "); err != nil {
			return "", err
		}
	}
	if _, err := b.WriteString(text); err != nil {
		return "", err
	}
	if strings.Contains(text, "\n") {
		if !strings.HasSuffix(text, "\n") {
			if _, err := b.WriteString("\n"); err != nil {
				return "", err
			}
		}
		if _, err := b.WriteString("*/"); err != nil {
			return "", err
		}
	}
	return b.String(), nil
}

type correlationIDExtractionStep struct {
	CodeLines       []string
	VarName         string
	VarValue        string
	VarValueVarName string
	VarType         common.GolangType
}

// templateCorrelationIDExtractionCode generates Go code to extract a variable from a struct for the correlation id
// getter and setter method.
//
// The function returns a list of extract steps. Each step contains one or more lines of Go code and some meta information.
func templateCorrelationIDExtractionCode(mng *manager.TemplateRenderManager, c lang.BaseRuntimeExpression, varStruct *lang.GoStruct, addValidationCode bool) (items []correlationIDExtractionStep, err error) {
	// TODO: consider also AdditionalProperties in object
	logger := log.GetLogger(log.LoggerPrefixRendering)

	field, ok := lo.Find(varStruct.Fields, func(item lang.GoStructField) bool {
		return strings.EqualFold(item.OriginalName, string(c.StructFieldKind))
	})
	if !ok {
		return nil, fmt.Errorf("field %s not found in %s", c.StructFieldKind, varStruct)
	}

	locationPath := c.LocationPath
	baseType := field.Type
	for pathIdx := 0; pathIdx < len(locationPath); {
		var body []string
		var varValueStmts string

		// Anchor is a variable that holds the current value of the locationPath item
		anchor := fmt.Sprintf("v%d", pathIdx)
		nextAnchor := fmt.Sprintf("v%d", pathIdx+1)

		memberName, err2 := unescapeJSONPointerFragmentPart(locationPath[pathIdx])
		if err2 != nil {
			err = fmt.Errorf("cannot unescape CorrelationID locationPath %q, item %q: %w", locationPath, locationPath[pathIdx], err)
			return
		}

		switch typ := baseType.(type) {
		case *lang.GoStruct:
			logger.Trace("---> GoStruct", "locationPath", locationPath[:pathIdx], "member", memberName, "object", typ.String())
			fld, ok := lo.Find(typ.Fields, func(item lang.GoStructField) bool { return item.MarshalName == memberName })
			if !ok {
				err = fmt.Errorf(
					"field %q not found in struct %s, locationPath: /%s",
					memberName, typ.OriginalName, strings.Join(locationPath[:pathIdx], "/"),
				)
				return
			}
			varValueStmts = fmt.Sprintf("%s.%s", anchor, fld.OriginalName)
			baseType = fld.Type
			body = []string{fmt.Sprintf("%s := %s", nextAnchor, varValueStmts)}
		case *lang.GoMap:
			logger.Trace("---> GoMap", "locationPath", locationPath[:pathIdx], "member", memberName, "object", typ.String())
			varValueStmts = fmt.Sprintf("%s[%s]", anchor, toGoLiteral(memberName))
			baseType = typ.ValueType
			// TODO: replace templateGoUsage calls to smth another to remove import from impl, this is a potential circular import
			varExpr := fmt.Sprintf("var %s %s", nextAnchor, lo.Must(templateGoUsage(mng, typ.ValueType)))
			if typ.ValueType.CanBeAddressed() {
				// Append ` = new(TYPE)` to initialize a pointer
				varExpr += fmt.Sprintf(" = new(%s)", lo.Must(templateGoUsage(mng, typ.ValueType)))
			}

			ifExpr := fmt.Sprintf(`if v, ok := %s; ok {
				%s = v
			}`, varValueStmts, nextAnchor)
			if addValidationCode {
				fmtErrorf := templateGoQual(mng, "fmt", "Errorf")
				ifExpr += fmt.Sprintf(` else {
					err = %s("key %%q not found in map on locationPath /%s", %s)
					return
				}`, fmtErrorf, strings.Join(locationPath[:pathIdx], "/"), toGoLiteral(memberName))
			}
			body = []string{
				fmt.Sprintf(`if %s == nil { 
					%s = make(%s) 
				}`, anchor, anchor, lo.Must(templateGoUsage(mng, typ))),
				varExpr,
				ifExpr,
			}
		case *lang.GoArray:
			logger.Trace("---> GoArray", "locationPath", locationPath[:pathIdx], "member", memberName, "object", typ.String())
			if _, ok := memberName.(string); ok {
				err = fmt.Errorf(
					"index %q is not a number, array %s, locationPath: /%s",
					memberName,
					typ.OriginalName,
					strings.Join(locationPath[:pathIdx], "/"),
				)
				return
			}
			if addValidationCode {
				fmtErrorf := templateGoQual(mng, "fmt", "Errorf")
				body = append(body, fmt.Sprintf(`if len(%s) <= %s {
					err = %s("index %%q is out of range in array of length %%d on locationPath /%s", %s, len(%s))
					return
				}`, anchor, toGoLiteral(memberName), fmtErrorf, strings.Join(locationPath[:pathIdx], "/"), toGoLiteral(memberName), anchor))
			}
			varValueStmts = fmt.Sprintf("%s[%s]", anchor, toGoLiteral(memberName))
			baseType = typ.ItemsType
			body = append(body, fmt.Sprintf("%s := %s", nextAnchor, varValueStmts))
		case *lang.GoSimple: // Should be a terminal type in chain, raise error otherwise (if any locationPath parts left to resolve)
			logger.Trace("---> GoSimple", "locationPath", locationPath[:pathIdx], "member", memberName, "object", typ.String())
			if pathIdx >= len(locationPath)-1 { // Primitive types should get addressed by the last locationPath item
				err = fmt.Errorf(
					"primitive type %q does not contain any fields, locationPath: /%s",
					typ.Name(),
					strings.Join(locationPath[:pathIdx], "/"),
				)
				return
			}
			baseType = typ
		case lang.GolangWrappedType:
			logger.Trace(
				"---> GolangWrappedType",
				"locationPath", locationPath[:pathIdx], "member", memberName, "object", baseType.String(), "type", fmt.Sprintf("%T", typ),
			)
			t := typ.UnwrapGolangType()
			if lo.IsNil(t) {
				err = fmt.Errorf(
					"wrapper type %T contains nil, locationPath: /%s",
					typ,
					strings.Join(locationPath[:pathIdx], "/"),
				)
				return
			}
			baseType = t
			continue
		case lang.GolangReferenceType:
			logger.Trace(
				"---> GolangReferenceType",
				"locationPath", locationPath[:pathIdx], "member", memberName, "object", baseType.String(), "type", fmt.Sprintf("%T", typ),
			)
			t := typ.DerefGolangType()
			if lo.IsNil(t) {
				err = fmt.Errorf(
					"reference type %T contains nil, locationPath: /%s",
					typ,
					strings.Join(locationPath[:pathIdx], "/"),
				)
				return
			}
			baseType = t
			continue
		default:
			logger.Trace("---> Unknown type", "locationPath", locationPath[:pathIdx], "object", typ.String(), "type", fmt.Sprintf("%T", typ))
			err = fmt.Errorf(
				"type %s is not addressable, locationPath: /%s",
				typ.Name(),
				strings.Join(locationPath[:pathIdx], "/"),
			)
			return
		}

		pathIdx++
		item := correlationIDExtractionStep{
			CodeLines:       body,
			VarName:         nextAnchor,
			VarValue:        varValueStmts,
			VarValueVarName: anchor,
			VarType:         baseType,
		}
		logger.Trace("---> Add step", "lines", body, "varName", nextAnchor, "varValue", varValueStmts, "varType", baseType.String())

		items = append(items, item)
	}

	return
}

// qualifiedTypeGeneratedPackage returns the package name or alias of module where the object is defined to use this name
// further in the generated code. If object is already defined in *current module*, returns empty string with no error.
// If we don't know where the object is defined, returns ErrNotPinned.
func qualifiedTypeGeneratedPackage(mng *manager.TemplateRenderManager, obj common.Artifact) (string, error) {
	d, found := mng.NamespaceManager.FindArtifact(obj)
	if !found {
		if v, ok := obj.(pinnable); ok && v.Pinnable() {
			return "", ErrNotPinned
		}
		return "", nil // Type is not supposed to be found in the generated code (e.g. Go built-in types)
	}

	// Use the package path from reuse config if it is found
	if d.Layout.ReusePackagePath != "" {
		return mng.ImportsManager.AddImport(d.Layout.ReusePackagePath, ""), nil
	}

	// Check if the object is found in the same directory (assuming the directory is equal to package)
	nsTypeFileDir := path.Dir(d.FileName)
	if nsTypeFileDir == path.Dir(mng.FileName) {
		return "", nil // Object is found in the current package, its name doesn't require a package name
	}

	pkgPath := path.Join(mng.RenderOpts.ImportBase, nsTypeFileDir)
	return mng.ImportsManager.AddImport(pkgPath, d.PackageName), nil
}

func qualifiedImplementationGeneratedPackage(mng *manager.TemplateRenderManager, obj common.ImplementationObject) (string, error) {
	defInfo, found := lo.Find(mng.Implementations, func(def manager.ImplementationItem) bool {
		return def.Object == obj
	})
	if !found {
		return "", ErrNotPinned
	}

	// Use the package path from reuse config if it is defined
	if defInfo.Object.Config.ReusePackagePath != "" {
		return mng.ImportsManager.AddImport(defInfo.Object.Config.ReusePackagePath, ""), nil
	}

	if defInfo.Directory == path.Dir(mng.FileName) {
		return "", nil // Object is defined in the current package, its name doesn't require a package name
	}

	pkgPath := path.Join(mng.RenderOpts.ImportBase, defInfo.Directory)
	return mng.ImportsManager.AddImport(pkgPath, defInfo.Object.Config.Package), nil
}

// templateGoQual returns a qualified name of the object in the generated code. Adds the import to the current file
// if needed.
//
// Receives the import path and the object name in format “path/to/package.name“. For example, “net/url.URL” or
// “golang.org/x/net/ipv4.Conn“. This could be a single string or a sequence of strings that are joined together.
//
// Returns the qualified name of the object that is used to access it in the generated code. For example, “url.URL“ or
// “ipv4.Conn“ for the examples above.
func templateGoQual(mng *manager.TemplateRenderManager, parts ...string) string {
	pkgPath, pkgName, n := qualifiedToImport(parts)
	return fmt.Sprintf("%s.%s", mng.ImportsManager.AddImport(pkgPath, pkgName), n)
}

// templateGoQualRuntime returns a qualified name of the object in the generated code. Adds the import to the runtime.
// Works like templateGoQual but prepends the runtime package to the import path.
func templateGoQualRuntime(mng *manager.TemplateRenderManager, parts ...string) string {
	p := append([]string{mng.RenderOpts.RuntimeModule}, parts...)
	return templateGoQual(mng, p...)
}
