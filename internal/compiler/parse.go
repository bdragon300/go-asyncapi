package compiler

import (
	"github.com/bdragon300/go-asyncapi/internal/asyncapi"
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

type anyDecoder interface {
	Decode(v any) error
}

// guessDocumentKind tries to guess the document kind by its contents.
func guessDocumentKind(decoder anyDecoder) (DocumentKind, compiledObject, error) {
	test := documentFormatTester{}

	if err := decoder.Decode(&test); err != nil {
		return "", nil, err
	}
	switch {
	case test.Asyncapi != "":
		return DocumentKindAsyncapi, &asyncapi.AsyncAPI{}, nil
	case test.Openapi != "":
		panic("openapi not implemented") // TODO
	}
	panic("jsonschema not implemented") // TODO
	// Assume that some data is jsonschema, TODO: maybe it's better to match more strict?
}
