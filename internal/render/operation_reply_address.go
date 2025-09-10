package render

import (
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render/lang"
)

type OperationReplyAddress struct {
	lang.BaseJSONPointed
	lang.BaseRuntimeExpression

	// OriginalName is the name of the operation reply address as it was defined in the AsyncAPI document.
	OriginalName string

	// Description is an optional address description. Renders as Go doc comment.
	Description string

	Dummy bool
}

func (o *OperationReplyAddress) Name() string {
	return o.OriginalName
}

func (o *OperationReplyAddress) Kind() common.ArtifactKind {
	return common.ArtifactKindOther
}

func (o *OperationReplyAddress) Selectable() bool {
	return false
}

func (o *OperationReplyAddress) Visible() bool {
	return !o.Dummy
}

func (o *OperationReplyAddress) String() string {
	return "OperationReplyAddress(" + o.OriginalName + ")"
}
