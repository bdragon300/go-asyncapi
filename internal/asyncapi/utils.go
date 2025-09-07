package asyncapi

import (
	"errors"
	"fmt"
	"strings"
	"unicode"

	"github.com/bdragon300/go-asyncapi/internal/render"
	"github.com/samber/lo"
)

// guessTagByContentType guesses the struct tag name by the MIME type. It returns the last
// word extracted from the content type string, e.g. for "application/xhtml+xml" it will return "xml".
func guessTagByContentType(contentType string) string {
	words := strings.FieldsFunc(contentType, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})
	if res, ok := lo.Last(words); ok {
		return res
	}

	return contentType
}

// parseRuntimeExpression parses a runtime expression in the form of "$message.source#fragment" and returns
// the struct field kind where the value is located and the fragment part, split by slashes.
// See: https://github.com/asyncapi/spec/blob/master/spec/asyncapi.md#runtimeExpression
func parseRuntimeExpression(location string) (render.RuntimeExpressionStructFieldKind, []string, error) {
	locationParts := strings.SplitN(location, "#", 2)
	if len(locationParts) < 2 {
		return "", nil, errors.New("no source or fragment")
	}

	var structField render.RuntimeExpressionStructFieldKind
	switch {
	case strings.HasSuffix(locationParts[0], "header"):
		structField = render.RuntimeExpressionStructFieldKindHeaders
	case strings.HasSuffix(locationParts[0], "payload"):
		structField = render.RuntimeExpressionStructFieldKindPayload
	default:
		return "", nil, fmt.Errorf("source can be header or payload, got %q", locationParts[0])
	}

	if !strings.HasPrefix(locationParts[1], "/") {
		return "", nil, errors.New("fragment must start with a slash")
	}
	if locationParts[1] == "/" {
		return "", nil, errors.New("insufficient fragment, must not point to root")
	}

	locationPath := strings.Split(locationParts[1], "/")[1:]
	return structField, locationPath, nil
}
