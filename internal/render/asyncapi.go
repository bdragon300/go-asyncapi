package render

import (
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render/lang"
	"github.com/samber/lo"
)

const fallbackContentType = "application/json" // Default content type if it omitted in spec


type AsyncAPI struct {
	AllMessages        *lang.ListPromise[*Message]
	DefaultContentType string
}

func (a AsyncAPI) Kind() common.ObjectKind {
	return common.ObjectKindAsyncAPI
}


func (a AsyncAPI) Selectable() bool {
	return true
}

func (a AsyncAPI) EffectiveDefaultContentType() string {
	res, _ := lo.Coalesce(a.DefaultContentType, fallbackContentType)
	return res
}

func (a AsyncAPI) String() string {
	return "AsyncAPI"
}

//// SpecEffectiveContentTypes returns a list of all unique content types used in the spec. This includes all content
//// types from all messages and the default content type.
//func (a AsyncAPI) SpecEffectiveContentTypes() []string {
//	return lo.Uniq(lo.Map(a.AllMessages.Targets(), func(item *Message, _ int) string {
//		contentType, _ := lo.Coalesce(item.ContentType, a.EffectiveDefaultContentType())
//		return contentType
//	}))
//}
