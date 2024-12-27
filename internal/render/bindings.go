package render

import (
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render/lang"
	"github.com/bdragon300/go-asyncapi/internal/types"
	"github.com/bdragon300/go-asyncapi/internal/utils"
)

// Bindings never renders itself, only as a part of other object
type Bindings struct {
	OriginalName string  // Actually it isn't used

	Values types.OrderedMap[string, *lang.GoValue] // Binding values by protocol
	// Value of jsonschema fields as json marshalled strings
	JSONValues types.OrderedMap[string, types.OrderedMap[string, string]] // Binbing values by protocol
}

func (b *Bindings) Kind() common.ObjectKind {
	return common.ObjectKindOther  // TODO: separate Bindings from Channel, leaving only the Promise, and make its own 4 ObjectKinds
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
	return utils.CapitalizeUnchanged(b.OriginalName)
}
