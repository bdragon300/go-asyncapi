package compiler

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
		panic("openapi not implemented") // TODO
	}
	panic("jsonschema not implemented") // TODO
}
