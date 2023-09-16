package assemble

import (
	"fmt"

	"github.com/bdragon300/asyncapi-codegen/internal/common"
	"github.com/dave/jennifer/jen"
	"github.com/samber/lo"
)

func NewRefLink[T any](ref string) *Link[T] {
	return &Link[T]{ref: ref}
}

func NewCbLink[T any](findCb func(item common.Assembler, path []string) bool) *Link[T] {
	return &Link[T]{findCb: findCb}
}

type Link[T any] struct {
	ref    string
	findCb func(item common.Assembler, path []string) bool

	target   T
	assigned bool
}

func (r *Link[T]) Assign(obj any) {
	r.target = obj.(T)
	r.assigned = true
}

func (r *Link[T]) Assigned() bool {
	return r.assigned
}

func (r *Link[T]) FindCallback() func(item common.Assembler, path []string) bool {
	return r.findCb
}

func (r *Link[T]) Target() T {
	return r.target
}

func (r *Link[T]) Ref() string {
	return r.ref
}

func NewListCbLink[T any](findCb func(item common.Assembler, path []string) bool) *LinkList[T] {
	return &LinkList[T]{findCb: findCb}
}

type LinkList[T any] struct {
	findCb func(item common.Assembler, path []string) bool

	targets  []T
	assigned bool
}

func (r *LinkList[T]) AssignList(objs []any) {
	var ok bool
	r.targets, ok = lo.FromAnySlice[T](objs)
	if !ok {
		panic(fmt.Sprintf("Cannot assign slice of %+v to %T", objs, r.targets))
	}
	r.assigned = true
}

func (r *LinkList[T]) Assigned() bool {
	return r.assigned
}

func (r *LinkList[T]) FindCallback() func(item common.Assembler, path []string) bool {
	return r.findCb
}

func (r *LinkList[T]) Targets() []T {
	return r.targets
}

func NewRefLinkAsAssembler(ref string) *LinkAsAssembler {
	return &LinkAsAssembler{
		Link: *NewRefLink[common.Assembler](ref),
	}
}

type LinkAsAssembler struct {
	Link[common.Assembler]
}

func (r *LinkAsAssembler) AssembleDefinition(ctx *common.AssembleContext) []*jen.Statement {
	return r.target.AssembleDefinition(ctx)
}

func (r *LinkAsAssembler) AssembleUsage(ctx *common.AssembleContext) []*jen.Statement {
	return r.target.AssembleUsage(ctx)
}

func (r *LinkAsAssembler) AllowRender() bool {
	return false // Prevent rendering the object we're point to for several times
}

func NewRefLinkAsGolangType(ref string) *LinkAsGolangType {
	return &LinkAsGolangType{
		Link: *NewRefLink[common.GolangType](ref),
	}
}

type LinkAsGolangType struct {
	Link[common.GolangType]
}

func (r *LinkAsGolangType) TypeName() string {
	return r.target.TypeName()
}

func (r *LinkAsGolangType) AllowRender() bool {
	return false // Prevent rendering the object we're point to for several times
}

func (r *LinkAsGolangType) AssembleDefinition(ctx *common.AssembleContext) []*jen.Statement {
	return r.target.AssembleDefinition(ctx)
}

func (r *LinkAsGolangType) AssembleUsage(ctx *common.AssembleContext) []*jen.Statement {
	return r.target.AssembleUsage(ctx)
}
