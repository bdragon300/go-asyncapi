package amqp

import (
	"encoding/json"
	"time"

	"github.com/bdragon300/go-asyncapi/internal/render/lang"
	"github.com/bdragon300/go-asyncapi/internal/utils"
	"github.com/samber/lo"

	"github.com/bdragon300/go-asyncapi/internal/asyncapi"
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render"
	"github.com/bdragon300/go-asyncapi/internal/types"
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

func (pb ProtoBuilder) BuildChannel(ctx *common.CompileContext, channel *asyncapi.Channel, parent *render.Channel) (*render.ProtoChannel, error) {
	golangName := utils.ToGolangName(parent.OriginalName+lo.Capitalize(pb.ProtoName), true)
	chanStruct, err := asyncapi.BuildProtoChannelStruct(ctx, channel, parent, pb.ProtoName, golangName)
	if err != nil {
		return nil, err
	}

	chanStruct.Fields = append(
		chanStruct.Fields,
		lang.GoStructField{Name: "exchange", Type: &lang.GoSimple{TypeName: "string"}},
		lang.GoStructField{Name: "queue", Type: &lang.GoSimple{TypeName: "string"}},
		lang.GoStructField{Name: "routingKey", Type: &lang.GoSimple{TypeName: "string"}},
	)

	return &render.ProtoChannel{
		Channel:  parent,
		Type:     chanStruct,
		Protocol: pb.ProtoName,
	}, nil
}

func (pb ProtoBuilder) BuildChannelBindings(ctx *common.CompileContext, rawData types.Union2[json.RawMessage, yaml.Node]) (vals *lang.GoValue, jsonVals types.OrderedMap[string, string], err error) {
	var bindings channelBindings
	if err = types.UnmarshalRawsUnion2(rawData, &bindings); err != nil {
		err = types.CompileError{Err: err, Path: ctx.PathStackRef(), Proto: pb.ProtoName}
		return
	}

	vals = &lang.GoValue{Type: &lang.GoSimple{TypeName: "ChannelBindings", Import: ctx.RuntimeModule(pb.ProtoName)}, EmptyCurlyBrackets: true}
	switch bindings.Is {
	case "queue":
		vals.StructValues.Set("ChannelType", &lang.GoSimple{TypeName: "ChannelTypeQueue", Import: ctx.RuntimeModule(pb.ProtoName)})
	default: // routingKey and any other value
		vals.StructValues.Set("ChannelType", &lang.GoSimple{TypeName: "ChannelTypeRoutingKey", Import: ctx.RuntimeModule(pb.ProtoName)})
	}
	if bindings.Exchange != nil {
		ecVals := lang.ConstructGoValue(*bindings.Exchange, []string{"Type"}, &lang.GoSimple{TypeName: "ExchangeConfiguration", Import: ctx.RuntimeModule(pb.ProtoName)})
		switch bindings.Exchange.Type {
		case "default":
			ecVals.StructValues.Set("Type", &lang.GoSimple{TypeName: "ExchangeTypeDefault", Import: ctx.RuntimeModule(pb.ProtoName)})
		case "topic":
			ecVals.StructValues.Set("Type", &lang.GoSimple{TypeName: "ExchangeTypeTopic", Import: ctx.RuntimeModule(pb.ProtoName)})
		case "direct":
			ecVals.StructValues.Set("Type", &lang.GoSimple{TypeName: "ExchangeTypeDirect", Import: ctx.RuntimeModule(pb.ProtoName)})
		case "fanout":
			ecVals.StructValues.Set("Type", &lang.GoSimple{TypeName: "ExchangeTypeFanout", Import: ctx.RuntimeModule(pb.ProtoName)})
		case "headers":
			ecVals.StructValues.Set("Type", &lang.GoSimple{TypeName: "ExchangeTypeHeaders", Import: ctx.RuntimeModule(pb.ProtoName)})
		}
		vals.StructValues.Set("ExchangeConfiguration", ecVals)
	}
	qVals := lang.ConstructGoValue(*bindings.Queue, nil, &lang.GoSimple{TypeName: "QueueConfiguration", Import: ctx.RuntimeModule(pb.ProtoName)})
	vals.StructValues.Set("QueueConfiguration", qVals)

	return
}

func (pb ProtoBuilder) BuildOperationBindings(ctx *common.CompileContext, rawData types.Union2[json.RawMessage, yaml.Node]) (vals *lang.GoValue, jsonVals types.OrderedMap[string, string], err error) {
	var bindings operationBindings
	if err = types.UnmarshalRawsUnion2(rawData, &bindings); err != nil {
		err = types.CompileError{Err: err, Path: ctx.PathStackRef(), Proto: pb.ProtoName}
		return
	}

	vals = lang.ConstructGoValue(bindings, []string{"Expiration", "DeliveryMode"}, &lang.GoSimple{TypeName: "OperationBindings", Import: ctx.RuntimeModule(pb.ProtoName)})
	if bindings.Expiration > 0 {
		v := lang.ConstructGoValue(bindings.Expiration*int(time.Millisecond), nil, &lang.GoSimple{TypeName: "Duration", Import: "time"})
		vals.StructValues.Set("Expiration", v)
	}
	switch bindings.DeliveryMode {
	case 1:
		vals.StructValues.Set("DeliveryMode", &lang.GoSimple{TypeName: "DeliveryModeTransient", Import: ctx.RuntimeModule(pb.ProtoName)})
	case 2:
		vals.StructValues.Set("DeliveryMode", &lang.GoSimple{TypeName: "DeliveryModePersistent", Import: ctx.RuntimeModule(pb.ProtoName)})
	}

	return
}
