package compiler

import (
	"github.com/bdragon300/asyncapi-codegen-go/internal/common"
	"github.com/bdragon300/asyncapi-codegen-go/internal/render"
)

const utilsPackageName = "utils"

func UtilsCompile(ctx *common.CompileContext) error {
	ctx.Logger.Trace("Utils package")
	ctx.ResultsStore.Add(utilsPackageName, ctx.PathStack(), buildSerializer(ctx))
	return nil
}

func buildSerializer(ctx *common.CompileContext) *render.UtilsSerializer {
	lnk := render.NewListCbLink[*render.Message](func(item common.Renderer, path []string) bool {
		_, ok := item.(*render.Message)
		return ok
	})
	ctx.Linker.AddMany(lnk)

	return &render.UtilsSerializer{
		AllMessages:        lnk,
		DefaultContentType: ctx.ResultsStore.DefaultContentType(),
	}
}
