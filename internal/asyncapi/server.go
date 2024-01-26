package asyncapi

import (
	"path"

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
	ctx.SetTopObjName(ctx.Stack.Top().PathItem)
	obj, err := s.build(ctx, ctx.Stack.Top().PathItem)
	if err != nil {
		return err
	}
	if obj == nil {
		return nil
	}
	ctx.PutObject(obj)
	// FIXME: its needed to consider servers only from `servers` section
	if ctx.CurrentPackage() == PackageScopeServers {
		ctx.Storage.AddProtocol(s.Protocol)
	}
	return nil
}

func (s Server) build(ctx *common.CompileContext, serverKey string) (common.Renderer, error) {
	ignore := s.XIgnore || !ctx.CompileOpts.ServerOpts.IsAllowedName(serverKey)
	if ignore {
		ctx.Logger.Debug("Server denoted to be ignored")
		return &render.Server{Dummy: true}, nil
	}
	if s.Ref != "" {
		ctx.Logger.Trace("Ref", "$ref", s.Ref)
		res := render.NewRendererPromise(s.Ref, common.PromiseOriginUser)
		ctx.PutPromise(res)
		return res, nil
	}

	protoBuilder, ok := ProtocolBuilders[s.Protocol]
	if !ok {
		ctx.Logger.Warn("Skip unsupported server protocol", "proto", s.Protocol)
		return nil, nil
	}
	srvName, _ := lo.Coalesce(s.XGoName, serverKey)
	res := render.Server{Name: srvName, Protocol: s.Protocol}

	// Bindings
	if s.Bindings != nil {
		ctx.Logger.Trace("Server bindings")
		res.BindingsStruct = &render.GoStruct{
			BaseType: render.BaseType{
				Name:         ctx.GenerateObjName(srvName, "Bindings"),
				DirectRender: true,
				Import:       ctx.CurrentPackage(),
			},
			Fields: nil,
		}

		ref := ctx.PathRef() + "/bindings"
		res.BindingsPromise = render.NewPromise[*render.Bindings](ref, common.PromiseOriginInternal)
		ctx.PutPromise(res.BindingsPromise)
	}

	// Server variables
	for _, v := range s.Variables.Entries() {
		ctx.Logger.Trace("Server variable", "name", v.Key)
		ref := path.Join(ctx.PathRef(), "variables", v.Key)
		prm := render.NewPromise[*render.ServerVariable](ref, common.PromiseOriginInternal)
		ctx.PutPromise(prm)
		res.Variables.Set(v.Key, prm)
	}

	var err error
	ctx.Logger.Trace("Server", "proto", protoBuilder.ProtocolName())
	res.ProtoServer, err = protoBuilder.BuildServer(ctx, &s, serverKey, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}
