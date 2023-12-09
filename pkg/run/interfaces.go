package run

import (
	"encoding/json"
	"fmt"
)

type Headers map[string]any

func (h Headers) ToByteValues() map[string][]byte {
	res := make(map[string][]byte, len(h))
	for k, v := range h {
		switch tv := v.(type) {
		case []byte:
			res[k] = tv
		case string:
			res[k] = []byte(tv)
		default:
			b, err := json.Marshal(tv) // FIXME: use special util function for type conversion
			if err != nil {
				panic(fmt.Sprintf("Cannot marshal header value of type %T: %v", v, err))
			}
			res[k] = b
		}
	}

	return res
}

type Parameter interface {
	Name() string
	String() string
}
