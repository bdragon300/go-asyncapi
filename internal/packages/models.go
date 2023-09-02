package packages

import (
	"github.com/bdragon300/asyncapi-codegen/internal/common"
	"github.com/dave/jennifer/jen"
	"github.com/samber/lo"
)

type PackageItem[T common.Assembler] struct {
	Typ  T
	Path []string
}

type ModelsPackage struct {
	Items []PackageItem[common.GolangType]
}

func (m *ModelsPackage) Put(ctx *common.CompileContext, item common.Assembler) {
	m.Items = append(m.Items, PackageItem[common.GolangType]{
		Typ:  item.(common.GolangType),
		Path: ctx.PathStack(),
	})
}

func (m *ModelsPackage) FindBy(cb func(item any, path []string) bool) (common.Assembler, bool) {
	return findItemBy(m.Items, cb)
}

func (m *ModelsPackage) ListBy(cb func(item any, path []string) bool) []common.Assembler {
	return listSubItemsBy(m.Items, cb)
}

func findItemBy[T common.Assembler](items []PackageItem[T], cb func(item any, path []string) bool) (common.Assembler, bool) {
	res, ok := lo.Find(items, func(item PackageItem[T]) bool {
		return cb(item.Typ, item.Path)
	})
	return res.Typ, ok
}

func listSubItemsBy[T common.Assembler](items []PackageItem[T], cb func(item any, path []string) bool) []common.Assembler {
	return lo.FilterMap(items, func(item PackageItem[T], index int) (common.Assembler, bool) {
		if cb(item.Typ, item.Path) {
			return item.Typ, true
		}
		return nil, false
	})
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

	for _, item := range pkg.Items {
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
