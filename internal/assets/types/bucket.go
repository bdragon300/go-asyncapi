package types

import (
	"reflect"

	"github.com/bdragon300/asyncapi-codegen/internal/scanner"
	"github.com/samber/lo"
)

type bucketItem struct {
	typ  LangType
	path []string
}

type LangTypeBucket struct {
	items []bucketItem
}

func (s *LangTypeBucket) Put(ctx *scanner.Context, item scanner.LangRenderer) {
	s.items = append(s.items, bucketItem{
		typ:  item.(LangType),
		path: ctx.PathStack(),
	})
}

func (s *LangTypeBucket) Find(path []string) (scanner.LangRenderer, bool) {
	res, ok := lo.Find(s.items, func(item bucketItem) bool {
		return reflect.DeepEqual(item.path, path)
	})
	return res.typ.(scanner.LangRenderer), ok
}

func (s *LangTypeBucket) Items() []scanner.LangRenderer {
	return lo.Map(s.items, func(item bucketItem, _ int) scanner.LangRenderer { return item.typ.(scanner.LangRenderer) })
}
