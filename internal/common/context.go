package common

import (
	"github.com/samber/lo"
)

type Package interface {
	Put(ctx *CompileContext, item Assembler)
	FindBy(cb func(item any, path []string) bool) (Assembler, bool)
	ListBy(cb func(item any, path []string) bool) []Assembler
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

type CompileContext struct {
	Packages map[PackageKind]Package
	Stack    []ContextStackItem
	Linker   Linker
}

func (c *CompileContext) Push(item ContextStackItem) {
	c.Stack = append(c.Stack, item)
}

func (c *CompileContext) Pop() ContextStackItem {
	top := c.Top()
	c.Stack = c.Stack[:len(c.Stack)-1]
	return top
}

func (c *CompileContext) Top() ContextStackItem {
	if len(c.Stack) == 0 {
		panic("Stack is empty")
	}
	return c.Stack[len(c.Stack)-1]
}

func (c *CompileContext) CurrentPackage() Package {
	return c.Packages[c.Top().PackageKind]
}

func (c *CompileContext) PathStack() []string {
	return lo.Map(c.Stack, func(item ContextStackItem, _ int) string { return item.Path })
}

func (c *CompileContext) Copy() *CompileContext {
	var stack []ContextStackItem
	copy(stack, c.Stack)

	return &CompileContext{
		Packages: nil,
		Stack:    stack,
		Linker:   c.Linker,
	}
}

type LinkQuerier interface {
	Assign(obj any)
	FindCallback() func(item any, path []string) bool
	Package() PackageKind
	Ref() string
}

type ListQuerier interface {
	AssignList(obj []any)
	FindCallback() func(item any, path []string) bool
	Package() PackageKind
}

