package packages

import (
	"github.com/bdragon300/asyncapi-codegen/internal/common"
	"github.com/bdragon300/asyncapi-codegen/internal/lang"
	"github.com/bdragon300/asyncapi-codegen/internal/render"
	"github.com/bdragon300/asyncapi-codegen/internal/scan"
	"github.com/dave/jennifer/jen"
)

type MessagesPackage struct {
	Types    []PackageItem[lang.LangType]
	Messages []PackageItem[*lang.Message]
}

func (m *MessagesPackage) Put(ctx *scan.Context, item render.LangRenderer) {
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

func (m *MessagesPackage) Find(path []string) (render.LangRenderer, bool) {
	if res, ok := findItem(m.Types, path); ok {
		return res, true
	}
	if res, ok := findItem(m.Messages, path); ok {
		return res, true
	}
	return nil, false
}

func (m *MessagesPackage) List(path []string) []render.LangRenderer {
	res := listByPath(m.Types, path)
	res = append(res, listByPath(m.Messages, path)...)
	return res
}

func RenderMessages(pkg *MessagesPackage, baseDir string) (files map[string]*jen.File, err error) {
	channelsGo := jen.NewFilePathName(baseDir, "messages") // FIXME: basedir is actually package path
	if err != nil {
		return
	}

	ctx := &render.Context{
		CurrentPackage: common.MessagesPackageKind,
		ImportBase:     "github.com/bdragon300/asyncapi-codegen/generated", // FIXME
		RuntimePackage: "github.com/bdragon300/asyncapi-codegen/runtime",   // FIXME
	}
	for _, item := range pkg.Messages {
		for _, stmt := range item.Typ.RenderDefinition(ctx) {
			channelsGo.Add(stmt)
		}
		channelsGo.Add(jen.Line())
	}

	files = map[string]*jen.File{
		"messages/messages.go": channelsGo,
	}

	return
}
