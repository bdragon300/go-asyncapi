package utils

import (
	"regexp"
	"strings"

	"github.com/samber/lo"

	"github.com/stoewer/go-strcase"
)

func ToGolangName(srcName string, exported bool) string {
	if srcName == "" {
		return ""
	}

	// Remove everything except alphanumerics and '_'
	re := regexp.MustCompile("[^a-zA-Z0-9_]+")
	srcName = string(re.ReplaceAll([]byte(srcName), []byte("_")))

	// Cut extra "_" that may appear at string endings
	srcName = strings.Trim(srcName, "_")

	// Cut numbers from string start
	srcName = strings.TrimLeft(srcName, "1234567890")

	// TODO: detect Go builtin words and replace them
	if exported {
		return strcase.UpperCamelCase(srcName)
	}
	return strcase.LowerCamelCase(srcName)
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
