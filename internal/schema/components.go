package schema

import (
	"github.com/bdragon300/asyncapi-codegen/internal/utils"
)

type ComponentsItem struct {
	Schemas  utils.OrderedMap[string, Object]  `json:"schemas" yaml:"schemas" cgen:"noinline,packageDown=models"`
	Messages utils.OrderedMap[string, Message] `json:"messages" yaml:"messages" cgen:"noinline,packageDown=message"`
}

