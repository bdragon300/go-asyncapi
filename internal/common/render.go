package common

import (
	"fmt"
	"path"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/dave/jennifer/jen"
)

type Renderer interface {
	DirectRendering() bool
	RenderDefinition(ctx *RenderContext) []*jen.Statement
	RenderUsage(ctx *RenderContext) []*jen.Statement
	String() string
}

func NewRenderContext(currentPackage, importBase string) *RenderContext {
	return &RenderContext{
		CurrentPackage: currentPackage,
		ImportBase:     importBase,
		logger:         log.Default().WithPrefix("Rendering 🎨"),
	}
}

type RenderContext struct {
	CurrentPackage string
	ImportBase     string
	logger         *log.Logger
	logCallLvl     int
}

func (c *RenderContext) RuntimePackage(subPackage string) string {
	return path.Join(RunPackagePath, subPackage)
}

func (c *RenderContext) GeneratedPackage(subPackage string) string {
	return path.Join(c.ImportBase, subPackage)
}

func (c *RenderContext) LogRender(kind, pkg, name, mode string, direct bool, args ...any) {
	l := c.logger
	args = append(args, "pkg", c.CurrentPackage, "mode", mode)
	if pkg != "" {
		name = pkg + "." + name
	}
	name = kind + " " + name
	if c.logCallLvl > 0 {
		name = fmt.Sprintf("%s> %s", strings.Repeat("-", c.logCallLvl), name) // Ex: prefix: --> Message...
	}
	if direct {
		l.Info(name, args...)
	} else {
		l.Debug(name, args...)
	}
	c.logCallLvl++
}

func (c *RenderContext) LogReturn() {
	if c.logCallLvl > 0 {
		c.logCallLvl--
	}
}
