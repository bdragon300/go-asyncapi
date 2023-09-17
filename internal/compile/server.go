package compile

import (
	"encoding/json"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/bdragon300/asyncapi-codegen/internal/assemble"
	"github.com/bdragon300/asyncapi-codegen/internal/common"
	"github.com/bdragon300/asyncapi-codegen/internal/utils"
)

type serverProtoBuilderFunc func(ctx *common.CompileContext, server *Server, name string) (common.Assembler, error)

var ProtoServerBuilders = map[string]serverProtoBuilderFunc{}

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
	ctx.SetObjName(ctx.Stack.Top().Path) // TODO: use title
	obj, err := s.build(ctx, ctx.Stack.Top().Path)
	if err != nil {
		return fmt.Errorf("error on %q: %w", strings.Join(ctx.PathStack(), "."), err)
	}
	ctx.CurrentPackage().Put(ctx, obj)
	return nil
}

func (s Server) build(ctx *common.CompileContext, serverKey string) (common.Assembler, error) {
	if s.Ref != "" {
		res := assemble.NewRefLinkAsAssembler(s.Ref)
		ctx.Linker.Add(res)
		return res, nil
	}

	protoBuilder, ok := ProtoServerBuilders[s.Protocol]
	if !ok {
		panic(fmt.Sprintf("Unknown protocol %q at path %s", s.Protocol, ctx.PathStack()))
	}
	protoServer, err := protoBuilder(ctx, &s, serverKey)
	if err != nil {
		return nil, fmt.Errorf("error build server at path %s: %w", ctx.PathStack(), err)
	}

	return &assemble.Server{Protocol: s.Protocol, ProtoServer: protoServer}, nil
}

// TODO: This object MAY be extended with Specification Extensions.
type ServerVariable struct {
	Enum        []string `json:"enum" yaml:"enum"`
	Default     string   `json:"default" yaml:"default"`
	Description string   `json:"description" yaml:"description"`
	Examples    []string `json:"examples" yaml:"examples"`

	Ref string `json:"$ref" yaml:"$ref"`
}
