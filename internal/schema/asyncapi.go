package schema

import (
	"github.com/bdragon300/asyncapi-codegen/internal/utils"
)

type AsyncAPI struct {
	Asyncapi           string                                `json:"asyncapi" yaml:"asyncapi"`
	ID                 string                                `json:"id" yaml:"id"`
	Info               InfoItem                              `json:"info" yaml:"info"`
	Servers            ServerItem                            `json:"servers" yaml:"servers"`
	DefaultContentType string                                `json:"defaultContentType" yaml:"defaultContentType"`
	Channels           utils.OrderedMap[string, ChannelItem] `json:"channels" yaml:"channels"`
	Components         ComponentsItem                        `json:"components" yaml:"components"`
	Tags               []TagItem                             `json:"tags" yaml:"tags"`
	ExternalDocs       ExternalDocsItem                      `json:"externalDocs" yaml:"externalDocs"`
}

type InfoItem struct {
	Title       string `json:"title" yaml:"title"`
	Version     string `json:"version" yaml:"version"`
	Description string `json:"description" yaml:"description"`
}

type ServerItem struct{}

type ChannelItem struct{}

type TagItem struct{}

type ExternalDocsItem struct{}

