package asyncapi

import (
	"github.com/bdragon300/go-asyncapi/internal/render/lang"
	"github.com/samber/lo"

	"github.com/bdragon300/go-asyncapi/internal/types"

	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render"
)

type Server struct {
	Host            string                                   `json:"host" yaml:"host"`
	Protocol        string                                   `json:"protocol" yaml:"protocol"`
	ProtocolVersion string                                   `json:"protocolVersion" yaml:"protocolVersion"`
	Pathname        string                                   `json:"pathname" yaml:"pathname"`
	Description     string                                   `json:"description" yaml:"description"`
	Variables       types.OrderedMap[string, ServerVariable] `json:"variables" yaml:"variables"`
	Security        []SecurityRequirement                    `json:"security" yaml:"security"`
	Tags            []Tag                                    `json:"tags" yaml:"tags"`
	ExternalDocs    *ExternalDocumentation                   `json:"externalDocs" yaml:"externalDocs"`
	Bindings        *ServerBindings                          `json:"bindings" yaml:"bindings"`

	XGoName string `json:"x-go-name" yaml:"x-go-name"`
	XIgnore bool   `json:"x-ignore" yaml:"x-ignore"`

	Ref string `json:"$ref" yaml:"$ref"`
}

func (s Server) Compile(ctx *common.CompileContext) error {
	ctx.RegisterNameTop(ctx.Stack.Top().PathItem)
	obj, err := s.build(ctx, ctx.Stack.Top().PathItem)
	if err != nil {
		return err
	}
	ctx.PutObject(obj)
	return nil
}

func (s Server) build(ctx *common.CompileContext, serverKey string) (common.Renderable, error) {
	_, isSelectable := ctx.Stack.Top().Flags[common.SchemaTagSelectable]
	if s.XIgnore {
		ctx.Logger.Debug("Server denoted to be ignored")
		return &render.Server{Dummy: true}, nil
	}
	if s.Ref != "" {
		// Make a promise selectable if it defined in `servers` section
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

	// Channels which are bound to this server
	prm := lang.NewListCbPromise[common.Renderable](func(item common.CompileArtifact, path []string) bool {
		if len(path) < 2 || len(path) >= 2 && path[0] != "channels" {
			return false
		}
		return item.Kind() == common.ObjectKindChannel && item.Visible()
	}, nil)
	ctx.PutListPromise(prm)
	res.AllActiveChannelsPromise = prm

	// Bindings
	if s.Bindings != nil {
		ctx.Logger.Trace("Server bindings")
		res.BindingsType = &lang.GoStruct{
			BaseType: lang.BaseType{
				OriginalName:  ctx.GenerateObjName(srvName, "Bindings"),
				HasDefinition: true,
			},
		}

		ref := ctx.PathStackRef("bindings")
		res.BindingsPromise = lang.NewPromise[*render.Bindings](ref, nil)
		ctx.PutPromise(res.BindingsPromise)
	}

	// Server variables
	for _, v := range s.Variables.Entries() {
		ctx.Logger.Trace("Server variable", "name", v.Key)
		ref := ctx.PathStackRef("variables", v.Key)
		prm := lang.NewPromise[*render.ServerVariable](ref, nil)
		ctx.PutPromise(prm)
		res.VariablesPromises.Set(v.Key, prm)
	}

	if _, ok := ProtocolBuilders[s.Protocol]; !ok {
		ctx.Logger.Warn("Skip unsupported server protocol", "proto", s.Protocol)
		res.ProtoServer = BuildProtoServer(ctx, &s, &res, "")
		return &res, nil
	}

	ctx.Logger.Trace("Server", "proto", s.Protocol)
	res.ProtoServer = BuildProtoServer(ctx, &s, &res, s.Protocol)

	return &res, nil
}
