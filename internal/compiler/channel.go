package compiler

import (
	"fmt"
	"strings"

	"github.com/bdragon300/asyncapi-codegen/internal/common"
	"github.com/bdragon300/asyncapi-codegen/internal/lang"
	"github.com/bdragon300/asyncapi-codegen/internal/render"
	"github.com/bdragon300/asyncapi-codegen/internal/scan"
	"github.com/bdragon300/asyncapi-codegen/internal/utils"
)

type Channel struct {
	Description string                              `json:"description" yaml:"description"`
	Servers     []string                            `json:"servers" yaml:"servers"`
	Subscribe   *Operation                          `json:"subscribe" yaml:"subscribe"`
	Publish     *Operation                          `json:"publish" yaml:"publish"`
	Parameters  utils.OrderedMap[string, Parameter] `json:"parameters" yaml:"parameters"`
	Bindings    utils.OrderedMap[string, any]       `json:"bindings" yaml:"bindings"` // TODO: replace any to common bindings object

	Ref string `json:"$ref" yaml:"$ref"`
}

func (c Channel) Build(ctx *scan.Context) error {
	obj, err := c.buildChannel(ctx, ctx.Top().Path)
	if err != nil {
		return fmt.Errorf("error on %q: %w", strings.Join(ctx.PathStack(), "."), err)
	}
	ctx.CurrentPackage().Put(ctx, obj)
	return nil
}

func (c Channel) buildChannel(ctx *scan.Context, name string) (render.LangRenderer, error) {
	if c.Ref != "" {
		res := lang.NewLinkerQueryRendererRef(common.ChannelsPackageKind, c.Ref)
		ctx.Linker.Add(res)
		return res, nil
	}
	res := &lang.Channel{SupportedProtocols: make(map[string]render.LangRenderer)}
	if len(c.Servers) > 0 {
		for _, srv := range c.Servers {
			path := []string{"servers", srv}
			lnk := lang.NewLinkerPathQuery[*lang.Server](common.ServersPackageKind, path)
			ctx.Linker.Add(lnk)
			res.AppliedServers = append(res.AppliedServers, lnk)
		}
	} else {
		lnk := lang.NewLinkerQueryList[*lang.Server](common.ServersPackageKind, []string{"servers"})
		ctx.Linker.AddMany(lnk)
		res.AppliedToAllServers = lnk
	}

	for pName, pBuild := range c.supportedProtocols() {
		obj, err := pBuild(ctx, name)
		if err != nil {
			return nil, fmt.Errorf("Unable to build %s protocol: %w", pName, err)
		}
		res.SupportedProtocols[pName] = obj
	}
	return res, nil
}

func (c Channel) supportedProtocols() map[string]protocolBuilderFunc {
	return map[string]protocolBuilderFunc{
		"kafka":        c.buildKafka,
		"kafka-secure": c.buildKafka,
	}
}

type Operation struct {
	OperationID  string                        `json:"operationId" yaml:"operationId"`
	Summary      string                        `json:"summary" yaml:"summary"`
	Description  string                        `json:"description" yaml:"description"`
	Security     []SecurityRequirement         `json:"security" yaml:"security"`
	Tags         []Tag                         `json:"tags" yaml:"tags"`
	ExternalDocs *ExternalDocumentation        `json:"externalDocs" yaml:"externalDocs"`
	Bindings     utils.OrderedMap[string, any] `json:"bindings" yaml:"bindings"` // TODO: replace any to common bindings object
	Traits       []OperationTrait              `json:"traits" yaml:"traits"`
	Message      *Message                      `json:"message" yaml:"message"`
}

type Parameter struct {
	Description string  `json:"description" yaml:"description"`
	Schema      *Object `json:"schema" yaml:"schema"`
	Location    string  `json:"location" yaml:"location"`

	Ref string `json:"$ref" yaml:"$ref"`
}

type OperationTrait struct {
	OperationID  string                        `json:"operationId" yaml:"operationId"`
	Summary      string                        `json:"summary" yaml:"summary"`
	Description  string                        `json:"description" yaml:"description"`
	Security     []SecurityRequirement         `json:"security" yaml:"security"`
	Tags         []Tag                         `json:"tags" yaml:"tags"`
	ExternalDocs *ExternalDocumentation        `json:"externalDocs" yaml:"externalDocs"`
	Bindings     utils.OrderedMap[string, any] `json:"bindings" yaml:"bindings"` // TODO: replace any to common bindings object

	Ref string `json:"$ref" yaml:"$ref"`
}

type SecurityRequirement struct {
	utils.OrderedMap[string, []string]
}
