package asyncapi

import (
	"fmt"

	"github.com/bdragon300/go-asyncapi/internal/common"
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
	if a.DefaultContentType != "" {
		ctx.Storage.SetDefaultContentType(a.DefaultContentType)
	}
	ctx.Logger.Trace(fmt.Sprintf("Default content type set to %s", ctx.Storage.DefaultContentType()))

	ctx.Storage.SetActiveServers(a.Servers.Keys())
	ctx.Logger.Trace("Active servers list", "servers", ctx.Storage.ActiveServers())

	ctx.Storage.SetActiveChannels(a.Channels.Keys())
	ctx.Logger.Trace("Active channels list", "channels", ctx.Storage.ActiveChannels())
	return nil
}

type InfoItem struct {
	Title       string `json:"title" yaml:"title"`
	Version     string `json:"version" yaml:"version"`
	Description string `json:"description" yaml:"description"`
}

type ExternalDocumentation struct{}
