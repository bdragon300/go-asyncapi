package packages

import (
	"github.com/bdragon300/asyncapi-codegen/internal/assemble"
	"github.com/bdragon300/asyncapi-codegen/internal/common"
	"github.com/dave/jennifer/jen"
)

type ServersPackage struct {
	Types   []PackageItem[common.GolangType]
	Servers []PackageItem[*assemble.Server]
}

func (c *ServersPackage) Put(ctx *common.CompileContext, item common.Assembler) {
	switch v := item.(type) {
	case *assemble.Server:
		c.Servers = append(c.Servers, PackageItem[*assemble.Server]{
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

func (c *ServersPackage) FindBy(cb func(item any, path []string) bool) (common.Assembler, bool) {
	if res, ok := findItemBy(c.Types, cb); ok {
		return res, true
	}
	if res, ok := findItemBy(c.Servers, cb); ok {
		return res, true
	}
	return nil, false
}

func (c *ServersPackage) ListBy(cb func(item any, path []string) bool) []common.Assembler {
	res := listSubItemsBy(c.Types, cb)
	res = append(res, listSubItemsBy(c.Servers, cb)...)
	return res
}

func RenderServers(pkg *ServersPackage, baseDir string) (files map[string]*jen.File, err error) {
	serversGo := jen.NewFilePathName(baseDir, "servers") // FIXME: basedir is actually package path
	if err != nil {
		return
	}

	ctx := &common.AssembleContext{
		CurrentPackage: common.ServersPackageKind,
		ImportBase:     "github.com/bdragon300/asyncapi-codegen/generated", // FIXME
	}
	for _, item := range pkg.Servers {
		for _, stmt := range item.Typ.AssembleDefinition(ctx) {
			serversGo.Add(stmt)
		}
		serversGo.Add(jen.Line())
	}

	files = map[string]*jen.File{
		"servers/servers.go": serversGo,
	}

	return
}
