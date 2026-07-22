package compiler

import (
	"encoding/json"
	"io"
	"path"
	"strings"

	"gopkg.in/yaml.v3"
)

type DocumentKind string

const (
	DocumentKindAsyncapi   DocumentKind = "asyncapi"
	DocumentKindJsonschema DocumentKind = "jsonschema"
	DocumentKindOpenapi    DocumentKind = "openapi"
)

type documentFormatTester struct {
	Asyncapi string `json:"asyncapi" yaml:"asyncapi"`
	Openapi  string `json:"openapi" yaml:"openapi"`
}

type AnyDecoder interface {
	Decode(v any) error
}

type AnyEncoder interface {
	Encode(v any) error
}

type DecoderFactory func(io.Reader) AnyDecoder

type EncoderFactory func(io.Writer) AnyEncoder

// GuessDocumentFormat tries to guess the document format by its extension and returns the encoder and decoder factories and format name.
// If the format is not recognized, it returns nil factories and an empty string.
func GuessDocumentFormat(fileName string, indent int) (EncoderFactory, DecoderFactory, string) {
	switch path.Ext(fileName) {
	case ".yaml", ".yml":
		ef := func(w io.Writer) AnyEncoder { e := yaml.NewEncoder(w); e.SetIndent(indent); return e }
		df := func(r io.Reader) AnyDecoder { return yaml.NewDecoder(r) }
		return ef, df, "yaml"
	case ".json":
		ef := func(w io.Writer) AnyEncoder {
			e := json.NewEncoder(w)
			e.SetIndent("", strings.Repeat(" ", indent))
			return e
		}
		df := func(r io.Reader) AnyDecoder { return json.NewDecoder(r) }
		return ef, df, "json"
	}
	return nil, nil, ""
}

// guessDocumentKind tries to guess the document kind by its contents.
func guessDocumentKind(decoder AnyDecoder) (DocumentKind, error) {
	test := documentFormatTester{}

	if err := decoder.Decode(&test); err != nil {
		return "", err
	}
	switch {
	case test.Asyncapi != "":
		return DocumentKindAsyncapi, nil
	case test.Openapi != "":
		panic("openapi not implemented")
	}
	panic("jsonschema not implemented")
}
