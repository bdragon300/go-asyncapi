package types

import (
	"encoding/json"
	"fmt"
	"reflect"

	"gopkg.in/yaml.v3"
)

// Union2 is a union type that can hold one of two types. Typically, this is used for handling the JSON or YAML objects,
// that can be marshalled/unmarshalled to different types.
//
// The union is represented as two fields, V0 and V1. The Selector field can have a value 0 or 1, indicating which of
// the two fields, V0 or V1 are currently active.
type Union2[T0, T1 any] struct {
	V0       T0
	V1       T1
	Selector uint8
}

func (ju *Union2[T0, T1]) MarshalJSON() ([]byte, error) {
	return json.Marshal(ju.CurrentValue())
}

func (ju *Union2[T0, T1]) MarshalYAML() (any, error) {
	return ju.CurrentValue(), nil
}

func (ju *Union2[T0, T1]) UnmarshalJSON(bytes []byte) error {
	if err := json.Unmarshal(bytes, &ju.V0); err == nil {
		ju.Selector = 0
	} else if err = json.Unmarshal(bytes, &ju.V1); err == nil {
		ju.Selector = 1
	} else {
		return err
	}
	return nil
}

func (ju *Union2[T0, T1]) UnmarshalYAML(value *yaml.Node) error {
	// Treat nulls as "null" string
	for _, n := range value.Content {
		if n.ShortTag() == "!!null" {
			n.SetString("null")
		}
	}

	if err := value.Decode(&ju.V0); err == nil {
		ju.Selector = 0
	} else if err = value.Decode(&ju.V1); err == nil {
		ju.Selector = 1
	} else {
		return err
	}
	return nil
}

// CurrentValue returns the currently active value in the union.
func (ju *Union2[T0, T1]) CurrentValue() any {
	switch ju.Selector {
	case 0:
		return ju.V0
	case 1:
		return ju.V1
	default:
		panic(fmt.Sprintf("Selector can be 0 or 1 only, got %d", ju.Selector))
	}
}

// ToUnion2 converts a value to an Union2 object. If v can get converted to T0 and T1, it creates a new Union2 object
// with the Selector set to 0 or 1, depending on the type. If the value cannot be converted to any of the types,
// the function panics.
func ToUnion2[T0, T1 any](v any) *Union2[T0, T1] {
	val := reflect.ValueOf(v)
	zero0 := new(T0)
	zero1 := new(T1)
	if val.CanConvert(reflect.TypeOf(zero0).Elem()) {
		return &Union2[T0, T1]{V0: v.(T0), V1: *zero1, Selector: 0}
	}
	if val.CanConvert(reflect.TypeOf(zero1).Elem()) {
		return &Union2[T0, T1]{V0: *zero0, V1: v.(T1), Selector: 1}
	}
	panic(fmt.Sprintf("v is not convertable neither to type %T nor to type %T", zero0, zero1))
}

// UnmarshalRawMessageUnion2 is utility that unmarshals the Union2 that holds the json.RawMessage and yaml.Node types
// to the target object.
func UnmarshalRawMessageUnion2(union Union2[json.RawMessage, yaml.Node], target any) error {
	switch union.Selector {
	case 0:
		if err := json.Unmarshal(union.V0, target); err != nil {
			return err
		}
	case 1:
		if err := union.V1.Decode(target); err != nil {
			return err
		}
	}
	return nil
}

// Union3 is a union type that can hold one of three types. Typically, this is used for handling the JSON or YAML objects,
// that can be marshalled/unmarshalled to different types.
//
// The union is represented as three fields, V0, V1 and V2. The Selector field can have a value 0, 1 or 2, indicating
// which of the three fields, V0, V1 or V2 are currently active.
type Union3[T0, T1, T2 any] struct {
	V0       T0
	V1       T1
	V2       T2
	Selector uint8
}

func (ju *Union3[T0, T1, T2]) MarshalJSON() ([]byte, error) {
	return json.Marshal(ju.CurrentValue())
}

func (ju *Union3[T0, T1, T2]) MarshalYAML() (any, error) {
	return ju.CurrentValue(), nil
}

func (ju *Union3[T0, T1, T2]) UnmarshalJSON(bytes []byte) error {
	if err := json.Unmarshal(bytes, &ju.V0); err == nil {
		ju.Selector = 0
	} else if err = json.Unmarshal(bytes, &ju.V1); err == nil {
		ju.Selector = 1
	} else if err = json.Unmarshal(bytes, &ju.V2); err == nil {
		ju.Selector = 2
	} else {
		return err
	}
	return nil
}

func (ju *Union3[T0, T1, T2]) UnmarshalYAML(value *yaml.Node) error {
	// Treat nulls as "null" string
	for _, n := range value.Content {
		if n.ShortTag() == "!!null" {
			n.SetString("null")
		}
	}

	if err := value.Decode(&ju.V0); err == nil {
		ju.Selector = 0
	} else if err = value.Decode(&ju.V1); err == nil {
		ju.Selector = 1
	} else if err = value.Decode(&ju.V2); err == nil {
		ju.Selector = 2
	} else {
		return err
	}
	return nil
}

// CurrentValue returns the currently active value in the union.
func (ju *Union3[T0, T1, T2]) CurrentValue() any {
	switch ju.Selector {
	case 0:
		return ju.V0
	case 1:
		return ju.V1
	case 2:
		return ju.V2
	default:
		panic(fmt.Sprintf("Selector can be 0, 1 or 2 only, got %d", ju.Selector))
	}
}
