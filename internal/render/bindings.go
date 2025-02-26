package render

import (
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render/lang"
	"github.com/bdragon300/go-asyncapi/internal/types"
)

// Bindings represents the bindings object. It's used for all binding types: channel, operation, message, and server.
type Bindings struct {
	OriginalName string

	// Values is constant bindings values by protocol
	Values types.OrderedMap[string, *lang.GoValue]
	// JSONValues is jsonschema bindings values by protocol
	JSONValues types.OrderedMap[string, types.OrderedMap[string, string]]
}

func (b *Bindings) Kind() common.ObjectKind {
	return common.ObjectKindOther // TODO: separate Bindings from Channel, leaving only the Promise, and make its own 4 ObjectKinds
}

func (b *Bindings) Selectable() bool {
	return false
}

func (b *Bindings) Visible() bool {
	return true
}

func (b *Bindings) String() string {
	return "Bindings " + b.OriginalName
}

func (b *Bindings) Name() string {
	return b.OriginalName
}
