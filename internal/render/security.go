package render

import (
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render/lang"
	"github.com/samber/lo"
)

type SecurityScheme struct {
	lang.BaseJSONPointed
	// OriginalName is the name of the security object as it was defined in the AsyncAPI document.
	OriginalName string

	SchemeType  string
	Description string

	// InitValues contains the constant initialization values for this security generated struct.
	InitValues *lang.GoValue

	// Dummy is true when security scheme is ignored (x-ignore: true)
	Dummy bool

	// AllSecuredServersPromise contains all servers with security scheme applied.
	AllSecuredServersPromise *lang.ListPromise[common.Artifact]
	// AllSecuredOperationsPromise contains all operations with security scheme applied.
	AllSecuredOperationsPromise *lang.ListPromise[common.Artifact]
}

func (s *SecurityScheme) BoundServers() []*Server {
	r := lo.FilterMap(s.AllSecuredServersPromise.T(), func(r common.Artifact, _ int) (*Server, bool) {
		srv := common.DerefArtifact(r).(*Server)
		return srv, lo.ContainsBy(srv.SecuritySchemes(), func(item *SecurityScheme) bool {
			return common.CheckSameArtifacts(s, item)
		})
	})
	return r
}

func (s *SecurityScheme) BoundOperations() []*Operation {
	r := lo.FilterMap(s.AllSecuredOperationsPromise.T(), func(r common.Artifact, _ int) (*Operation, bool) {
		op := common.DerefArtifact(r).(*Operation)
		return op, lo.ContainsBy(op.SecuritySchemes(), func(item *SecurityScheme) bool {
			return common.CheckSameArtifacts(s, item)
		})
	})
	return r
}

func (s *SecurityScheme) Name() string {
	return s.OriginalName
}

func (s *SecurityScheme) Kind() common.ArtifactKind {
	return common.ArtifactKindSecurity
}

func (s *SecurityScheme) Selectable() bool {
	return !s.Dummy
}

func (s *SecurityScheme) Visible() bool {
	return !s.Dummy
}

func (s *SecurityScheme) String() string {
	return "SecurityScheme(" + s.OriginalName + ")"
}

func (s *SecurityScheme) Pinnable() bool {
	return true
}
