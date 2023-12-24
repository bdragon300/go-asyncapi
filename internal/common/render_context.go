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
	String() string // Human-readable object identifier (name) as string, for logging purposes
}

type RenderContext struct {
	CurrentPackage string
	ImportBase     string
	Logger         *types.Logger
	logCallLvl     int
}

func (c *RenderContext) RuntimePackage(subPackage string) string {
	return path.Join(RunPackagePath, subPackage)
}

func (c *RenderContext) GeneratedPackage(subPackage string) string {
	return path.Join(c.ImportBase, subPackage)
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
