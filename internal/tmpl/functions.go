package tmpl

import (
	"errors"
	"fmt"
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/log"
	"github.com/bdragon300/go-asyncapi/internal/render"
	"github.com/bdragon300/go-asyncapi/internal/render/lang"
	"github.com/bdragon300/go-asyncapi/internal/tmpl/manager"
	"github.com/bdragon300/go-asyncapi/internal/utils"
	"github.com/samber/lo"
	"path"
	"strings"
	"text/template"
	"unicode"
)

func GetTemplateFunctions(renderManager *manager.TemplateRenderManager) template.FuncMap {
	type golangTypeExtractor interface {
		InnerGolangType() common.GolangType
	}

	extraFuncs := template.FuncMap{
		// Functions that return go code as string
		"golit": func(val any) (string, error) { return templateGoLit(val) },
		"goid": func(val any) string { return templateGoID(renderManager, val, true) },
		"goidorig": func(val any) string { return templateGoID(renderManager, val, false) },
		"gocomment": func(text string) (string, error) { return templateGoComment(text) },
		"goqual": func(parts ...string) string { return qualifiedName(renderManager, parts...) },
		"goqualrun": func(parts ...string) string { return qualifiedRuntimeName(renderManager, parts...) },
		"godef": func(r common.GolangType) (string, error) {
			tplName := path.Join(r.GoTemplate(), "definition")
			renderManager.NamespaceManager.AddType(r, renderManager.CurrentSelection, true)
			if v, ok := r.(golangTypeWrapper); ok {
				r = v.UnwrapGolangType()
			}
			res, err := templateExecTemplate(tplName, r)
			if err != nil {
				return "", fmt.Errorf("%s: %w", r, err)
			}
			return res, nil
		},
		"gopkg": func(obj any) (pkg string, err error) {
			switch v := obj.(type) {
			case common.GolangType:
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
			return lo.Ternary(pkg != "", pkg + ".", ""), nil
		},
		"gousage": func(r common.GolangType) (string, error) { return templateGoUsage(r) },

		// Type helpers
		"deref": func(r common.Renderable) common.Renderable {
			if r == nil {
				return nil
			}
			return common.DerefRenderable(r)
		},
		"innertype": func(val common.GolangType) common.GolangType {
			if v, ok := any(val).(golangTypeExtractor); ok {
				return v.InnerGolangType()
			}
			return nil
		},
		"visible": func(r common.Renderable) common.Renderable {
			return lo.Ternary(!lo.IsNil(r) && r.Visible(), r, nil)
		},
		"ptr": func(val common.GolangType) (common.GolangType, error) {
			if lo.IsNil(val) {
				return nil, fmt.Errorf("cannot get a pointer to nil")
			}
			return &lang.GoPointer{Type: val}, nil
		},
		"impl": func(protocol string) *common.ImplementationObject {
			impl, found := lo.Find(renderManager.Implementations, func(def manager.ImplementationItem) bool {
				return def.Object.Manifest.Protocol == protocol
			})
			if !found {
				return nil
			}
			return &impl.Object
		},

		// Templates calling
		"tmpl": func(templateName string, ctx any) (string, error) {
			return templateExecTemplate(templateName, ctx)
		},
		"trytmpl": func(templateName string, ctx any) (string, error) {
			res, err := templateExecTemplate(templateName, ctx)
			switch {
			case errors.Is(err, ErrTemplateNotFound):
				return "", nil
			case err != nil:
				return "", err
			}

			return res, nil
		},

		// Working with render namespace
		"def": func(objects ...any) string {
			for _, o := range objects {
				switch v := o.(type) {
				case common.GolangType:
					if !lo.IsNil(o) {
						renderManager.NamespaceManager.AddType(v, renderManager.CurrentSelection, false)
					}
				case string:
					if o != "" {
						renderManager.NamespaceManager.AddName(v)
					}
				}
			}
			return ""
		},
		"defined": func(r any) bool {
			return templateGoDefined(renderManager, r)
		},
		"ndefined": func(r any) bool {
			return !templateGoDefined(renderManager, r)
		},

		// Other
		"correlationIDExtractionCode": func(c *render.CorrelationID, varStruct *lang.GoStruct, addValidationCode bool) (items []correlationIDExtractionStep, err error) {
			return templateCorrelationIDExtractionCode(renderManager, c, varStruct, addValidationCode)
		},
		"debug": func(args ...any) string {
			logger := log.GetLogger(log.LoggerPrefixRendering)
			for _, arg := range args {
				logger.Debugf("debug: [%[1]p] %[1]v", arg)
			}
			return ""
		},
	}

	return lo.Assign(sproutFunctions, extraFuncs)
}

func templateGoDefined(mng *manager.TemplateRenderManager, r any) bool {
	if lo.IsNil(r) {
		return false
	}

	switch v := r.(type) {
	case common.GolangType:
		o, found := mng.NamespaceManager.FindType(v)
		return found && o.Actual
	case string:
		return mng.NamespaceManager.FindName(v)
	}

	panic(fmt.Sprintf("unsupported type %[1]T: %[1]v", r))
}

type golangTypeWrapper interface {
	UnwrapGolangType() common.GolangType
}

func templateGoUsage(r common.GolangType) (string, error) {
	tplName := path.Join(r.GoTemplate(), "usage")
	if v, ok := r.(golangTypeWrapper); ok {
		r = v.UnwrapGolangType()
	}
	res, err := templateExecTemplate(tplName, r)
	if err != nil {
		return "", fmt.Errorf("%s: %w", r, err)
	}
	return res, nil
}

func templateExecTemplate(templateName string, ctx any) (string, error) {
	var bld strings.Builder

	tpl, err := LoadTemplate(templateName)
	if err != nil {
		return "", fmt.Errorf("template %q: %w", templateName, err)
	}
	if err := tpl.Execute(&bld, ctx); err != nil {
		return "", fmt.Errorf("template %q: %w", templateName, err)
	}
	return bld.String(), nil
}

func templateGoLit(val any) (string, error) {
	if v, ok := val.(common.GolangType); ok {
		return templateGoUsage(v)
	}
	return utils.ToGoLiteral(val), nil
}

func templateGoID(mng *manager.TemplateRenderManager, val any, forceCapitalize bool) string {
	var res string

	switch v := val.(type) {
	case common.Renderable:
		// Prefer the name of the topObject over the name of the val if the topObject is a Ref and points to val.
		// Otherwise, use the name of the val.
		//
		// For example, the topObject is a lang.Ref defined in `servers.myServer`. val contains the render.Server
		// defined in `components.servers.reusableServer` that this Ref is points to. Then we'll use the "myServer"
		// as the server name in generated code: functions, structs, etc.
		topObject := mng.CurrentObject
		res = lo.Ternary(common.CheckSameRenderables(topObject, v), topObject.Name(), v.Name())
	case string:
		res = v
	default:
		panic(fmt.Sprintf("type is not supported %[1]T: %[1]v", val))
	}

	if res == "" {
		return ""
	}
	exported := true
	if !forceCapitalize {
		exported = unicode.IsUpper(rune(res[0]))
	}
	return utils.ToGolangName(res, exported)
}

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
	CodeLines []string
	VarName         string
	VarValue        string
	VarValueVarName string
	VarType         common.GolangType
}

func templateCorrelationIDExtractionCode(mng *manager.TemplateRenderManager, c *render.CorrelationID, varStruct *lang.GoStruct, addValidationCode bool) (items []correlationIDExtractionStep, err error) {
	// TODO: consider also AdditionalProperties in object
	logger := log.GetLogger(log.LoggerPrefixRendering)

	field, ok := lo.Find(varStruct.Fields, func(item lang.GoStructField) bool {
		return strings.ToLower(item.Name) == strings.ToLower(string(c.StructFieldKind))
	})
	if !ok {
		return nil, fmt.Errorf("field %s not found in %s", c.StructFieldKind, varStruct)
	}

	locationPath := c.LocationPath
	baseType := field.Type
	for pathIdx:=0; pathIdx < len(locationPath); pathIdx++ {
		var body []string
		var varValueStmts string

		// Anchor is a variable that holds the current value of the locationPath item
		anchor := fmt.Sprintf("v%d", pathIdx)
		nextAnchor := fmt.Sprintf("v%d", pathIdx+1)

		memberName, err2 := unescapeCorrelationIDPathItem(locationPath[pathIdx])
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
			varValueStmts = fmt.Sprintf("%s.%s", anchor, fld.Name)
			baseType = fld.Type
			body = []string{fmt.Sprintf("%s := %s", nextAnchor, varValueStmts)}
		case *lang.GoMap:
			// TODO: x-parameter in correlationIDs spec section to set numbers as "0" for string keys or 0 for int keys
			logger.Trace("---> GoMap", "locationPath", locationPath[:pathIdx], "member", memberName, "object", typ.String())
			varValueStmts = fmt.Sprintf("%s[%s]", anchor, utils.ToGoLiteral(memberName))
			baseType = typ.ValueType
			// TODO: replace templateGoUsage calls to smth another to remove import from impl, this is a potential circular import
			varExpr := fmt.Sprintf("var %s %s", nextAnchor, lo.Must(templateGoUsage(typ.ValueType)))
			if typ.ValueType.Addressable() {
				// Append ` = new(TYPE)` to initialize a pointer
				varExpr += fmt.Sprintf(" = new(%s)", lo.Must(templateGoUsage(typ.ValueType)))
			}

			ifExpr := fmt.Sprintf(`if v, ok := %s; ok {
				%s = v
			}`, varValueStmts, nextAnchor)
			if addValidationCode {
				fmtErrorf := qualifiedName(mng, "fmt", "Errorf")
				ifExpr += fmt.Sprintf(` else {
					err = %s("key %%q not found in map on locationPath /%s", %s)
					return
				}`, fmtErrorf, strings.Join(locationPath[:pathIdx], "/"), utils.ToGoLiteral(memberName))
			}
			body = []string{
				fmt.Sprintf(`if %s == nil { 
					%s = make(%s) 
				}`, anchor, anchor, lo.Must(templateGoUsage(typ))),
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
				fmtErrorf := qualifiedName(mng, "fmt", "Errorf")
				body = append(body, fmt.Sprintf(`if len(%s) <= %s {
					err = %s("index %%q is out of range in array of length %%d on locationPath /%s", %s, len(%s))
					return
				}`, anchor, utils.ToGoLiteral(memberName), fmtErrorf, strings.Join(locationPath[:pathIdx], "/"), utils.ToGoLiteral(memberName), anchor))
			}
			varValueStmts = fmt.Sprintf("%s[%s]", anchor, utils.ToGoLiteral(memberName))
			baseType = typ.ItemsType
			body = append(body, fmt.Sprintf("%s := %s", nextAnchor, varValueStmts))
		case *lang.GoSimple: // Should be a terminal type in chain, raise error otherwise (if any locationPath parts left to resolve)
			logger.Trace("---> GoSimple", "locationPath", locationPath[:pathIdx], "member", memberName, "object", typ.String())
			if pathIdx >= len(locationPath)-1 { // Primitive types should get addressed by the last locationPath item
				err = fmt.Errorf(
					"type %q cannot be resolved further, locationPath: /%s",
					typ.Name(), // TODO: check if this is correct
					strings.Join(locationPath[:pathIdx], "/"),
				)
				return
			}
			baseType = typ
		case lang.GolangTypeExtractor:
			logger.Trace(
				"---> GolangTypeExtractor",
				"locationPath", locationPath[:pathIdx], "member", memberName, "object", baseType.String(), "type", fmt.Sprintf("%T", typ),
			)
			t := typ.InnerGolangType()
			if lo.IsNil(t) {
				err = fmt.Errorf(
					"type %T does not contain a wrapped GolangType, locationPath: /%s",
					typ,
					strings.Join(locationPath[:pathIdx], "/"),
				)
				return
			}
			baseType = t
			continue
		case lang.GolangTypeWrapper:
			logger.Trace(
				"---> GolangTypeWrapper",
				"locationPath", locationPath[:pathIdx], "member", memberName, "object", baseType.String(), "type", fmt.Sprintf("%T", typ),
			)
			t := typ.UnwrapGolangType()
			if lo.IsNil(t) {
				err = fmt.Errorf(
					"type %T does not contain a wrapped GolangType, locationPath: /%s",
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
				typ.Name(), // TODO: check if this is correct
				strings.Join(locationPath[:pathIdx], "/"),
			)
			return
		}

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

type definable interface {
	ObjectHasDefinition() bool
}

// qualifiedTypeGeneratedPackage adds the package where the obj is defined to the import list of the current module
// (if it's not there yet). Returns the import alias.
// If import is not needed (obj is defined in current package), returns empty string. If obj was not defined anywhere
// yet, returns ErrNotDefined.
func qualifiedTypeGeneratedPackage(mng *manager.TemplateRenderManager, obj common.GolangType) (string, error) {
	defInfo, defined := mng.NamespaceManager.FindType(obj)
	if !defined {
		if v, ok := obj.(definable); ok && v.ObjectHasDefinition() { // TODO: replace to Selectable?
			return "", ErrNotDefined
		}
		return "", nil // Type is not supposed to be defined in the generated code (e.g. Go built-in types)
	}

	// Use the package path from reuse config if it is defined
	if defInfo.Selection.ReusePackagePath != "" {
		return mng.ImportsManager.AddImport(defInfo.Selection.ReusePackagePath, ""), nil
	}

	// Check if the object is defined in the same directory (assuming the directory is equal to package)
	fileDir := path.Dir(defInfo.Selection.Render.File)
	if fileDir == path.Dir(mng.CurrentSelection.Render.File) {
		return "", nil // Object is defined in the current package, its name doesn't require a package name
	}

	pkgPath := path.Join(mng.RenderOpts.ImportBase, fileDir)
	return mng.ImportsManager.AddImport(pkgPath, defInfo.Selection.Render.Package), nil
}

func qualifiedImplementationGeneratedPackage(mng *manager.TemplateRenderManager, obj common.ImplementationObject) (string, error) {
	defInfo, found := lo.Find(mng.Implementations, func(def manager.ImplementationItem) bool {
		return def.Object == obj
	})
	if !found {
		return "", ErrNotDefined
	}

	// Use the package path from reuse config if it is defined
	if defInfo.Object.Config.ReusePackagePath != "" {
		return mng.ImportsManager.AddImport(defInfo.Object.Config.ReusePackagePath, ""), nil
	}

	if defInfo.Directory == path.Dir(mng.CurrentSelection.Render.File) {
		return "", nil // Object is defined in the current package, its name doesn't require a package name
	}

	pkgPath := path.Join(mng.RenderOpts.ImportBase, defInfo.Directory)
	return mng.ImportsManager.AddImport(pkgPath, defInfo.Object.Config.Package), nil
}

func qualifiedName(mng *manager.TemplateRenderManager, parts ...string) string {
	pkgPath, pkgName, n := qualifiedToImport(parts)
	var name string
	if n != "" {
		name = utils.ToGolangName(n, unicode.IsUpper(rune(n[0])))
	}
	return fmt.Sprintf("%s.%s", mng.ImportsManager.AddImport(pkgPath, pkgName), name)
}

func qualifiedRuntimeName(mng *manager.TemplateRenderManager, parts ...string) string {
	p := append([]string{mng.RenderOpts.RuntimeModule}, parts...)
	pkgPath, pkgName, n := qualifiedToImport(p)
	var name string
	if n != "" {
		name = utils.ToGolangName(n, unicode.IsUpper(rune(n[0])))
	}
	return fmt.Sprintf("%s.%s", mng.ImportsManager.AddImport(pkgPath, pkgName), name)
}
