package packages

import (
	"github.com/bdragon300/asyncapi-codegen/internal/common"
	"github.com/bdragon300/asyncapi-codegen/internal/lang"
	"github.com/bdragon300/asyncapi-codegen/internal/render"
	"github.com/bdragon300/asyncapi-codegen/internal/scan"
	"github.com/bdragon300/asyncapi-codegen/internal/utils"
	"github.com/dave/jennifer/jen"
	"github.com/samber/lo"
)

type PackageItem[T render.LangRenderer] struct {
	Typ  T
	Path []string
}

type ModelsPackage struct {
	Items []PackageItem[lang.LangType]
}

func (s *ModelsPackage) Put(ctx *scan.Context, item render.LangRenderer) {
	s.Items = append(s.Items, PackageItem[lang.LangType]{
		Typ:  item.(lang.LangType),
		Path: ctx.PathStack(),
	})
}

func (s *ModelsPackage) Find(path []string) (render.LangRenderer, bool) {
	return findItem(s.Items, path)
}

func (s *ModelsPackage) List(path []string) []render.LangRenderer {
	return listByPath(s.Items, path)
}

func findItem[T render.LangRenderer](items []PackageItem[T], path []string) (render.LangRenderer, bool) {
	res, ok := lo.Find(items, func(item PackageItem[T]) bool {
		return utils.SlicesEqual(item.Path, path)
	})
	return res.Typ, ok
}

func listByPath[T render.LangRenderer](items []PackageItem[T], path []string) []render.LangRenderer {
	return lo.FilterMap(items, func(item PackageItem[T], index int) (render.LangRenderer, bool) {
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

	ctx := &render.Context{
		CurrentPackage: common.ModelsPackageKind,
		ImportBase:     "github.com/bdragon300/asyncapi-codegen/generated", // FIXME
		RuntimePackage: "github.com/bdragon300/asyncapi-codegen/runtime",   // FIXME
	}

	for _, item := range pkg.Items {
		if !item.Typ.AllowRender() {
			continue
		}

		for _, stmt := range item.Typ.RenderDefinition(ctx) {
			modelsGo.Add(stmt)
		}
		modelsGo.Add(jen.Line())
	}

	files = map[string]*jen.File{
		"models/models.go": modelsGo,
	}

	return
}
