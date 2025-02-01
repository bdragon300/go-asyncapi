package tmpl

import (
	"net/url"
	"path"
	"strconv"
	"strings"
)

func unescapeCorrelationIDPathItem(value string) (any, error) {
	if v, err := strconv.Atoi(value); err == nil {
		return v, nil // Number path items are treated as integers
	}

	// Unquote path item if it is quoted. Quoted forces a path item to be treated as a string, not number.
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

// qualifiedToImport converts the qual* template function parameters to qualified name and import package path.
// And also it returns the package name (the last part of the package path).
func qualifiedToImport(exprParts []string) (pkgPath string, pkgName string, name string) {
	// exprParts["a"] -> ["a", "a", ""]
	// exprParts["", "a"] -> ["", "", "a"]
	// exprParts["a.x"] -> ["a", "a", "x"]
	// exprParts["a/b/c"] -> ["a/b/c", "c", ""]
	// exprParts["a", "x"] -> ["a", "a", "x"]
	// exprParts["a/b.c", "x"] -> ["a/b.c", "bc", "x"]
	// exprParts["n", "d", "a/b.c", "x"] -> ["n/d/a/b.c-e", "b.c-e", "x"]
	switch len(exprParts) {
	case 0:
		panic("Empty parameters, at least one is required")
	case 1:
		pkgPath = exprParts[0]
	default:
		pkgPath = path.Join(exprParts[:len(exprParts)-1]...) + "." + exprParts[len(exprParts)-1]
	}
	// Split the whole expression into package path and name.
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
