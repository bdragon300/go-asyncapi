package render

import (
	"fmt"
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render/lang"
	"github.com/bdragon300/go-asyncapi/internal/types"
	"github.com/bdragon300/go-asyncapi/run"
	"github.com/samber/lo"
	"net/url"
)

type Server struct {
	OriginalName string
	Dummy        bool // x-ignore is set
	IsSelectable bool // true if server is defined in `components` section
	IsPublisher bool
	IsSubscriber bool

	AllActiveChannelsPromise *lang.ListPromise[common.Renderable]

	Host string
	Pathname string
	Protocol        string
	ProtocolVersion string

	VariablesPromises  types.OrderedMap[string, *lang.Promise[*ServerVariable]]

	BindingsType    *lang.GoStruct           // nil if bindings are not defined for server
	BindingsPromise *lang.Promise[*Bindings] // nil if bindings are not defined for server as well

	ProtoServer *ProtoServer // nil if server is dummy or has unsupported protocol
}

func (s *Server) Kind() common.ObjectKind {
	return common.ObjectKindServer
}

func (s *Server) Selectable() bool {
	return !s.Dummy && s.IsSelectable // Select only the servers defined in the `channels` section`
}

func (s *Server) Visible() bool {
	return !s.Dummy
}

func (s *Server) SelectProtoObject(protocol string) common.Renderable {
	if s.ProtoServer.Selectable() && s.ProtoServer.Protocol == protocol {
		return s.ProtoServer
	}
	return nil
}

func (s *Server) Name() string {
	return s.OriginalName
}

func (s *Server) URL(variables []common.ConfigServerVariable) (*url.URL, error) {
	if len(variables) == 0 {
		return &url.URL{Scheme: s.Protocol, Host: s.Host, Path: s.Pathname}, nil
	}

	res := &url.URL{Scheme: s.Protocol, Path: s.Pathname}
	params := lo.SliceToMap(variables, func(v common.ConfigServerVariable) (string, string) {
		return v.Name, v.Value
	})
	h, err := run.ParamString{Expr: s.Host, Parameters: params}.Expand()
	if err != nil {
		return nil, fmt.Errorf("expand host %q: %w", s.Host, err)
	}
	res.Host = h
	return res, nil
}

func (s *Server) BoundChannels() []common.Renderable {
	r := lo.Filter(s.AllActiveChannelsPromise.T(), func(r common.Renderable, _ int) bool {
		ch := common.DerefRenderable(r).(*Channel)
		return lo.ContainsBy(ch.BoundServers(), func(item common.Renderable) bool {
			return common.CheckSameRenderables(s, item)
		})
	})
	return r
}

func (s *Server) BoundOperations() []common.Renderable {
	chans := s.BoundChannels()
	ops := lo.FlatMap(chans, func(c common.Renderable, _ int) []common.Renderable {
		ch := common.DerefRenderable(c).(*Channel)
		return ch.BoundOperations()
	})
	return ops
}

func (s *Server) String() string {
	return fmt.Sprintf("Server[%s] %s", s.Protocol, s.OriginalName)
}

func (s *Server) BindingsProtocols() (res []string) {
	if s.BindingsType == nil {
		return nil
	}
	if s.BindingsPromise != nil {
		res = append(res, s.BindingsPromise.T().Values.Keys()...)
		res = append(res, s.BindingsPromise.T().JSONValues.Keys()...)
	}
	return lo.Uniq(res)
}

func (s *Server) ProtoBindingsValue(protoName string) common.Renderable {
	res := &lang.GoValue{
		Type:               &lang.GoSimple{TypeName: "ServerBindings", Import: protoName, RuntimeImport: true},
		EmptyCurlyBrackets: true,
	}
	if s.BindingsPromise != nil {
		if b, ok := s.BindingsPromise.T().Values.Get(protoName); ok {
			res = b
		}
	}
	return res
}

func (s *Server) Variables() (res types.OrderedMap[string, *ServerVariable]) {
	for _, entry := range s.VariablesPromises.Entries() {
		res.Set(entry.Key, entry.Value.T())
	}
	return
}

func (s *Server) Bindings() (res *Bindings) {
	if s.BindingsPromise != nil {
		return s.BindingsPromise.T()
	}
	return nil
}

type ProtoServer struct {
	*Server
	Type *lang.GoStruct // Nil if server is dummy or has unsupported protocol
}

func (p *ProtoServer) String() string {
	return "ProtoServer " + p.OriginalName
}

func (p *ProtoServer) Selectable() bool {
	return !p.Dummy
}
