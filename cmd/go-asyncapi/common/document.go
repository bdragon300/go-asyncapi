package common2

import (
	"github.com/bdragon300/go-asyncapi/internal/jsonpointer"
	"github.com/bdragon300/go-asyncapi/internal/types"
)

func NewDocumentTree(originDocument *jsonpointer.JSONPointer) *DocumentTree {
	return &DocumentTree{RawNode: types.NewEmptyRawNode(types.RawNodeKindObject, nil, originDocument)}
}

type DocumentTree struct {
	*types.RawNode
}
