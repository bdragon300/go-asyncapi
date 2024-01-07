package asyncapi

import (
	"github.com/samber/lo"

	"github.com/bdragon300/asyncapi-codegen-go/internal/types"

	"github.com/bdragon300/asyncapi-codegen-go/internal/common"
	"github.com/bdragon300/asyncapi-codegen-go/internal/render"
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
	ctx.SetTopObjName(ctx.Stack.Top().Path) // TODO: use title
	obj, err := s.build(ctx, ctx.Stack.Top().Path)
	if err != nil {
		return err
	}
	if obj == nil {
		return nil
	}
	ctx.PutObject(obj)
	if ctx.CurrentPackage() == "servers" { // FIXME: optimize somehow
		ctx.Storage.AddProtocol(s.Protocol)
	}
	return nil
}

func (s Server) build(ctx *common.CompileContext, serverKey string) (common.Renderer, error) {
	if s.XIgnore {
		ctx.Logger.Debug("Server denoted to be ignored")
		return &render.GoSimple{Name: "any", IsIface: true}, nil
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

	if s.Bindings != nil {
		ctx.Logger.Trace("Server bindings")
		res.BindingsStruct = &render.GoStruct{
			BaseType: render.BaseType{
				Name:         ctx.GenerateObjName(srvName, "Bindings"),
				DirectRender: true,
				PackageName:  ctx.CurrentPackage(),
			},
			Fields: nil,
		}

		ref := ctx.PathRef() + "/bindings"
		res.BindingsPromise = render.NewPromise[*render.Bindings](ref, common.PromiseOriginInternal)
		ctx.PutPromise(res.BindingsPromise)
	}

	var err error
	ctx.Logger.Trace("Server", "proto", protoBuilder.ProtocolName())
	res.ProtoServer, err = protoBuilder.BuildServer(ctx, &s, serverKey)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

// TODO: This object MAY be extended with Specification Extensions.
type ServerVariable struct {
	Enum        []string `json:"enum" yaml:"enum"`
	Default     string   `json:"default" yaml:"default"`
	Description string   `json:"description" yaml:"description"`
	Examples    []string `json:"examples" yaml:"examples"`

	Ref string `json:"$ref" yaml:"$ref"`
}
