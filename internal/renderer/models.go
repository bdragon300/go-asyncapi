package renderer

import (
	"github.com/bdragon300/asyncapi-codegen/internal/buckets"
	"github.com/bdragon300/asyncapi-codegen/internal/scanner"
	"github.com/dave/jennifer/jen"
)

type ModelsTplArgs struct {
	Definitions string
}

func RenderTypes(bucket *buckets.Schema, baseDir string) (files map[string]*jen.File, err error) {
	modelsGo := jen.NewFilePathName(baseDir, "models") // FIXME: basedir is actually package path
	if err != nil {
		return
	}

	names := make(map[string]scanner.LangRenderer)
	var itemsToRender []scanner.LangRenderer
	for _, item := range bucket.Items() {
		if item.SkipRender() {
			continue
		}
		name := getUniqueName(item, names)
		item.PrepareRender(name)
		itemsToRender = append(itemsToRender, item)
	}

	for _, item := range itemsToRender {
		rendered := item.RenderDefinition()
		for _, stmt := range rendered {
			modelsGo.Add(stmt)
		}
		modelsGo.Add(jen.Line())
	}

	files = map[string]*jen.File{
		"models/models.go": modelsGo,
	}

	return
}

func getUniqueName(typ scanner.LangRenderer, names map[string]scanner.LangRenderer) string {
	langName := typ.GetDefaultName()
	findName := langName

	// Use type's name or append a number such as MyType2, MyType3, ...
	for i := 1; ; i++ {
		if _, ok := names[findName]; !ok {
			names[findName] = typ
			return findName
		}
		findName = langName + strconv.Itoa(i)
	}
}
