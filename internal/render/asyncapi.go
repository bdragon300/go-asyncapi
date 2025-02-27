package render

import (
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render/lang"
	"github.com/samber/lo"
)

// DefaultContentType is the default content type to use if none is set.
const DefaultContentType = "application/json"

// AsyncAPI represents the root of the AsyncAPI document.
type AsyncAPI struct {
	lang.BasePositioned
	DefaultContentType string
}

func (a *AsyncAPI) Name() string {
	return ""
}

func (a *AsyncAPI) Kind() common.ArtifactKind {
	return common.ArtifactKindAsyncAPI
}

func (a *AsyncAPI) Selectable() bool {
	return true
}

func (a *AsyncAPI) Visible() bool {
	return true
}

func (a *AsyncAPI) EffectiveDefaultContentType() string {
	res, _ := lo.Coalesce(a.DefaultContentType, DefaultContentType)
	return res
}

func (a *AsyncAPI) String() string {
	return "AsyncAPI"
}
