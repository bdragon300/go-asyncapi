package buckets

import (
	"reflect"

	"github.com/bdragon300/asyncapi-codegen/internal/lang"

	"github.com/bdragon300/asyncapi-codegen/internal/scanner"
	"github.com/samber/lo"
)

type schemaBucketItem struct {
	typ  lang.LangType
	path []string
}

type Schema struct {
	items []schemaBucketItem
}

func (s *Schema) Put(ctx *scanner.Context, item scanner.LangRenderer) {
	s.items = append(s.items, schemaBucketItem{
		typ:  item.(lang.LangType),
		path: ctx.PathStack(),
	})
}

func (s *Schema) Find(path []string) (scanner.LangRenderer, bool) {
	res, ok := lo.Find(s.items, func(item schemaBucketItem) bool {
		return reflect.DeepEqual(item.path, path)
	})
	return res.typ.(scanner.LangRenderer), ok
}

func (s *Schema) Items() []scanner.LangRenderer {
	return lo.Map(s.items, func(item schemaBucketItem, _ int) scanner.LangRenderer { return item.typ.(scanner.LangRenderer) })
}
