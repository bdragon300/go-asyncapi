package packages

import (
	"github.com/bdragon300/asyncapi-codegen/internal/common"
	"github.com/dave/jennifer/jen"
)

type ModelsPackage struct {
	items []common.PackageItem[common.Assembler]
}

func (m *ModelsPackage) Put(ctx *common.CompileContext, item common.Assembler) {
	m.items = append(m.items, common.PackageItem[common.Assembler]{
		Typ:  item,
		Path: ctx.PathStack(),
	})
}

func (m *ModelsPackage) Items() []common.PackageItem[common.Assembler] {
	return m.items
}

func RenderModels(pkg *ModelsPackage, baseDir string) (files map[string]*jen.File, err error) {
	modelsGo := jen.NewFilePathName(baseDir, "models") // FIXME: basedir is actually package path
	if err != nil {
		return
	}

	ctx := &common.AssembleContext{
		CurrentPackage: common.ModelsPackageKind,
		ImportBase:     "github.com/bdragon300/asyncapi-codegen/generated", // FIXME
	}

	for _, item := range pkg.items {
		if !item.Typ.AllowRender() {
			continue
		}

		for _, stmt := range item.Typ.AssembleDefinition(ctx) {
			modelsGo.Add(stmt)
		}
		modelsGo.Add(jen.Line())
	}

	files = map[string]*jen.File{
		"models/models.go": modelsGo,
	}

	return
}
