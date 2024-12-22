package lang

import (
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/utils"
	"github.com/samber/lo"
)

func NewRef(ref string, name string, selectable *bool) *Ref {
	return &Ref{
		Promise:    *newAssignCbPromise[common.Renderable](ref, common.PromiseOriginUser, nil, nil),
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
	return "Ref -> " + r.ref
}

func (r *Ref) Name() string {
	n, _ := lo.Coalesce(utils.CapitalizeUnchanged(r.name), r.target.Name())
	return n
}

func (r *Ref) UnwrapRenderable() common.Renderable {
	return unwrapRenderablePromiseOrRef(r.target)
}

