package scancontext

func NewRefQuery[T any](ctx *Context, ref string) *RefQuery[T] {
	return &RefQuery[T]{
		ref:         ref,
		ctxSnapshot: ctx.Copy(),
	}
}

type RefQuery[T any] struct {
	ref         string
	Link        T
	ctxSnapshot *Context
}

func (r *RefQuery[T]) AssignLink(typ any) {
	r.Link = typ.(T)
}

func (r *RefQuery[T]) Ref() string {
	return r.ref
}
