package compile

import (
	"encoding/json"

	"gopkg.in/yaml.v3"

	"github.com/bdragon300/asyncapi-codegen-go/internal/common"
	"github.com/bdragon300/asyncapi-codegen-go/internal/render"
	"github.com/bdragon300/asyncapi-codegen-go/internal/utils"
)

type protoServerCompilerFunc func(ctx *common.CompileContext, server *Server, name string) (common.Renderer, error)

var ProtoServerCompiler = map[string]protoServerCompilerFunc{}

type Server struct {
	URL             string                                                             `json:"url" yaml:"url"`
	Protocol        string                                                             `json:"protocol" yaml:"protocol"`
	ProtocolVersion string                                                             `json:"protocolVersion" yaml:"protocolVersion"`
	Description     string                                                             `json:"description" yaml:"description"`
	Variables       utils.OrderedMap[string, ServerVariable]                           `json:"variables" yaml:"variables"`
	Security        []SecurityRequirement                                              `json:"security" yaml:"security"`
	Tags            []Tag                                                              `json:"tags" yaml:"tags"`
	Bindings        utils.OrderedMap[string, utils.Union2[json.RawMessage, yaml.Node]] `json:"bindings" yaml:"bindings"`

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
	ctx.PutToCurrentPkg(obj)
	if ctx.TopPackageName() == "servers" {
		ctx.AddProtocol(s.Protocol)
	}
	return nil
}

func (s Server) build(ctx *common.CompileContext, serverKey string) (common.Renderer, error) {
	if s.Ref != "" {
		ctx.LogDebug("Ref", "$ref", s.Ref)
		res := render.NewRefLinkAsRenderer(s.Ref, common.LinkOriginUser)
		ctx.Linker.Add(res)
		return res, nil
	}

	protoBuilder, ok := ProtoServerCompiler[s.Protocol]
	if !ok {
		ctx.LogWarn("Skip unsupported server protocol", "proto", s.Protocol)
		return nil, nil
	}
	protoServer, err := protoBuilder(ctx, &s, serverKey)
	if err != nil {
		return nil, err
	}

	return &render.Server{
		Name:        serverKey,
		Protocol:    s.Protocol,
		ProtoServer: protoServer,
		BindingsStruct: &render.Struct{
			BaseType: render.BaseType{
				Name:        ctx.GenerateObjName(serverKey, "Bindings"),
				Render:      true,
				PackageName: ctx.TopPackageName(),
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
