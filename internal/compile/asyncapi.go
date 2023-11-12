package compile

import (
	"fmt"

	"github.com/bdragon300/asyncapi-codegen-go/internal/common"
	"github.com/bdragon300/asyncapi-codegen-go/internal/utils"
)

type AsyncAPI struct {
	Asyncapi           string                            `json:"asyncapi" yaml:"asyncapi"`
	ID                 string                            `json:"id" yaml:"id"`
	Info               InfoItem                          `json:"info" yaml:"info"`
	Servers            utils.OrderedMap[string, Server]  `json:"servers" yaml:"servers" cgen:"noinline,packageDown=servers"`
	DefaultContentType string                            `json:"defaultContentType" yaml:"defaultContentType"`
	Channels           utils.OrderedMap[string, Channel] `json:"channels" yaml:"channels" cgen:"noinline,packageDown=channels"`
	Components         ComponentsItem                    `json:"components" yaml:"components"`
	Tags               []Tag                             `json:"tags" yaml:"tags"`
	ExternalDocs       ExternalDocumentation             `json:"externalDocs" yaml:"externalDocs"`
}

func (a AsyncAPI) Compile(ctx *common.CompileContext) error {
	if a.DefaultContentType != "" {
		ctx.DefaultContentType = a.DefaultContentType
	}
	ctx.Logger.Trace(fmt.Sprintf("Default content type set to %s", ctx.DefaultContentType))
	return nil
}

type InfoItem struct {
	Title       string `json:"title" yaml:"title"`
	Version     string `json:"version" yaml:"version"`
	Description string `json:"description" yaml:"description"`
}

type ExternalDocumentation struct{}

