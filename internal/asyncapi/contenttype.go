package asyncapi

import (
	"github.com/samber/lo"
	"strings"
	"unicode"
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
