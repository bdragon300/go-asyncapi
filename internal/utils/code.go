package utils

import (
	"fmt"
	"strings"

	"github.com/dave/jennifer/jen"
)

func ToCode(in []*jen.Statement) []jen.Code {
	result := make([]jen.Code, len(in))
	for i, item := range in {
		result[i] = any(item).(jen.Code)
	}
	return result
}

func QualSprintf(format string, args ...any) jen.Code {
	res := &jen.Statement{}
	format = strings.ReplaceAll(format, "%Q(", "%%Q(")
	s := fmt.Sprintf(format, args...)

	// Expression: %Q(encoding/json,Marshal)
	blocks := strings.Split(s, "%Q(")
	if len(blocks) == 0 {
		return jen.Op("")
	}

	res = res.Add(jen.Op(blocks[0]))
	for _, p := range blocks[1:] {
		parts := strings.SplitN(p, ")", 2)
		params := strings.SplitN(parts[0], ",", 2)
		code := jen.Qual(params[0], params[1])
		if len(parts) > 1 {
			code = code.Op(parts[1])
		}
		res = res.Add(code)
	}

	return res
}
