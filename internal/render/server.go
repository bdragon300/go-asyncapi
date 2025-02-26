package render

import (
	"fmt"
	"net/url"

	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render/lang"
	"github.com/bdragon300/go-asyncapi/internal/types"
	"github.com/bdragon300/go-asyncapi/run"
	"github.com/samber/lo"
)

// Server represents the server object.
type Server struct {
	// OriginalName is the name of the server as it was defined in the AsyncAPI document.
	OriginalName string
	// Host is the server host raw value.
	Host string
	// Pathname is the server pathname value.
	Pathname string
	// Protocol is the server protocol value, even if it isn't supported by the tool.
	Protocol string
	// ProtocolVersion is the server protocol version value.
	ProtocolVersion string

	// Dummy is true when server is ignored (x-ignore: true)
	Dummy bool
	// IsSelectable is true if server should get to selections
	IsSelectable bool
	// IsPublisher is true if the generation of publisher code is enabled
	IsPublisher bool
	// IsSubscriber is true if the generation of subscriber code is enabled
	IsSubscriber bool

	// AllActiveChannelsPromise contains all active channels in the document.
	//
	// On compiling stage we don't know which channels are bound to a particular server.
	// So we just collect all channels to this field and postpone filtering them until the rendering stage.
	//
	// We could use a promise callback to filter channels by server, but the server in channel is also a promise,
	// and the order of promises resolving is not guaranteed.
	AllActiveChannelsPromise *lang.ListPromise[common.Renderable]

	// VariablesPromises is a list of server variables defined for this server.
	VariablesPromises types.OrderedMap[string, *lang.Promise[*ServerVariable]]

	// BindingsType is a Go struct for server bindings. Nil if no bindings are set.
	BindingsType *lang.GoStruct
	// BindingsPromise is a promise to server bindings contents. Nil if no bindings are set.
	BindingsPromise *lang.Promise[*Bindings]

	// ProtoServer is a ProtoServer object built based on the server protocol.
	ProtoServer *ProtoServer
}

// Variables returns the [types.OrderedMap] with server variables by name. Returns empty [types.OrderedMap] if variables
// are not set.
func (s *Server) Variables() (res types.OrderedMap[string, *ServerVariable]) {
	for _, entry := range s.VariablesPromises.Entries() {
		res.Set(entry.Key, entry.Value.T())
	}
	return
}

// Bindings returns the Bindings object or nil if bindings are not set.
func (s *Server) Bindings() (res *Bindings) {
	if s.BindingsPromise != nil {
		return s.BindingsPromise.T()
	}
	return nil
}

// SelectProtoObject returns the ProtoServer object for the requested protocol. Returns nil if the server is not
// selectable or has another protocol
func (s *Server) SelectProtoObject(protocol string) common.Renderable {
	if s.ProtoServer.Selectable() && s.ProtoServer.Protocol == protocol {
		return s.ProtoServer
	}
	return nil
}

// BoundChannels returns a list of channels that are bound to this server.
func (s *Server) BoundChannels() []common.Renderable {
	r := lo.Filter(s.AllActiveChannelsPromise.T(), func(r common.Renderable, _ int) bool {
		ch := common.DerefRenderable(r).(*Channel)
		return lo.ContainsBy(ch.BoundServers(), func(item common.Renderable) bool {
			return common.CheckSameRenderables(s, item)
		})
	})
	return r
}

// BoundOperations returns a list of operations that are bound to this server.
func (s *Server) BoundOperations() []common.Renderable {
	chans := s.BoundChannels()
	ops := lo.FlatMap(chans, func(c common.Renderable, _ int) []common.Renderable {
		ch := common.DerefRenderable(c).(*Channel)
		return ch.BoundOperations()
	})
	return ops
}

// BindingsProtocols returns a list of protocols that have bindings defined for this server.
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

// ProtoBindingsValue returns the struct initialization [lang.GoValue] of BindingsType for the given protocol.
// The returned value contains all constant bindings values defined in document for the protocol.
// If no bindings are set for the protocol, returns an empty [lang.GoValue].
func (s *Server) ProtoBindingsValue(protoName string) common.Renderable {
	res := &lang.GoValue{
		Type:               &lang.GoSimple{TypeName: "ServerBindings", Import: protoName, IsRuntimeImport: true},
		EmptyCurlyBrackets: true,
	}
	if s.BindingsPromise != nil {
		if b, ok := s.BindingsPromise.T().Values.Get(protoName); ok {
			res = b
		}
	}
	return res
}

// URL typically is used in templates to get the inflated server url by server variables values from the tool config.
//
// By default, the function returns the full server url made from Host and Pathname for any input value. However, if
// input is *non-empty* []common.ConfigServerVariable, it additionally substitutes the given variables to the host
// expression in the url. See asyncapi standard for more info.
func (s *Server) URL(input any) (*url.URL, error) {
	variables, ok := input.([]common.ConfigServerVariable)
	if lo.IsNil(input) || !ok || len(variables) == 0 {
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

func (s *Server) Name() string {
	return s.OriginalName
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

func (s *Server) String() string {
	return fmt.Sprintf("Server[%s](%s)", s.Protocol, s.OriginalName)
}

type ProtoServer struct {
	*Server
	// Type is a server's Go struct.
	Type *lang.GoStruct
}

func (p *ProtoServer) Selectable() bool {
	return !p.Dummy
}

func (p *ProtoServer) String() string {
	return "ProtoServer(" + p.OriginalName + ")"
}
