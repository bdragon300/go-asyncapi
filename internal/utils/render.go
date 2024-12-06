package utils

import (
	"fmt"
	"strings"
)

//func ToCode(in []*jen.Statement) []jen.Code {
//	result := make([]jen.Code, len(in))
//	for i, item := range in {
//		result[i] = any(item).(jen.Code)
//	}
//	return result
//}
//
//func QualSprintf(format string, args ...any) jen.Code {
//	res := &jen.Statement{}
//	format = strings.ReplaceAll(format, "%Q(", "%%Q(")
//	s := fmt.Sprintf(format, args...)
//
//	// Expression: %Q(encoding/json,Marshal)
//	blocks := strings.Split(s, "%Q(")
//	if len(blocks) == 0 {
//		return jen.Op("")
//	}
//
//	res = res.Add(jen.Op(blocks[0]))
//	for _, p := range blocks[1:] {
//		parts := strings.SplitN(p, ")", 2)
//		params := strings.SplitN(parts[0], ",", 2)
//		code := jen.Qual(params[0], params[1])
//		if len(parts) > 1 {
//			code = code.Op(parts[1])
//		}
//		res = res.Add(code)
//	}
//
//	return res
//}

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
			// See:
			// https://github.com/dave/jennifer/issues/39
			// https://github.com/golang/go/issues/26363
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