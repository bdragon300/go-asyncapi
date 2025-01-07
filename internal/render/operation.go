package render

import (
	"fmt"
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render/lang"
	"github.com/samber/lo"
)

type Operation struct {
	OriginalName string  // Actually it isn't used
	Dummy        bool
	IsSelectable bool // true if channel should get to selections
	IsPublisher  bool
	IsSubscriber bool

	ChannelPromise *lang.Promise[*Channel] // Channel this operation bound with

	BindingsType             *lang.GoStruct           // nil if no bindings are set for operation at all
	BindingsPromise *lang.Promise[*Bindings] // nil if no bindings are set for operation at all

	MessagesPromises []*lang.Promise[*Message]

	ProtoOperations []*ProtoOperation // Proto operations for each protocol that ChannelPromise has
}

func (o *Operation) Kind() common.ObjectKind {
	return common.ObjectKindOperation  // TODO: separate Bindings from Channel, leaving only the Promise, and make its own 4 ObjectKinds
}

func (o *Operation) Selectable() bool {
	// Proto channels for each supported protocol
	// If bound channel is not selectable, then operation is not selectable as well
	return !o.Dummy && o.ChannelPromise.T().Selectable() && o.IsSelectable
}

func (o *Operation) Visible() bool {
	return !o.Dummy && o.ChannelPromise.T().Visible()
}

func (o *Operation) String() string {
	return "Operation " + o.OriginalName
}

func (o *Operation) Name() string {
	return o.OriginalName
}

func (o *Operation) SelectProtoObject(protocol string) common.Renderable {
	res := lo.Filter(o.ProtoOperations, func(p *ProtoOperation, _ int) bool {
		return p.Selectable() && p.Protocol == protocol
	})
	if len(res) > 0 {
		return res[0]
	}
	return nil
}

func (o *Operation) Channel() *Channel {
	return o.ChannelPromise.T()
}

func (o *Operation) Bindings() *Bindings {
	return o.BindingsPromise.T()
}

func (o *Operation) Messages() []common.Renderable {
	return lo.Map(o.MessagesPromises, func(prm *lang.Promise[*Message], _ int) common.Renderable { return prm.T() })
}

func (o *Operation) BoundMessages() []common.Renderable {
	return o.Messages()
}

func (o *Operation) BindingsProtocols() (res []string) {
	if o.BindingsType == nil {
		return nil
	}
	if o.BindingsPromise != nil {
		res = append(res, o.BindingsPromise.T().Values.Keys()...)
		res = append(res, o.BindingsPromise.T().JSONValues.Keys()...)
	}
	return lo.Uniq(res)
}

func (o *Operation) ProtoBindingsValue(protoName string) common.Renderable {
	res := &lang.GoValue{
		Type:               &lang.GoSimple{TypeName: "OperationBindings", Import: common.GetContext().RuntimeModule(protoName)},
		EmptyCurlyBrackets: true,
	}
	if o.BindingsPromise != nil {
		if b, ok := o.BindingsPromise.T().Values.Get(protoName); ok {
			res = b
		}
	}
	return res
}

type ProtoOperation struct {
	*Operation

	ProtoChannelPromise *lang.Promise[*ProtoChannel]
	Type                *lang.GoStruct
	Protocol string
}

func (p *ProtoOperation) Selectable() bool {
	return !p.Dummy && p.isBound()
}

func (p *ProtoOperation) String() string {
	return fmt.Sprintf("ProtoOperation[%s] %s", p.Protocol, p.OriginalName)
}

func (p *ProtoOperation) ProtoChannel() *ProtoChannel {
	return p.ProtoChannelPromise.T()
}

// isBound returns true if operation is bound to at least one server with supported protocol
func (p *ProtoOperation) isBound() bool {
	protos := lo.Map(p.ChannelPromise.T().BoundServers(), func(s common.Renderable, _ int) string {
		srv := common.DerefRenderable(s).(*Server)
		return srv.Protocol
	})
	r := lo.Contains(protos, p.Protocol)
	return r
}
