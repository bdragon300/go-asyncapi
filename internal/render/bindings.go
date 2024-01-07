package render

import (
	"github.com/bdragon300/asyncapi-codegen-go/internal/common"
	"github.com/bdragon300/asyncapi-codegen-go/internal/types"
	"github.com/bdragon300/asyncapi-codegen-go/internal/utils"
	j "github.com/dave/jennifer/jen"
)

// Bindings never renders itself, only as a part of other object
type Bindings struct {
	Name   string
	Values types.OrderedMap[string, *GoValue] // Binding values by protocol
	// Value of jsonschema fields as json marshalled strings
	JSONValues types.OrderedMap[string, types.OrderedMap[string, string]] // Binbing values by protocol
}

func (b *Bindings) DirectRendering() bool {
	return false
}

func (b *Bindings) RenderDefinition(_ *common.RenderContext) []*j.Statement {
	panic("not implemented")
}

func (b *Bindings) RenderUsage(_ *common.RenderContext) []*j.Statement {
	panic("not implemented")
}

func (b *Bindings) ID() string {
	return b.Name
}

func (b *Bindings) String() string {
	return "Bindings " + b.Name
}

func (b *Bindings) RenderBindingsMethod(
	ctx *common.RenderContext,
	bindingsStruct *GoStruct,
	protoName, protoAbbr string,
) []*j.Statement {
	ctx.LogRender("Bindings.RenderBindingsMethod", "", bindingsStruct.Name, "definition", false)
	defer ctx.LogReturn()

	receiver := j.Id(bindingsStruct.ReceiverName()).Add(utils.ToCode(bindingsStruct.RenderUsage(ctx))...)
	pv, ok := b.Values.Get(protoName)
	if !ok {
		ctx.Logger.Debug("Skip render bindings method", "name", bindingsStruct.Name, "proto", protoName)
		return nil
	}

	return []*j.Statement{
		j.Func().Params(receiver.Clone()).Id(protoAbbr).
			Params().
			Add(utils.ToCode(pv.Type.RenderUsage(ctx))...).
			BlockFunc(func(bg *j.Group) {
				bg.Id("b").Op(":=").Add(utils.ToCode(pv.RenderUsage(ctx))...)
				for _, e := range b.JSONValues.GetOr(protoName, types.OrderedMap[string, string]{}).Entries() {
					n := utils.ToLowerFirstLetter(e.Key)
					bg.Id(n).Op(":=").Lit(e.Value)
					bg.Add(utils.QualSprintf("_ = %Q(encoding/json,Unmarshal)([]byte(%[1]s), &b.%[2]s)", n, e.Key))
				}
				bg.Return(j.Id("b"))
			}),
	}
}

func renderChannelAndOperationBindingsMethod(
	ctx *common.RenderContext,
	bindingsStruct *GoStruct,
	channelBindings, publishBindings, subscribeBindings *Bindings,
	protoName, protoAbbr string,
) []*j.Statement {
	ctx.LogRender("renderChannelAndOperationBindingsMethod", "", bindingsStruct.Name, "definition", false)
	defer ctx.LogReturn()

	receiver := j.Id(bindingsStruct.ReceiverName()).Add(utils.ToCode(bindingsStruct.RenderUsage(ctx))...)

	return []*j.Statement{
		j.Func().Params(receiver.Clone()).Id(protoAbbr).
			Params().
			Qual(ctx.RuntimeModule(protoName), "ChannelBindings").
			BlockFunc(func(bg *j.Group) {
				cb := &GoValue{Type: &GoSimple{Name: "ChannelBindings", Package: ctx.RuntimeModule(protoName)}, NilCurlyBrakets: true}
				if channelBindings != nil {
					if b, ok := channelBindings.Values.Get(protoName); ok {
						ctx.Logger.Debug("Channel bindings", "proto", protoName)
						cb = b
					}
				}
				if publishBindings != nil {
					if v, ok := publishBindings.Values.Get(protoName); ok {
						ctx.Logger.Debug("Publish operation bindings", "proto", protoName)
						cb.StructVals.Set("PublisherBindings", v)
					}
				}
				if subscribeBindings != nil {
					if v, ok := subscribeBindings.Values.Get(protoName); ok {
						ctx.Logger.Debug("Subscribe operation bindings", "proto", protoName)
						cb.StructVals.Set("SubscriberBindings", v)
					}
				}
				bg.Id("b").Op(":=").Add(utils.ToCode(cb.RenderUsage(ctx))...)

				if channelBindings != nil {
					ctx.Logger.Debug("Channel jsonschema bindings", "proto", protoName)
					for _, e := range channelBindings.JSONValues.GetOr(protoName, types.OrderedMap[string, string]{}).Entries() {
						n := utils.ToLowerFirstLetter(e.Key)
						bg.Id(n).Op(":=").Lit(e.Value)
						bg.Add(utils.QualSprintf("_ = %Q(encoding/json,Unmarshal)([]byte(%[1]s), &b.%[2]s)", n, e.Key))
					}
				}
				if publishBindings != nil {
					ctx.Logger.Debug("Publish operation jsonschema bindings", "proto", protoName)
					for _, e := range publishBindings.JSONValues.GetOr(protoName, types.OrderedMap[string, string]{}).Entries() {
						n := utils.ToLowerFirstLetter(e.Key)
						bg.Id(n).Op(":=").Lit(e.Value)
						bg.Add(utils.QualSprintf("_ = %Q(encoding/json,Unmarshal)([]byte(%[1]s), &b.PublisherBindings.%[2]s)", n, e.Key))
					}
				}
				if subscribeBindings != nil {
					ctx.Logger.Debug("Subscribe operation jsonschema bindings", "proto", protoName)
					for _, e := range subscribeBindings.JSONValues.GetOr(protoName, types.OrderedMap[string, string]{}).Entries() {
						n := utils.ToLowerFirstLetter(e.Key)
						bg.Id(n).Op(":=").Lit(e.Value)
						bg.Add(utils.QualSprintf("_ = %Q(encoding/json,Unmarshal)([]byte(%[1]s), &b.SubscriberBindings.%[2]s)", n, e.Key))
					}
				}
				bg.Return(j.Id("b"))
			}),
	}
}
