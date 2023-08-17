package utils

import (
	"regexp"
	"strings"

	"github.com/stoewer/go-strcase"
)

func NormalizeGolangName(srcName string) string {
	if srcName == "" {
		return ""
	}

	// Remove everything except alphanumerics and '_'
	re := regexp.MustCompile("[^a-zA-Z0-9_]+")
	srcName = string(re.ReplaceAll([]byte(srcName), []byte("_")))

	// Cut extra "_" that may appear at string endings
	srcName = strings.Trim(srcName, "_")

	// Cut numbers from string start
	srcName = strings.TrimLeft(srcName, "1234567890")

	return strcase.UpperCamelCase(srcName)
}
