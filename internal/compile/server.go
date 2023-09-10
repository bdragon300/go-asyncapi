package compile

import (
	"fmt"
	"strings"

	"github.com/bdragon300/asyncapi-codegen/internal/assemble"
	"github.com/bdragon300/asyncapi-codegen/internal/common"
	"github.com/bdragon300/asyncapi-codegen/internal/utils"
)

type serverProtoBuilderFunc func(ctx *common.CompileContext, server *Server, name string) (assemble.ServerParts, error)

var ProtoServerBuilders = map[string]serverProtoBuilderFunc{}

type Server struct {
	URL             string                                   `json:"url" yaml:"url"`
	Protocol        string                                   `json:"protocol" yaml:"protocol"`
	ProtocolVersion string                                   `json:"protocolVersion" yaml:"protocolVersion"`
	Description     string                                   `json:"description" yaml:"description"`
	Variables       utils.OrderedMap[string, ServerVariable] `json:"variables" yaml:"variables"`
	Security        []SecurityRequirement                    `json:"security" yaml:"security"`
	Tags            []Tag                                    `json:"tags" yaml:"tags"`
	Bindings        utils.OrderedMap[string, any]            `json:"bindings" yaml:"bindings"` // TODO: replace any to common protocols object

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
		res := assemble.NewRefLinkAsAssembler(common.ServersPackageKind, s.Ref)
		ctx.Linker.Add(res)
		return res, nil
	}

	res := &assemble.Server{Protocol: s.Protocol}
	protoBuilder, ok := ProtoServerBuilders[s.Protocol]
	if !ok {
		panic(fmt.Sprintf("Unknown protocol %q at path %s", s.Protocol, ctx.PathStack()))
	}
	var err error
	res.Parts, err = protoBuilder(ctx, &s, serverKey)
	if err != nil {
		return nil, fmt.Errorf("error build server at path %s: %w", ctx.PathStack(), err)
	}

	return res, nil
}

// TODO: This object MAY be extended with Specification Extensions.
type ServerVariable struct {
	Enum        []string `json:"enum" yaml:"enum"`
	Default     string   `json:"default" yaml:"default"`
	Description string   `json:"description" yaml:"description"`
	Examples    []string `json:"examples" yaml:"examples"`

	Ref string `json:"$ref" yaml:"$ref"`
}
