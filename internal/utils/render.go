package utils

import (
	"fmt"
	"strings"
)

func ToGoLiteral(val any) string {
	var res string
	switch val.(type) {
	case bool, string, int, complex128:
		// default constant types can be left bare
		return fmt.Sprintf("%#v", val)
	case float64:
		res = fmt.Sprintf("%#v", val)
		if !strings.Contains(res, ".") && !strings.Contains(res, "e") {
			// If the formatted value is not in scientific notation, and does not have a dot, then
			// we add ".0". Otherwise, it will be interpreted as an int.
			// See: https://github.com/golang/go/issues/26363
			res += ".0"
		}
		return res
	case float32, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr:
		// other built-in types need specific type info
		return fmt.Sprintf("%T(%#v)", val, val)
	case complex64:
		// fmt package already renders parenthesis for complex64
		return fmt.Sprintf("%T%#v", val, val)
	}

	panic(fmt.Sprintf("unsupported type for literal: %T", val))
}