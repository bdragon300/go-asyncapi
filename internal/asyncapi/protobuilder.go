package asyncapi

import (
	"fmt"

	"github.com/bdragon300/go-asyncapi/internal/compiler/compile"

	"github.com/bdragon300/go-asyncapi/internal/render/lang"
	"github.com/bdragon300/go-asyncapi/internal/utils"
	"github.com/samber/lo"

	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render"
)

// BuildProtoChannelStruct builds a Go struct for the channel. This function is called from protocol builders.
func BuildProtoChannelStruct(ctx *compile.Context, source *Channel, target *render.Channel, protoName, golangName string) *lang.GoStruct {
	chanStruct := lang.GoStruct{
		BaseType: lang.BaseType{
			OriginalName:  golangName,
			Description:   source.Description,
			HasDefinition: true,
		},
		Fields: []lang.GoStructField{
			{Name: "address", Type: &lang.GoSimple{TypeName: "ParamString", Import: ctx.RuntimeModule("")}},
		},
	}

	// Publisher stuff
	if target.IsPublisher {
		ctx.Logger.Trace("Channel publish operation", "proto", protoName)
		chanStruct.Fields = append(chanStruct.Fields, lang.GoStructField{
			Name: "publisher",
			Type: &lang.GoSimple{
				TypeName:    "Publisher",
				Import:      ctx.RuntimeModule(protoName),
				IsInterface: true,
			},
		})
	}

	// Subscriber stuff
	if target.IsSubscriber {
		ctx.Logger.Trace("Channel subscribe operation", "proto", protoName)
		chanStruct.Fields = append(chanStruct.Fields, lang.GoStructField{
			Name: "subscriber",
			Type: &lang.GoSimple{
				TypeName:    "Subscriber",
				Import:      ctx.RuntimeModule(protoName),
				IsInterface: true,
			},
		})
	}

	return &chanStruct
}

func BuildProtoServer(ctx *compile.Context, source *Server, target *render.Server, protoName string) *render.ProtoServer {
	srvStruct := lang.GoStruct{
		BaseType: lang.BaseType{
			OriginalName:  target.OriginalName,
			Description:   source.Description,
			HasDefinition: true,
		},
	}
	// TODO: handle when protoName is empty (it appears when we build ProtoServer for unsupported protocol)
	// TODO: consider IsPublisher and IsSubscriber (and also in templates)
	// Producer/consumer
	ctx.Logger.Trace("Server producer", "proto", protoName)
	fld := lang.GoStructField{
		Name: "producer",
		Type: &lang.GoSimple{TypeName: "Producer", Import: ctx.RuntimeModule(protoName), IsInterface: true},
	}
	srvStruct.Fields = append(srvStruct.Fields, fld)

	ctx.Logger.Trace("Server consumer", "proto", protoName)
	fld = lang.GoStructField{
		Name: "consumer",
		Type: &lang.GoSimple{TypeName: "Consumer", Import: ctx.RuntimeModule(protoName), IsInterface: true},
	}
	srvStruct.Fields = append(srvStruct.Fields, fld)

	return &render.ProtoServer{
		Server: target,
		Type:   &srvStruct,
	}
}

func BuildProtoOperation(ctx *compile.Context, source *Operation, target *render.Operation, proto string) *render.ProtoOperation {
	prmChType := lang.NewGolangTypePromise(source.Channel.Ref, func(obj common.Artifact) common.GolangType {
		ch := obj.(*render.Channel)
		if ch.Dummy {
			return &lang.GoSimple{TypeName: "any", IsInterface: true} // Dummy type
		}
		protoCh, found := lo.Find(ch.ProtoChannels, func(p *render.ProtoChannel) bool {
			return p.Protocol == proto
		})
		if !found {
			panic(fmt.Sprintf("ProtoChannel[%s] not found in %s. This is a bug", proto, ch))
		}
		return protoCh.Type
	})
	ctx.PutPromise(prmChType)
	prmCh := lang.NewPromise[*render.ProtoChannel](source.Channel.Ref, func(obj common.Artifact) *render.ProtoChannel {
		ch := obj.(*render.Channel)
		if ch.Dummy {
			return &render.ProtoChannel{Channel: ch, Protocol: proto} // Dummy channel
		}
		protoCh, found := lo.Find(ch.ProtoChannels, func(p *render.ProtoChannel) bool {
			return p.Protocol == proto
		})
		if !found {
			panic(fmt.Sprintf("ProtoChannel[%s] not found in %s. This is a bug", proto, ch))
		}
		return protoCh
	})
	ctx.PutPromise(prmCh)

	return &render.ProtoOperation{
		Operation: target,
		Type: &lang.GoStruct{
			BaseType: lang.BaseType{
				OriginalName:  utils.ToGolangName(target.OriginalName+lo.Capitalize(proto), true),
				HasDefinition: true,
			},
			Fields: []lang.GoStructField{
				{Name: "Channel", Type: &lang.GoPointer{Type: prmChType}},
			},
		},
		ProtoChannelPromise: prmCh,
		Protocol:            proto,
	}
}
