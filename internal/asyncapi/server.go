package asyncapi

import (
	"encoding/json"

	"github.com/samber/lo"

	"github.com/bdragon300/asyncapi-codegen-go/internal/types"

	"gopkg.in/yaml.v3"

	"github.com/bdragon300/asyncapi-codegen-go/internal/common"
	"github.com/bdragon300/asyncapi-codegen-go/internal/render"
)

type protoServerCompilerFunc func(ctx *common.CompileContext, server *Server, name string) (common.Renderer, error)

var ProtoServerCompiler = map[string]protoServerCompilerFunc{}

type Server struct {
	URL             string                                                             `json:"url" yaml:"url"`
	Protocol        string                                                             `json:"protocol" yaml:"protocol"`
	ProtocolVersion string                                                             `json:"protocolVersion" yaml:"protocolVersion"`
	Description     string                                                             `json:"description" yaml:"description"`
	Variables       types.OrderedMap[string, ServerVariable]                           `json:"variables" yaml:"variables"`
	Security        []SecurityRequirement                                              `json:"security" yaml:"security"`
	Tags            []Tag                                                              `json:"tags" yaml:"tags"`
	Bindings        types.OrderedMap[string, types.Union2[json.RawMessage, yaml.Node]] `json:"bindings" yaml:"bindings"`

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
	if ctx.TopPackageName() == "servers" {
		ctx.Storage.AddProtocol(s.Protocol)
	}
	return nil
}

func (s Server) build(ctx *common.CompileContext, serverKey string) (common.Renderer, error) {
	if s.XIgnore {
		ctx.Logger.Debug("Server denoted to be ignored")
		return &render.Simple{Name: "any", IsIface: true}, nil
	}
	if s.Ref != "" {
		ctx.Logger.Trace("Ref", "$ref", s.Ref)
		res := render.NewRendererPromise(s.Ref, common.PromiseOriginUser)
		ctx.PutPromise(res)
		return res, nil
	}

	protoBuilder, ok := ProtoServerCompiler[s.Protocol]
	if !ok {
		ctx.Logger.Warn("Skip unsupported server protocol", "proto", s.Protocol)
		return nil, nil
	}
	protoServer, err := protoBuilder(ctx, &s, serverKey)
	if err != nil {
		return nil, err
	}

	srvName, _ := lo.Coalesce(s.XGoName, serverKey)
	return &render.Server{
		Name:        srvName,
		Protocol:    s.Protocol,
		ProtoServer: protoServer,
		BindingsStruct: &render.Struct{
			BaseType: render.BaseType{
				Name:         ctx.GenerateObjName(srvName, "Bindings"),
				DirectRender: true,
				PackageName:  ctx.TopPackageName(),
			},
			Fields: nil,
		},
	}, nil
}

// TODO: This object MAY be extended with Specification Extensions.
type ServerVariable struct {
	Enum        []string `json:"enum" yaml:"enum"`
	Default     string   `json:"default" yaml:"default"`
	Description string   `json:"description" yaml:"description"`
	Examples    []string `json:"examples" yaml:"examples"`

	Ref string `json:"$ref" yaml:"$ref"`
}
