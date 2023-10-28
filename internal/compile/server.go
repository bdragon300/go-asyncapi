package compile

import (
	"encoding/json"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/bdragon300/asyncapi-codegen-go/internal/assemble"
	"github.com/bdragon300/asyncapi-codegen-go/internal/common"
	"github.com/bdragon300/asyncapi-codegen-go/internal/utils"
)

type protoServerCompilerFunc func(ctx *common.CompileContext, server *Server, name string) (common.Assembler, error)

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
		return fmt.Errorf("error on %q: %w", strings.Join(ctx.PathStack(), "."), err)
	}
	ctx.PutToCurrentPkg(obj)
	if ctx.TopPackageName() == "servers" {
		ctx.NotifyProtocol(s.Protocol)
	}
	return nil
}

func (s Server) build(ctx *common.CompileContext, serverKey string) (common.Assembler, error) {
	if s.Ref != "" {
		res := assemble.NewRefLinkAsAssembler(s.Ref)
		ctx.Linker.Add(res)
		return res, nil
	}

	protoBuilder, ok := ProtoServerCompiler[s.Protocol]
	if !ok {
		panic(fmt.Sprintf("Unknown protocol %q at path %s", s.Protocol, ctx.PathStack()))
	}
	protoServer, err := protoBuilder(ctx, &s, serverKey)
	if err != nil {
		return nil, fmt.Errorf("error build server at path %s: %w", ctx.PathStack(), err)
	}

	return &assemble.Server{
		Protocol:    s.Protocol,
		ProtoServer: protoServer,
		BindingsStruct: &assemble.Struct{
			BaseType: assemble.BaseType{
				Name:    ctx.GenerateObjName("", "Bindings"),
				Render:  true,
				Package: ctx.TopPackageName(),
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
