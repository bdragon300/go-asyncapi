package asyncapi

import (
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render"
	"github.com/bdragon300/go-asyncapi/internal/types"
)

type AsyncAPI struct {
	Asyncapi           string                              `json:"asyncapi" yaml:"asyncapi"`
	ID                 string                              `json:"id" yaml:"id"`
	Info               InfoItem                            `json:"info" yaml:"info"`
	Servers            types.OrderedMap[string, Server]    `json:"servers" yaml:"servers" cgen:"selectable"`
	DefaultContentType string                              `json:"defaultContentType" yaml:"defaultContentType"`
	Channels           types.OrderedMap[string, Channel]   `json:"channels" yaml:"channels" cgen:"selectable"`
	Operations         types.OrderedMap[string, Operation] `json:"operations" yaml:"operations" cgen:"selectable"`
	Components         ComponentsItem                      `json:"components" yaml:"components"`
}

func (a AsyncAPI) Compile(ctx *common.CompileContext) error {
	obj := a.build(ctx)
	ctx.PutObject(obj)
	return nil
}

func (a AsyncAPI) build(ctx *common.CompileContext) *render.AsyncAPI {
	ctx.Logger.Trace("AsyncAPI root object")
	res := &render.AsyncAPI{
		DefaultContentType: a.DefaultContentType,
	}
	return res
}

type InfoItem struct {
	Title          string                `json:"title" yaml:"title"`
	Version        string                `json:"version" yaml:"version"`
	Description    string                `json:"description" yaml:"description"`
	TermsOfService string                `json:"termsOfService" yaml:"termsOfService"`
	Contact        ContactItem           `json:"contact" yaml:"contact"`
	License        LicenseItem           `json:"license" yaml:"license"`
	Tags           []Tag                 `json:"tags" yaml:"tags"`
	ExternalDocs   ExternalDocumentation `json:"externalDocs" yaml:"externalDocs"`
}

type ContactItem struct {
	Name  string `json:"name" yaml:"name"`
	URL   string `json:"url" yaml:"url"`
	Email string `json:"email" yaml:"email"`
}

type LicenseItem struct {
	Name string `json:"name" yaml:"name"`
	URL  string `json:"url" yaml:"url"`
}

type ExternalDocumentation struct {
	Description string `json:"description" yaml:"description"`
	URL         string `json:"url" yaml:"url"`
}
