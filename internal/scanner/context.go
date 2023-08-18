package scanner

import (
	"github.com/bdragon300/asyncapi-codegen/internal/common"
	"github.com/dave/jennifer/jen"
	"github.com/samber/lo"
)

type LangRenderer interface {
	SkipRender() bool
	PrepareRender(name string)
	RenderDefinition() []*jen.Statement
	RenderUsage() []*jen.Statement
	GetDefaultName() string
}

type Bucket interface {
	Put(ctx *Context, item LangRenderer)
	Find(path []string) (LangRenderer, bool)
}

type ContextStackItem struct {
	Path  string
	Flags map[string]string
}

type Context struct {
	Buckets map[common.BucketKind]Bucket
	Stack   []ContextStackItem
	RefMgr  *RefManager
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

func (c *Context) PathStack() []string {
	return lo.Map(c.Stack, func(item ContextStackItem, _ int) string { return item.Path })
}

func (c *Context) Copy() *Context {
	var stack []ContextStackItem
	copy(stack, c.Stack)

	return &Context{
		Buckets: nil,
		Stack:   stack,
		RefMgr:  c.RefMgr,
	}
}
