package utils

import (
	"crypto/md5"
	"encoding/base32"
	"go/token"
	"os"
	"path"
	"regexp"
	"strings"
	"unicode"

	ahocorasick "github.com/BobuSumisu/aho-corasick"

	"github.com/samber/lo"
)

const defaultPackage = "main"

var (
	golangTypeReplaceRe = regexp.MustCompile("[^a-zA-Z0-9_]+")
	fileNameReplaceRe   = regexp.MustCompile("[^a-zA-Z0-9_-]+")
)

// Initialisms are the commonly used acronyms inside identifiers, that code linters want they to be in upper case
var initialisms = []string{
	// Got from https://github.com/golang/lint/blob/6edffad5e6160f5949cdefc81710b2706fbcd4f6/lint.go#L770
	"Acl", "Api", "Ascii", "Cpu", "Css", "Dns", "Eof", "Guid", "Html", "Http", "Https", "Id", "Ip", "Json", "Lhs",
	"Qps", "Ram", "Rhs", "Rpc", "Sla", "Smtp", "Sql", "Ssh", "Tcp", "Tls", "Ttl", "Udp", "Ui", "Uid", "Uuid", "Uri",
	"Url", "Utf8", "Vm", "Xml", "Xmpp", "Xsrf", "Xss",
	// Additional initialisms used in the generated code
	"Amqp", "Ip", "Mqtt",
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

	var camel string
	if exported {
		camel = lo.PascalCase(rawString)
	} else {
		camel = lo.CamelCase(rawString)
	}

	str := transformInitialisms(camel)

	// Avoid conflict with Golang reserved keywords
	if token.IsKeyword(str) {
		return str + "_"
	}
	return str
}

// transformInitialisms transforms possible initialisms to upper case in a name in camel case or pascal case.
func transformInitialisms(name string) string {
	source := []byte(name)
	res := make([]byte, len(source))

	var last int64
	initialismsTrie.Walk(source, func(end, n, pattern int64) bool {
		// end: index of the last character of the matched pattern
		// n: length of the matched pattern
		// pattern: index of the matched pattern in the initialisms slice

		right := end + 1 // `end` is inclusive, so we need to add 1
		left := right - n
		// Write everything before the matched pattern.
		// `left` may point before `last` here on the repeated match on the same position. E.g. when "http" and "https"
		// initialisms found.
		if left > last {
			copy(res[last:], source[last:left])
		}

		// Determine if the matched pattern is a word and transform it to uppercase if it is.
		// For example, the transformation of string "httpsSmthId" gives "HTTPSSmthID" as a result of the following:
		// 1. First match is "http". It's not a word (next letter is lowercase), write without transformation.
		// 2. On next iteration the match is "https". It's a word (next letter is in uppercase). Transform it
		//    and write "HTTPS" over "http" from the previous iteration
		// 3. Final match is "id". It's a word (end of string). Transform it to "ID" and write over "id".
		if right == int64(len(source)) || unicode.IsUpper(rune(source[right])) {
			copy(res[left:], strings.ToUpper(initialisms[pattern]))
		} else {
			copy(res[left:], source[left:right])
		}
		last = right
		return true
	})
	copy(res[last:], source[last:])

	return string(res)
}

func JoinNonemptyStrings(sep string, s ...string) string {
	return strings.Join(lo.Compact(s), sep)
}

func ToGoFilePath(pathString string) string {
	if pathString == "" {
		hsh := md5.New()
		return "empty" + normalizePathItem(base32.StdEncoding.EncodeToString(hsh.Sum([]byte(pathString))))
	}

	directory, file := path.Split(path.Clean(pathString))
	if directory != "" {
		dirs := strings.Split(path.Clean(directory), string(os.PathSeparator))
		normDirs := lo.Map(dirs, func(s string, _ int) string {
			return normalizePathItem(s)
		})
		directory = path.Join(normDirs...)
	}
	fileName := strings.TrimSuffix(file, path.Ext(file))
	normFile := normalizePathItem(fileName)

	return path.Join(directory, normFile+".go")
}

func normalizePathItem(name string) string {
	// Replace everything except alphanumerics to underscores
	newString := string(fileNameReplaceRe.ReplaceAll([]byte(name), []byte("_")))

	// Cut underscores that may appear at string endings
	newString = strings.Trim(newString, "_")

	// newString may become empty after normalization, because rawPath is empty or be "/", "___" strings.
	// In this case, the filename will be the md5 hash from original rawPath in base32 form
	if newString == "" {
		hsh := md5.New()
		newString = "empty" + normalizePathItem(base32.StdEncoding.EncodeToString(hsh.Sum([]byte(name))))
	}

	return lo.SnakeCase(newString)
}

func GetPackageName(directory string) string {
	directory = path.Clean(directory)
	_, pkgName := path.Split(directory)

	if pkgName == "" || pkgName == "." {
		return defaultPackage
	}
	return pkgName
}
