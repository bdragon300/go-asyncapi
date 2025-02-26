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

// GoStruct represents a Go struct.
type GoStruct struct {
	BaseType
	// Fields is a list of fields in the struct. If empty, then the struct is empty.
	Fields []GoStructField
}

func (s *GoStruct) String() string {
	if s.Import != "" {
		return "GoStruct /" + s.Import + "." + s.OriginalName
	}
	return "GoStruct " + s.OriginalName
}

func (s *GoStruct) GoTemplate() string {
	return "code/lang/gostruct"
}

func (s *GoStruct) IsStruct() bool {
	return true
}

// GoStructField represents a field in a Go struct (without generics support).
type GoStructField struct {
	// Name is the name of the field.
	Name string
	// MarshalName is the name of the field for marshaling/unmarshaling. It appears in the struct tag, like `json:"marshal_name"`.
	MarshalName string
	// Description is an optional field description. Renders as Go doc comment.
	Description string
	// Type is the type of the field.
	Type common.GolangType
	// ContentTypesFunc callback returns a list of content types associated with the struct. Used to compose a struct tag on the rendering stage.
	ContentTypesFunc func() []string
	// ExtraTags is extra tags and their values to append to the struct tag. If a tag already exists, it is overwritten.
	ExtraTags types.OrderedMap[string, string]
	// ExtraTagNames are extra tags to append to the struct tag, their values will be filled with the same name as other tags.
	ExtraTagNames []string
	// ExtraTagValues are the values to append as comma-separated string to all tags (excluding ExtraTags).
	// E.g. []{"omitempty", "string"} will be appended as ``,omitempty,string'' to all tags.
	ExtraTagValues []string
}

// RenderTags returns the Go struct field tag contents. E.g. “json:"name,omitempty,string" xml:"name"”.
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
