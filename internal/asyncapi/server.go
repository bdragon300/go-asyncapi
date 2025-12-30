package asyncapi

import (
	"strconv"

	"github.com/bdragon300/go-asyncapi/internal/compiler/compile"
	"github.com/bdragon300/go-asyncapi/internal/render/lang"
	"github.com/samber/lo"

	"github.com/bdragon300/go-asyncapi/internal/types"

	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render"
)

type Server struct {
	Host            string                                   `json:"host,omitzero" yaml:"host"`
	Protocol        string                                   `json:"protocol,omitzero" yaml:"protocol"`
	ProtocolVersion string                                   `json:"protocolVersion,omitzero" yaml:"protocolVersion"`
	Pathname        string                                   `json:"pathname,omitzero" yaml:"pathname"`
	Description     string                                   `json:"description,omitzero" yaml:"description"`
	Variables       types.OrderedMap[string, ServerVariable] `json:"variables,omitzero" yaml:"variables"`
	Security        []SecurityScheme                         `json:"security,omitzero" yaml:"security"`
	Tags            []Tag                                    `json:"tags,omitzero" yaml:"tags"`
	ExternalDocs    *ExternalDocumentation                   `json:"externalDocs,omitzero" yaml:"externalDocs"`
	Bindings        *Bindings                                `json:"bindings,omitzero" yaml:"bindings"`

	XGoName string `json:"x-go-name,omitzero" yaml:"x-go-name"`
	XIgnore bool   `json:"x-ignore,omitzero" yaml:"x-ignore"`

	Ref string `json:"$ref,omitzero" yaml:"$ref"`
}

func (s Server) Compile(ctx *compile.Context) error {
	obj, err := s.build(ctx, ctx.Stack.Top().Key)
	if err != nil {
		return err
	}
	ctx.PutArtifact(obj)
	return nil
}

func (s Server) build(ctx *compile.Context, serverKey string) (common.Artifact, error) {
	_, isSelectable := ctx.Stack.Top().Flags[common.SchemaTagSelectable]
	if s.XIgnore {
		ctx.Logger.Debug("Server denoted to be ignored")
		return &render.Server{Dummy: true}, nil
	}
	if s.Ref != "" {
		// Make a server selectable if it defined in `servers` section
		return registerRef(ctx, s.Ref, serverKey, lo.Ternary(isSelectable, lo.ToPtr(true), nil)), nil
	}

	srvName, _ := lo.Coalesce(s.XGoName, serverKey)
	res := render.Server{
		OriginalName:    srvName,
		Host:            s.Host,
		Pathname:        s.Pathname,
		Protocol:        s.Protocol,
		ProtocolVersion: s.ProtocolVersion,
		IsSelectable:    isSelectable,
		IsPublisher:     ctx.CompileOpts.GeneratePublishers,
		IsSubscriber:    ctx.CompileOpts.GenerateSubscribers,
	}

	// All active channels
	prm := lang.NewListCbPromise[common.Artifact](func(item common.Artifact) bool {
		path := item.Pointer().Pointer
		if len(path) < 2 || len(path) >= 2 && path[0] != "channels" {
			return false
		}
		return item.Kind() == common.ArtifactKindChannel && item.Visible()
	}, nil)
	ctx.PutListPromise(prm)
	res.AllActiveChannelsPromise = prm

	// Bindings
	if s.Bindings != nil {
		ctx.Logger.Trace("Server bindings")
		ref := ctx.CurrentRefPointer("bindings")
		res.BindingsPromise = lang.NewPromise[*render.Bindings](ref, nil)
		ctx.PutPromise(res.BindingsPromise)
	}

	// Security
	if len(s.Security) > 0 {
		ctx.Logger.Trace("Server security schemes", "count", len(s.Security))
		for ind := range s.Security {
			ref := ctx.CurrentRefPointer("security", strconv.Itoa(ind))
			secPrm := lang.NewPromise[*render.SecurityScheme](ref, nil)
			ctx.PutPromise(secPrm)
			res.SecuritySchemePromises = append(res.SecuritySchemePromises, secPrm)
		}
	}

	// Server variables
	for _, v := range s.Variables.Entries() {
		ctx.Logger.Trace("Server variable", "name", v.Key)
		ref := ctx.CurrentRefPointer("variables", v.Key)
		prm := lang.NewPromise[*render.ServerVariable](ref, nil)
		ctx.PutPromise(prm)
		res.VariablesPromises.Set(v.Key, prm)
	}

	ctx.Logger.Trace("Server", "proto", s.Protocol)
	return &res, nil
}
