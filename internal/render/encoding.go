package render

import (
	"github.com/bdragon300/go-asyncapi/internal/render/lang"
	"strings"

	"github.com/samber/lo"
)

//var encodingEncoders = map[string]j.Code{
//	"json": j.Op(`func(w io.Writer) Encoder`).Block(j.Return(j.Qual("encoding/json", "NewEncoder").Call(j.Id("w")))),
//	"yaml": j.Op(`func(w io.Writer) Encoder`).Block(j.Return(j.Qual("gopkg.in/yaml.v3", "NewEncoder").Call(j.Id("w")))),
//	// TODO: add other encoders: protobuf, avro, etc.
//}

//var encodingDecoders = map[string]j.Code{
//	"json": j.Op(`func(r io.Reader) Decoder`).Block(j.Return(j.Qual("encoding/json", "NewDecoder").Call(j.Id("r")))),
//	"yaml": j.Op(`func(r io.Reader) Decoder`).Block(j.Return(j.Qual("gopkg.in/yaml.v3", "NewDecoder").Call(j.Id("r")))),
//	// TODO: add other decoders: protobuf, avro, etc.
//}

type EncodingEncode struct {
	AllMessages        *lang.ListPromise[*Message]
	DefaultContentType string
}

//func (e EncodingEncode) RenderDefinition(ctx *common.RenderContext) []*j.Statement {
//	ctx.LogStartRender("EncodingEncode", "", "", "definition", e.DirectRendering())
//	defer ctx.LogFinishRender()
//
//	contentTypes := lo.Uniq(lo.FilterMap(e.AllMessages.Targets(), func(item *Message, _ int) (string, bool) {
//		return item.ContentType, item.ContentType != ""
//	}))
//	return []*j.Statement{
//		j.Op(`
//			type Encoder interface {
//				Encode(v any) error
//			}`),
//
//		j.Add(utils.QualSprintf(`var Encoders = map[string]func(w %Q(io,Writer)) Encoder`)).Values(j.DictFunc(func(d j.Dict) {
//			for _, ct := range contentTypes {
//				format := getFormatByContentType(ct)
//				if format != "" {
//					if v, ok := encodingEncoders[format]; ok {
//						d[j.Lit(ct)] = v
//					}
//				} else {
//					d[j.Lit(ct)] = j.Add(utils.QualSprintf(`func(_ %Q(io,Writer)) Encoder { panic("No encoder is set for content type %s") }`, ct))
//				}
//			}
//		})),
//
//		j.Add(utils.QualSprintf(`
//			func NewEncoder(contentType string, w %Q(io,Writer)) Encoder {
//				if v, ok := Encoders[contentType]; ok {
//					return v(w)
//				}
//				panic("Unknown content type " + contentType)
//			}`)),
//	}
//}

func (m EncodingEncode) Selectable() bool {
	return true
}

func (e EncodingEncode) MessageContentTypes() []string {
	return lo.Uniq(lo.FilterMap(e.AllMessages.Targets(), func(item *Message, _ int) (string, bool) {
		return item.ContentType, item.ContentType != ""
	}))
}

func (e EncodingEncode) FormatByContentType(contentType string) string {
	return getFormatByContentType(contentType)
}

func (e EncodingEncode) ID() string {
	return "Encode"
}

func (e EncodingEncode) String() string {
	return "EncodingEncode"
}

type EncodingDecode struct {
	AllMessages        *lang.ListPromise[*Message]
	DefaultContentType string
}

//func (e EncodingDecode) RenderDefinition(ctx *common.RenderContext) []*j.Statement {
//	ctx.LogStartRender("EncodingDecode", "", "", "definition", e.DirectRendering())
//	defer ctx.LogFinishRender()
//
//	contentTypes := lo.Uniq(lo.FilterMap(e.AllMessages.Targets(), func(item *Message, _ int) (string, bool) {
//		return item.ContentType, item.ContentType != ""
//	}))
//	return []*j.Statement{
//		j.Op(`
//			type Decoder interface {
//				Decode(v any) error
//			}`),
//
//		j.Add(utils.QualSprintf(`var Decoders = map[string]func(r %Q(io,Reader)) Decoder`)).Values(j.DictFunc(func(d j.Dict) {
//			for _, ct := range contentTypes {
//				format := getFormatByContentType(ct)
//				if format != "" {
//					if v, ok := encodingDecoders[format]; ok {
//						d[j.Lit(ct)] = v
//					}
//				} else {
//					d[j.Lit(ct)] = j.Add(utils.QualSprintf(`func(_ %Q(io,Reader)) Decoder { panic("No decoder is set for content type %s") }`, ct))
//				}
//			}
//		})),
//
//		j.Add(utils.QualSprintf(`
//			func NewDecoder(contentType string, r %Q(io,Reader)) Decoder {
//				if v, ok := Decoders[contentType]; ok {
//					return v(r)
//				}
//				panic("Unknown content type " + contentType)
//			}`)),
//	}
//}

func (e EncodingDecode) MessageContentTypes() []string {
	return lo.Uniq(lo.FilterMap(e.AllMessages.Targets(), func(item *Message, _ int) (string, bool) {
		return item.ContentType, item.ContentType != ""
	}))
}

func (e EncodingDecode) FormatByContentType(contentType string) string {
	return getFormatByContentType(contentType)
}

func (e EncodingDecode) ID() string {
	return "Decode"
}

func (e EncodingDecode) String() string {
	return "EncodingDecode"
}

func getFormatByContentType(contentType string) string {
	// TODO: add other formats: protobuf, avro, etc.
	switch {
	case strings.HasSuffix(contentType, "json"):
		return "json"
	case strings.HasSuffix(contentType, "yaml"):
		return "yaml"
	}
	return ""
}
