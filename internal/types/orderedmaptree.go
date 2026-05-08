package types

import (
	"encoding/json"

	"github.com/buger/jsonparser"
	"gopkg.in/yaml.v3"
)

// OrderedMapTree is OrderedMap, that also unmarshals all nested objects into OrderedMapTree tree.
// Also, it preserves the raw contents on the first level to be marshalled without losing the original formatting and comments.
type OrderedMapTree struct {
	OrderedMap[string, any]
	rawYAML *yaml.Node
	rawJSON []byte
}

func (r *OrderedMapTree) UnmarshalJSON(data []byte) error {
	res, err := r.unmarshalJSONValue(data, jsonparser.Object)
	if err != nil {
		return err
	}
	*r = res.(OrderedMapTree)
	r.rawJSON = data
	return err
}

func (r *OrderedMapTree) unmarshalJSONValue(data []byte, valType jsonparser.ValueType) (res any, err error) {
	switch valType {
	case jsonparser.Object:
		var m OrderedMapTree
		err = jsonparser.ObjectEach(data, func(keyData []byte, valueData []byte, valueType jsonparser.ValueType, _ int) error {
			key := string(keyData)
			val, err := r.unmarshalJSONValue(valueData, valueType)
			if err != nil {
				return err
			}
			m.Set(key, val)
			return nil
		})
		res = m
	case jsonparser.Array:
		var arr []any
		_, err = jsonparser.ArrayEach(data, func(d []byte, t jsonparser.ValueType, _ int, _ error) {
			val, err2 := r.unmarshalJSONValue(d, t)
			if err2 != nil {
				err = err2 // No other way to deliver the error from the callback
				return
			}
			arr = append(arr, val)
		})
		res = arr
	default:
		err = json.Unmarshal(data, &res)
	}
	return
}

func (r OrderedMapTree) MarshalJSON() ([]byte, error) {
	if r.rawJSON != nil {
		return r.rawJSON, nil
	}
	return r.OrderedMap.MarshalJSON()
}

func (r *OrderedMapTree) UnmarshalYAML(value *yaml.Node) error {
	res, err := r.unmarshalYAMLValue(value, yaml.MappingNode)
	if err != nil {
		return err
	}
	*r = res.(OrderedMapTree)
	r.rawYAML = value
	return nil
}

func (r *OrderedMapTree) unmarshalYAMLValue(node *yaml.Node, valType yaml.Kind) (res any, err error) {
	switch valType {
	case yaml.MappingNode:
		var m OrderedMapTree
		for i := 0; i < len(node.Content); i += 2 {
			keyNode, valueNode := node.Content[i], node.Content[i+1]

			key := keyNode.Value
			val, err2 := r.unmarshalYAMLValue(valueNode, valueNode.Kind)
			if err2 != nil {
				return nil, err2
			}
			m.Set(key, val)
		}
		res = m
	case yaml.SequenceNode:
		var arr []any
		for _, itemNode := range node.Content {
			val, err2 := r.unmarshalYAMLValue(itemNode, itemNode.Kind)
			if err2 != nil {
				return nil, err2
			}
			arr = append(arr, val)
		}
		res = arr
	case yaml.AliasNode:
		res, err = r.unmarshalYAMLValue(node.Alias, node.Alias.Kind)
	default:
		err = node.Decode(&res)
	}
	return
}

func (r OrderedMapTree) MarshalYAML() (any, error) {
	if r.rawYAML != nil {
		return r.rawYAML, nil
	}
	return r.OrderedMap.MarshalYAML()
}
