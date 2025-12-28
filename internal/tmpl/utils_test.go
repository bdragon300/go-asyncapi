package tmpl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQualifiedToImport(t *testing.T) {
	t.Run("Empty expression", func(t *testing.T) {
		assert.Panics(t, func() {
			qualifiedToImport([]string{})
		})
	})
	t.Run("With parameters", func(t *testing.T) {
		tests := []struct {
			title     string
			exprParts []string
			pkgPath   string
			pkgName   string
			name      string
		}{
			{
				title:     "Single parameter",
				exprParts: []string{"a"},
				pkgPath:   "a",
				pkgName:   "a",
				name:      "",
			},
			{
				title:     "Empty package",
				exprParts: []string{"", "a"},
				pkgPath:   "",
				pkgName:   "",
				name:      "a",
			},
			{
				title:     "Empty package and name",
				exprParts: []string{"", "a", "b"},
				pkgPath:   "a",
				pkgName:   "a",
				name:      "b",
			},
			{
				title:     "Single level package and name",
				exprParts: []string{"a", "x"},
				pkgPath:   "a",
				pkgName:   "a",
				name:      "x",
			},
			{
				title:     "Single level package and name with dot",
				exprParts: []string{"a.x"},
				pkgPath:   "a",
				pkgName:   "a",
				name:      "x",
			},
			{
				title:     "Package with slash without name",
				exprParts: []string{"a/b/c"},
				pkgPath:   "a/b/c",
				pkgName:   "c",
				name:      "",
			},
			{
				title:     "Package with slash with dot and name",
				exprParts: []string{"a/b.c", "x"},
				pkgPath:   "a/b.c",
				pkgName:   "b.c",
				name:      "x",
			},
			{
				title:     "Package with slash in several parts without name",
				exprParts: []string{"n", "d", "a/b.x"},
				pkgPath:   "n/d/a/b",
				pkgName:   "b",
				name:      "x",
			},
			{
				title:     "Package with slash in several parts without name",
				exprParts: []string{"n", "d", "a/b.c.x"},
				pkgPath:   "n/d/a/b.c",
				pkgName:   "b.c",
				name:      "x",
			},
			{
				title:     "Package with slash in several parts without name",
				exprParts: []string{"n", "d", "b.c.x"},
				pkgPath:   "n/d/b.c",
				pkgName:   "b.c",
				name:      "x",
			},
			{
				title:     "Package with slash in several parts and name",
				exprParts: []string{"n", "d", "a/b.c", "x"},
				pkgPath:   "n/d/a/b.c",
				pkgName:   "b.c",
				name:      "x",
			},
		}
		for _, tt := range tests {
			t.Run(tt.title, func(t *testing.T) {
				pkgPath, pkgName, name := qualifiedToImport(tt.exprParts)
				assert.Equal(t, tt.pkgPath, pkgPath)
				assert.Equal(t, tt.pkgName, pkgName)
				assert.Equal(t, tt.name, name)
			})
		}
	})
}
