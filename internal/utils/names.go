package utils

import (
	"crypto/md5"
	"encoding/base32"
	"fmt"
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

// initialisms are the commonly used acronyms inside identifiers, that should be written upper case according to the
// Go code style guide.
var initialisms = []string{
	// Got from https://github.com/golang/lint/blob/6edffad5e6160f5949cdefc81710b2706fbcd4f6/lint.go#L770
	"Acl", "Api", "Ascii", "Cpu", "Css", "Dns", "Eof", "Guid", "Html", "Http", "Https", "Id", "Ip", "Json", "Lhs",
	"Qps", "Ram", "Rhs", "Rpc", "Sla", "Smtp", "Sql", "Ssh", "Tcp", "Tls", "Ttl", "Udp", "Ui", "Uid", "Uuid", "Uri",
	"Url", "Utf8", "Vm", "Xml", "Xmpp", "Xsrf", "Xss",
	// Additional initialisms used in the generated code
	"Amqp", "Ip", "Mqtt",
}
var initialismsTrie = ahocorasick.NewTrieBuilder().AddStrings(initialisms).Build()

// ToGolangName converts any string to a valid Golang name. The exported argument determines if the result
// should be the exported name (start with an uppercase letter) or not.
//
// This function removes the invalid characters from source string, converts it to camel case or pascal case,
// makes initialisms uppercase, ensuring that the result doesn't conflict with Golang reserved keywords.
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

// transformInitialisms receives a string in camel case or pascal case and transforms all possible initialisms to upper case.
func transformInitialisms(s string) string {
	source := []byte(s)
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

// JoinNonemptyStrings joins non-empty strings with a separator.
func JoinNonemptyStrings(sep string, s ...string) string {
	return strings.Join(lo.Compact(s), sep)
}

// ToGoFilePath converts any path-looking string to the valid path to Go source file path and returns it.
//
// If the path is empty, the function returns a constant file name.
//
// While converting, the function shortens the path by eliminating the dot parts, removes invalid characters from
// every part, converting the rest to snake case.
// If an item contains invalid characters only, it is replaced by hash string of the original string to make it non-empty.
// The last part also will have ".go" file extension.
func ToGoFilePath(pathString string) string {
	if pathString == "" {
		hsh := md5.New()
		return fmt.Sprintf("empty%s.go", normalizePathItem(base32.StdEncoding.EncodeToString(hsh.Sum([]byte(pathString)))))
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

// normalizePathItem converts a string to a valid path item.
//
// The function removes invalid characters from the string, converts it to snake case. If string contains the
// invalid characters only, it is replaced by hash string of the original string to make it non-empty.
func normalizePathItem(s string) string {
	// Replace everything except alphanumerics to underscores
	newString := string(fileNameReplaceRe.ReplaceAll([]byte(s), []byte("_")))

	// Cut underscores that may appear at string endings
	newString = strings.Trim(newString, "_")

	// newString may become empty after normalization, because rawPath is empty or be "/", "___" strings.
	// In this case, the filename will be the md5 hash from original rawPath in base32 form
	if newString == "" {
		hsh := md5.New()
		newString = "empty" + normalizePathItem(base32.StdEncoding.EncodeToString(hsh.Sum([]byte(s))))
	}

	return lo.SnakeCase(newString)
}

// GetPackageName returns the last part of the directory path as a package name. If the directory is empty, the function
// returns "main".
func GetPackageName(directory string) string {
	directory = path.Clean(directory)
	_, pkgName := path.Split(directory)

	if pkgName == "" || pkgName == "." {
		return defaultPackage
	}
	return pkgName
}
