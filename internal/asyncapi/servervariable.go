package asyncapi

import (
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/compiler/compile"
	"github.com/bdragon300/go-asyncapi/internal/render"
)

type ServerVariable struct {
	Enum        []string `json:"enum,omitzero" yaml:"enum"`
	Default     string   `json:"default,omitzero" yaml:"default"`
	Description string   `json:"description,omitzero" yaml:"description"`
	Examples    []string `json:"examples,omitzero" yaml:"examples"`

	Ref string `json:"$ref,omitzero" yaml:"$ref"`
}

func (sv ServerVariable) Compile(ctx *compile.Context) error {
	obj, err := sv.build(ctx, ctx.Stack.Top().Key)
	if err != nil {
		return err
	}
	ctx.PutArtifact(obj)
	return nil
}

func (sv ServerVariable) build(ctx *compile.Context, serverVariableKey string) (common.Artifact, error) {
	if sv.Ref != "" {
		return registerRef(ctx, sv.Ref, serverVariableKey, nil), nil
	}

	res := &render.ServerVariable{
		OriginalName: serverVariableKey,
		Enum:         sv.Enum,
		Default:      sv.Default,
		Description:  sv.Description,
	}

	return res, nil
}
