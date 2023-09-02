package packages

import (
	"github.com/bdragon300/asyncapi-codegen/internal/assemble"
	"github.com/bdragon300/asyncapi-codegen/internal/common"
	"github.com/dave/jennifer/jen"
)

type MessagesPackage struct {
	Types    []PackageItem[common.GolangType]
	Messages []PackageItem[*assemble.Message]
}

func (m *MessagesPackage) Put(ctx *common.CompileContext, item common.Assembler) {
	switch v := item.(type) {
	case *assemble.Message:
		m.Messages = append(m.Messages, PackageItem[*assemble.Message]{
			Typ:  v,
			Path: ctx.PathStack(),
		})
	default:
		m.Types = append(m.Types, PackageItem[common.GolangType]{
			Typ:  v.(common.GolangType),
			Path: ctx.PathStack(),
		})
	}
}

// TODO: refactor the methods below
func (m *MessagesPackage) Find(path []string) (common.Assembler, bool) {
	if res, ok := findItem(m.Types, path); ok {
		return res, true
	}
	if res, ok := findItem(m.Messages, path); ok {
		return res, true
	}
	return nil, false
}

func (m *MessagesPackage) FindBy(cb func(item any, path []string) bool) (common.Assembler, bool) {
	if res, ok := findItemBy(m.Types, cb); ok {
		return res, true
	}
	if res, ok := findItemBy(m.Messages, cb); ok {
		return res, true
	}
	return nil, false
}

func (m *MessagesPackage) List(path []string) []common.Assembler {
	res := listSubItems(m.Types, path)
	res = append(res, listSubItems(m.Messages, path)...)
	return res
}

func (m *MessagesPackage) ListBy(cb func(item any, path []string) bool) []common.Assembler {
	res := listSubItemsBy(m.Types, cb)
	res = append(res, listSubItemsBy(m.Messages, cb)...)
	return res
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
	for _, item := range pkg.Messages {
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
