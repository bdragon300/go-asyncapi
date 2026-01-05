package tmpl

import (
	"net/url"
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
//	https://example.com/resource#/foo/"123"/bar
//	https://example.com/resource#/foo/'123'/bar
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
