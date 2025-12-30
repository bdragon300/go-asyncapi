package lang

import (
	"fmt"

	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/types"
	"github.com/samber/lo"
)

// GoValue represents a Go value of any other type that can be rendered in Go code.
//
// This value can be a constant, a struct, array or map initialization expression. This is suitable when some data
// from the AsyncAPI document should get to the code as initialization value of some type. For example, the AsyncAPI bindings.
type GoValue struct {
	BaseJSONPointed
	// Type is the type that is rendered before the value, e.g. ``int(123)'' or ``map[string]string{"123", "456"}''.
	// If nil, the value will be rendered as a bare untyped value, like ``123'' or ``{"123", "456"}''.
	Type common.GolangType
	// EmptyCurlyBrackets affects to the rendering if Empty returns true. If true, then empty value will
	// be rendered as ``{}``, otherwise as ``nil``.
	EmptyCurlyBrackets bool // If GoValue is empty: `{}` if true, or `nil` otherwise
	// LiteralValue is a value that should be rendered as a literal value, like ``123`` or ``"hello"``.
	LiteralValue any
	// ArrayValues is a list of values that should be rendered as an array/slice initialization in curly brackets.
	ArrayValues []any
	// StructValues is a list of key-value pairs that should be rendered as a struct initialization in curly brackets.
	StructValues types.OrderedMap[string, any]
	// MapValues is a list of key-value pairs that should be rendered as a map initialization in curly brackets.
	MapValues types.OrderedMap[string, any]
}

// Empty returns true if the GoValue represents nil or empty or zero value.
func (gv *GoValue) Empty() bool {
	return gv.LiteralValue == nil && gv.StructValues.Len() == 0 && gv.MapValues.Len() == 0 && gv.ArrayValues == nil
}

func (gv *GoValue) Name() string {
	return ""
}

func (gv *GoValue) Kind() common.ArtifactKind {
	return common.ArtifactKindOther
}

func (gv *GoValue) Selectable() bool {
	return false
}

func (gv *GoValue) Visible() bool {
	return true
}

func (gv *GoValue) String() string {
	switch {
	case gv.LiteralValue != nil:
		return fmt.Sprintf("GoValue:%v", gv.LiteralValue)
	case gv.StructValues.Len() > 0:
		return fmt.Sprintf("GoValue:{%v...}", lo.Slice(gv.StructValues.Entries(), 0, 2))
	case gv.MapValues.Len() > 0:
		return fmt.Sprintf("GoValue:{%v...}", lo.Slice(gv.MapValues.Entries(), 0, 2))
	case gv.ArrayValues != nil:
		return fmt.Sprintf("GoValue:{%v...}", lo.Slice(gv.ArrayValues, 0, 2))
	}
	return "GoValue:nil"
}

func (gv *GoValue) CanBeAddressed() bool {
	return gv.Type != nil && gv.Type.CanBeAddressed()
}

func (gv *GoValue) CanBeDereferenced() bool {
	return gv.Type != nil && gv.Type.CanBeDereferenced()
}

func (gv *GoValue) GoTemplate() string {
	return "code/lang/govalue"
}
