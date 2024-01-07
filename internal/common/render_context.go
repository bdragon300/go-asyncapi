package common

import (
	"fmt"
	"path"
	"strings"

	"github.com/bdragon300/asyncapi-codegen-go/internal/types"

	"github.com/dave/jennifer/jen"
)

type Renderer interface {
	DirectRendering() bool
	RenderDefinition(ctx *RenderContext) []*jen.Statement
	RenderUsage(ctx *RenderContext) []*jen.Statement
	ID() string // Human-readable object identifier (typically it's the name), for logging, go file name, etc.
	String() string
}

type ProtocolRenderer interface {
	ProtocolName() string
	ProtocolTitle() string
}

type PackageScope int

const (
	PackageScopeType PackageScope = iota
	PackageScopeAll
)

type FileScope int

const (
	FileScopeName FileScope = iota
	FileScopeType
)

type RenderOpts struct {
	RuntimeModule string
	ImportBase    string
	TargetPackage string
	TargetDir     string
	PackageScope  PackageScope
	FileScope     FileScope
}

type RenderContext struct {
	ProtoRenderers map[string]ProtocolRenderer
	CurrentPackage string
	Logger         *types.Logger
	RenderOpts     RenderOpts
	logCallLvl     int
}

func (c *RenderContext) RuntimeModule(subPackage string) string {
	return path.Join(c.RenderOpts.RuntimeModule, subPackage)
}

func (c *RenderContext) GeneratedModule(subPackage string) string {
	switch c.RenderOpts.PackageScope {
	case PackageScopeAll:
		return c.RenderOpts.ImportBase // Everything in one package
	case PackageScopeType:
		return path.Join(c.RenderOpts.ImportBase, subPackage) // Everything split up into packages by entity type
	}
	panic(fmt.Sprintf("Unknown package scope %q", c.RenderOpts.PackageScope))
}

func (c *RenderContext) LogRender(kind, pkg, name, mode string, directRendering bool, args ...any) {
	l := c.Logger
	args = append(args, "pkg", c.CurrentPackage, "mode", mode)
	if pkg != "" {
		name = pkg + "." + name
	}
	name = kind + " " + name
	if c.logCallLvl > 0 {
		name = fmt.Sprintf("%s> %s", strings.Repeat("-", c.logCallLvl), name) // Ex: prefix: --> Message...
	}
	if directRendering && mode == "definition" {
		l.Debug(name, args...)
	} else {
		l.Trace(name, args...)
	}
	c.logCallLvl++
}

func (c *RenderContext) LogReturn() {
	if c.logCallLvl > 0 {
		c.logCallLvl--
	}
}
