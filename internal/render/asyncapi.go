package render

import (
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/samber/lo"
)

const fallbackContentType = "application/json" // Default content type if it omitted in spec

type AsyncAPI struct {
	DefaultContentType string
}

func (a *AsyncAPI) Kind() common.ObjectKind {
	return common.ObjectKindAsyncAPI
}

func (a *AsyncAPI) Selectable() bool {
	return true
}

func (a *AsyncAPI) Visible() bool {
	return true
}

func (a *AsyncAPI) EffectiveDefaultContentType() string {
	res, _ := lo.Coalesce(a.DefaultContentType, fallbackContentType)
	return res
}

func (a *AsyncAPI) String() string {
	return "AsyncAPI"
}

func (a *AsyncAPI) Name() string {
	return ""
}
