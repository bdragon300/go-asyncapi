package lang

import (
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/samber/lo"
	"strconv"
	"strings"
)

func NewRef(ref string, name string, selectable *bool) *Ref {
	return &Ref{
		Promise:    *newPromise[common.Renderable](ref, common.PromiseOriginUser, nil, nil),
		selectable: selectable,
		name:       name,
	}
}

type Ref struct {
	Promise[common.Renderable]
	selectable *bool
	name       string
}

func (r *Ref) Kind() common.ObjectKind {
	return r.target.Kind()
}

func (r *Ref) Selectable() bool {
	if r.selectable == nil {
		return r.origin == common.PromiseOriginUser && r.target.Selectable()
	}
	return r.origin == common.PromiseOriginUser && *r.selectable
}

func (r *Ref) Visible() bool {
	return r.origin == common.PromiseOriginUser && r.target.Visible()
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

	b.WriteString(" -> ")
	b.WriteString(r.ref)
	return b.String()
}

func (r *Ref) Name() string {
	n, _ := lo.Coalesce(r.name, r.target.Name())
	return n
}

func (r *Ref) UnwrapRenderable() common.Renderable {
	return common.DerefRenderable(r.target)
}

