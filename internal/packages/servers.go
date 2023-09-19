package packages

import (
	"github.com/bdragon300/asyncapi-codegen/internal/common"
	"github.com/dave/jennifer/jen"
)

func RenderServers(pkg *Package, baseDir string) (files map[string]*jen.File, err error) {
	serversGo := jen.NewFilePathName(baseDir, "servers") // FIXME: basedir is actually package path
	if err != nil {
		return
	}

	ctx := &common.AssembleContext{
		CurrentPackage: common.ServersPackageKind,
		ImportBase:     "github.com/bdragon300/asyncapi-codegen/generated", // FIXME
	}
	for _, item := range pkg.items {
		if !item.Typ.AllowRender() {
			continue
		}
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
