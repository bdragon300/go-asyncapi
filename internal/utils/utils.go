package utils

import (
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/bdragon300/asyncapi-codegen-go/internal/types"

	"github.com/samber/lo"
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

func SlicesEqual[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	_, ok := IsSubsequence(a, b, 0)
	return ok
}

func StructToOrderedMap(value any, target *types.OrderedMap[string, any], marshalFields []string) error {
	if target == nil {
		return errors.New("target is nil")
	}
	rval := reflect.ValueOf(value)
	if rval.Kind() != reflect.Struct {
		return fmt.Errorf("expected %v (Struct), got %v", reflect.Struct, rval.Kind())
	}

	rtyp := rval.Type()
	for i := 0; i < rval.NumField(); i++ {
		fld := rtyp.Field(i)
		if lo.Contains(marshalFields, fld.Name) {
			fldVal := rval.Field(i)
			if fldVal.IsValid() && !fldVal.IsZero() {
				if fldVal.Kind() == reflect.Pointer {
					fldVal = reflect.Indirect(fldVal)
				}
				target.Set(fld.Name, fldVal.Interface())
			}
		}
	}

	return nil
}

func SplitSpecPath(path string) (specID, pointer string) {
	path = strings.TrimSpace(path)
	if strings.HasPrefix(path, "#/") {
		return "", path[1:]
	}

	specID = path
	if u, err := url.Parse(path); err == nil {
		pointer = u.Fragment
		u.Fragment = ""

		switch {
		case u.Scheme == "file" || u.Host == "" && u.User == nil && u.Scheme == "": // Ref to a file on the local machine
			// Cut out the optional scheme, assuming that the rest is a filename
			u.Scheme = ""
			specID, _ = filepath.Abs(u.String())
		default: // Ref to a remote file
			specID = u.String()
		}
	}
	return
}

func IsRemoteSpecID(specID string) bool {
	return strings.HasPrefix(specID, "http://") || strings.HasPrefix(specID, "https://")
}
