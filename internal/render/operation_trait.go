package render

import (
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render/lang"
	"github.com/samber/lo"
)

// OperationTrait represents an operation trait object
type OperationTrait struct {
	lang.BaseJSONPointed
	// OriginalName is the name of the operation trait itself, typically this is a key where the operation trait is registered in the document.
	OriginalName string

	// SecuritySchemePromises is a promises to the security scheme objects defined for this operation.
	SecuritySchemePromises []*lang.Promise[*SecurityScheme]

	// BindingsPromise is a promise to message trait bindings contents. Nil if message trait bindings are not set.
	BindingsPromise *lang.Promise[*Bindings]
}

// SecuritySchemes returns the list of security schemes defined for this operation.
func (o *OperationTrait) SecuritySchemes() []*SecurityScheme {
	r := lo.Map(o.SecuritySchemePromises, func(item *lang.Promise[*SecurityScheme], _ int) *SecurityScheme {
		return item.T()
	})
	return r
}

// Bindings returns the Bindings object or nil if no bindings are set.
func (o *OperationTrait) Bindings() *Bindings {
	if o.BindingsPromise != nil {
		return o.BindingsPromise.T()
	}
	return nil
}

func (o *OperationTrait) Name() string {
	return o.OriginalName
}

func (o *OperationTrait) Kind() common.ArtifactKind {
	return common.ArtifactKindOther
}

func (o *OperationTrait) Selectable() bool {
	return false // OperationTrait only affects to the Operation rendering and does not produce additional code
}

func (o *OperationTrait) Visible() bool {
	return false
}

func (o *OperationTrait) String() string {
	return "OperationTrait(" + o.OriginalName + ")"
}
