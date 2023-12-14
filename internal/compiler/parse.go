package compiler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/bdragon300/asyncapi-codegen-go/internal/asyncapi"
	"github.com/bdragon300/asyncapi-codegen-go/internal/common"
	"gopkg.in/yaml.v3"
)

var logger = common.NewLogger("")

type SchemaKind string

const (
	SchemaKindAsyncapi   SchemaKind = "asyncapi"
	SchemaKindJsonschema SchemaKind = "jsonschema"
	SchemaKindOpenapi    SchemaKind = "openapi"
)

type specTypeTest struct {
	Asyncapi string `json:"asyncapi" yaml:"asyncapi"`
	Openapi  string `json:"openapi" yaml:"openapi"`
}

func parseSpecFile(fileName string) (SchemaKind, compiledObject, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return "", nil, fmt.Errorf("open file %s: %w", fileName, err)
	}
	defer func() { _ = f.Close() }()

	switch {
	case strings.HasSuffix(fileName, ".yaml") || strings.HasSuffix(fileName, ".yml"):
		logger.Debug("Found YAML spec file", "filename", fileName)
		schemaKind, spec, err := guessSpecKind(yaml.NewDecoder(f))
		if err != nil {
			return "", nil, fmt.Errorf("guess spec kind: %w", err)
		}
		if _, err = f.Seek(0, io.SeekStart); err != nil {
			return "", nil, fmt.Errorf("file seek: %w", err)
		}
		err = yaml.NewDecoder(f).Decode(spec)
		return schemaKind, spec, err
	case strings.HasSuffix(fileName, ".json"):
		logger.Debug("Found JSON spec file", "filename", fileName)
		schemaKind, spec, err := guessSpecKind(json.NewDecoder(f))
		if err != nil {
			return "", nil, fmt.Errorf("guess spec kind: %w", err)
		}
		if _, err = f.Seek(0, io.SeekStart); err != nil {
			return "", nil, fmt.Errorf("file seek: %w", err)
		}
		err = json.NewDecoder(f).Decode(spec)
		return schemaKind, spec, err
	}

	return "", nil, errors.New("cannot determine format of a spec file: unknown filename extension")
}

type decoder interface {
	Decode(v any) error
}

func guessSpecKind(decoder decoder) (SchemaKind, compiledObject, error) {
	test := specTypeTest{}

	if err := decoder.Decode(&test); err != nil {
		return "", nil, err
	}
	switch {
	case test.Asyncapi != "":
		return SchemaKindAsyncapi, &asyncapi.AsyncAPI{}, nil
	case test.Openapi != "":
		panic("openapi not implemented")
	}
	panic("jsonschema not implemented")
	// Assume that some data is jsonschema, TODO: maybe it's better to match more strict?
}
