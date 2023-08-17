package buckets

import (
	"reflect"
	"strconv"

	"github.com/bdragon300/asyncapi-codegen/internal/scancontext"
	"github.com/dave/jennifer/jen"
	"github.com/samber/lo"
)

type bucketItem struct {
	typ  scancontext.LangRenderer
	path []string
}

type LangTypeBucket struct {
	Items []bucketItem
}

func (s *LangTypeBucket) Put(ctx *scancontext.Context, item scancontext.LangRenderer) {
	s.Items = append(s.Items, bucketItem{
		typ:  item,
		path: ctx.PathStack(),
	})
}

func (s *LangTypeBucket) RenderUsage() []*jen.Statement {
	// A type with the highest level (the last one added to bucket) is a main type
	if len(s.Items) == 0 {
		panic("Empty LangTypeBucket")
	}

	names := make(map[string]scancontext.LangRenderer)
	for _, item := range s.Items {
		if item.typ.SkipRender() {
			continue
		}
		name := s.getLangName(item.typ, names)
		item.typ.PrepareRender(name)
	}
	res := lo.Flatten(lo.Map(s.Items, func(item bucketItem, index int) []*jen.Statement {
		if item.typ.SkipRender() {
			return nil
		}
		return item.typ.RenderUsage()
	}))

	return res
}

func (s *LangTypeBucket) RenderDefinitions() []*jen.Statement {
	names := make(map[string]scancontext.LangRenderer)
	for _, item := range s.Items {
		if item.typ.SkipRender() {
			continue
		}
		name := s.getLangName(item.typ, names)
		item.typ.PrepareRender(name)
	}
	res := lo.Flatten(lo.Map(s.Items, func(item bucketItem, index int) []*jen.Statement {
		if item.typ.SkipRender() {
			return nil
		}
		return item.typ.RenderDefinition()
	}))

	return res
}

func (s *LangTypeBucket) Find(path []string) (scancontext.LangRenderer, bool) {
	res, ok := lo.Find(s.Items, func(item bucketItem) bool {
		return reflect.DeepEqual(item.path, path)
	})
	return res.typ, ok
}

func (s *LangTypeBucket) getLangName(typ scancontext.LangRenderer, names map[string]scancontext.LangRenderer) string {
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
