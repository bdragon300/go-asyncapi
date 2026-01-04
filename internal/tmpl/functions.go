package tmpl

import (
	"errors"
	"fmt"
	"go/token"
	"maps"
	"path"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"text/template"

	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/log"
	"github.com/bdragon300/go-asyncapi/internal/render/lang"
	"github.com/bdragon300/go-asyncapi/internal/tmpl/manager"
	"github.com/bdragon300/go-asyncapi/internal/utils"
	"github.com/bdragon300/go-asyncapi/templates/codeextra"
	"github.com/samber/lo"
)

type pinnable interface {
	// Pinnable is true if the object may be pinned to generated file to be referenced from other generated code.
	Pinnable() bool
}

type UtilCodeInfo struct {
	Protocol    string
	FileName    string
	PackageName string
}

type ImplementationCodeInfo struct {
	Protocol string
	// ImplementationManifest denotes which built-in implementation manifest was used to generate implementation code.
	// Nil for ordinary files or if implementation is user-defined.
	ImplementationManifest *codeextra.ImplementationManifest
	// ImplementationConfig is configuration for the implementation code, both for built-in and user-defined implementations.
	// Nil for ordinary files.
	ImplementationConfig common.ImplementationCodeCustomOpts
}

func (i ImplementationCodeInfo) Name() string {
	if i.ImplementationManifest != nil {
		return i.ImplementationManifest.Name
	}
	return i.ImplementationConfig.Name
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
		"goComment": func(text string) string { traceCall("goComment", text); return templateGoComment(text) },
		"goPkg": func(obj common.Artifact) (pkg string, err error) {
			traceCall("goPkg", obj)
			pkg, err = importPinnedArtifact(renderManager, obj)
			if err != nil {
				return "", fmt.Errorf("%s: %w", obj, err)
			}
			return lo.Ternary(pkg != "", pkg+".", ""), nil
		},
		"goPkgUtil": func(parts ...string) (string, error) {
			traceCall("goPkgUtil", parts)
			pkg, err := importCodeExtraPackage(renderManager, parts, false)
			return lo.Ternary(pkg != "", pkg+".", ""), err
		},
		"goPkgImpl": func(parts ...string) (string, error) {
			traceCall("goPkgImpl", parts)
			pkg, err := importCodeExtraPackage(renderManager, parts, true)
			return lo.Ternary(pkg != "", pkg+".", ""), err
		},
		"goPkgRun": func(parts ...string) string {
			traceCall("goPkgRun", parts)
			pkg := importRuntimeSubpackage(renderManager, parts)
			return lo.Ternary(pkg != "", pkg+".", "")
		},
		"goPkgExt": func(parts ...string) string {
			traceCall("goPkgExt", parts)
			pkg := importExternalPackage(renderManager, parts)
			return lo.Ternary(pkg != "", pkg+".", "")
		},
		"goDef": func(r common.GolangType) (string, error) {
			traceCall("goDef", r)
			tplName := path.Join(r.GoTemplate(), "definition")
			if v, ok := r.(pinnable); ok && v.Pinnable() {
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
		"pin": func(a common.Artifact) (string, error) {
			traceCall("pin", a)
			if lo.IsNil(a) {
				return "", fmt.Errorf("cannot pin nil value")
			}
			if v, ok := a.(pinnable); !ok || !v.Pinnable() {
				return "", fmt.Errorf("type %T is not pinnable", a)
			}
			renderManager.NamespaceManager.DeclareArtifact(a, renderManager, false)
			return "", nil
		},
		"once": func(r any) any {
			traceCall("once", r)
			return templateOnce(renderManager, r)
		},

		// Other
		"impl": func(protocol string) *ImplementationCodeInfo {
			traceCall("impl", protocol)
			s, ok := getExtraCodeFileByProtocol(protocol, renderManager, true)
			if !ok {
				return nil
			}
			return &ImplementationCodeInfo{
				Protocol:               protocol,
				ImplementationManifest: s.ImplementationManifest,
				ImplementationConfig:   *s.ImplementationConfig,
			}
		},
		"utilCode": func(protocol string) *UtilCodeInfo {
			traceCall("utilCode", protocol)
			s, ok := getExtraCodeFileByProtocol(protocol, renderManager, false)
			if !ok {
				return nil
			}
			return &UtilCodeInfo{
				Protocol:    protocol,
				FileName:    s.FileName,
				PackageName: s.PackageName,
			}
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
		"runtimeExpressionCode": func(c lang.BaseRuntimeExpression, target *lang.GoStruct, addValidationCode bool) (items []runtimeExpressionCodeStep, err error) {
			traceCall("runtimeExpressionCode", c, target, addValidationCode)
			return templateRuntimeExpressionCode(renderManager, c, target, addValidationCode)
		},
		"mapping": func(v string, variantPairs ...any) (any, error) {
			traceCall("mapping", v, variantPairs)
			if len(variantPairs)%2 != 0 {
				return "", fmt.Errorf("mapping requires even number of variantPairs, got %d", len(variantPairs))
			}
			for i := 0; i < len(variantPairs); i += 2 {
				if variantPairs[i] == v {
					return variantPairs[i+1], nil
				}
			}
			return "", fmt.Errorf("unknown value %q", v)
		},
		"toList": func(v any) ([]any, error) {
			traceCall("toList", v)
			rval := reflect.ValueOf(v)
			if rval.Kind() != reflect.Slice && rval.Kind() != reflect.Array {
				return nil, fmt.Errorf("argument is not a slice or an array, got %[1]T(%[1]v)", v)
			}
			res := make([]any, rval.Len())
			for i := 0; i < rval.Len(); i++ {
				res[i] = rval.Index(i).Interface()
			}
			return res, nil
		},
		"hasKey": func(key string, m any) (bool, error) { // Overwrites sprout's hasKey to accept any mapping type
			traceCall("hasKey", key, m)
			rval := reflect.ValueOf(m)
			if rval.Kind() != reflect.Map {
				return false, fmt.Errorf("argument is not a map, got %[1]T(%[1]v)", m)
			}
			kval := reflect.ValueOf(key)
			if !kval.Type().AssignableTo(rval.Type().Key()) {
				return false, fmt.Errorf("key type %s is not assignable to map key type %s", kval.Type(), rval.Type().Key())
			}
			val := rval.MapIndex(kval)
			return val.IsValid(), nil
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
		// For example, suppose that the topObject is a lang.Ref located in `servers.myServer` that points to a
		// val which is a server object in `components.servers.reusableServer`. In this case result will be "myServer".
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
func templateGoComment(text string) string {
	if strings.HasPrefix(text, "//") || strings.HasPrefix(text, "/*") {
		// automatic formatting disabled.
		return text
	}

	var b strings.Builder
	if strings.Contains(text, "\n") {
		lo.Must(b.WriteString("/*\n"))
	} else {
		lo.Must(b.WriteString("// "))
	}
	lo.Must(b.WriteString(text))
	if strings.Contains(text, "\n") {
		if !strings.HasSuffix(text, "\n") {
			lo.Must(b.WriteString("\n"))
		}
		lo.Must(b.WriteString("*/"))
	}
	return b.String()
}

type runtimeExpressionCodeStep struct {
	CodeLines       []string
	VarName         string
	VarValue        string
	VarValueVarName string
	VarType         common.GolangType
}

// templateRuntimeExpressionCode returns the Go code that extracts the value from the targetStruct according to the
// runtime expression c. If addValidationCode is true, the result also contains the additional error handing code,
// that is typically used for property getter functions.
//
// The function returns a list of extract steps. Each step contains one or more lines of Go code and some meta information.
func templateRuntimeExpressionCode(mng *manager.TemplateRenderManager, c lang.BaseRuntimeExpression, targetStruct *lang.GoStruct, addValidationCode bool) (items []runtimeExpressionCodeStep, err error) {
	// TODO: consider also AdditionalProperties in object
	logger := log.GetLogger(log.LoggerPrefixRendering)

	field, ok := lo.Find(targetStruct.Fields, func(item lang.GoStructField) bool {
		return strings.EqualFold(item.OriginalName, string(c.StructFieldKind))
	})
	if !ok {
		return nil, fmt.Errorf("field %s not found in %s", c.StructFieldKind, targetStruct)
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
			err = fmt.Errorf("cannot unescape runtime expression, locationPath %q, item %q: %w", locationPath, locationPath[pathIdx], err)
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
				fmtErrorf := importExternalPackage(mng, []string{"fmt"}) + ".Errorf"
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
				fmtErrorf := importExternalPackage(mng, []string{"fmt"}) + ".Errorf"
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
		item := runtimeExpressionCodeStep{
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

// getExtraCodeFileByProtocol returns the FileRenderState for the given protocol and implementation flag. If not found, returns false.
func getExtraCodeFileByProtocol(protocol string, renderManager *manager.TemplateRenderManager, isImplementation bool) (manager.FileRenderState, bool) {
	states := slices.SortedStableFunc(maps.Values(renderManager.CommittedStates()), func(a, b manager.FileRenderState) int {
		return strings.Compare(a.FileName, b.FileName)
	})
	// Assuming all extra code files has the flat structure per protocol, and we may pick any file
	return lo.Find(states, func(state manager.FileRenderState) bool {
		return state.ExtraCodeProtocol == protocol && isImplementation == !lo.IsNil(state.ImplementationConfig)
	})
}

// importExternalPackage imports the external package and returns its alias or package name as prefix to prepend to the
// imported object in the generated code.
func importExternalPackage(mng *manager.TemplateRenderManager, parts []string) string {
	_, pkgPath, pkgName := getImportPath(parts, false)
	return mng.ImportsManager.AddImport(pkgPath, pkgName)
}

// importRuntimeSubpackage imports the given runtime subpackage and returns its alias or package name as prefix to prepend to the
// imported object in the generated code.
func importRuntimeSubpackage(mng *manager.TemplateRenderManager, parts []string) string {
	p := append([]string{mng.RenderOpts.RuntimeModule}, parts...)
	return importExternalPackage(mng, p)
}

func importCodeExtraPackage(mng *manager.TemplateRenderManager, parts []string, isImplementation bool) (string, error) {
	protocol, middle, pkg := getImportPath(parts, true)
	state, found := getExtraCodeFileByProtocol(protocol, mng, isImplementation)
	if !found {
		return "", ErrNotPinned
	}

	// Check if the import is current package (assuming the directory is equal to package)
	if path.Dir(state.FileName) == path.Dir(mng.FileName) && pkg == "" {
		return "", nil // Import from the current package
	}

	pkgPath := path.Join(mng.RenderOpts.ImportBase, path.Dir(state.FileName), middle, pkg)
	return mng.ImportsManager.AddImport(pkgPath, state.PackageName), nil
}

// importPinnedArtifact imports the pinned artifact into the current file and returns imported package alias or name.
// If artifact is pinned to the current package, returns empty string. If artifact has not been pinned yet, returns ErrNotPinned.
func importPinnedArtifact(mng *manager.TemplateRenderManager, obj common.Artifact) (string, error) {
	d, found := mng.NamespaceManager.FindArtifact(obj)
	if !found {
		if v, ok := obj.(pinnable); ok && v.Pinnable() {
			return "", ErrNotPinned
		}
		return "", nil // Type is not supposed to be found in the generated code (e.g. Go built-in types)
	}

	// Files in the same directory (i.e. the same package) does not need to be imported
	if path.Dir(d.FileName) == path.Dir(mng.FileName) {
		return "", nil
	}

	pkgPath := path.Join(mng.RenderOpts.ImportBase, path.Dir(d.FileName))
	return mng.ImportsManager.AddImport(pkgPath, d.PackageName), nil
}

// getImportPath joins path parts into an import path and splits the result into three components:
// first part, middle part, and last part. If split is true, the first and last parts are removed from the middle part.
func getImportPath(parts []string, split bool) (first string, middle string, last string) {
	if len(parts) == 0 {
		panic("import path cannot be empty")
	}

	p := path.Join(parts...)
	parts = strings.Split(p, "/")
	if len(parts) > 0 {
		first = parts[0]
		if split {
			parts = parts[1:]
		}
	}
	if len(parts) > 0 {
		last = parts[len(parts)-1]
		if split {
			parts = parts[:len(parts)-1]
		}
	}
	middle = path.Join(parts...)

	return
}
