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
	lang.BaseJSONPointed
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
	AllActiveChannelsPromise *lang.ListPromise[common.Artifact]

	// VariablesPromises is a list of server variables defined for this server.
	VariablesPromises types.OrderedMap[string, *lang.Promise[*ServerVariable]]

	// BindingsPromise is a promise to server bindings contents. Nil if no bindings are set.
	BindingsPromise *lang.Promise[*Bindings]

	// SecuritySchemePromises is a promises to the security scheme objects defined for this server.
	SecuritySchemePromises []*lang.Promise[*SecurityScheme]
}

// Variables returns the [types.OrderedMap] with server variables by name. Returns empty [types.OrderedMap] if variables
// are not set.
func (s *Server) Variables() (res types.OrderedMap[string, *ServerVariable]) {
	for k, v := range s.VariablesPromises.Entries() {
		res.Set(k, v.T())
	}
	return
}

// Bindings returns the Bindings object or nil if bindings are not set.
func (s *Server) Bindings() *Bindings {
	if s.BindingsPromise != nil {
		return s.BindingsPromise.T()
	}
	return nil
}

// SecuritySchemes returns the list of security schemes defined for this server.
func (s *Server) SecuritySchemes() []*SecurityScheme {
	return lo.Map(s.SecuritySchemePromises, func(item *lang.Promise[*SecurityScheme], _ int) *SecurityScheme {
		return item.T()
	})
}

// BoundChannels returns a list of channels that are bound to this server.
func (s *Server) BoundChannels() []*Channel {
	r := lo.FilterMap(s.AllActiveChannelsPromise.T(), func(r common.Artifact, _ int) (*Channel, bool) {
		ch := common.DerefArtifact(r).(*Channel)
		return ch, lo.ContainsBy(ch.BoundServers(), func(item *Server) bool {
			return common.CheckSameArtifacts(s, item)
		})
	})
	return r
}

// BoundOperations returns a list of operations that are bound to this server through its channels.
func (s *Server) BoundOperations() []*Operation {
	r := lo.FlatMap(s.BoundChannels(), func(ch *Channel, _ int) []*Operation {
		return ch.BoundOperations()
	})
	return r
}

// ActiveProtocols returns a protocol of this server. This method exists mainly for internal use in selectors.
func (s *Server) ActiveProtocols() []string {
	return []string{s.Protocol}
}

// BindingsProtocols returns a list of protocols that have bindings defined for this server.
func (s *Server) BindingsProtocols() []string {
	if s.Bindings() != nil {
		return lo.Uniq(s.Bindings().Protocols())
	}
	return nil
}

// URL typically is used in templates to get the inflated server url by server variables values from the tool config.
//
// By default, the function returns the full server url made from Host and Pathname for any input value. However, if
// input is *non-empty* []common.InfraServerVariableOpts, it additionally substitutes the given variables to the host
// expression in the url. See asyncapi standard for more info.
func (s *Server) URL(input any) (*url.URL, error) {
	variables, ok := input.([]common.InfraServerVariableOpts)
	if lo.IsNil(input) || !ok || len(variables) == 0 {
		return &url.URL{Scheme: s.Protocol, Host: s.Host, Path: s.Pathname}, nil
	}

	res := &url.URL{Scheme: s.Protocol, Path: s.Pathname}
	params := lo.SliceToMap(variables, func(v common.InfraServerVariableOpts) (string, string) {
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

func (s *Server) Kind() common.ArtifactKind {
	return common.ArtifactKindServer
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

func (s *Server) Pinnable() bool {
	return true
}
