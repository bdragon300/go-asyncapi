package packages

import (
	"github.com/bdragon300/asyncapi-codegen/internal/common"
	"github.com/dave/jennifer/jen"
)

type MessagesPackage struct {
	items []common.PackageItem[common.Assembler]
}

func (m *MessagesPackage) Put(ctx *common.CompileContext, item common.Assembler) {
	m.items = append(m.items, common.PackageItem[common.Assembler]{
		Typ:  item,
		Path: ctx.PathStack(),
	})
}

func (m *MessagesPackage) Items() []common.PackageItem[common.Assembler] {
	return m.items
}

func RenderMessages(pkg *MessagesPackage, baseDir string) (files map[string]*jen.File, err error) {
	channelsGo := jen.NewFilePathName(baseDir, "messages") // FIXME: basedir is actually package path
	if err != nil {
		return
	}

	ctx := &common.AssembleContext{
		CurrentPackage: common.MessagesPackageKind,
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
		"messages/messages.go": channelsGo,
	}

	return
}
