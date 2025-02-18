package specurl

import (
	"net/url"
	"slices"
	"strings"

	"github.com/samber/lo"
)

func Parse(ref string) *URL {
	specFile, pointer, isRemote := parseRef(ref)
	pointerParts := strings.Split(pointer, "/")
	if len(pointerParts) > 0 && pointerParts[0] == "" {
		pointerParts = pointerParts[1:]
	}
	return &URL{
		SpecID:   specFile,
		Pointer:  pointerParts,
		isRemote: isRemote,
	}
}

type URL struct {
	SpecID   string   // URL to the spec file (schema, domain, path)
	Pointer  []string // URL fragment segments (escaped)
	isRemote bool     // Ref to a file on remote resource. When true, then IsExternal is also true
}

func (r URL) MatchPointer(unescapedPath []string) bool {
	escapedPath := lo.Map(unescapedPath, func(item string, _ int) string {
		return url.PathEscape(item)
	})
	return slices.Compare(r.Pointer, escapedPath) == 0
}

func (r URL) IsExternal() bool {
	return r.SpecID != ""
}

func (r URL) IsRemote() bool {
	return r.isRemote
}

func (r URL) String() (s string) {
	if len(r.Pointer) > 0 {
		s = r.PointerRef()
	}
	if r.IsExternal() {
		s = r.SpecID + s
	}
	return
}

func (r URL) PointerRef() string {
	return "#/" + strings.Join(r.Pointer, "/")
}

func parseRef(ref string) (specFile, pointer string, isRemote bool) {
	ref = strings.TrimSpace(ref)
	if strings.HasPrefix(ref, "#/") {
		return "", ref[1:], false
	}

	specFile = ref
	if u, err := url.Parse(ref); err == nil {
		pointer = u.Fragment
		u.Fragment = ""

		switch {
		case u.Scheme == "file" || u.Host == "" && u.User == nil && u.Scheme == "":
			// Ref points to a local file
			// Cut out the optional 'file://' scheme, assuming that the rest is a filename
			u.Scheme = ""
		default: // Ref points to a remote file by URL
			isRemote = true
		}
		specFile = u.String()
	}
	return
}

func BuildRef(parts ...string) string {
	escapedParts := lo.Map(parts, func(item string, _ int) string {
		return url.PathEscape(item)
	})
	return "#/" + strings.Join(escapedParts, "/")
}
