package types

import (
	"encoding/json"
	"fmt"
	"iter"

	"github.com/samber/lo"

	"github.com/buger/jsonparser"
	"gopkg.in/yaml.v3"
)

// OrderedMap is a map that preserves the keys insertion order.
//
// This is a naive and inefficient implementation hastily crafted just for the project needs. It has the JSON and YAML
// (un)marshalling support and basic map operations. Not thread-safe.
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

func (o OrderedMap[K, V]) MarshalJSON() ([]byte, error) {
	buf := make([]byte, 0)
	buf = append(buf, '{')

	for i, key := range o.keys {
		keyBytes, err := json.Marshal(key)
		if err != nil {
			return nil, err
		}
		valBytes, err := json.Marshal(o.data[key])
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

	buf = append(buf, '}')
	return buf, nil
}

// IsZero returns true if the map is empty. Primarily use is to make the "omitzero" tag work in JSON/YAML serialization.
func (o OrderedMap[K, V]) IsZero() bool {
	return len(o.data) == 0
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

// Has returns true if the key is found in the map.
func (o OrderedMap[K, V]) Has(key K) bool {
	_, ok := o.data[key]
	return ok
}

// Set sets the value for a given key. Always places the entry to the last position, even if it already exists.
func (o *OrderedMap[K, V]) Set(key K, value V) {
	if o.data == nil {
		o.data = make(map[K]V)
	}
	if _, ok := o.data[key]; ok {
		o.keys = lo.Filter(o.keys, func(item K, _ int) bool { return item != key })
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

	o.keys = lo.Filter(o.keys, func(item K, _ int) bool { return item != key })
	delete(o.data, key)
	return true
}

// Keys returns the keys of the map in the order they were added.
func (o OrderedMap[K, V]) Keys() []K {
	return o.keys
}

// Entries returns an iterator over the map entries in insertion/modification order -- key-value pairs.
func (o OrderedMap[K, V]) Entries() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for _, key := range o.keys {
			if !yield(key, o.data[key]) {
				return
			}
		}
	}
}

// Map returns the underlying map. Note that the order of keys is not preserved in the returned map.
func (o OrderedMap[K, V]) Map() map[K]V {
	return o.data
}

// Len returns the number of items in the map.
func (o OrderedMap[K, V]) Len() int {
	return len(o.data)
}

func (o OrderedMap[K, V]) OrderedMap() {}
