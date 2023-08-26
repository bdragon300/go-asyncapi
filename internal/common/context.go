package common

import (
	"github.com/samber/lo"
)

type Package interface {
	Put(ctx *Context, item Assembled)
	Find(path []string) (Assembled, bool)
	List(path []string) []Assembled
}

type Linker interface {
	Add(query LinkQuerier)
	AddMany(query ListQuerier)
}

type ContextStackItem struct {
	Path        string
	Flags       map[SchemaTag]string
	PackageKind PackageKind
}

type Context struct {
	Packages map[PackageKind]Package
	Stack    []ContextStackItem
	Linker   Linker
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
		Linker:   c.Linker,
	}
}

type LinkQuerier interface {
	Assign(obj any)
	Package() PackageKind
	Path() []string
	Ref() string
}

type ListQuerier interface {
	AssignList(obj []any)
	Package() PackageKind
	Path() []string
}

