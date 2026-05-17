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
	SetValue(value *RawNode)
}

type yamlSlot struct {
	key   any
	value *RawNode

	keyComments, valueComments [3]string // (headComment, lineComment, footComment) for the key and value nodes respectively
	keyStyle, valueStyle       yaml.Style
}

func (y yamlSlot) Key() any {
	return y.key
}

func (y yamlSlot) Value() *RawNode {
	return y.value
}

func (y yamlSlot) SetValue(value *RawNode) {
	*y.value = *value
}

type plainSlot struct {
	key   any
	value *RawNode
}

func (j plainSlot) Key() any {
	return j.key
}

func (j plainSlot) Value() *RawNode {
	return j.value
}

func (j plainSlot) SetValue(value *RawNode) {
	*j.value = *value
}

func NewScalarRawNode(path []string, value any) *RawNode {
	return &RawNode{
		kind:        RawNodeKindScalar,
		path:        path,
		scalarValue: value,
	}
}

// RawNode is a variant data type, which can represent a scalar, array or object value unmarshalled from JSON or YAML.
// It marshals back to the original content preserving comments and key ordering in objects.
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
			r.slots[i].SetValue(value)
			return
		}
	}
	r.slots = append(r.slots, plainSlot{key: key, value: value})
}

// SetByPath sets the value at the given path, creating intermediate nodes if necessary.
func (r *RawNode) SetByPath(path []string, value *RawNode) {
	if len(path) == 0 {
		panic("path cannot be empty")
	}
	if r.kind != RawNodeKindObject && r.kind != RawNodeKindArray {
		panic("not an object or array node")
	}
	for _, sl := range r.slots {
		if sl.Key() == path[0] {
			sl.Value().SetByPath(path[1:], value)
			return
		}
	}
	newNode := &RawNode{kind: RawNodeKindObject, path: append(r.path, path[0])}
	newNode.SetByPath(path[1:], value)
	r.slots = append(r.slots, plainSlot{key: path[0], value: newNode})
}

// DeleteByPath deletes the node at the given path and returns true if the node was found and deleted, false otherwise.
func (r *RawNode) DeleteByPath(path []string) bool {
	if r.kind != RawNodeKindObject && r.kind != RawNodeKindArray {
		panic("not an object or array node")
	}
	if len(path) == 0 {
		return false
	}
	for i, sl := range r.slots {
		if sl.Key() == path[0] {
			if len(path) == 1 {
				r.slots = append(r.slots[:i], r.slots[i+1:]...)
				return true
			}
			return sl.Value().DeleteByPath(path[1:])
		}
	}
	return false
}

// CloneZero returns the copy of r with zero value, keeping path, kind and metainfo.
func (r RawNode) CloneZero() *RawNode {
	r.slots = nil
	r.scalarValue = nil
	return &r
}

// Len returns the number of entries in the object or array node. Panics if called on a scalar node.
func (r RawNode) Len() int {
	if r.kind != RawNodeKindObject && r.kind != RawNodeKindArray {
		panic("not an object or array node")
	}
	return len(r.slots)
}

// AsScalar converts r to a scalar value. Panics if called on an array or object node.
func (r RawNode) AsScalar() any {
	if r.kind != RawNodeKindScalar {
		panic("not a scalar node")
	}
	return r.scalarValue
}

// AsSlice converts r to a slice value, recursively converting all nested nodes to their corresponding Go types (scalar, slice or map).
// Panics if called on a scalar or object node.
func (r RawNode) AsSlice() []any {
	if r.kind != RawNodeKindArray {
		panic("not an array node")
	}
	var res []any
	for _, sl := range r.slots {
		switch sl.Value().kind {
		case RawNodeKindScalar:
			res = append(res, sl.Value().AsScalar())
		case RawNodeKindArray:
			res = append(res, sl.Value().AsSlice())
		case RawNodeKindObject:
			res = append(res, sl.Value().AsOrderedMap())
		}
	}
	return res
}

// AsOrderedMap converts r to an OrderedMap value, recursively converting all nested nodes to their corresponding Go types (scalar, slice or map).
// Panics if called on a scalar or array node.
func (r RawNode) AsOrderedMap() OrderedMap[any, any] {
	if r.kind != RawNodeKindObject {
		panic("not an object node")
	}
	var res OrderedMap[any, any]
	for _, sl := range r.slots {
		switch sl.Value().kind {
		case RawNodeKindScalar:
			res.Set(sl.Key(), sl.Value().AsScalar())
		case RawNodeKindArray:
			res.Set(sl.Key(), sl.Value().AsSlice())
		case RawNodeKindObject:
			res.Set(sl.Key(), sl.Value().AsOrderedMap())
		}
	}
	return res
}

// Equal checks if the r and n are equal, recursively comparing all nested nodes. It ignores comments and key ordering in maps.
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
			res.slots = append(res.slots, plainSlot{key: key, value: val})
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
			res.slots = append(res.slots, plainSlot{key: idx, value: val})
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
	switch r.kind {
	case RawNodeKindArray:
		return json.Marshal(r.AsSlice())
	case RawNodeKindObject:
		return json.Marshal(r.AsOrderedMap())
	case RawNodeKindScalar:
		return json.Marshal(r.AsScalar())
	}
	return nil, fmt.Errorf("unknown node kind: %s", r.kind)
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
				keyStyle:      keyNode.Style,
				valueStyle:    valueNode.Style,
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
				valueStyle:    itemNode.Style,
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
			if v, ok := sl.(yamlSlot); ok {
				keyNode.HeadComment, keyNode.LineComment, keyNode.FootComment = v.keyComments[0], v.keyComments[1], v.keyComments[2]
				valNode.HeadComment, valNode.LineComment, valNode.FootComment = v.valueComments[0], v.valueComments[1], v.valueComments[2]
				keyNode.Style, valNode.Style = v.keyStyle, v.valueStyle
			}

			n.Content = append(n.Content, keyNode, valNode)
		}
	case RawNodeKindArray:
		n.Kind = yaml.SequenceNode
		for _, sl := range r.slots {
			valNode := &yaml.Node{}
			if err := valNode.Encode(sl.Value()); err != nil {
				return nil, err
			}
			if v, ok := sl.(yamlSlot); ok {
				valNode.HeadComment, valNode.LineComment, valNode.FootComment = v.valueComments[0], v.valueComments[1], v.valueComments[2]
			}

			n.Content = append(n.Content, valNode)
		}
	case RawNodeKindScalar:
		if err := n.Encode(r.scalarValue); err != nil {
			return nil, err
		}
	}

	return n, nil
}
