package packages

import (
	"github.com/bdragon300/asyncapi-codegen/internal/assemble"
	"github.com/bdragon300/asyncapi-codegen/internal/common"
	"github.com/dave/jennifer/jen"
)

type ChannelsPackage struct {
	Types    []PackageItem[common.GolangType]
	Channels []PackageItem[*assemble.Channel]
}

func (c *ChannelsPackage) Put(ctx *common.Context, item common.Assembled) {
	switch v := item.(type) {
	case *assemble.Channel:
		c.Channels = append(c.Channels, PackageItem[*assemble.Channel]{
			Typ:  v,
			Path: ctx.PathStack(),
		})
	default:
		c.Types = append(c.Types, PackageItem[common.GolangType]{
			Typ:  v.(common.GolangType),
			Path: ctx.PathStack(),
		})
	}
}

func (c *ChannelsPackage) Find(path []string) (common.Assembled, bool) {
	if res, ok := findItem(c.Types, path); ok {
		return res, true
	}
	if res, ok := findItem(c.Channels, path); ok {
		return res, true
	}
	return nil, false
}

func (c *ChannelsPackage) List(path []string) []common.Assembled {
	res := listByPath(c.Types, path)
	res = append(res, listByPath(c.Channels, path)...)
	return res
}

func RenderChannels(pkg *ChannelsPackage, baseDir string) (files map[string]*jen.File, err error) {
	channelsGo := jen.NewFilePathName(baseDir, "channels") // FIXME: basedir is actually package path
	if err != nil {
		return
	}

	ctx := &common.AssembleContext{
		CurrentPackage: common.MessagesPackageKind,
		ImportBase:     "github.com/bdragon300/asyncapi-codegen/generated", // FIXME
		RuntimePackage: "github.com/bdragon300/asyncapi-codegen/runtime",   // FIXME
	}
	for _, item := range pkg.Channels {
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
