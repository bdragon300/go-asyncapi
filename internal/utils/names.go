package utils

import (
	"go/token"
	"regexp"
	"strings"
	"unicode"

	ahocorasick "github.com/BobuSumisu/aho-corasick"

	"github.com/samber/lo"

	"github.com/stoewer/go-strcase"
)

var (
	golangTypeReplaceRe = regexp.MustCompile("[^a-zA-Z0-9_]+")
	fileNameReplaceRe   = regexp.MustCompile("[^a-zA-Z0-9_-]+")
)

// Initialisms are the commonly used acronyms inside identifiers, that code linters want they to be in upper case
var initialisms = []string{
	"Acl", "Api", "Ascii", "Cpu", "Css", "Dns", "Eof", "Guid", "Html", "Http", "Https", "Id", "Ip", "Json", "Lhs",
	"Qps", "Ram", "Rhs", "Rpc", "Sla", "Smtp", "Sql", "Ssh", "Tcp", "Tls", "Ttl", "Udp", "Ui", "Uid", "Uuid", "Uri",
	"Url", "Utf8", "Vm", "Xml", "Xmpp", "Xsrf", "Xss",
}
var initialismsTrie = ahocorasick.NewTrieBuilder().AddStrings(initialisms).Build()

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

	var camel []byte
	if exported {
		camel = []byte(strcase.UpperCamelCase(rawString))
	} else {
		camel = []byte(strcase.LowerCamelCase(rawString))
	}

	// Transform possible initialisms to upper case if they appear in string
	var bld strings.Builder
	initialismsTrie.Walk(camel, func(end, n, pattern int64) bool {
		right := end + 1 // `end` is inclusive, so we need to add 1
		bld.Write(camel[bld.Len() : right-n])
		if right == int64(len(camel)) || unicode.IsUpper(rune(camel[right])) {
			bld.WriteString(strings.ToUpper(initialisms[pattern]))
		} else {
			bld.Write(camel[right-n : right])
		}
		return true
	})
	if bld.Len() < len(camel) {
		bld.Write(camel[bld.Len():])
	}
	res := bld.String()

	// Avoid to conflict with Golang reserved keywords
	if token.IsKeyword(res) {
		return res + "_"
	}
	return res
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
}
