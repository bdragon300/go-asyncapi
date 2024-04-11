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
// Got from https://github.com/golang/lint/blob/6edffad5e6160f5949cdefc81710b2706fbcd4f6/lint.go#L770
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
	res := make([]byte, len(camel))
	var last int64
	initialismsTrie.Walk(camel, func(end, n, pattern int64) bool {
		// end: index of the last character of the matched pattern
		// n: length of the matched pattern
		// pattern: index of the matched pattern in the initialisms slice

		right := end + 1 // `end` is inclusive, so we need to add 1
		left := right - n
		// Write everything before the matched pattern
		// `left` may point before `last` here on the repeated match on the same position. E.g. when "http" and "https"
		// initialisms found.
		if left > last {
			copy(res[last:], camel[last:left])
		}

		// Transform only the whole word, not a part of other word
		// For example, "httpsSmthId":
		// 1. First matches "http", it's written without transform since it's not a whole word
		// 2. On next iteration matches "https" as whole word (next letter is in uppercase), transforms it
		//    and writes "HTTPS" over "http" from the previous iteration
		// 3. Matches "id" as whole word (end of string), transforms it to "ID"
		if right == int64(len(camel)) || unicode.IsUpper(rune(camel[right])) {
			copy(res[left:], strings.ToUpper(initialisms[pattern]))
		} else {
			copy(res[left:], camel[left:right])
		}
		last = right
		return true
	})
	copy(res[last:], camel[last:])
	str := string(res)

	// Avoid conflict with Golang reserved keywords
	if token.IsKeyword(str) {
		return str + "_"
	}
	return str
}

func ToLowerFirstLetter(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToLower(string(s[0])) + s[1:]
}

func JoinNonemptyStrings(sep string, s ...string) string {
	return strings.Join(lo.Compact(s), sep)
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
