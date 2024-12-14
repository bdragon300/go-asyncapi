package lang

import (
	"github.com/bdragon300/go-asyncapi/internal/common"
)

type GoTypeAlias struct {
	BaseType
	AliasedType common.GolangType
}

func (p *GoTypeAlias) GoTemplate() string {
	return "lang/gotypealias"
}

func (p *GoTypeAlias) UnwrapGolangType() (common.GolangType, bool) {
	if v, ok := p.AliasedType.(GolangTypeWrapperType); ok {
		return v.UnwrapGolangType()
	}
	return p.AliasedType, p.AliasedType != nil
}

func (p *GoTypeAlias) IsPointer() bool {
	return p.AliasedType.IsPointer()
}

func (p *GoTypeAlias) IsStruct() bool {
	if v, ok := any(p.AliasedType).(golangStructType); ok {
		return v.IsStruct()
	}
	return false
}

func (p *GoTypeAlias) String() string {
	if p.Import != "" {
		return "GoTypeAlias /" + p.Import + "." + p.OriginalName
	}
	return "GoTypeAlias " + p.OriginalName
}