package packages

import (
	"github.com/bdragon300/asyncapi-codegen/internal/common"
	"github.com/dave/jennifer/jen"
)

func RenderParameters(pkg *Package, baseDir string) (files map[string]*jen.File, err error) {
	channelsGo := jen.NewFilePathName(baseDir, "parameters") // FIXME: basedir is actually package path
	if err != nil {
		return
	}

	ctx := &common.AssembleContext{
		CurrentPackage: common.ParametersPackageKind,
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
		"parameters/parameters.go": channelsGo,
	}

	return
}
