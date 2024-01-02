package amqp

import (
	"encoding/json"
	"time"

	"github.com/bdragon300/asyncapi-codegen-go/internal/asyncapi"
	"github.com/bdragon300/asyncapi-codegen-go/internal/common"
	"github.com/bdragon300/asyncapi-codegen-go/internal/render"
	renderAmqp "github.com/bdragon300/asyncapi-codegen-go/internal/render/amqp"
	"github.com/bdragon300/asyncapi-codegen-go/internal/types"
	"gopkg.in/yaml.v3"
)

type channelBindings struct {
	Is       string          `json:"is" yaml:"is"`
	Exchange *exchangeParams `json:"exchange" yaml:"exchange"`
	Queue    *queueParams    `json:"queue" yaml:"queue"`
}

type exchangeParams struct {
	Name       *string `json:"name" yaml:"name"` // Empty string means "default amqp exchange"
	Type       string  `json:"type" yaml:"type"`
	Durable    *bool   `json:"durable" yaml:"durable"`
	AutoDelete *bool   `json:"autoDelete" yaml:"autoDelete"`
	VHost      string  `json:"vhost" yaml:"vhost"`
}

type queueParams struct {
	Name       string `json:"name" yaml:"name"`
	Durable    *bool  `json:"durable" yaml:"durable"`
	Exclusive  *bool  `json:"exclusive" yaml:"exclusive"`
	AutoDelete *bool  `json:"autoDelete" yaml:"autoDelete"`
	VHost      string `json:"vhost" yaml:"vhost"`
}

type operationBindings struct {
	Expiration   int      `json:"expiration" yaml:"expiration"`
	UserID       string   `json:"userId" yaml:"userId"`
	CC           []string `json:"cc" yaml:"cc"`
	Priority     int      `json:"priority" yaml:"priority"`
	DeliveryMode int      `json:"deliveryMode" yaml:"deliveryMode"`
	Mandatory    bool     `json:"mandatory" yaml:"mandatory"`
	BCC          []string `json:"bcc" yaml:"bcc"`
	ReplyTo      string   `json:"replyTo" yaml:"replyTo"`
	Timestamp    bool     `json:"timestamp" yaml:"timestamp"`
	Ack          bool     `json:"ack" yaml:"ack"`
}

func (pb ProtoBuilder) BuildChannel(ctx *common.CompileContext, channel *asyncapi.Channel, channelKey string, abstractChannel *render.Channel) (common.Renderer, error) {
	baseChan, err := pb.BuildBaseProtoChannel(ctx, channel, channelKey, abstractChannel)
	if err != nil {
		return nil, err
	}

	baseChan.Struct.Fields = append(
		baseChan.Struct.Fields,
		render.GoStructField{Name: "exchange", Type: &render.GoSimple{Name: "string"}},
		render.GoStructField{Name: "queue", Type: &render.GoSimple{Name: "string"}},
	)

	return &renderAmqp.ProtoChannel{BaseProtoChannel: *baseChan}, nil
}

func (pb ProtoBuilder) BuildChannelBindings(ctx *common.CompileContext, rawData types.Union2[json.RawMessage, yaml.Node]) (vals *render.GoValue, jsonVals types.OrderedMap[string, string], err error) {
	var bindings channelBindings
	if err = types.UnmarshalRawsUnion2(rawData, &bindings); err != nil {
		err = types.CompileError{Err: err, Path: ctx.PathRef(), Proto: pb.ProtoName}
		return
	}

	vals = &render.GoValue{Type: &render.GoSimple{Name: "ChannelBindings", Package: ctx.RuntimePackage(pb.ProtoName)}, NilCurlyBrakets: true}
	switch bindings.Is {
	case "queue":
		vals.StructVals.Set("ChannelType", &render.GoSimple{Name: "ChannelTypeQueue", Package: ctx.RuntimePackage(pb.ProtoName)})
	default: // routingKey and any other value
		vals.StructVals.Set("ChannelType", &render.GoSimple{Name: "ChannelTypeRoutingKey", Package: ctx.RuntimePackage(pb.ProtoName)})
	}
	if bindings.Exchange != nil {
		ecVals := render.ConstructGoValue(
			bindings.Exchange, []string{"Type"}, &render.GoSimple{Name: "ExchangeConfiguration", Package: ctx.RuntimePackage(pb.ProtoName)},
		)
		switch bindings.Exchange.Type {
		case "default":
			ecVals.StructVals.Set("Type", &render.GoSimple{Name: "ExchangeTypeDefault", Package: ctx.RuntimePackage(pb.ProtoName)})
		case "topic":
			ecVals.StructVals.Set("Type", &render.GoSimple{Name: "ExchangeTypeTopic", Package: ctx.RuntimePackage(pb.ProtoName)})
		case "direct":
			ecVals.StructVals.Set("Type", &render.GoSimple{Name: "ExchangeTypeDirect", Package: ctx.RuntimePackage(pb.ProtoName)})
		case "fanout":
			ecVals.StructVals.Set("Type", &render.GoSimple{Name: "ExchangeTypeFanout", Package: ctx.RuntimePackage(pb.ProtoName)})
		case "headers":
			ecVals.StructVals.Set("Type", &render.GoSimple{Name: "ExchangeTypeHeaders", Package: ctx.RuntimePackage(pb.ProtoName)})
		}
		vals.StructVals.Set("ExchangeConfiguration", ecVals)
	}
	qVals := render.ConstructGoValue(
		bindings.Queue, nil, &render.GoSimple{Name: "QueueConfiguration", Package: ctx.RuntimePackage(pb.ProtoName)},
	)
	vals.StructVals.Set("QueueConfiguration", &qVals)

	return
}

func (pb ProtoBuilder) BuildOperationBindings(ctx *common.CompileContext, rawData types.Union2[json.RawMessage, yaml.Node]) (vals *render.GoValue, jsonVals types.OrderedMap[string, string], err error) {
	var bindings operationBindings
	if err = types.UnmarshalRawsUnion2(rawData, &bindings); err != nil {
		err = types.CompileError{Err: err, Path: ctx.PathRef(), Proto: pb.ProtoName}
		return
	}

	vals = render.ConstructGoValue(
		bindings, []string{"Expiration", "DeliveryMode"}, &render.GoSimple{Name: "OperationBindings", Package: ctx.RuntimePackage(pb.ProtoName)},
	)
	if bindings.Expiration > 0 {
		v := render.ConstructGoValue(
			bindings.Expiration*int(time.Millisecond), nil, &render.GoSimple{Name: "Duration", Package: "time"},
		)
		vals.StructVals.Set("Expiration", v)
	}
	switch bindings.DeliveryMode {
	case 1:
		vals.StructVals.Set("DeliveryMode", &render.GoSimple{Name: "DeliveryModeTransient", Package: ctx.RuntimePackage(pb.ProtoName)})
	case 2:
		vals.StructVals.Set("DeliveryMode", &render.GoSimple{Name: "DeliveryModePersistent", Package: ctx.RuntimePackage(pb.ProtoName)})
	}

	return
}
