package compile

import (
	"github.com/bdragon300/asyncapi-codegen-go/internal/assemble"
	"github.com/bdragon300/asyncapi-codegen-go/internal/common"
)

const utilsPackageName = "utils"

func UtilsCompile(ctx *common.CompileContext) error {
	ctx.LogInfo("Utils package")
	pkg := common.Package{}
	if _, ok := ctx.Packages[utilsPackageName]; !ok {
		ctx.Packages[utilsPackageName] = &pkg
	}

	ctx.Packages[utilsPackageName].Put(buildSerializer(ctx), nil)

	return nil
}

func buildSerializer(ctx *common.CompileContext) *assemble.UtilsSerializer {
	lnk := assemble.NewListCbLink[*assemble.Message](func(item common.Assembler, path []string) bool {
		_, ok := item.(*assemble.Message)
		return ok
	})
	ctx.Linker.AddMany(lnk)

	return &assemble.UtilsSerializer{
		AllMessages:        lnk,
		DefaultContentType: ctx.DefaultContentType,
	}
}
