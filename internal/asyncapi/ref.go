package asyncapi

import (
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/render/lang"
)

type StandaloneRef struct {
	Ref string `json:"$ref" yaml:"$ref"`
}

func (c StandaloneRef) String() string {
	return "StandaloneRef -> " + c.Ref
}


func registerRef(ctx *common.CompileContext, ref string, name string, selectable *bool) common.Renderable {
	ctx.Logger.Trace("Ref", "$ref", ref)
	prm := lang.NewRef(ref, name, selectable)
	ctx.PutPromise(prm)

	return prm
}