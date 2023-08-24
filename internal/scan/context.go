package scan

import (
	"github.com/bdragon300/asyncapi-codegen/internal/common"
	"github.com/bdragon300/asyncapi-codegen/internal/render"
	"github.com/samber/lo"
)

type Package interface {
	Put(ctx *Context, item render.LangRenderer)
	Find(path []string) (render.LangRenderer, bool)
	MustFind(path []string) render.LangRenderer
}

type ContextStackItem struct {
	Path        string
	Flags       map[SchemaTag]string
	PackageKind common.PackageKind
}

type Context struct {
	Packages map[common.PackageKind]Package
	Stack    []ContextStackItem
	RefMgr   *RefManager
}

func (c *Context) Push(item ContextStackItem) {
	c.Stack = append(c.Stack, item)
}

func (c *Context) Pop() ContextStackItem {
	top := c.Top()
	c.Stack = c.Stack[:len(c.Stack)-1]
	return top
}

func (c *Context) Top() ContextStackItem {
	if len(c.Stack) == 0 {
		panic("Stack is empty")
	}
	return c.Stack[len(c.Stack)-1]
}

func (c *Context) CurrentPackage() Package {
	return c.Packages[c.Top().PackageKind]
}

func (c *Context) PathStack() []string {
	return lo.Map(c.Stack, func(item ContextStackItem, _ int) string { return item.Path })
}

func (c *Context) Copy() *Context {
	var stack []ContextStackItem
	copy(stack, c.Stack)

	return &Context{
		Packages: nil,
		Stack:    stack,
		RefMgr:   c.RefMgr,
	}
}
