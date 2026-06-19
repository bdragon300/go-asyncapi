package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"iter"
	"path"
	"path/filepath"
	"slices"
	"strconv"

	"github.com/bdragon300/go-asyncapi/internal/jsonpointer"
	"github.com/buger/jsonparser"
	"github.com/samber/lo"
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

func (y *yamlSlot) SetValue(value *RawNode) {
	y.value = value
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

func (j *plainSlot) SetValue(value *RawNode) {
	j.value = value
}

func NewEmptyRawNode(kind RawNodeKind, nodePath []string, originDocument *jsonpointer.JSONPointer) *RawNode {
	if originDocument.FSPath != "" && !path.IsAbs(originDocument.FSPath) {
		panic(fmt.Errorf("originDocument should be an absolute path, got %q", originDocument.FSPath))
	}
	if len(originDocument.Pointer) > 0 {
		panic(fmt.Errorf("originDocument should point to the root of the document, got %q", originDocument.PointerString()))
	}
	return &RawNode{originDocument: originDocument, kind: kind, path: nodePath}
}

func NewScalarRawNode(nodePath []string, value any, originDocument *jsonpointer.JSONPointer) *RawNode {
	if originDocument.FSPath != "" && !path.IsAbs(originDocument.FSPath) {
		panic(fmt.Errorf("originDocument should be an absolute path, got %q", originDocument.FSPath))
	}
	if len(originDocument.Pointer) > 0 {
		panic(fmt.Errorf("originDocument should point to the root of the document, got %q", originDocument.PointerString()))
	}
	return &RawNode{
		kind:           RawNodeKindScalar,
		path:           nodePath,
		scalarValue:    value,
		originDocument: originDocument,
	}
}

// RawNode is a variant data type, which can represent a scalar, array or object value unmarshalled from JSON or YAML.
// It marshals back to the original content preserving comments and key ordering in objects.
type RawNode struct {
	kind        RawNodeKind
	path        []string
	slots       []slot
	scalarValue any

	// originDocument is a document *absolute* location where this node is initially was parsed from.
	originDocument *jsonpointer.JSONPointer
}

func (r RawNode) Kind() RawNodeKind {
	return r.kind
}

func (r RawNode) Path() []string {
	return r.path
}

func (r RawNode) AbsOriginDocumentPath() *jsonpointer.JSONPointer {
	return r.originDocument
}

// IsZero returns true if r is an empty, which means zero value for scalar nodes and zero length for array and object nodes.
func (r RawNode) IsZero() bool {
	switch r.kind {
	case RawNodeKindScalar:
		return r.scalarValue == nil
	case RawNodeKindArray, RawNodeKindObject:
		return r.Len() == 0
	default:
		panic("unknown node kind: " + string(r.kind))
	}
}

// DeepCopy creates a deep copy of r, recursively cloning all nested nodes.
// It preserves the original node's kind, path and origin document.
func (r RawNode) DeepCopy() *RawNode {
	if r.IsZero() {
		return NewEmptyRawNode(r.kind, r.path, r.originDocument)
	}
	switch r.kind {
	case RawNodeKindScalar:
		return NewScalarRawNode(r.path, r.scalarValue, r.originDocument)
	case RawNodeKindArray, RawNodeKindObject:
		res := NewEmptyRawNode(r.kind, r.path, r.originDocument)
		for _, sl := range r.slots {
			res.slots = append(res.slots, &plainSlot{key: sl.Key(), value: sl.Value().DeepCopy()})
		}
		return res
	default:
		panic(fmt.Sprintf("unknown node kind: %s", r.kind))
	}
}

// AbsPointerString returns the JSON pointer string representation of the node's path with the document's absolute
// location related to the current directory.
func (r RawNode) AbsPointerString() string {
	if r.originDocument == nil {
		return jsonpointer.PointerString(r.path...)
	}
	if r.originDocument.FSPath != "" {
		return lo.Must(filepath.Abs(r.originDocument.FSPath)) + jsonpointer.PointerString(r.path...)
	}
	return r.originDocument.Location() + jsonpointer.PointerString(r.path...)
}

func (r RawNode) Entries() iter.Seq2[any, *RawNode] {
	return func(yield func(any, *RawNode) bool) {
		if r.kind != RawNodeKindObject && r.kind != RawNodeKindArray {
			return
		}
		for _, sl := range r.slots {
			if !yield(sl.Key(), sl.Value()) {
				return
			}
		}
	}
}

func (r RawNode) Get(key any) (*RawNode, bool) {
	if r.kind != RawNodeKindObject && r.kind != RawNodeKindArray {
		return nil, false
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
	r.slots = append(r.slots, &plainSlot{key: key, value: value})
}

func (r *RawNode) SetAtStart(key any, value *RawNode) {
	if r.kind != RawNodeKindObject && r.kind != RawNodeKindArray {
		panic("not an object or array node")
	}
	for i, sl := range r.slots {
		if sl.Key() == key {
			r.slots[i].SetValue(value)
			return
		}
	}
	r.slots = append([]slot{&plainSlot{key: key, value: value}}, r.slots...)
}

func (r *RawNode) Append(value *RawNode) {
	if r.kind != RawNodeKindArray {
		panic("not an array node")
	}
	r.slots = append(r.slots, &plainSlot{key: len(r.slots), value: value})
}

// GetByPath returns the node at the given path, or nil if the path does not exist. Panics if called on a scalar node.
func (r *RawNode) GetByPath(path []string) *RawNode {
	if len(path) == 0 {
		return r
	}
	if r.kind != RawNodeKindObject && r.kind != RawNodeKindArray {
		panic("not an object or array node")
	}
	for _, sl := range r.slots {
		k := sl.Key()
		if r.kind == RawNodeKindArray {
			k = fmt.Sprintf("%d", k)
		}
		if k == path[0] {
			return sl.Value().GetByPath(path[1:])
		}
	}
	return nil
}

// SetNodeByPath inserts the given node in r at the same path or replaces the existing one.
// If the path does not exist, it is created. If r or any existing node on the path is not an object, an error is returned.
func (r *RawNode) SetNodeByPath(node *RawNode) error {
	if r.kind != RawNodeKindObject {
		return fmt.Errorf("not an object or array node at path %q", jsonpointer.PointerString(r.path...))
	}
	if len(node.path) == 0 {
		*r = *node
		return nil
	}

	return r.setNodeByPath(node.path, node)
}

func (r *RawNode) setNodeByPath(path []string, value *RawNode) error {
	if r.kind != RawNodeKindObject {
		return fmt.Errorf("not an object node at path %q", jsonpointer.PointerString(r.path...))
	}

	var node *RawNode
	sl, found := lo.Find(r.slots, func(sl slot) bool {
		k := sl.Key()
		if r.kind == RawNodeKindArray {
			k = fmt.Sprintf("%v", k)
		}
		return k == path[0]
	})
	if found {
		node = sl.Value()
	} else {
		node = &RawNode{
			kind:           RawNodeKindObject,
			path:           append(r.path, path[0]),
			originDocument: r.originDocument,
		}
		sl = &plainSlot{key: path[0], value: node}
		r.slots = append(r.slots, sl)
	}
	if len(path) <= 1 {
		sl.SetValue(value)
		return nil
	}

	if node.kind != RawNodeKindObject {
		return fmt.Errorf("node is not an object on path %s", jsonpointer.PointerString(path[:len(path)-1]...))
	}
	return node.setNodeByPath(path[1:], value)
}

// DeleteByPath deletes the node at the given path and returns true if the node was found and deleted, false otherwise.
func (r *RawNode) DeleteByPath(path []string) (bool, error) {
	if r.kind != RawNodeKindObject && r.kind != RawNodeKindArray {
		return false, fmt.Errorf("not an object or array node at path %q", jsonpointer.PointerString(r.path...))
	}
	if len(path) == 0 {
		return false, fmt.Errorf("empty path")
	}

	for i, sl := range r.slots {
		k := sl.Key()
		if r.kind == RawNodeKindArray {
			k = fmt.Sprintf("%d", k)
		}
		if k == path[0] {
			if len(path) == 1 {
				r.slots = append(r.slots[:i], r.slots[i+1:]...)
				return true, nil
			}
			return sl.Value().DeleteByPath(path[1:])
		}
	}
	return false, nil
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

// asSlice converts r to a slice value, recursively converting all nested nodes to their corresponding Go types (scalar, slice or map).
// Panics if called on a scalar or object node.
func (r RawNode) asSlice() []any {
	if r.kind != RawNodeKindArray {
		panic("not an array node")
	}
	var res []any
	for _, sl := range r.slots {
		switch sl.Value().kind {
		case RawNodeKindScalar:
			res = append(res, sl.Value().AsScalar())
		case RawNodeKindArray:
			res = append(res, sl.Value().asSlice())
		case RawNodeKindObject:
			res = append(res, sl.Value().asOrderedMap())
		}
	}
	return res
}

// asOrderedMap converts r to an OrderedMap value, recursively converting all nested nodes to their corresponding Go
// types (scalar, slice or map).
// Panics if called on a scalar or array node.
func (r RawNode) asOrderedMap() OrderedMap[any, any] {
	if r.kind != RawNodeKindObject {
		panic("not an object node")
	}
	var res OrderedMap[any, any]
	for _, sl := range r.slots {
		switch sl.Value().kind {
		case RawNodeKindScalar:
			res.Set(sl.Key(), sl.Value().AsScalar())
		case RawNodeKindArray:
			res.Set(sl.Key(), sl.Value().asSlice())
		case RawNodeKindObject:
			res.Set(sl.Key(), sl.Value().asOrderedMap())
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
	nodePath = slices.Clone(nodePath)
	res = &RawNode{path: nodePath, originDocument: r.originDocument}

	switch valType {
	case jsonparser.Object:
		res.kind = RawNodeKindObject
		err = jsonparser.ObjectEach(data, func(keyData []byte, valueData []byte, valueType jsonparser.ValueType, _ int) error {
			key := string(keyData)
			val, err := r.unmarshalJSONValue(valueData, valueType, append(nodePath, key))
			if err != nil {
				return err
			}
			res.slots = append(res.slots, &plainSlot{key: key, value: val})
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
			res.slots = append(res.slots, &plainSlot{key: idx, value: val})
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
		return json.Marshal(r.asSlice())
	case RawNodeKindObject:
		return json.Marshal(r.asOrderedMap())
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
	nodePath = slices.Clone(nodePath)
	res = &RawNode{path: nodePath, originDocument: r.originDocument}

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
			res.slots = append(res.slots, &sl)
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
			res.slots = append(res.slots, &sl)
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
			if v, ok := sl.(*yamlSlot); ok {
				keyNode.HeadComment, keyNode.LineComment, keyNode.FootComment = v.keyComments[0], v.keyComments[1], v.keyComments[2]
				valNode.HeadComment, valNode.LineComment, valNode.FootComment = v.valueComments[0], v.valueComments[1], v.valueComments[2]
				keyNode.Style, valNode.Style = v.keyStyle, v.valueStyle

				// YAML supports writing the values in JSON-like syntax, this is called "flow style".
				// Empty mappings and arrays are always rendered in flow style in YAML as `{}` and `[]`.
				// And once we unmarshalled an empty mapping/array and added new values to it, it's better to revert the block style.
				// So, we reset the flow style bit here.
				keyNode.Style &= ^yaml.FlowStyle
				valNode.Style &= ^yaml.FlowStyle
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
			if v, ok := sl.(*yamlSlot); ok {
				valNode.HeadComment, valNode.LineComment, valNode.FootComment = v.valueComments[0], v.valueComments[1], v.valueComments[2]
				valNode.Style = v.valueStyle

				// YAML supports writing the values in JSON-like syntax, this is called "flow style".
				// Empty mappings and arrays are always rendered in flow style in YAML as `{}` and `[]`.
				// And once we unmarshalled an empty mapping/array and added new values to it, it's better to revert the block style.
				// So, we reset the flow style bit here.
				valNode.Style &= ^yaml.FlowStyle
			}

			n.Content = append(n.Content, valNode)
		}
	case RawNodeKindScalar:
		if err := n.Encode(r.scalarValue); err != nil {
			return nil, err
		}
	}

	n.Style &= ^yaml.FlowStyle
	if n.Kind == yaml.ScalarNode && n.Tag == "!!str" {
		// Force quote strings
		n.Style |= yaml.SingleQuotedStyle
	}
	return n, nil
}
