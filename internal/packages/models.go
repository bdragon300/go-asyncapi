package packages

import (
	"github.com/bdragon300/asyncapi-codegen/internal/common"
	"github.com/bdragon300/asyncapi-codegen/internal/utils"
	"github.com/dave/jennifer/jen"
	"github.com/samber/lo"
)

type PackageItem[T common.Assembled] struct {
	Typ  T
	Path []string
}

type ModelsPackage struct {
	Items []PackageItem[common.GolangType]
}

func (s *ModelsPackage) Put(ctx *common.Context, item common.Assembled) {
	s.Items = append(s.Items, PackageItem[common.GolangType]{
		Typ:  item.(common.GolangType),
		Path: ctx.PathStack(),
	})
}

func (s *ModelsPackage) Find(path []string) (common.Assembled, bool) {
	return findItem(s.Items, path)
}

func (s *ModelsPackage) List(path []string) []common.Assembled {
	return listByPath(s.Items, path)
}

func findItem[T common.Assembled](items []PackageItem[T], path []string) (common.Assembled, bool) {
	res, ok := lo.Find(items, func(item PackageItem[T]) bool {
		return utils.SlicesEqual(item.Path, path)
	})
	return res.Typ, ok
}

func listByPath[T common.Assembled](items []PackageItem[T], path []string) []common.Assembled {
	return lo.FilterMap(items, func(item PackageItem[T], index int) (common.Assembled, bool) {
		if _, ok := utils.IsSubsequence(path, item.Path, 0); ok && len(item.Path) == len(path)+1 {
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
		RuntimePackage: "github.com/bdragon300/asyncapi-codegen/runtime",   // FIXME
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
