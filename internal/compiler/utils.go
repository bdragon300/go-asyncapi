package compiler

import (
	"github.com/bdragon300/asyncapi-codegen-go/internal/common"
	"github.com/bdragon300/asyncapi-codegen-go/internal/render"
)

const utilsPackageName = "utils"

func UtilsCompile(ctx *common.CompileContext) error {
	ctx.Logger.Trace("Utils package")
	ctx.Storage.AddObject(utilsPackageName, ctx.PathStack(), buildSerializer(ctx))
	return nil
}

func buildSerializer(ctx *common.CompileContext) *render.UtilsSerializer {
	allMessagesPrm := render.NewListCbPromise[*render.Message](func(item common.Renderer, path []string) bool {
		_, ok := item.(*render.Message)
		return ok
	})
	ctx.PutListPromise(allMessagesPrm)

	return &render.UtilsSerializer{
		AllMessages:        allMessagesPrm,
		DefaultContentType: ctx.Storage.DefaultContentType(),
	}
}
