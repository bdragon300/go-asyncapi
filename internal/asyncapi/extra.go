package asyncapi

import (
	"github.com/bdragon300/asyncapi-codegen-go/internal/common"
	"github.com/bdragon300/asyncapi-codegen-go/internal/render"
	"github.com/bdragon300/asyncapi-codegen-go/internal/types"
)

type xGoTypeHint struct {
	Kind    string `json:"kind" yaml:"kind"`
	Pointer bool   `json:"pointer" yaml:"pointer"`
}

type xGoTypeImportPackage struct {
	Package string `json:"package" yaml:"package"`
}

type xGoType struct {
	Type     string               `json:"type" yaml:"type"`
	Import   xGoTypeImportPackage `json:"import" yaml:"import"`
	Embedded bool                 `json:"embedded" yaml:"embedded"`
	Hint     xGoTypeHint          `json:"hint" yaml:"hint"`
}

func buildXGoType(xGoTypeValue *types.Union2[string, xGoType]) (golangType common.GolangType) {
	t := &render.GoSimple{}

	switch xGoTypeValue.Selector {
	case 0:
		t.Name = xGoTypeValue.V0
	case 1:
		t.Name = xGoTypeValue.V1.Type
		t.Package = xGoTypeValue.V1.Import.Package
		t.IsIface = xGoTypeValue.V1.Hint.Kind == "interface"

		if xGoTypeValue.V1.Hint.Pointer {
			return &render.GoPointer{Type: t}
		}
	}

	golangType = t
	return
}
