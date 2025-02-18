package render

import (
	"cmp"
	"fmt"
	"slices"

	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render/lang"
	"github.com/samber/lo"
)

type Message struct {
	OriginalName string
	OutType      *lang.GoStruct
	InType       *lang.GoStruct
	Dummy        bool
	IsSelectable bool // true if message should get to selections
	IsPublisher  bool
	IsSubscriber bool

	HeadersFallbackType *lang.GoMap
	HeadersTypePromise  *lang.Promise[*lang.GoStruct]

	AllActiveChannelsPromise   *lang.ListPromise[common.Renderable]
	AllActiveOperationsPromise *lang.ListPromise[common.Renderable]

	BindingsType    *lang.GoStruct           // nil if message bindings are not defined for message
	BindingsPromise *lang.Promise[*Bindings] // nil if message bindings are not defined for message as well

	ContentType          string                        // Message's content type
	CorrelationIDPromise *lang.Promise[*CorrelationID] // nil if correlationID is not defined for message
	PayloadType          common.GolangType             // `any` or a particular type
	AsyncAPIPromise      *lang.Promise[*AsyncAPI]

	ProtoMessages []*ProtoMessage
}

func (m *Message) Kind() common.ObjectKind {
	return common.ObjectKindMessage
}

func (m *Message) Selectable() bool {
	return !m.Dummy && m.IsSelectable // Select only the messages defined in the `channels` section`
}

func (m *Message) Visible() bool {
	return !m.Dummy
}

func (m *Message) SelectProtoObject(protocol string) common.Renderable {
	objects := lo.Filter(m.ProtoMessages, func(p *ProtoMessage, _ int) bool {
		return p.Selectable() && p.Protocol == protocol
	})
	return lo.FirstOr(objects, nil)
}

func (m *Message) Name() string {
	return m.OriginalName
}

func (m *Message) EffectiveContentType() string {
	if m.Dummy {
		return ""
	}
	res, _ := lo.Coalesce(m.ContentType, m.AsyncAPIPromise.T().EffectiveDefaultContentType())
	return res
}

func (m *Message) BindingsProtocols() (res []string) {
	if m.BindingsType == nil {
		return nil
	}
	if m.BindingsPromise != nil {
		res = append(res, m.BindingsPromise.T().Values.Keys()...)
		res = append(res, m.BindingsPromise.T().JSONValues.Keys()...)
	}
	return lo.Uniq(res)
}

func (m *Message) HeadersType() *lang.GoStruct {
	if m.HeadersTypePromise != nil {
		return m.HeadersTypePromise.T()
	}
	return nil
}

func (m *Message) Bindings() *Bindings {
	if m.BindingsPromise != nil {
		return m.BindingsPromise.T()
	}
	return nil
}

func (m *Message) CorrelationID() *CorrelationID {
	if m.CorrelationIDPromise != nil {
		return m.CorrelationIDPromise.T()
	}
	return nil
}

func (m *Message) AsyncAPI() *AsyncAPI {
	return m.AsyncAPIPromise.T()
}

func (m *Message) BoundChannels() []common.Renderable {
	r := lo.Filter(m.AllActiveChannelsPromise.T(), func(c common.Renderable, _ int) bool {
		ch := common.DerefRenderable(c).(*Channel)
		return lo.ContainsBy(ch.BoundMessages(), func(item common.Renderable) bool {
			return common.CheckSameRenderables(item, m)
		})
	})
	// ListPromise is filled up by linker, which doesn't guarantee the order. So, sort items by name
	slices.SortFunc(r, func(a, b common.Renderable) int { return cmp.Compare(a.Name(), b.Name()) })
	return r
}

func (m *Message) BoundOperations() []common.Renderable {
	r := lo.Filter(m.AllActiveOperationsPromise.T(), func(o common.Renderable, _ int) bool {
		op := common.DerefRenderable(o).(*Operation)
		return lo.ContainsBy(op.BoundMessages(), func(item common.Renderable) bool {
			return common.CheckSameRenderables(item, m)
		})
	})
	// ListPromise is filled up by linker, which doesn't guarantee the order. So, sort items by name
	slices.SortFunc(r, func(a, b common.Renderable) int { return cmp.Compare(a.Name(), b.Name()) })
	return r
}

func (m *Message) String() string {
	return "Message " + m.OriginalName
}

func (m *Message) ProtoBindingsValue(protoName string) common.Renderable {
	res := &lang.GoValue{
		Type:               &lang.GoSimple{TypeName: "ServerBindings", Import: protoName, RuntimeImport: true},
		EmptyCurlyBrackets: true,
	}
	if m.BindingsPromise != nil {
		if b, ok := m.BindingsPromise.T().Values.Get(protoName); ok {
			res = b
		}
	}
	return res
}

type ProtoMessage struct {
	*Message
	Protocol string
}

func (p *ProtoMessage) Selectable() bool {
	return !p.Dummy && p.isBound()
}

func (p *ProtoMessage) String() string {
	return fmt.Sprintf("ProtoMessage[%s] %s", p.Protocol, p.OriginalName)
}

// isBound returns true if the message is bound to the protocol
func (p *ProtoMessage) isBound() bool {
	res := lo.SomeBy(p.BoundChannels(), func(c common.Renderable) bool {
		ch := common.DerefRenderable(c).(*Channel)
		return !lo.IsNil(ch.SelectProtoObject(p.Protocol))
	}) || lo.SomeBy(p.BoundOperations(), func(o common.Renderable) bool {
		op := common.DerefRenderable(o).(*Operation)
		return !lo.IsNil(op.SelectProtoObject(p.Protocol))
	})

	return res
}
