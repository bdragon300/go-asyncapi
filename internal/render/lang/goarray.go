package lang

import (
	"github.com/bdragon300/go-asyncapi/internal/common"
)

type GoArray struct {
	BaseType
	ItemsType common.GolangType
	Size      int
}

func (a *GoArray) GoTemplate() string {
	return "code/lang/goarray"
}

func (a *GoArray) String() string {
	if a.Import != "" {
		return "GoArray /" + a.Import + "." + a.OriginalName
	}
	return "GoArray " + a.OriginalName
}
