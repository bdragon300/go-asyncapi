package render

import (
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render/lang"
	"github.com/bdragon300/go-asyncapi/internal/types"
)

// Bindings never renders itself, only as a part of other object
type Bindings struct {
	Name   string

	Values types.OrderedMap[string, *lang.GoValue] // Binding values by protocol
	// Value of jsonschema fields as json marshalled strings
	JSONValues types.OrderedMap[string, types.OrderedMap[string, string]] // Binbing values by protocol
}

func (b *Bindings) Kind() common.ObjectKind {
	return common.ObjectKindOther  // TODO: separate Bindings from Channel, leaving only the Promise, and make its own 4 ObjectKinds
}

func (b *Bindings) Selectable() bool {
	return false
}


//func (b *Bindings) ID() string {
//	return b.Name
//}

func (b *Bindings) String() string {
	return "Bindings " + b.Name
}

//func (b *Bindings) RenderBindingsMethod(
//	ctx *common.RenderContext,
//	bindingsStruct *GoStruct,
//	protoName, protoTitle string,
//) []*j.Statement {
//	ctx.LogStartRender("Bindings.RenderBindingsMethod", "", bindingsStruct.Name, "definition", false)
//	defer ctx.LogFinishRender()
//
//	receiver := j.Id(bindingsStruct.ReceiverName()).Add(utils.ToCode(bindingsStruct.U(ctx))...)
//	pv, ok := b.Values.Get(protoName)
//	if !ok {
//		ctx.Logger.Debug("Skip render bindings method", "name", bindingsStruct.Name, "proto", protoName)
//		return nil
//	}
//
//	return []*j.Statement{
//		j.Func().Params(receiver.Clone()).Id(protoTitle).
//			Params().
//			Add(utils.ToCode(pv.Type.U(ctx))...).
//			BlockFunc(func(bg *j.Group) {
//				bg.Id("b").Op(":=").Add(utils.ToCode(pv.U(ctx))...)
//				for _, e := range b.JSONValues.GetOr(protoName, types.OrderedMap[string, string]{}).Entries() {
//					n := utils.ToLowerFirstLetter(e.Key)
//					bg.Id(n).Op(":=").Lit(e.Value)
//					bg.Add(utils.QualSprintf("_ = %Q(encoding/json,Unmarshal)([]byte(%[1]s), &b.%[2]s)", n, e.Key))
//				}
//				bg.Return(j.Id("b"))
//			}),
//	}
//}
//
//func renderChannelAndOperationBindingsMethod(
//	ctx *common.RenderContext,
//	bindingsStruct *GoStruct,
//	channelBindings, publishBindings, subscribeBindings *Bindings,
//	protoName, protoTitle string,
//) []*j.Statement {
//	ctx.LogStartRender("renderChannelAndOperationBindingsMethod", "", bindingsStruct.Name, "definition", false)
//	defer ctx.LogFinishRender()
//
//	receiver := j.Id(bindingsStruct.ReceiverName()).Add(utils.ToCode(bindingsStruct.U(ctx))...)
//
//	return []*j.Statement{
//		j.Func().Params(receiver.Clone()).Id(protoTitle).
//			Params().
//			Qual(ctx.RuntimeModule(protoName), "ChannelBindings").
//			BlockFunc(func(bg *j.Group) {
//				cb := &GoValue{Type: &GoSimple{Name: "ChannelBindings", Import: ctx.RuntimeModule(protoName)}, EmptyCurlyBrackets: true}
//				if channelBindings != nil {
//					if b, ok := channelBindings.Values.Get(protoName); ok {
//						ctx.Logger.Debug("Channel bindings", "proto", protoName)
//						cb = b
//					}
//				}
//				if publishBindings != nil {
//					if v, ok := publishBindings.Values.Get(protoName); ok {
//						ctx.Logger.Debug("Publish operation bindings", "proto", protoName)
//						cb.StructValues.Set("PublisherBindings", v)
//					}
//				}
//				if subscribeBindings != nil {
//					if v, ok := subscribeBindings.Values.Get(protoName); ok {
//						ctx.Logger.Debug("Subscribe operation bindings", "proto", protoName)
//						cb.StructValues.Set("SubscriberBindings", v)
//					}
//				}
//				bg.Id("b").Op(":=").Add(utils.ToCode(cb.U(ctx))...)
//
//				if channelBindings != nil {
//					ctx.Logger.Debug("Channel jsonschema bindings", "proto", protoName)
//					for _, e := range channelBindings.JSONValues.GetOr(protoName, types.OrderedMap[string, string]{}).Entries() {
//						n := utils.ToLowerFirstLetter(e.Key)
//						bg.Id(n).Op(":=").Lit(e.Value)
//						bg.Add(utils.QualSprintf("_ = %Q(encoding/json,Unmarshal)([]byte(%[1]s), &b.%[2]s)", n, e.Key))
//					}
//				}
//				if publishBindings != nil {
//					ctx.Logger.Debug("Publish operation jsonschema bindings", "proto", protoName)
//					for _, e := range publishBindings.JSONValues.GetOr(protoName, types.OrderedMap[string, string]{}).Entries() {
//						n := utils.ToLowerFirstLetter(e.Key) + "Pub"
//						bg.Id(n).Op(":=").Lit(e.Value)
//						bg.Add(utils.QualSprintf("_ = %Q(encoding/json,Unmarshal)([]byte(%[1]s), &b.PublisherBindings.%[2]s)", n, e.Key))
//					}
//				}
//				if subscribeBindings != nil {
//					ctx.Logger.Debug("Subscribe operation jsonschema bindings", "proto", protoName)
//					for _, e := range subscribeBindings.JSONValues.GetOr(protoName, types.OrderedMap[string, string]{}).Entries() {
//						n := utils.ToLowerFirstLetter(e.Key) + "Sub"
//						bg.Id(n).Op(":=").Lit(e.Value)
//						bg.Add(utils.QualSprintf("_ = %Q(encoding/json,Unmarshal)([]byte(%[1]s), &b.SubscriberBindings.%[2]s)", n, e.Key))
//					}
//				}
//				bg.Return(j.Id("b"))
//			}),
//	}
//}
