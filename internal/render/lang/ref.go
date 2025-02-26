package lang

import (
	"strconv"
	"strings"

	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/samber/lo"
)

// NewRef returns a new Ref promise object. It receives the $ref URL, the name of entity where this $ref is located and
// the selectable flag. If selectable is nil, it doesn't affect anything, otherwise the Selectable method returns the
// value this flag.
func NewRef(ref string, name string, selectable *bool) *Ref {
	return &Ref{
		Promise:    *newPromise[common.Renderable](ref, common.PromiseOriginRef, nil, nil),
		selectable: selectable,
		name:       name,
	}
}

// Ref is promise object that is used specially for $ref urls in documents. It points to a single object, addresses
// the object only by ref.
//
// Ref matches to the [common.Renderable] interface and gets to the render selections, where it is commonly used in
// the templates to render the object that are referenced by $ref.
type Ref struct {
	Promise[common.Renderable]
	selectable *bool
	name       string
}

// Name returns the Ref.name if it is set, otherwise the target name. This is useful when we want to render the target
// object with the name of $ref.
//
// For example, we want to render the server defined in “components.servers” section with any foobar name, but
// referenced by $ref in “servers” root section. In this case the server should be named as in the root section.
func (r *Ref) Name() string {
	n, _ := lo.Coalesce(r.name, r.target.Name())
	return n
}

func (r *Ref) Kind() common.ObjectKind {
	return r.target.Kind()
}

func (r *Ref) Selectable() bool {
	if r.selectable == nil {
		return r.origin == common.PromiseOriginRef && r.target.Selectable()
	}
	return r.origin == common.PromiseOriginRef && *r.selectable
}

func (r *Ref) Visible() bool {
	return r.origin == common.PromiseOriginRef && r.target.Visible()
}

func (r *Ref) String() string {
	var b strings.Builder
	b.WriteString("Ref")
	if r.name != "" {
		b.WriteString("[name=")
		b.WriteString(r.name)
		b.WriteString("]")
	}
	if r.selectable != nil {
		b.WriteString("[selectable=")
		b.WriteString(strconv.FormatBool(*r.selectable))
		b.WriteString("]")
	}

	b.WriteString("->")
	b.WriteString(r.ref)
	return b.String()
}

func (r *Ref) UnwrapRenderable() common.Renderable {
	return common.DerefRenderable(r.target)
}
