package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"iter"
	"strconv"

	"github.com/buger/jsonparser"
	"gopkg.in/yaml.v3"
)

type RawNodeKind string

const (
	RawNodeKindScalar RawNodeKind = "scalar"
	RawNodeKindArray  RawNodeKind = "array"
	RawNodeKindObject RawNodeKind = "object"
)

type slot interface {
	Key() any
	Value() *RawNode
}

type yamlSlot struct {
	key   any
	value *RawNode

	keyComments, valueComments [3]string // (headComment, lineComment, footComment) for the key and value nodes respectively
}

func (y yamlSlot) Key() any {
	return y.key
}

func (y yamlSlot) Value() *RawNode {
	return y.value
}

type jsonSlot struct {
	key   any
	value *RawNode
}

func (j jsonSlot) Key() any {
	return j.key
}

func (j jsonSlot) Value() *RawNode {
	return j.value
}

func NewScalarRawNode(path []string, value any) *RawNode {
	return &RawNode{
		kind:        RawNodeKindScalar,
		path:        path,
		scalarValue: value,
	}
}

// RawNode is a variant data type, which can represent a scalar, array or object value unmarshalled from JSON or YAML.
// It marshals back to the original content preserving possible comments and key ordering in objects.
type RawNode struct {
	kind        RawNodeKind
	path        []string
	slots       []slot
	scalarValue any

	// Meta is a place to store any additional information related to this node.
	Meta any
}

func (r RawNode) Kind() RawNodeKind {
	return r.kind
}

func (r RawNode) Path() []string {
	return r.path
}

func (r RawNode) Entries() iter.Seq2[any, *RawNode] {
	if r.kind != RawNodeKindObject && r.kind != RawNodeKindArray {
		panic("not an object or array node")
	}
	return func(yield func(any, *RawNode) bool) {
		for _, sl := range r.slots {
			if !yield(sl.Key(), sl.Value()) {
				return
			}
		}
	}
}

func (r RawNode) Get(key any) (*RawNode, bool) {
	if r.kind != RawNodeKindObject && r.kind != RawNodeKindArray {
		panic("not an object or array node")
	}
	for _, sl := range r.slots {
		if sl.Key() == key {
			return sl.Value(), true
		}
	}
	return nil, false
}

func (r RawNode) Has(key any) bool {
	_, ok := r.Get(key)
	return ok
}

func (r *RawNode) Set(key any, value *RawNode) {
	if r.kind != RawNodeKindObject && r.kind != RawNodeKindArray {
		panic("not an object or array node")
	}
	for i, sl := range r.slots {
		if sl.Key() == key {
			r.slots[i] = jsonSlot{key: key, value: value}
			return
		}
	}
	r.slots = append(r.slots, jsonSlot{key: key, value: value})
}

func (r RawNode) Len() int {
	if r.kind != RawNodeKindObject && r.kind != RawNodeKindArray {
		panic("not an object or array node")
	}
	return len(r.slots)
}

func (r RawNode) AsScalar() any {
	if r.kind != RawNodeKindScalar {
		panic("not a scalar node")
	}
	return r.scalarValue
}

func (r RawNode) Equal(n *RawNode) bool {
	if n == nil || r.kind != n.kind {
		return false
	}
	switch r.kind {
	case RawNodeKindScalar:
		return r.scalarValue == n.scalarValue
	case RawNodeKindArray, RawNodeKindObject:
		if len(r.slots) != len(n.slots) {
			return false
		}
		for _, sl := range r.slots {
			nVal, ok := n.Get(sl.Key())
			if !ok || !sl.Value().Equal(nVal) {
				return false
			}
		}
		return true
	}
	return false
}

func (r *RawNode) UnmarshalJSON(data []byte) error {
	_, typ, _, err := jsonparser.Get(data)
	if err != nil {
		return fmt.Errorf("get type of root JSON value: %w", err)
	}
	res, err := r.unmarshalJSONValue(data, typ, nil)
	if err != nil {
		return err
	}

	*r = *res
	return err
}

func (r RawNode) unmarshalJSONValue(data []byte, valType jsonparser.ValueType, nodePath []string) (res *RawNode, err error) {
	res = &RawNode{path: nodePath}

	switch valType {
	case jsonparser.Object:
		res.kind = RawNodeKindObject
		err = jsonparser.ObjectEach(data, func(keyData []byte, valueData []byte, valueType jsonparser.ValueType, _ int) error {
			key := string(keyData)
			val, err := r.unmarshalJSONValue(valueData, valueType, append(nodePath, key))
			if err != nil {
				return err
			}
			res.slots = append(res.slots, jsonSlot{key: key, value: val})
			return nil
		})
	case jsonparser.Array:
		res.kind = RawNodeKindArray
		var innerErr error
		var idx int
		_, err = jsonparser.ArrayEach(data, func(d []byte, t jsonparser.ValueType, _ int, _ error) {
			val, err2 := r.unmarshalJSONValue(d, t, append(nodePath, strconv.Itoa(idx)))
			if err2 != nil {
				innerErr = err2 // The only way to deliver the error from the callback
				return
			}
			res.slots = append(res.slots, jsonSlot{key: idx, value: val})
			idx++
		})
		return res, errors.Join(err, innerErr)
	default:
		res.kind = RawNodeKindScalar
		err = json.Unmarshal(data, &res.scalarValue)
	}
	return
}

func (r RawNode) MarshalJSON() ([]byte, error) {
	var buf []byte

	switch r.kind {
	case RawNodeKindArray:
		buf = append(buf, '[')
	case RawNodeKindObject:
		buf = append(buf, '{')
	case RawNodeKindScalar:
		return json.Marshal(r.scalarValue)
	}

	for i, sl := range r.slots {
		keyBytes, err := json.Marshal(sl.Key())
		if err != nil {
			return nil, err
		}
		valBytes, err := json.Marshal(sl.Value())
		if err != nil {
			return nil, err
		}

		if i > 0 {
			buf = append(buf, ',')
		}
		buf = append(buf, keyBytes...)
		buf = append(buf, ':')
		buf = append(buf, valBytes...)
	}

	switch r.kind {
	case RawNodeKindArray:
		buf = append(buf, ']')
	case RawNodeKindObject:
		buf = append(buf, '}')
	}
	return buf, nil
}

func (r *RawNode) UnmarshalYAML(value *yaml.Node) error {
	res, err := r.unmarshalYAMLValue(value, nil)
	if err != nil {
		return err
	}

	*r = *res
	return nil
}

func (r RawNode) unmarshalYAMLValue(node *yaml.Node, nodePath []string) (res *RawNode, err error) {
	res = &RawNode{path: nodePath}

	switch node.Kind {
	case yaml.MappingNode:
		res.kind = RawNodeKindObject
		for i := 0; i < len(node.Content); i += 2 {
			keyNode, valueNode := node.Content[i], node.Content[i+1]

			key := keyNode.Value
			val, err2 := r.unmarshalYAMLValue(valueNode, append(nodePath, key))
			if err2 != nil {
				err = err2
				return
			}
			sl := yamlSlot{
				key:           key,
				value:         val,
				keyComments:   [3]string{keyNode.HeadComment, keyNode.LineComment, keyNode.FootComment},
				valueComments: [3]string{valueNode.HeadComment, valueNode.LineComment, valueNode.FootComment},
			}
			res.slots = append(res.slots, sl)
		}
	case yaml.SequenceNode:
		res.kind = RawNodeKindArray
		for i, itemNode := range node.Content {
			val, err2 := r.unmarshalYAMLValue(itemNode, append(nodePath, strconv.Itoa(i)))
			if err2 != nil {
				err = err2
				return
			}
			sl := yamlSlot{
				key:           i,
				value:         val,
				valueComments: [3]string{itemNode.HeadComment, itemNode.LineComment, itemNode.FootComment},
			}
			res.slots = append(res.slots, sl)
		}
	case yaml.AliasNode:
		return r.unmarshalYAMLValue(node.Alias, nodePath)
	case yaml.ScalarNode:
		res.kind = RawNodeKindScalar
		err = node.Decode(&res.scalarValue)
	}

	return
}

func (r RawNode) MarshalYAML() (any, error) {
	n := &yaml.Node{}

	switch r.kind {
	case RawNodeKindObject:
		n.Kind = yaml.MappingNode
		for _, sl := range r.slots {
			keyNode := &yaml.Node{}
			if err := keyNode.Encode(sl.Key()); err != nil {
				return nil, err
			}
			valNode := &yaml.Node{}
			if err := valNode.Encode(sl.Value()); err != nil {
				return nil, err
			}

			n.Content = append(n.Content, keyNode, valNode)

			if v, ok := sl.(yamlSlot); ok {
				keyNode.HeadComment, keyNode.LineComment, keyNode.FootComment = v.keyComments[0], v.keyComments[1], v.keyComments[2]
				valNode.HeadComment, valNode.LineComment, valNode.FootComment = v.valueComments[0], v.valueComments[1], v.valueComments[2]
			}
		}
	case RawNodeKindArray:
		n.Kind = yaml.SequenceNode
		for _, sl := range r.slots {
			valNode := &yaml.Node{}
			if err := valNode.Encode(sl.Value()); err != nil {
				return nil, err
			}

			n.Content = append(n.Content, valNode)

			if v, ok := sl.(yamlSlot); ok {
				valNode.HeadComment, valNode.LineComment, valNode.FootComment = v.valueComments[0], v.valueComments[1], v.valueComments[2]
			}
		}
	case RawNodeKindScalar:
		if err := n.Encode(r.scalarValue); err != nil {
			return nil, err
		}
	}

	return n, nil
}
