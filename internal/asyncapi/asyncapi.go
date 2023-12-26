package asyncapi

import (
	"fmt"

	"github.com/bdragon300/asyncapi-codegen-go/internal/types"

	"github.com/bdragon300/asyncapi-codegen-go/internal/common"
)

type AsyncAPI struct {
	Asyncapi           string                            `json:"asyncapi" yaml:"asyncapi"`
	ID                 string                            `json:"id" yaml:"id"`
	Info               InfoItem                          `json:"info" yaml:"info"`
	Servers            types.OrderedMap[string, Server]  `json:"servers" yaml:"servers" cgen:"directRender,packageDown=servers"`
	DefaultContentType string                            `json:"defaultContentType" yaml:"defaultContentType"`
	Channels           types.OrderedMap[string, Channel] `json:"channels" yaml:"channels" cgen:"directRender,packageDown=channels"`
	Components         ComponentsItem                    `json:"components" yaml:"components"`
	Tags               []Tag                             `json:"tags" yaml:"tags"`
	ExternalDocs       ExternalDocumentation             `json:"externalDocs" yaml:"externalDocs"`
}

func (a AsyncAPI) Compile(ctx *common.CompileContext) error {
	if a.DefaultContentType != "" {
		ctx.Storage.SetDefaultContentType(a.DefaultContentType)
	}
	ctx.Logger.Trace(fmt.Sprintf("Default content type set to %s", ctx.Storage.DefaultContentType()))
	return nil
}

type InfoItem struct {
	Title       string `json:"title" yaml:"title"`
	Version     string `json:"version" yaml:"version"`
	Description string `json:"description" yaml:"description"`
}

type ExternalDocumentation struct{}
