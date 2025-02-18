package lang

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/bdragon300/go-asyncapi/internal/log"
	"github.com/samber/lo"

	"github.com/bdragon300/go-asyncapi/internal/types"

	"github.com/bdragon300/go-asyncapi/internal/common"
)

// GoStruct defines the data required to generate a struct in Go.
type GoStruct struct {
	BaseType
	Fields []GoStructField
}

func (s *GoStruct) GoTemplate() string {
	return "code/lang/gostruct"
}

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
	Name             string
	MarshalName      string
	Description      string
	Type             common.GolangType
	ContentTypesFunc func() []string                  // Returns list of content types associated with the struct
	ExtraTags        types.OrderedMap[string, string] // Just append these tags as constant, overwrite other tags on overlap
	ExtraTagNames    []string                         // Append these tags and fill them the same value as others
	ExtraTagValues   []string                         // Add these comma-separated values to all tags (excluding ExtraTags)
}

func (f *GoStructField) RenderTags() string {
	logger := log.GetLogger(log.LoggerPrefixRendering)
	tags := f.getTags()
	logger.Trace("--> Rendering field tags", "field", f.Name, "values", lo.FromEntries(tags.Entries()))
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
