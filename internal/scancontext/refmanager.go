package scancontext

import (
	"fmt"
	"strings"
)

type RefQuerier interface {
	AssignLink(typ any)
	Ref() string
}

func NewRefManager() *RefManager {
	return &RefManager{refs: make(map[BucketKind][]RefQuerier)}
}

type RefManager struct {
	refs map[BucketKind][]RefQuerier
}

func (m *RefManager) Add(refQuery RefQuerier, fromBucket BucketKind) {
	if _, ok := m.refs[fromBucket]; !ok {
		m.refs[fromBucket] = nil
	}
	m.refs[fromBucket] = append(m.refs[fromBucket], refQuery)
}

func (m *RefManager) ProcessRefs(ctx *Context) {
	for bktKind, queries := range m.refs {
		bucket := ctx.Buckets[bktKind]
		for _, query := range queries {
			if !strings.HasPrefix(query.Ref(), "#/") {
				panic("We don't support external refs yet")
			}
			path, _ := strings.CutPrefix(query.Ref(), "#/")
			parts := strings.Split(path, "/")
			item, ok := bucket.Find(parts)
			if !ok {
				panic(fmt.Sprintf("Cannot find %s ref path in the document", query.Ref()))
			}
			query.AssignLink(item)
		}
	}
}
