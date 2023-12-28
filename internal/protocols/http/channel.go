package http

import (
	"encoding/json"

	"github.com/bdragon300/asyncapi-codegen-go/internal/types"

	"gopkg.in/yaml.v3"

	"github.com/bdragon300/asyncapi-codegen-go/internal/asyncapi"
	"github.com/bdragon300/asyncapi-codegen-go/internal/common"
	"github.com/bdragon300/asyncapi-codegen-go/internal/protocols"
	"github.com/bdragon300/asyncapi-codegen-go/internal/render"
	"github.com/bdragon300/asyncapi-codegen-go/internal/utils"
	j "github.com/dave/jennifer/jen"
)

type operationBindings struct {
	Type   string `json:"type" yaml:"type"`
	Method string `json:"method" yaml:"method"`
	Query  any    `json:"query" yaml:"query"` // jsonschema object
}

func BuildChannel(ctx *common.CompileContext, channel *asyncapi.Channel, channelKey string) (common.Renderer, error) {
	baseChan, err := protocols.BuildChannel(ctx, channel, channelKey, ProtoName, protoAbbr)
	if err != nil {
		return nil, err
	}

	baseChan.Struct.Fields = append(baseChan.Struct.Fields, render.StructField{Name: "path", Type: &render.Simple{Name: "string"}})

	chanResult := &ProtoChannel{BaseProtoChannel: *baseChan}

	// Channel bindings
	ctx.Logger.Trace("Channel bindings")
	ctx.Logger.NextCallLevel()
	bindingsStruct := &render.Struct{ // TODO: remove in favor of parent channel
		BaseType: render.BaseType{
			Name:         ctx.GenerateObjName(channelKey, "Bindings"),
			DirectRender: true,
			PackageName:  ctx.TopPackageName(),
		},
	}
	method, err := buildChannelBindingsMethod(ctx, channel, bindingsStruct)
	ctx.Logger.PrevCallLevel()
	if err != nil {
		return nil, err
	}
	if method != nil {
		chanResult.BindingsStructNoRender = bindingsStruct
		chanResult.BindingsMethod = method
	}

	return chanResult, nil
}

func buildChannelBindingsMethod(ctx *common.CompileContext, channel *asyncapi.Channel, bindingsStruct *render.Struct) (*render.Func, error) {
	structValues := &render.StructInit{Type: &render.Simple{Name: "ChannelBindings", Package: ctx.RuntimePackage(ProtoName)}}
	var hasBindings bool

	// Publish channel bindings
	var publisherJSON types.OrderedMap[string, any]
	if channel.Publish != nil && !channel.Publish.XIgnore {
		ctx.Logger.Trace("Channel publish operation bindings")
		if b, ok := channel.Publish.Bindings.Get(ProtoName); ok {
			hasBindings = true
			var err error
			if publisherJSON, err = buildOperationBindings(ctx, b); err != nil {
				return nil, types.CompileError{Err: err, Path: ctx.PathRef()}
			}
		}
	}

	// Subscribe channel bindings
	var subscriberJSON types.OrderedMap[string, any]
	if channel.Subscribe != nil && !channel.Subscribe.XIgnore {
		ctx.Logger.Trace("Channel subscribe operation bindings")
		if b, ok := channel.Subscribe.Bindings.Get(ProtoName); ok {
			hasBindings = true
			var err error
			if subscriberJSON, err = buildOperationBindings(ctx, b); err != nil {
				return nil, types.CompileError{Err: err, Path: ctx.PathRef()}
			}
		}
	}

	if !hasBindings {
		return nil, nil
	}

	// Method HTTP() http.ChannelBindings
	return &render.Func{
		FuncSignature: render.FuncSignature{
			Name: protoAbbr,
			Args: nil,
			Return: []render.FuncParam{
				{Type: render.Simple{Name: "ChannelBindings", Package: ctx.RuntimePackage(ProtoName)}},
			},
		},
		Receiver:     bindingsStruct,
		PackageName:  ctx.TopPackageName(),
		BodyRenderer: protocols.ChannelBindingsMethodBody(structValues, &publisherJSON, &subscriberJSON),
	}, nil
}

func buildOperationBindings(ctx *common.CompileContext, opBindings types.Union2[json.RawMessage, yaml.Node]) (res types.OrderedMap[string, any], err error) {
	var bindings operationBindings
	if err = types.UnmarshalRawsUnion2(opBindings, &bindings); err != nil {
		return res, types.CompileError{Err: err, Path: ctx.PathRef()}
	}
	if bindings.Query != nil {
		v, err := json.Marshal(bindings.Query)
		if err != nil {
			return res, err
		}
		res.Set("Query", string(v))
	}

	return
}

type ProtoChannel struct {
	protocols.BaseProtoChannel
	BindingsStructNoRender *render.Struct // nil if bindings not set FIXME: remove in favor of struct in parent channel
	BindingsMethod         *render.Func
}

func (p ProtoChannel) DirectRendering() bool {
	return true
}

func (p ProtoChannel) RenderDefinition(ctx *common.RenderContext) []*j.Statement {
	var res []*j.Statement
	if p.BindingsMethod != nil {
		res = append(res, p.BindingsMethod.RenderDefinition(ctx)...)
	}
	res = append(res, p.ServerIface.RenderDefinition(ctx)...)
	if !p.IsEmpty {
		res = append(res, protocols.RenderChannelOpenFunc(
			ctx, p.Struct, p.Name, p.ServerIface, p.ParametersStructNoRender, p.BindingsStructNoRender,
			p.Publisher, p.Subscriber, ProtoName, protoAbbr,
		)...)
	}
	res = append(res, p.renderNewFunc(ctx)...)
	res = append(res, p.Struct.RenderDefinition(ctx)...)
	res = append(res, protocols.RenderChannelCommonMethods(ctx, p.Struct, p.Publisher, p.Subscriber)...)
	res = append(res, p.renderCommonMethods(ctx)...)
	if p.Publisher {
		res = append(res, protocols.RenderChannelPublisherMethods(ctx, p.Struct, ProtoName)...)
		res = append(res, p.renderPublisherMethods(ctx)...)
	}
	if p.Subscriber {
		res = append(res, protocols.RenderChannelSubscriberMethods(
			ctx, p.Struct, p.SubMessagePromise, p.FallbackMessageType, ProtoName, protoAbbr,
		)...)
	}
	return res
}

func (p ProtoChannel) RenderUsage(ctx *common.RenderContext) []*j.Statement {
	return p.Struct.RenderUsage(ctx)
}

func (p ProtoChannel) String() string {
	return p.BaseProtoChannel.Name
}

func (p ProtoChannel) renderNewFunc(ctx *common.RenderContext) []*j.Statement {
	return []*j.Statement{
		// NewChannel1Proto(params Channel1Parameters, publisher proto.Publisher, subscriber proto.Subscriber) *Channel1Proto
		j.Func().Id(p.Struct.NewFuncName()).
			ParamsFunc(func(g *j.Group) {
				if p.ParametersStructNoRender != nil {
					g.Id("params").Add(utils.ToCode(p.ParametersStructNoRender.RenderUsage(ctx))...)
				}
				if p.Publisher {
					g.Id("publisher").Qual(ctx.RuntimePackage(ProtoName), "Publisher")
				}
				if p.Subscriber {
					g.Id("subscriber").Qual(ctx.RuntimePackage(ProtoName), "Subscriber")
				}
			}).
			Op("*").Add(utils.ToCode(p.Struct.RenderUsage(ctx))...).
			BlockFunc(func(bg *j.Group) {
				bg.Op("res := ").Add(utils.ToCode(p.Struct.RenderUsage(ctx))...).Values(j.DictFunc(func(d j.Dict) {
					d[j.Id("name")] = j.Id(utils.ToGolangName(p.Name, true) + "Name").CallFunc(func(g *j.Group) {
						if p.ParametersStructNoRender != nil {
							g.Id("params")
						}
					})
					if p.Publisher {
						d[j.Id("publisher")] = j.Id("publisher")
					}
					if p.Subscriber {
						d[j.Id("subscriber")] = j.Id("subscriber")
					}
				}))
				bg.Op("res.path = res.name.String()")
				bg.Op(`return &res`)
			}),
	}
}

func (p ProtoChannel) renderCommonMethods(_ *common.RenderContext) []*j.Statement {
	rn := p.Struct.ReceiverName()
	receiver := j.Id(rn).Id(p.Struct.Name)

	return []*j.Statement{
		// Method Path() string
		j.Func().Params(receiver.Clone()).Id("Path").
			Params().
			String().
			Block(
				j.Return(j.Id(rn).Dot("path")),
			),
	}
}

func (p ProtoChannel) renderPublisherMethods(ctx *common.RenderContext) []*j.Statement {
	rn := p.Struct.ReceiverName()
	receiver := j.Id(rn).Id(p.Struct.Name)

	var msgTyp common.GolangType = render.Pointer{Type: p.FallbackMessageType, DirectRender: true}
	if p.PubMessagePromise != nil {
		msgTyp = render.Pointer{Type: p.PubMessagePromise.Target().OutStruct, DirectRender: true}
	}

	var msgBindings *render.Struct
	if p.PubMessagePromise != nil {
		if _, ok := p.PubMessagePromise.Target().BindingsStructProtoMethods.Get(ProtoName); ok {
			msgBindings = p.PubMessagePromise.Target().BindingsStruct
		}
	}

	return []*j.Statement{
		// Method MakeEnvelope(envelope http.EnvelopeWriter, message *Message1Out) error
		j.Func().Params(receiver.Clone()).Id("MakeEnvelope").
			ParamsFunc(func(g *j.Group) {
				g.Id("envelope").Qual(ctx.RuntimePackage(ProtoName), "EnvelopeWriter")
				g.Id("message").Add(utils.ToCode(msgTyp.RenderUsage(ctx))...)
			}).
			Error().
			BlockFunc(func(bg *j.Group) {
				bg.Op("envelope.ResetPayload()")
				if p.PubMessagePromise == nil { // No Message set for Channel in spec
					bg.Empty().Add(utils.QualSprintf(`
						enc := %Q(encoding/json,NewEncoder)(envelope)
						if err := enc.Encode(message); err != nil {
							return err
						}`))
				} else { // Message is set for Channel in spec
					bg.Op(`
						if err := message.MarshalHTTPEnvelope(envelope); err != nil {
							return err
						}`)
				}
				bg.Op("envelope.SetPath").Call(j.Id(rn).Dot("path"))
				if msgBindings != nil {
					bg.Op("envelope.SetBindings").Call(
						j.Add(utils.ToCode(msgBindings.RenderUsage(ctx))...).Values().Dot("HTTP()"),
					)
				}
				bg.Return(j.Nil())
			}),
	}
}
