package common

import (
	"path"
	"strings"

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

type SimpleStack[T any] struct {
	stack []T
}

func (s *SimpleStack[T]) Top() T {
	if len(s.stack) == 0 {
		panic("Stack is empty")
	}
	return s.stack[len(s.stack)-1]
}

func (s *SimpleStack[T]) Pop() T {
	top := s.Top()
	s.stack = s.stack[:len(s.stack)-1]
	return top
}

func (s *SimpleStack[T]) Push(v T) {
	s.stack = append(s.stack, v)
}

func (s *SimpleStack[T]) Items() []T {
	return s.stack
}

func (s *SimpleStack[T]) replaceTop(v T) {
	if len(s.stack) == 0 {
		panic("Stack is empty")
	}
	s.stack[len(s.stack)-1] = v
}

type ContextStackItem struct {
	Path        string
	Flags       map[SchemaTag]string
	PackageKind PackageKind
	ObjName     string
}

type CompileContext struct {
	Packages map[PackageKind]Package
	Stack    SimpleStack[ContextStackItem]
	Linker   Linker
}

func (c *CompileContext) CurrentPackage() Package {
	return c.Packages[c.Stack.Top().PackageKind]
}

func (c *CompileContext) PathStack() []string {
	return lo.Map(c.Stack.Items(), func(item ContextStackItem, _ int) string { return item.Path })
}

func (c *CompileContext) PathRef() string {
	return "#/" + path.Join(c.PathStack()...)
}

func (c *CompileContext) SetObjName(n string) {
	t := c.Stack.Top()
	t.ObjName = n
	c.Stack.replaceTop(t)
}

func (c *CompileContext) CurrentObjName() string {
	items := lo.FilterMap(c.Stack.Items(), func(item ContextStackItem, index int) (string, bool) {
		return item.ObjName, item.ObjName != ""
	})
	if len(items) == 0 {
		items = c.PathStack()
	}
	return strings.Join(items, "_")
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

