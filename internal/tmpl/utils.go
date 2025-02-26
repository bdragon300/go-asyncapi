package tmpl

import (
	"github.com/samber/lo"
	"net/url"
	"path"
	"strconv"
	"strings"
)

// unescapeJSONPointerFragmentPart unescapes a path item in JSON Pointer. Returns the unescaped string if it is a string,
// or int if it is a number.
//
// JSON Pointer path part is encoded according the [RFC3986 Section 2] and [RFC6901 Section 3]. This function decodes
// the path part according to these specifications.
//
// Additionally, a number part can explicitly be quoted (double or single quotes) to force interpret them as a string.
// It helps to specify if this part addresses an array index or a numeric object key.
// For examples below, this function returns "123" as a string:
//
//  https://example.com/resource#/foo/"123"/bar
//  https://example.com/resource#/foo/'123'/bar
//
// Such quoting is *not recommended* as the common practice because it does not comply with the JSON Pointer specification,
// but may be used as a workaround for some rare cases.
//
// [RFC6901 Section 3]: https://tools.ietf.org/html/rfc6901#section-3)
// [RFC3986 Section 2]: https://tools.ietf.org/html/rfc3986#section-2
func unescapeJSONPointerFragmentPart(value string) (any, error) {
	// Number path items are treated as integers
	if v, err := strconv.Atoi(value); err == nil {
		return v, nil
	}

	// Unquote quoted numbers, which forces them to be treated as a string, not as a number.
	quoted := strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") ||
		strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")
	if quoted {
		value = value[1 : len(value)-1] // Unquote
	}

	// RFC3986 URL unescape
	value, err := url.PathUnescape(value)
	if err != nil {
		return nil, err
	}

	// RFC6901 JSON Pointer unescape: replace `~1` to `/` and `~0` to `~`
	return strings.ReplaceAll(strings.ReplaceAll(value, "~1", "/"), "~0", "~"), nil
}

// qualifiedToImport accepts the import expression and splits it into package path and imported name.
// Additionally, it returns the package name (the last part of the package path).
//
// This function accepts the import expression as a single string or a sequence of strings that are joined together.
// The last part if it is not a path, is considered as a name.
//
// Expression syntax is `path/to/package[.name]`.
//
// Examples:
//
//   qualifiedToImport("a") -> "a", "a", ""
//   qualifiedToImport("", "a") -> "", "", "a"
//   qualifiedToImport("a.x") -> "a", "a", "x"
//   qualifiedToImport("a/b/c") -> "a/b/c", "c", ""
//   qualifiedToImport("a", "x") -> "a", "a", "x"
//   qualifiedToImport("a/b.c", "x") -> "a/b.c", "b.c", "x"
//   qualifiedToImport("n", "d", "a/b.x") -> "n/d/a/b", "b", "x"
//   qualifiedToImport("n", "d", "a/b.c", "x") -> "n/d/a/b.c", "b.c", "x"
func qualifiedToImport(exprParts []string) (pkgPath string, pkgName string, name string) {
	switch len(exprParts) {
	case 0:
		panic("Empty parameters, at least one is required")
	case 1:
		pkgPath = exprParts[0]
	default:
		lastPart := exprParts[len(exprParts)-1]
		sep := lo.Ternary(strings.Contains(lastPart, "/") || strings.Contains(lastPart, "."), "/", ".")
		pkgPath = path.Join(exprParts[:len(exprParts)-1]...) + sep + lastPart
	}

	// Split the expression into package path and name.
	// The name is the sequence after the last dot (package path can contain dots in last part).
	if pos := strings.LastIndex(pkgPath, "."); pos >= 0 {
		name = pkgPath[pos+1:]
		pkgPath = pkgPath[:pos]
	}
	pkgName = pkgPath
	if pos := strings.LastIndex(pkgPath, "/"); pos >= 0 {
		pkgName = pkgPath[pos+1:]
	}
	return
}
