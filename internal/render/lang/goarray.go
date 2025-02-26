package lang

import (
	"fmt"

	"github.com/bdragon300/go-asyncapi/internal/common"
)

// GoArray is a type representing a Go array or slice.
type GoArray struct {
	BaseType
	// ItemsType is the type of the array items.
	ItemsType common.GolangType
	// If Size is set, then GoArray is a Go array with fixed-size. Otherwise, it is a slice.
	Size int
}

func (a *GoArray) String() string {
	if a.Import != "" {
		return fmt.Sprintf("GoArray(%s.%s)", a.Import, a.OriginalName)
	}
	return "GoArray(" + a.OriginalName + ")"
}

func (a *GoArray) GoTemplate() string {
	return "code/lang/goarray"
}
