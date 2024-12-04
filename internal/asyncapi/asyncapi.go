package asyncapi

import (
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render"
	"github.com/bdragon300/go-asyncapi/internal/render/lang"
	"github.com/bdragon300/go-asyncapi/internal/types"
)


type AsyncAPI struct {
	Asyncapi           string                            `json:"asyncapi" yaml:"asyncapi"`
	ID                 string                            `json:"id" yaml:"id"`
	Info               InfoItem                          `json:"info" yaml:"info"`
	Servers            types.OrderedMap[string, Server]  `json:"servers" yaml:"servers"`
	DefaultContentType string                            `json:"defaultContentType" yaml:"defaultContentType"`
	Channels           types.OrderedMap[string, Channel] `json:"channels" yaml:"channels"`
	Components         ComponentsItem                    `json:"components" yaml:"components"`
	Tags               []Tag                             `json:"tags" yaml:"tags"`
	ExternalDocs       ExternalDocumentation             `json:"externalDocs" yaml:"externalDocs"`
}

// TODO: make AsyncAPI as one of object kinds
func (a AsyncAPI) Compile(ctx *common.CompileContext) error {
	ctx.RegisterNameTop(ctx.Stack.Top().PathItem)
	obj := a.build(ctx)
	ctx.PutObject(obj)
	return nil
}

func (a AsyncAPI) build(ctx *common.CompileContext) *render.AsyncAPI {
	allMessagesPrm := lang.NewListCbPromise[*render.Message](func(item common.Renderable, _ []string) bool {
		_, ok := item.(*render.Message)
		return ok
	})
	ctx.PutListPromise(allMessagesPrm)

	res := &render.AsyncAPI{
		AllMessages:        allMessagesPrm,
		DefaultContentType: a.DefaultContentType,
	}
	return res
}

type InfoItem struct {
	Title       string `json:"title" yaml:"title"`
	Version     string `json:"version" yaml:"version"`
	Description string `json:"description" yaml:"description"`
}

type ExternalDocumentation struct{}
