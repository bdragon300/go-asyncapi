package utils

import (
	"regexp"
	"strings"

	"github.com/samber/lo"

	"github.com/stoewer/go-strcase"
)

var (
	golangTypeReplaceRe = regexp.MustCompile("[^a-zA-Z0-9_]+")
	fileNameReplaceRe   = regexp.MustCompile("[^a-zA-Z0-9_-]+")
)

func ToGolangName(rawString string, exported bool) string {
	if rawString == "" {
		return ""
	}

	// Remove everything except alphanumerics and '_'
	rawString = string(golangTypeReplaceRe.ReplaceAll([]byte(rawString), []byte("_")))

	// Cut extra "_" that may appear at string endings
	rawString = strings.Trim(rawString, "_")

	// Cut numbers from string start
	rawString = strings.TrimLeft(rawString, "1234567890")

	// TODO: detect Go builtin words and replace them
	// TODO: transform words such as ID, URL, etc., to upper case
	if exported {
		return strcase.UpperCamelCase(rawString)
	}
	return strcase.LowerCamelCase(rawString)
}

func ToLowerFirstLetter(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToLower(string(s[0])) + s[1:]
}

func JoinNonemptyStrings(sep string, s ...string) string {
	s = lo.Filter(s, func(item string, _ int) bool { return item != "" })
	return strings.Join(s, sep)
}

func ToFileName(rawString string) string {
	if rawString == "" {
		return ""
	}

	// Remove everything except alphanumerics and '/'
	rawString = string(fileNameReplaceRe.ReplaceAll([]byte(rawString), []byte("_")))

	// Cut extra "/" that may appear at string endings
	rawString = strings.Trim(rawString, "_")

	return strcase.SnakeCase(rawString)
	// return strings.Join(lo.Map(parts, func(item string, index int) string {
	//	return strings.ToLower(item)
	// }), "_")
}
