package utils

import (
	"encoding/json"
	"fmt"

	"github.com/buger/jsonparser"
	"gopkg.in/yaml.v3"
)

type OrderedMap[K comparable, V any] struct {
	data map[K]V
	keys []K
}

func (o *OrderedMap[K, V]) UnmarshalJSON(bytes []byte) error {
	if o.data == nil {
		o.data = make(map[K]V)
	}

	return jsonparser.ObjectEach(
		bytes,
		func(keyData []byte, valueData []byte, dataType jsonparser.ValueType, offset int) error {
			var key K
			if err := json.Unmarshal(keyData, &key); err != nil {
				return err
			}

			var val V
			if err := json.Unmarshal(valueData, &val); err != nil {
				return err
			}

			if _, ok := o.data[key]; !ok {
				o.keys = append(o.keys, key)
			}
			o.data[key] = val
			return nil
		},
	)
}

func (o *OrderedMap[K, V]) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.MappingNode {
		return fmt.Errorf("pipeline must contain YAML mapping, got %v", value.Kind)
	}

	if o.data == nil {
		o.data = make(map[K]V)
	}

	for index := 0; index < len(value.Content); index += 2 {
		var key K
		var val V

		if err := value.Content[index].Decode(&key); err != nil {
			return err
		}
		if err := value.Content[index+1].Decode(&val); err != nil {
			return err
		}

		if _, ok := o.data[key]; !ok {
			o.keys = append(o.keys, key)
		}
		o.data[key] = val
	}

	return nil
}

func (o OrderedMap[K, V]) Keys() []K {
	return o.keys
}

func (o OrderedMap[K, V]) MustGet(key K) V {
	return o.data[key]
}

func (o OrderedMap[K, V]) Get(key K) (V, bool) {
	v, ok := o.data[key]
	return v, ok
}

func (o OrderedMap[K, V]) Len() int {
	return len(o.data)
}

func (o OrderedMap[K, V]) OrderedMap() {
}

