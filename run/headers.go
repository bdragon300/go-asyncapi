package run

import (
	"encoding/json"
	"fmt"
)

// Headers is the key-value pairs that keeps the message headers.
//
// The key can be any string, value is any value that can be represented as a byte sequence.
type Headers map[string]any

// ToByteValues converts the headers to the map of string keys and byte values.
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
