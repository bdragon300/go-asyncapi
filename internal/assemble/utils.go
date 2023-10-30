package assemble

import (
	"strings"

	"github.com/samber/lo"

	"github.com/bdragon300/asyncapi-codegen-go/internal/common"
	"github.com/bdragon300/asyncapi-codegen-go/internal/utils"
	j "github.com/dave/jennifer/jen"
)

var serializerEncoders = map[string]j.Code{
	"json": j.Op(`func(w io.Writer) Encoder`).Block(j.Return(j.Qual("encoding/json", "NewEncoder").Call(j.Id("w")))),
	"yaml": j.Op(`func(w io.Writer) Encoder`).Block(j.Return(j.Qual("gopkg.in/yaml.v3", "NewEncoder").Call(j.Id("w")))),
	// TODO: add other encoders: protobuf, avro, etc.
}

var serializerDecoders = map[string]j.Code{
	"json": j.Op(`func(r io.Reader) Decoder`).Block(j.Return(j.Qual("encoding/json", "NewDecoder").Call(j.Id("r")))),
	"yaml": j.Op(`func(r io.Reader) Decoder`).Block(j.Return(j.Qual("gopkg.in/yaml.v3", "NewDecoder").Call(j.Id("r")))),
	// TODO: add other decoders: protobuf, avro, etc.
}

type UtilsSerializer struct {
	AllMessages        *LinkList[*Message]
	DefaultContentType string
}

func (u UtilsSerializer) AllowRender() bool {
	return true
}

func (u UtilsSerializer) AssembleDefinition(_ *common.AssembleContext) []*j.Statement {
	contentTypes := lo.Uniq(lo.FilterMap(u.AllMessages.Targets(), func(item *Message, index int) (string, bool) {
		return item.ContentType, item.ContentType != ""
	}))
	return []*j.Statement{
		j.Op(`
			type Encoder interface {
				Encode(v any) error
			}
			
			type Decoder interface {
				Decode(v any) error
			}`),

		j.Add(utils.QualSprintf(`var Encoders = map[string]func(w %Q(io,Writer)) Encoder`)).Values(j.DictFunc(func(d j.Dict) {
			for _, ct := range contentTypes {
				format := getFormatByContentType(ct)
				if format != "" {
					if v, ok := serializerEncoders[format]; ok {
						d[j.Lit(ct)] = v
					}
				} else {
					d[j.Lit(ct)] = j.Add(utils.QualSprintf(`func(_ %Q(io,Writer)) Encoder { panic("No encoder is set for content type %s") }`, ct))
				}
			}
		})),

		j.Add(utils.QualSprintf(`var Decoders = map[string]func(r %Q(io,Reader)) Decoder`)).Values(j.DictFunc(func(d j.Dict) {
			for _, ct := range contentTypes {
				format := getFormatByContentType(ct)
				if format != "" {
					if v, ok := serializerDecoders[format]; ok {
						d[j.Lit(ct)] = v
					}
				} else {
					d[j.Lit(ct)] = j.Add(utils.QualSprintf(`func(_ %Q(io,Reader)) Decoder { panic("No decoder is set for content type %s") }`, ct))
				}
			}
		})),

		j.Add(utils.QualSprintf(`
			func NewEncoder(contentType string, w %Q(io,Writer)) Encoder {
				if v, ok := Encoders[contentType]; ok {
					return v(w)
				}
				panic("Unknown content type " + contentType)
			}
			
			func NewDecoder(contentType string, r %Q(io,Reader)) Decoder {
				if v, ok := Decoders[contentType]; ok {
					return v(r)
				}
				panic("Unknown content type " + contentType)
			}`)),
	}
}

func (u UtilsSerializer) AssembleUsage(_ *common.AssembleContext) []*j.Statement {
	panic("not implemented")
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
