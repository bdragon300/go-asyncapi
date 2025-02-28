package types

import (
	"encoding/json"
	"fmt"

	"github.com/samber/lo"

	"github.com/buger/jsonparser"
	"gopkg.in/yaml.v3"
)

// OrderedMap is a map that preserves the global order of keys.
// This is important for keeping the tool's result the same across runs.
// This map contains the unmarshaling logic for JSON and YAML. Not thread-safe.
type OrderedMap[K comparable, V any] struct {
	data map[K]V
	keys []K
}

func (o *OrderedMap[K, V]) UnmarshalJSON(bytes []byte) error {
	return jsonparser.ObjectEach(
		bytes,
		func(keyData []byte, valueData []byte, _ jsonparser.ValueType, _ int) error {
			var key K
			if err := json.Unmarshal(keyData, &key); err != nil {
				return err
			}

			var val V
			if err := json.Unmarshal(valueData, &val); err != nil {
				return err
			}

			o.Set(key, val)
			return nil
		},
	)
}

func (o *OrderedMap[K, V]) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.MappingNode {
		return fmt.Errorf("pipeline must contain YAML mapping, got %v", value.Kind)
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

		o.Set(key, val)
	}

	return nil
}

// Get returns the value for the given key or false if the key is not found.
func (o OrderedMap[K, V]) Get(key K) (V, bool) {
	v, ok := o.data[key]
	return v, ok
}

// MustGet returns the value for the given key or panics if the key is not found.
func (o OrderedMap[K, V]) MustGet(key K) V {
	v, ok := o.Get(key)
	if !ok {
		panic(fmt.Sprintf("key %v not found", key))
	}
	return v
}

// GetOrEmpty returns the value for the given key or zero value if the key is not found.
func (o OrderedMap[K, V]) GetOrEmpty(key K) (res V) {
	if v, ok := o.Get(key); ok {
		return v
	}
	return
}

// Has returns true if the key is found in the map.
func (o OrderedMap[K, V]) Has(key K) bool {
	_, ok := o.data[key]
	return ok
}

// Set sets the value for a given key. Places the key to the end of the order, even if it already exists.
func (o *OrderedMap[K, V]) Set(key K, value V) {
	if o.data == nil {
		o.data = make(map[K]V)
	}
	if _, ok := o.data[key]; ok {
		o.keys = lo.DropWhile(o.keys, func(item K) bool { return item == key })
	}
	o.data[key] = value
	o.keys = append(o.keys, key)
}

// Delete removes the key from the map and returns true if key was found.
func (o *OrderedMap[K, V]) Delete(key K) bool {
	if o.data == nil {
		return false
	}
	if _, ok := o.data[key]; !ok {
		return false
	}

	o.keys = lo.DropWhile(o.keys, func(item K) bool { return item == key })
	delete(o.data, key)
	return true
}

// Keys returns the keys of the map in the order they were added.
func (o OrderedMap[K, V]) Keys() []K {
	return o.keys
}

// Entries returns the entries of the map in the order the keys were added.
func (o OrderedMap[K, V]) Entries() []lo.Entry[K, V] {
	return lo.Map(o.keys, func(item K, _ int) lo.Entry[K, V] {
		return lo.Entry[K, V]{Key: item, Value: o.data[item]}
	})
}

// Len returns the number of items in the map.
func (o OrderedMap[K, V]) Len() int {
	return len(o.data)
}

func (o OrderedMap[K, V]) OrderedMap() {}
