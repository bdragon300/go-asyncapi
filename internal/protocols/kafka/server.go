package kafka

import (
	"github.com/bdragon300/asyncapi-codegen/internal/assemble"
	"github.com/bdragon300/asyncapi-codegen/internal/common"
	"github.com/bdragon300/asyncapi-codegen/internal/compile"
	"github.com/bdragon300/asyncapi-codegen/internal/utils"
	j "github.com/dave/jennifer/jen"
	"github.com/samber/lo"
)

type serverBindings struct {
	SchemaRegistryURL    *string `json:"schemaRegistryUrl" yaml:"schemaRegistryUrl"`
	SchemaRegistryVendor *string `json:"schemaRegistryVendor" yaml:"schemaRegistryVendor"`
}

func BuildServer(ctx *common.CompileContext, server *compile.Server, serverKey string) (common.Assembler, error) {
	const buildProducer = true
	const buildConsumer = true

	srvResult := ProtoServer{
		Name:            serverKey,
		URL:             server.URL,
		ProtocolVersion: server.ProtocolVersion,
		Struct: &assemble.Struct{
			BaseType: assemble.BaseType{
				Name:        ctx.GenerateObjName("", ""),
				Description: server.Description,
				Render:      true,
				Package:     ctx.TopPackageName(),
			},
		},
	}

	// Server variables
	for _, v := range server.Variables.Entries() {
		srvResult.Variables.Set(v.Key, ProtoServerVariable{
			ArgName:     utils.ToGolangName(v.Key, false),
			Enum:        v.Value.Enum,
			Default:     v.Value.Default,
			Description: v.Value.Description,
		})
	}

	// Channels which are connected to this server
	channelsLnks := assemble.NewListCbLink[*assemble.Channel](func(item common.Assembler, path []string) bool {
		ch, ok := item.(*assemble.Channel)
		if !ok {
			return false
		}
		if len(ch.AppliedServers) > 0 {
			return lo.Contains(ch.AppliedServers, serverKey)
		}
		return ch.AppliedToAllServersLinks != nil
	})
	srvResult.ChannelLinkList = channelsLnks
	ctx.Linker.AddMany(channelsLnks)

	// Producer/consumer
	if buildProducer {
		fld := assemble.StructField{
			Name: "producer",
			Type: &assemble.Simple{Type: "Producer", Package: ctx.RuntimePackage(protoName), IsIface: true},
		}
		srvResult.Struct.Fields = append(srvResult.Struct.Fields, fld)
		srvResult.Producer = true
	}
	if buildConsumer {
		fld := assemble.StructField{
			Name: "consumer",
			Type: &assemble.Simple{Type: "Consumer", Package: ctx.RuntimePackage(protoName), IsIface: true},
		}
		srvResult.Struct.Fields = append(srvResult.Struct.Fields, fld)
		srvResult.Consumer = true
	}

	// Server bindings
	if server.Bindings.Len() > 0 {
		srvResult.BindingsStructNoAssemble = &assemble.Struct{
			BaseType: assemble.BaseType{
				Name:    srvResult.Struct.Name + "Bindings",
				Render:  true,
				Package: ctx.TopPackageName(),
			},
		}
		if srvBindings, ok := server.Bindings.Get(protoName); ok {
			var bindings serverBindings
			if err := utils.UnmarshalRawsUnion2(srvBindings, &bindings); err != nil { // TODO: implement $ref
				return nil, err
			}
			marshalFields := []string{"SchemaRegistryURL", "SchemaRegistryVendor"}
			if err := utils.StructToOrderedMap(bindings, &srvResult.BindingsValues, marshalFields); err != nil {
				return nil, err
			}
		}
	}

	return srvResult, nil
}

type ProtoServerVariable struct {
	ArgName     string
	Enum        []string // TODO: implement validation
	Default     string
	Description string // TODO
}

type ProtoServer struct {
	Name                     string
	URL                      string
	ProtocolVersion          string
	Producer                 bool
	Consumer                 bool
	Struct                   *assemble.Struct
	ChannelLinkList          *assemble.LinkList[*assemble.Channel]
	BindingsStructNoAssemble *assemble.Struct // nil if no bindings set in spec FIXME: replace this on parent's server struct
	BindingsValues           utils.OrderedMap[string, any]
	Variables                utils.OrderedMap[string, ProtoServerVariable]
}

func (p ProtoServer) AllowRender() bool {
	return true
}

func (p ProtoServer) AssembleDefinition(ctx *common.AssembleContext) []*j.Statement {
	var res []*j.Statement
	if p.BindingsStructNoAssemble != nil {
		res = append(res, p.assembleBindings(ctx)...)
	}
	res = append(res, p.assembleProtocolVersionConst(ctx)...)
	res = append(res, p.assembleURLFunc(ctx)...)
	res = append(res, p.assembleNewFunc(ctx)...)
	res = append(res, p.Struct.AssembleDefinition(ctx)...)
	res = append(res, p.assembleCommonMethods(ctx)...)
	res = append(res, p.assembleChannelMethods(ctx)...)
	if p.Producer {
		res = append(res, p.assembleProducerMethods(ctx)...)
	}
	if p.Consumer {
		res = append(res, p.assembleConsumerMethods(ctx)...)
	}
	return res
}

func (p ProtoServer) AssembleUsage(ctx *common.AssembleContext) []*j.Statement {
	return p.Struct.AssembleUsage(ctx)
}

func (p ProtoServer) assembleBindings(ctx *common.AssembleContext) []*j.Statement {
	receiver := j.Id(p.BindingsStructNoAssemble.ReceiverName()).Id(p.Struct.Name)

	return []*j.Statement{
		j.Func().Params(receiver.Clone()).Id("Kafka").
			Params().
			Qual(ctx.RuntimePackage(protoName), "ServerBindings").
			Block(
				j.Return(j.Qual(ctx.RuntimePackage(protoName), "ServerBindings")).Values(j.DictFunc(func(d j.Dict) {
					for _, e := range p.BindingsValues.Entries() {
						d[j.Id(e.Key)] = j.Lit(e.Value)
					}
				})),
			),
	}
}

func (p ProtoServer) assembleProtocolVersionConst(_ *common.AssembleContext) []*j.Statement {
	if p.ProtocolVersion == "" {
		return nil
	}

	return []*j.Statement{
		j.Const().Id(p.Struct.Name + "ProtocolVersion").Op("=").Lit(p.ProtocolVersion),
	}
}

func (p ProtoServer) assembleURLFunc(ctx *common.AssembleContext) []*j.Statement {
	// Server1URL(param1 string, param2 string) run.ParamString
	return []*j.Statement{
		j.Func().Id(p.Struct.Name+"URL").
			ParamsFunc(func(g *j.Group) {
				for _, entry := range p.Variables.Entries() {
					g.Id(entry.Value.ArgName).String()
				}
			}).
			Qual(ctx.RuntimePackage(""), "ParamString").
			BlockFunc(func(blockGroup *j.Group) {
				if p.Variables.Len() > 0 {
					blockGroup.Op("paramMap := map[string]string").Values(j.DictFunc(func(d j.Dict) {
						for _, entry := range p.Variables.Entries() {
							d[j.Lit(entry.Key)] = j.Id(entry.Value.ArgName)
						}
					}))
					for _, entry := range p.Variables.Entries() {
						if entry.Value.Default != "" {
							blockGroup.If(j.Id(entry.Value.ArgName).Op("==").Lit("")).
								Block(
									j.Id(entry.Value.ArgName).Op("=").Lit(entry.Value.Default),
								)
						}
					}
					blockGroup.Return(j.Qual(ctx.RuntimePackage(""), "ParamString").Values(j.Dict{
						j.Id("Expr"):       j.Lit(p.URL),
						j.Id("Parameters"): j.Id("paramMap"),
					}))
				} else {
					blockGroup.Return(j.Qual(ctx.RuntimePackage(""), "ParamString").Values(j.Dict{
						j.Id("Expr"): j.Lit(p.URL),
					}))
				}
			}),
	}
}

func (p ProtoServer) assembleNewFunc(ctx *common.AssembleContext) []*j.Statement {
	return []*j.Statement{
		// NewServer1(producer kafka.Producer, consumer kafka.Consumer) *Server1
		j.Func().Id(p.Struct.NewFuncName()).
			ParamsFunc(func(g *j.Group) {
				if p.Producer {
					g.Id("producer").Qual(ctx.RuntimePackage(protoName), "Producer")
				}
				if p.Consumer {
					g.Id("consumer").Qual(ctx.RuntimePackage(protoName), "Consumer")
				}
			}).
			Op("*").Add(utils.ToCode(p.Struct.AssembleUsage(ctx))...).
			Block(
				j.Return(j.Op("&").Add(utils.ToCode(p.Struct.AssembleUsage(ctx))...).Values(j.DictFunc(func(d j.Dict) {
					if p.Producer {
						d[j.Id("producer")] = j.Id("producer")
					}
					if p.Consumer {
						d[j.Id("consumer")] = j.Id("consumer")
					}
				}))),
			),
	}
}

func (p ProtoServer) assembleCommonMethods(ctx *common.AssembleContext) []*j.Statement {
	receiver := j.Id(p.Struct.ReceiverName()).Id(p.Struct.Name)

	return []*j.Statement{
		// Method Name() string
		j.Func().Params(receiver.Clone()).Id("Name").
			Params().
			String().
			Block(
				j.Return(j.Lit(p.Name)),
			),

		// Protocol() run.Protocol
		j.Func().Params(receiver.Clone()).Id("Protocol").
			Params().
			Qual(ctx.RuntimePackage(""), "Protocol").
			Block(
				j.Return(j.Qual(ctx.RuntimePackage(""), "ProtocolKafka")),
			),
	}
}

func (p ProtoServer) assembleChannelMethods(ctx *common.AssembleContext) []*j.Statement {
	var res []*j.Statement
	rn := p.Struct.ReceiverName()
	receiver := j.Id(rn).Id(p.Struct.Name)

	for _, ch := range p.ChannelLinkList.Targets() {
		// Method OpenChannel1Kafka(params Channel1Parameters) (*Channel1Kafka, error)
		protoChan := ch.AllProtocols[protoName]
		protoChanKafka := protoChan.(*ProtoChannel)

		res = append(res,
			j.Func().Params(receiver.Clone()).Id("Open"+protoChanKafka.Struct.Name).
				ParamsFunc(func(g *j.Group) {
					if protoChanKafka.ParametersStructNoAssemble != nil {
						g.Id("params").Add(utils.ToCode(protoChanKafka.ParametersStructNoAssemble.AssembleUsage(ctx))...)
					}
				}).
				Params(j.Op("*").Add(utils.ToCode(protoChan.AssembleUsage(ctx))...), j.Error()).
				Block(
					j.Return(j.Add(utils.ToCode(protoChanKafka.Struct.AssembleNameUsage(ctx, "Open"+protoChanKafka.Struct.Name))...).CallFunc(func(g *j.Group) {
						if protoChanKafka.ParametersStructNoAssemble != nil {
							g.Id("params")
						}
						g.Id(rn)
					})),
				),
		)
	}
	return res
}

func (p ProtoServer) assembleProducerMethods(ctx *common.AssembleContext) []*j.Statement {
	rn := p.Struct.ReceiverName()
	receiver := j.Id(rn).Id(p.Struct.Name)
	return []*j.Statement{
		j.Func().Params(receiver.Clone()).Id("Producer").
			Params().
			Qual(ctx.RuntimePackage(protoName), "Producer").
			Block(
				j.Return(j.Id(rn).Dot("producer")),
			),
	}
}

func (p ProtoServer) assembleConsumerMethods(ctx *common.AssembleContext) []*j.Statement {
	rn := p.Struct.ReceiverName()
	receiver := j.Id(rn).Id(p.Struct.Name)
	return []*j.Statement{
		j.Func().Params(receiver.Clone()).Id("Consumer").
			Params().
			Qual(ctx.RuntimePackage(protoName), "Consumer").
			Block(
				j.Return(j.Id(rn).Dot("consumer")),
			),
	}
}
