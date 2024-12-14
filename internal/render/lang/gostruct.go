package lang

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/bdragon300/go-asyncapi/internal/types"

	"github.com/bdragon300/go-asyncapi/internal/common"
)

// GoStruct defines the data required to generate a struct in Go.
type GoStruct struct {
	BaseType
	Fields []GoStructField
	// Typically it's ObjectKindOther or ObjectKindSchema
	ObjectKind common.ObjectKind
}

func (s *GoStruct) Kind() common.ObjectKind {
	return s.ObjectKind
}

func (s *GoStruct) GoTemplate() string {
	return "lang/gostruct"
}

//func (s GoStruct) NewFuncName() string {
//	return "New" + s.GetOriginalName
//}

//func (s GoStruct) ReceiverName() string {
//	return strings.ToLower(string(s.GetOriginalName[0]))
//}

//func (s GoStruct) MustGetField(name string) GoStructField {
//	f, ok := lo.Find(s.Fields, func(item GoStructField) bool {
//		return item.GetOriginalName == name
//	})
//	if !ok {
//		panic(fmt.Errorf("field %s.%s not found", s.GetOriginalName, name))
//	}
//	return f
//}

func (s *GoStruct) IsStruct() bool {
	return true
}

func (s *GoStruct) String() string {
	if s.Import != "" {
		return "GoStruct /" + s.Import + "." + s.OriginalName
	}
	return "GoStruct " + s.OriginalName
}

// GoStructField defines the data required to generate a field in Go.
type GoStructField struct {
	Name        string
	MarshalName string
	Description      string
	Type             common.GolangType
	ContentTypesFunc func() []string // Returns list of content types associated with the struct
	ExtraTags        types.OrderedMap[string, string]        // Just append these tags as constant, overwrite other tags on overlap
	ExtraTagNames    []string                                // Append these tags and fill them the same value as others
	ExtraTagValues   []string                                // Add these comma-separated values to all tags (excluding ExtraTags)
}

//func (f GoStructField) renderDefinition() []*jen.Statement {
//	var res []*jen.Statement
//	ctx.LogStartRender("GoStructField", "", f.GetOriginalName, "definition", false)
//	defer ctx.LogFinishRender()
//
//	if f.Description != "" {
//		res = append(res, jen.Comment(f.GetOriginalName+" -- "+utils.ToLowerFirstLetter(f.Description)))
//	}
//
//	stmt := jen.Id(f.GetOriginalName)
//
//	items := utils.ToCode(f.Type.U())
//	stmt = stmt.Add(items...)
//
//	if f.ContentTypesFunc != nil {
//		tagValues := append([]string{f.MarshalName}, f.ExtraTagValues...)
//		tagNames := lo.Uniq(lo.FilterMap(f.ContentTypesFunc.Targets(), func(item *render.Message, _ int) (string, bool) {
//			format := render.getFormatByContentType(item.ContentType)  // FIXME: rework this
//			return format, format != ""
//		}))
//		tagNames = append(tagNames, f.ExtraTagNames...)
//
//		tags := lo.SliceToMap(tagNames, func(item string) (string, string) {
//			return item, strings.Join(tagValues, ",")
//		})
//		tags = lo.Assign(tags, lo.FromEntries(f.ExtraTags.Entries()))
//		stmt = stmt.Tag(tags)
//	}
//
//	res = append(res, stmt)
//
//	return res
//}

func(f *GoStructField) RenderTags() string {
	tags := f.getTags()
	if tags.Len() == 0 {
		return ""
	}

	var b strings.Builder
	for _, e := range tags.Entries() {
		if b.Len() > 0 {
			b.WriteRune(' ')
		}
		b.WriteString(fmt.Sprintf(`%s:%q`, e.Key, e.Value))
	}

	str := b.String()
	if strconv.CanBackquote(str) {
		str = "`" + str + "`"
	} else {
		str = strconv.Quote(str)
	}

	return str
}

func (f *GoStructField) getTags() types.OrderedMap[string, string] {
	tagValues := append([]string{f.MarshalName}, f.ExtraTagValues...)
	var tagNames []string
	if f.ContentTypesFunc != nil {
		tagNames = f.ContentTypesFunc()
	}
	tagNames = append(tagNames, f.ExtraTagNames...)
	slices.Sort(tagNames)

	var res types.OrderedMap[string, string]
	for _, item := range tagNames {
		res.Set(item, strings.Join(tagValues, ","))
	}
	for _, item := range f.ExtraTags.Entries() {
		res.Set(item.Key, item.Value)
	}
	return res
}
