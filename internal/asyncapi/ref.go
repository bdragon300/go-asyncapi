package asyncapi

import (
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/compiler/compile"
	"github.com/bdragon300/go-asyncapi/internal/render/lang"
)

// StandaloneRef represents an object in AsyncAPI document, that could only be a $ref and nothing else.
// E.g. operation.channel, channel.servers, etc.
type StandaloneRef struct {
	Ref string `json:"$ref,omitzero" yaml:"$ref"`
}

func (c StandaloneRef) String() string {
	return "StandaloneRef -> " + c.Ref
}

// registerRef is helper function that adds a $ref to the compile context and returns it.
// This function is intended to be called from the compilation code.
func registerRef(ctx *compile.Context, ref string, name string, selectable *bool) common.Artifact {
	ctx.Logger.Trace("Ref", "$ref", ref)
	prm := lang.NewRef(ref, name, selectable)
	ctx.PutPromise(prm)

	return prm
}
