package scan

import (
	"fmt"
	"strings"

	"github.com/bdragon300/asyncapi-codegen/internal/common"
)

type RefQuerier interface {
	AssignLink(typ any)
	Ref() string
}

func NewRefManager() *RefManager {
	return &RefManager{refs: make(map[common.PackageKind][]RefQuerier)}
}

type RefManager struct {
	refs map[common.PackageKind][]RefQuerier
}

func (m *RefManager) Add(refQuery RefQuerier, fromPackage common.PackageKind) {
	if _, ok := m.refs[fromPackage]; !ok {
		m.refs[fromPackage] = nil
	}
	m.refs[fromPackage] = append(m.refs[fromPackage], refQuery)
}

func (m *RefManager) ProcessRefs(ctx *Context) {
	for bktKind, queries := range m.refs {
		pkg := ctx.Packages[bktKind]
		for _, query := range queries {
			if !strings.HasPrefix(query.Ref(), "#/") {
				panic("We don't support external refs yet")
			}
			path, _ := strings.CutPrefix(query.Ref(), "#/")
			parts := strings.Split(path, "/")
			item, ok := pkg.Find(parts)
			if !ok {
				panic(fmt.Sprintf("Cannot find %s ref path in the document", query.Ref()))
			}
			query.AssignLink(item)
		}
	}
}
