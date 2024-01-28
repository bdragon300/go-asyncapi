package utils

import (
	"net/url"
	"strings"
)

func IsSubsequence[T comparable](subseq, iterable []T, searchIndex int) (int, bool) {
	if searchIndex+len(subseq) > len(iterable) {
		return 0, false
	}
	for ind, item := range subseq {
		if item != iterable[searchIndex+ind] {
			return ind, false
		}
	}
	return searchIndex + len(subseq), true
}

func SlicesEqual[T comparable](a, b []T) bool { // TODO: use slices.Compare
	if len(a) != len(b) {
		return false
	}
	_, ok := IsSubsequence(a, b, 0)
	return ok
}

func SplitRefToPathPointer(ref string) (specID, pointer string, remote bool) {
	ref = strings.TrimSpace(ref)
	if strings.HasPrefix(ref, "#/") {
		return "", ref[1:], false
	}

	specID = ref
	if u, err := url.Parse(ref); err == nil {
		pointer = u.Fragment
		u.Fragment = ""

		switch {
		case u.Scheme == "file" || u.Host == "" && u.User == nil && u.Scheme == "": // Ref points to a local file
			// Cut out the optional 'file://' scheme, assuming that the rest is a filename
			u.Scheme = ""
		default: // Ref points to a remote file by URL
			remote = true
		}
		specID = u.String()
	}
	return
}
