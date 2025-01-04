package asyncapi

import (
	"github.com/bdragon300/go-asyncapi/internal/render/lang"
	"github.com/samber/lo"

	"github.com/bdragon300/go-asyncapi/internal/types"

	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render"
)

type Server struct {
	URL             string                                   `json:"url" yaml:"url"`  // DEPRECATED
	Host 		  string                                   `json:"host" yaml:"host"`
	Protocol        string                                   `json:"protocol" yaml:"protocol"`
	ProtocolVersion string                                   `json:"protocolVersion" yaml:"protocolVersion"`
	Pathname		string                                   `json:"pathname" yaml:"pathname"`
	Description     string                                   `json:"description" yaml:"description"`
	Variables       types.OrderedMap[string, ServerVariable] `json:"variables" yaml:"variables"`
	Security        []SecurityRequirement                    `json:"security" yaml:"security"`
	Tags            []Tag                                    `json:"tags" yaml:"tags"`
	ExternalDocs   *ExternalDocumentation                            `json:"externalDocs" yaml:"externalDocs"`
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
	if s.XIgnore {
		ctx.Logger.Debug("Server denoted to be ignored")
		return &render.Server{Dummy: true}, nil
	}
	if s.Ref != "" {
		// Make a promise selectable if it defined in `servers` section
		return registerRef(ctx, s.Ref, serverKey, lo.Ternary(isComponent, nil, lo.ToPtr(true))), nil
	}

	srvName, _ := lo.Coalesce(s.XGoName, serverKey)
	res := render.Server{
		OriginalName:    srvName,
		URL:             s.URL,
		Protocol:        s.Protocol,
		ProtocolVersion: s.ProtocolVersion,
		IsComponent:     isComponent,
	}

	// Channels which are bound to this server
	prm := lang.NewListCbPromise[common.Renderable](func(item common.CompileObject, path []string) bool {
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
		res.BindingsPromise = lang.NewPromise[*render.Bindings](ref)
		ctx.PutPromise(res.BindingsPromise)
	}

	// Server variables
	for _, v := range s.Variables.Entries() {
		ctx.Logger.Trace("Server variable", "name", v.Key)
		ref := ctx.PathStackRef("variables", v.Key)
		prm := lang.NewPromise[*render.ServerVariable](ref)
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
		res.ProtoServer = &render.ProtoServer{Server: &res, Type: protoStruct}
		return &res, nil
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

	return &res, nil
}
