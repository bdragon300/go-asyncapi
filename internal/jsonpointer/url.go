package jsonpointer

import (
	"fmt"
	"net/url"
	"slices"
	"strings"

	"github.com/samber/lo"
)

func Parse(uri string) (*JSONPointer, error) {
	parsedURI, fsPath, pointer := parse(uri)
	ptrParts := strings.Split(pointer, "/")
	if len(ptrParts) > 0 && ptrParts[0] == "" {
		ptrParts = ptrParts[1:] // Cut the first empty part appeared after splitting
	}
	for i := 0; i < len(ptrParts); i++ {
		var err error
		if ptrParts[i], err = url.PathUnescape(ptrParts[i]); err != nil {
			return nil, fmt.Errorf("url fragment unescape %q: %w", ptrParts[i], err)
		}
	}

	return &JSONPointer{
		URI:     parsedURI,
		FSPath:  fsPath,
		Pointer: ptrParts,
	}, nil
}

// JSONPointer represents the parsed JSON Pointer expression [JSON Pointer IETF Draft], with additional support of
// the filesystem paths in location.
//
// The JSON Pointer expression format is:
//
//	[location][#/path]
//
// The “location” can be a URI or filesystem path. “Path” part is optional. Here are some JSON Pointer examples:
//
//	http://example.com/myfile?format=json#/path/to/field
//	/home/user/myfile.json#/path/to/field
//	#/path/to/field
//	https://example.com/schemas/myfile.json
//	file:///home/user/myfile.json
//
// One exception is that URI's starting with schema “file://” are treated as local filesystem paths.
//
// Typically, JSONPointer is used for representation of $ref expressions, read from the document, but may be used as
// usual URL.
//
// [JSON Pointer IETF Draft]: https://datatracker.ietf.org/doc/html/draft-pbryan-zyp-json-ref-03
type JSONPointer struct {
	// URI is a location as parsed url.URL object if the location is a valid URI.
	URI *url.URL
	// FSPath is a location as string if the location is a filesystem path.
	FSPath string
	// Pointer is a list of URL-unescaped JSON pointer parts. Basically, it is everything after ``#/'' split by ``/''
	Pointer []string
}

// MatchPointer returns true if pointers are equal without considering the location. Receives the URL-unescaped pointer.
func (r JSONPointer) MatchPointer(unescapedPointer []string) bool {
	escapedPointer := lo.Map(unescapedPointer, func(item string, _ int) string {
		return url.PathEscape(item)
	})
	return slices.Compare(r.Pointer, escapedPointer) == 0
}

// Location returns string representation of the location. If the location is a URI, returns the URI string. Otherwise,
// returns the filesystem path (or empty string if no location is set).
func (r JSONPointer) Location() string {
	if r.URI != nil {
		return r.URI.String()
	}
	return r.FSPath
}

// PointerString returns the url-escaped JSON Pointer string representation without the location. E.g. “#/path/to/field”.
func (r JSONPointer) PointerString() string {
	return PointerString(r.Pointer...)
}

func (r JSONPointer) String() (s string) {
	if len(r.Pointer) > 0 {
		s = r.PointerString()
	}
	if l := r.Location(); l != "" {
		s = l + s
	}
	return
}

// PointerString returns the JSON Pointer string representation from the given parts. E.g. “#/path/to/field”.
func PointerString(parts ...string) string {
	escapedParts := lo.Map(parts, func(item string, _ int) string {
		return url.PathEscape(item)
	})
	return "#/" + strings.Join(escapedParts, "/")
}

func parse(ptr string) (uri *url.URL, fsPath, pointer string) {
	ptr = strings.TrimSpace(ptr)
	if strings.HasPrefix(ptr, "#/") {
		return nil, "", ptr[1:]
	}

	if u, err := url.Parse(ptr); err == nil {
		pointer = u.Fragment
		u.Fragment = ""

		// ``file://'' scheme or URI with path only are treated as local filesystem paths
		if u.Scheme == "file" || u.Host == "" && u.User == nil && u.Scheme == "" {
			return nil, u.Path, pointer
		}

		// ptr points to a remote file URI
		return u, "", pointer
	}

	return nil, ptr, ""
}
