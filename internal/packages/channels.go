package packages

import (
	"github.com/bdragon300/asyncapi-codegen/internal/common"
	"github.com/dave/jennifer/jen"
)

type ChannelsPackage struct {
	items []common.PackageItem[common.Assembler]
}

func (c *ChannelsPackage) Put(ctx *common.CompileContext, item common.Assembler) {
	c.items = append(c.items, common.PackageItem[common.Assembler]{
		Typ:  item,
		Path: ctx.PathStack(),
	})
}

func (c *ChannelsPackage) Items() []common.PackageItem[common.Assembler] {
	return c.items
}

func RenderChannels(pkg *ChannelsPackage, baseDir string) (files map[string]*jen.File, err error) {
	channelsGo := jen.NewFilePathName(baseDir, "channels") // FIXME: basedir is actually package path
	if err != nil {
		return
	}

	ctx := &common.AssembleContext{
		CurrentPackage: common.ChannelsPackageKind,
		ImportBase:     "github.com/bdragon300/asyncapi-codegen/generated", // FIXME
	}
	for _, item := range pkg.items {
		if !item.Typ.AllowRender() {
			continue
		}
		for _, stmt := range item.Typ.AssembleDefinition(ctx) {
			channelsGo.Add(stmt)
		}
		channelsGo.Add(jen.Line())
	}

	files = map[string]*jen.File{
		"channels/channels.go": channelsGo,
	}

	return
}
