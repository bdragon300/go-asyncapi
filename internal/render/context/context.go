package context

import (
	"github.com/bdragon300/go-asyncapi/internal/common"
	"path"
)

var Context common.RenderContext

// TODO: add object path?
type RenderContextImpl struct {
	RenderOpts     common.RenderOpts
}

func (c *RenderContextImpl) RuntimeModule(subPackage string) string {
	return path.Join(c.RenderOpts.RuntimeModule, subPackage)
}

func (c *RenderContextImpl) GeneratedModule(subPackage string) string {
	//switch c.RenderOpts.PackageScope {
	//case PackageScopeAll:
	//	return c.RenderOpts.ImportBase // Everything in one package
	//case PackageScopeType:
	//	return path.Join(c.RenderOpts.ImportBase, subPackage) // Everything split up into packages by entity type
	//}
	//panic(fmt.Sprintf("Unknown package scope %q", c.RenderOpts.PackageScope))
	panic("not implemented")
}

func (c *RenderContextImpl) QualifiedName(packageExpr string) string  {
	panic("not implemented")
}

func (c *RenderContextImpl) QualifiedGeneratedName(subPackage, name string) string {
	panic("not implemented")
}

func (c *RenderContextImpl) QualifiedRuntimeName(subPackage, name string) string {
	panic("not implemented")
}

func (c *RenderContextImpl) SpecProtocols() []string {
	panic("not implemented")
}

//// LogStartRender is typically called at the beginning of a D or U method and logs that the
//// object is started to be rendered. It also logs the object's name, type, and the current package.
//// Every call to LogStartRender should be followed by a call to LogFinishRender which logs that the object is finished to be
//// rendered.
//func (c *RenderContext) LogStartRender(kind, pkg, name, mode string, directRendering bool, args ...any) {
//	l := c.Logger
//	args = append(args, "pkg", c.CurrentPackage, "mode", mode)
//	if pkg != "" {
//		name = pkg + "." + name
//	}
//	name = kind + " " + name
//	if c.logCallLvl > 0 {
//		name = fmt.Sprintf("%s> %s", strings.Repeat("-", c.logCallLvl), name) // Ex: prefix: --> Message...
//	}
//	if directRendering && mode == "definition" {
//		l.Debug(name, args...)
//	} else {
//		l.Trace(name, args...)
//	}
//	c.logCallLvl++
//}
//
//func (c *RenderContext) LogFinishRender() {
//	if c.logCallLvl > 0 {
//		c.logCallLvl--
//	}
//}
