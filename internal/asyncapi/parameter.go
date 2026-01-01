package asyncapi

import (
	"github.com/bdragon300/go-asyncapi/internal/compiler/compile"
	"github.com/samber/lo"

	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render"
)

type Parameter struct {
	Enum        []string `json:"enum,omitzero" yaml:"enum"`
	Default     string   `json:"default,omitzero" yaml:"default"`
	Description string   `json:"description,omitzero" yaml:"description"`
	Examples    []string `json:"examples,omitzero" yaml:"examples"`
	Location    string   `json:"location,omitzero" yaml:"location"` // TODO: implement

	XGoName string `json:"x-go-name,omitzero" yaml:"x-go-name"`

	Ref string `json:"$ref,omitzero" yaml:"$ref"`
}

func (p Parameter) Compile(ctx *compile.Context) error {
	obj, err := p.build(ctx, ctx.Stack.Top().Key)
	if err != nil {
		return err
	}
	ctx.PutArtifact(obj)
	return nil
}

func (p Parameter) build(ctx *compile.Context, parameterKey string) (common.Artifact, error) {
	if p.Ref != "" {
		return registerRef(ctx, p.Ref, parameterKey, nil), nil
	}

	parName, _ := lo.Coalesce(p.XGoName, parameterKey)
	res := &render.Parameter{OriginalName: ctx.GenerateObjName(parName, "")}

	return res, nil
}
