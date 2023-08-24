package packages

import (
	"fmt"
	"strings"

	"github.com/bdragon300/asyncapi-codegen/internal/common"
	"github.com/bdragon300/asyncapi-codegen/internal/lang"
	"github.com/bdragon300/asyncapi-codegen/internal/render"
	"github.com/bdragon300/asyncapi-codegen/internal/scan"
	"github.com/dave/jennifer/jen"
)

type MessagePackage struct {
	Types    []PackageItem[lang.LangType]
	Messages []PackageItem[*lang.Message]
}

func (m *MessagePackage) Put(ctx *scan.Context, item render.LangRenderer) {
	switch v := item.(type) {
	case *lang.Message:
		m.Messages = append(m.Messages, PackageItem[*lang.Message]{
			Typ:  v,
			Path: ctx.PathStack(),
		})
	default:
		m.Types = append(m.Types, PackageItem[lang.LangType]{
			Typ:  v.(lang.LangType),
			Path: ctx.PathStack(),
		})
	}
}

func (m *MessagePackage) Find(path []string) (render.LangRenderer, bool) {
	if res, ok := findItem(m.Types, path); ok {
		return res, true
	}
	if res, ok := findItem(m.Messages, path); ok {
		return res, true
	}
	return nil, false
}

func (m *MessagePackage) MustFind(path []string) render.LangRenderer {
	res, ok := m.Find(path)
	if !ok {
		panic(fmt.Sprintf("Object %s not found", strings.Join(path, ".")))
	}
	return res
}

func RenderMessages(pkg *MessagePackage, baseDir string) (files map[string]*jen.File, err error) {
	modelsGo := jen.NewFilePathName(baseDir, "messages") // FIXME: basedir is actually package path
	if err != nil {
		return
	}

	ctx := &render.Context{
		CurrentPackage: common.MessagePackageKind,
		ImportBase:     "github.com/bdragon300/asyncapi-codegen/generated", // FIXME
	}
	for _, item := range pkg.Messages {
		for _, stmt := range item.Typ.RenderDefinition(ctx) {
			modelsGo.Add(stmt)
		}
		modelsGo.Add(jen.Line())
		modelsGo.ImportNames(item.Typ.AdditionalImports())
	}

	files = map[string]*jen.File{
		"messages/messages.go": modelsGo,
	}

	return
}
