package packages

import (
	"github.com/bdragon300/asyncapi-codegen/internal/common"
	"github.com/bdragon300/asyncapi-codegen/internal/lang"
	"github.com/bdragon300/asyncapi-codegen/internal/render"
	"github.com/bdragon300/asyncapi-codegen/internal/scan"
	"github.com/dave/jennifer/jen"
)

type ServersPackage struct {
	Types   []PackageItem[lang.LangType]
	Servers []PackageItem[*lang.Server]
}

func (c *ServersPackage) Put(ctx *scan.Context, item render.LangRenderer) {
	switch v := item.(type) {
	case *lang.Server:
		c.Servers = append(c.Servers, PackageItem[*lang.Server]{
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

func (c *ServersPackage) Find(path []string) (render.LangRenderer, bool) {
	if res, ok := findItem(c.Types, path); ok {
		return res, true
	}
	if res, ok := findItem(c.Servers, path); ok {
		return res, true
	}
	return nil, false
}

func (c *ServersPackage) List(path []string) []render.LangRenderer {
	res := listByPath(c.Types, path)
	res = append(res, listByPath(c.Servers, path)...)
	return res
}

func RenderServers(pkg *ServersPackage, baseDir string) (files map[string]*jen.File, err error) {
	serversGo := jen.NewFilePathName(baseDir, "servers") // FIXME: basedir is actually package path
	if err != nil {
		return
	}

	ctx := &render.Context{
		CurrentPackage: common.ServersPackageKind,
		ImportBase:     "github.com/bdragon300/asyncapi-codegen/generated", // FIXME
		RuntimePackage: "github.com/bdragon300/asyncapi-codegen/runtime",   // FIXME
	}
	for _, item := range pkg.Servers {
		for _, stmt := range item.Typ.RenderDefinition(ctx) {
			serversGo.Add(stmt)
		}
		serversGo.Add(jen.Line())
	}

	files = map[string]*jen.File{
		"servers/servers.go": serversGo,
	}

	return
}
