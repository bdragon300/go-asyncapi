package asyncapi

import (
	"github.com/bdragon300/go-asyncapi/internal/render/lang"
	"github.com/samber/lo"

	"github.com/bdragon300/go-asyncapi/internal/types"

	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render"
)

type Server struct {
	URL             string                                   `json:"url" yaml:"url"`
	Protocol        string                                   `json:"protocol" yaml:"protocol"`
	ProtocolVersion string                                   `json:"protocolVersion" yaml:"protocolVersion"`
	Description     string                                   `json:"description" yaml:"description"`
	Variables       types.OrderedMap[string, ServerVariable] `json:"variables" yaml:"variables"`
	Security        []SecurityRequirement                    `json:"security" yaml:"security"`
	Tags            []Tag                                    `json:"tags" yaml:"tags"`
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
	_, isComponent := ctx.Stack.Top().Flags[common.SchemaTagComponent]
	ignore := s.XIgnore //|| !ctx.CompileOpts.ServerOpts.IsAllowedName(serverKey)
	if ignore {
		ctx.Logger.Debug("Server denoted to be ignored")
		return &render.Server{Dummy: true}, nil
	}
	if s.Ref != "" {
		ctx.Logger.Trace("Ref", "$ref", s.Ref)
		prm := lang.NewRenderablePromise(s.Ref, common.PromiseOriginUser)
		// Set a server to be rendered if we reference it from `servers` document section
		ctx.PutPromise(prm)
		return prm, nil
	}

	srvName, _ := lo.Coalesce(s.XGoName, serverKey)
	// Render only the servers defined directly in `servers` document section, not in `components`
	res := render.Server{
		Name:            srvName,
		SpecKey:         serverKey,
		TypeNamePrefix:  ctx.GenerateObjName(srvName, ""),
		URL:             s.URL,
		Protocol:        s.Protocol,
		ProtocolVersion: s.ProtocolVersion,
	}

	// Channels which are connected to this server
	prm := lang.NewListCbPromise[*render.Channel](func(item common.Renderable, path []string) bool {
		_, ok := item.(*render.Channel)
		if !ok {
			return false
		}
		return len(path) >= 2 && path[0] == "channels"
	})
	ctx.PutListPromise(prm)
	res.AllChannelsPromise = prm

	// Bindings
	if s.Bindings != nil {
		ctx.Logger.Trace("Server bindings")
		res.BindingsType = &lang.GoStruct{
			BaseType: lang.BaseType{
				Name:          ctx.GenerateObjName(srvName, "Bindings"),
				HasDefinition: true,
			},
		}

		ref := ctx.PathStackRef("bindings")
		res.BindingsPromise = lang.NewPromise[*render.Bindings](ref, common.PromiseOriginInternal)
		ctx.PutPromise(res.BindingsPromise)
	}

	// Server variables
	for _, v := range s.Variables.Entries() {
		ctx.Logger.Trace("Server variable", "name", v.Key)
		ref := ctx.PathStackRef("variables", v.Key)
		prm := lang.NewPromise[*render.ServerVariable](ref, common.PromiseOriginInternal)
		ctx.PutPromise(prm)
		res.VariablesPromises.Set(v.Key, prm)
	}

	protoBuilder, ok := ProtocolBuilders[s.Protocol]
	if !ok {
		ctx.Logger.Warn("Skip unsupported server protocol", "proto", s.Protocol)
		protoStruct, err := BuildProtoServerStruct(ctx, &s, &res, "")
		if err != nil {
			return nil, err
		}
		return &render.ProtoServer{Server: &res, Type: protoStruct}, nil
	}

	ctx.Logger.Trace("Server", "proto", protoBuilder.ProtocolName())

	protoServer, err := protoBuilder.BuildServer(ctx, &s, &res)
	if err != nil {
		return nil, err
	}
	res.ProtoServer = protoServer

	// Register protocol only for servers in `servers` document section, not in `components`
	if !isComponent {
		ctx.Storage.RegisterProtocol(s.Protocol)
	}

	return res, nil
}
