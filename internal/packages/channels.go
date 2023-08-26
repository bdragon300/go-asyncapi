package packages

import (
	"github.com/bdragon300/asyncapi-codegen/internal/common"
	"github.com/bdragon300/asyncapi-codegen/internal/lang"
	"github.com/bdragon300/asyncapi-codegen/internal/render"
	"github.com/bdragon300/asyncapi-codegen/internal/scan"
	"github.com/dave/jennifer/jen"
)

type ChannelsPackage struct {
	Types    []PackageItem[lang.LangType]
	Channels []PackageItem[*lang.Channel]
}

func (c *ChannelsPackage) Put(ctx *scan.Context, item render.LangRenderer) {
	switch v := item.(type) {
	case *lang.Channel:
		c.Channels = append(c.Channels, PackageItem[*lang.Channel]{
			Typ:  v,
			Path: ctx.PathStack(),
		})
	default:
		c.Types = append(c.Types, PackageItem[lang.LangType]{
			Typ:  v.(lang.LangType),
			Path: ctx.PathStack(),
		})
	}
}

func (c *ChannelsPackage) Find(path []string) (render.LangRenderer, bool) {
	if res, ok := findItem(c.Types, path); ok {
		return res, true
	}
	if res, ok := findItem(c.Channels, path); ok {
		return res, true
	}
	return nil, false
}

func (c *ChannelsPackage) List(path []string) []render.LangRenderer {
	res := listByPath(c.Types, path)
	res = append(res, listByPath(c.Channels, path)...)
	return res
}

func RenderChannels(pkg *ChannelsPackage, baseDir string) (files map[string]*jen.File, err error) {
	channelsGo := jen.NewFilePathName(baseDir, "channels") // FIXME: basedir is actually package path
	if err != nil {
		return
	}

	ctx := &render.Context{
		CurrentPackage: common.MessagesPackageKind,
		ImportBase:     "github.com/bdragon300/asyncapi-codegen/generated", // FIXME
		RuntimePackage: "github.com/bdragon300/asyncapi-codegen/runtime",   // FIXME
	}
	for _, item := range pkg.Channels {
		for _, stmt := range item.Typ.RenderDefinition(ctx) {
			channelsGo.Add(stmt)
		}
		channelsGo.Add(jen.Line())
	}

	files = map[string]*jen.File{
		"channels/channels.go": channelsGo,
	}

	return
}
