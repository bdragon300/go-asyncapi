package lang

import (
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/log"
	"github.com/bdragon300/go-asyncapi/internal/types"
)

// GoStruct represents a Go struct.
type GoStruct struct {
	BaseType
	// Fields is a list of fields in the struct. If empty, then the struct is empty.
	Fields []GoStructField

	// AdditionalPropertiesType points to the schema of additionalProperties field when it either set to schema or
	// set to true. Nil if additionalProperties is not set or set to false.
	AdditionalPropertiesType common.GolangType

	StructFieldRenderInfo StructFieldRenderInfo

	// ContentTypesFunc callback returns a list of content types that are used in document. Nil if struct is not
	// a data model (not marked by common.SchemaTagDataModel tag).
	ContentTypesFunc func() []string

	// ContentTypeTagsFunc callback returns a list of tags associated with every content type that is used in document.
	// Nil if struct is not a data model (not marked by common.SchemaTagDataModel tag).
	ContentTypeTagsFunc func() []string
}

func (s *GoStruct) String() string {
	if s.Import != "" {
		return fmt.Sprintf("GoStruct(%s.%s)", s.Import, s.OriginalName)
	}
	return "GoStruct(" + s.OriginalName + ")"
}

func (s *GoStruct) GoTemplate() string {
	return "code/lang/gostruct"
}

func (s *GoStruct) IsStruct() bool {
	return true
}

func (s *GoStruct) StructRenderInfo() StructFieldRenderInfo {
	return s.StructFieldRenderInfo
}

func (s *GoStruct) AllContentTypes() []string {
	if s.ContentTypesFunc != nil {
		return s.ContentTypesFunc()
	}
	return nil
}

// AdditionalPropertiesIsStruct returns true if additionalProperties field is set to an object schema.
func (s *GoStruct) AdditionalPropertiesIsStruct() bool {
	if s.AdditionalPropertiesType == nil {
		return false
	}

	v := common.DerefArtifact[common.GolangType](s.AdditionalPropertiesType)
	v = common.UnwrapGolangType[common.GolangType](v)
	r, ok := v.(golangStructType)
	return ok && r.IsStruct()
}

func (s *GoStruct) AdditionalPropertiesIsField() bool {
	return s.AdditionalPropertiesType != nil && !s.AdditionalPropertiesIsStruct()
}

// RenderAllContentTypeTagsOmitted returns a Go expression with field tag for all content types with "-" value,
// which is used to ignore these tags in struct fields. E.g. `json:"-" xml:"-"`.
func (s *GoStruct) RenderAllContentTypeTagsOmitted() string {
	var tagNames types.OrderedMap[string, string]
	if s.ContentTypeTagsFunc != nil {
		for _, tag := range s.ContentTypeTagsFunc() {
			tagNames.Set(tag, "-")
		}
	}
	return toFieldTagsExpression(tagNames)
}

// GoStructField represents a field in a Go struct (without generics support).
type GoStructField struct {
	// OriginalName is the name of the field.
	OriginalName string
	// MarshalName is the name of the field for marshaling/unmarshaling. It appears in the struct tag, like `json:"marshal_name"`.
	MarshalName string
	// Description is an optional field description. Renders as Go doc comment.
	Description string
	// Type is the type of the field.
	Type common.GolangType
	// ContentTypeTagsFunc callback returns a list of content type tags associated with the struct.
	// Used to compose a struct tag on the rendering stage.
	ContentTypeTagsFunc func() []string
}

func (f *GoStructField) Name() string {
	if v, ok := f.Type.(structFieldRenderer); ok && v.StructRenderInfo().IsEmbeddedType {
		// Type is embedded, so no field name
		return ""
	}
	return f.OriginalName
}

// RenderTags returns the struct field tag Go expression in backquotes. E.g. `json:"name,omitempty,string" xml:"name"`.
func (f *GoStructField) RenderTags() string {
	fieldRenderer, ok := f.Type.(structFieldRenderer)
	if !ok {
		return ""
	}

	logger := log.GetLogger(log.LoggerPrefixRendering)
	tags := f.getTags(fieldRenderer.StructRenderInfo())
	logger.Trace("--> Rendering field tags", "field", f.OriginalName, "values", maps.Collect(tags.Entries()))
	return toFieldTagsExpression(tags)
}

func (f *GoStructField) getTags(structRenderInfo StructFieldRenderInfo) types.OrderedMap[string, string] {
	tagValues := append([]string{f.MarshalName}, structRenderInfo.TagValues...)
	var tagNames []string
	if f.ContentTypeTagsFunc != nil {
		tagNames = f.ContentTypeTagsFunc()
	}
	tagNames = append(tagNames, structRenderInfo.TagNames...)
	slices.Sort(tagNames)

	var res types.OrderedMap[string, string]
	for _, item := range tagNames {
		res.Set(item, strings.Join(tagValues, ","))
	}
	for k, v := range structRenderInfo.Tags.Entries() {
		res.Set(k, v)
	}
	return res
}

// StructFieldRenderInfo contains extra information for rendering a type in struct fields.
type StructFieldRenderInfo struct {
	// IsEmbeddedType is true if this type is rendered as embedded field in a struct (i.e. field without a name).
	// Comes from x-go-type.embedded extra field.
	IsEmbeddedType bool

	/* x-go-tags* fields */

	// Tags is extra tags and their values to append to the struct tag. If a tag already exists, it is overwritten.
	Tags types.OrderedMap[string, string]
	// TagNames are extra tags to append to the struct tag, their values will be filled with the same name as other tags.
	TagNames []string
	// TagValues are the values to append as comma-separated string to all tags (excluding ExtraTags).
	// E.g. []{"omitempty", "string"} will be appended as ``,omitempty,string'' to all tags.
	TagValues []string
}

// toFieldTagsExpression converts a map of tags to a Go struct field tag expression in backquotes, that is meant to
// be placed after the struct field declaration.
func toFieldTagsExpression(tags types.OrderedMap[string, string]) string {
	if tags.Len() == 0 {
		return ""
	}

	var b strings.Builder
	for k, v := range tags.Entries() {
		if b.Len() > 0 {
			b.WriteRune(' ')
		}
		b.WriteString(fmt.Sprintf(`%s:%q`, k, v))
	}

	str := b.String()
	if strconv.CanBackquote(str) {
		str = "`" + str + "`"
	} else {
		str = strconv.Quote(str)
	}
	return str
}
