package render

import (
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render/lang"
	"github.com/bdragon300/go-asyncapi/internal/types"
)

// Bindings represents the bindings object. It's used for all binding types: channel, operation, message, and server.
type Bindings struct {
	lang.BaseJSONPointed
	OriginalName string

	// Values is constant bindings values by protocol
	Values types.OrderedMap[string, any]
}

func (b *Bindings) Protocols() []string {
	return b.Values.Keys()
}

func (b *Bindings) Name() string {
	return b.OriginalName
}

func (b *Bindings) Kind() common.ArtifactKind {
	return common.ArtifactKindOther
}

func (b *Bindings) Selectable() bool {
	return false
}

func (b *Bindings) Visible() bool {
	return true
}

func (b *Bindings) String() string {
	return "Bindings(" + b.OriginalName + ")"
}
