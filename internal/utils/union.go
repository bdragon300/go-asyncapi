package utils

import (
	"encoding/json"
	"fmt"
	"reflect"

	"gopkg.in/yaml.v3"
)

type Union2[T0, T1 any] struct {
	V0       T0
	V1       T1
	Selector uint8
}

func (ju *Union2[T0, T1]) MarshalJSON() ([]byte, error) {
	switch ju.Selector {
	case 0:
		return json.Marshal(ju.V0)
	case 1:
		return json.Marshal(ju.V1)
	default:
		panic(fmt.Sprintf("Selector can be 0 or 1 only, got %d", ju.Selector))
	}
}

func (ju *Union2[T0, T1]) MarshalYAML() (any, error) {
	switch ju.Selector {
	case 0:
		return ju.V0, nil
	case 1:
		return ju.V1, nil
	default:
		panic(fmt.Sprintf("Selector can be 0 or 1 only, got %d", ju.Selector))
	}
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
	if err := value.Decode(&ju.V0); err == nil {
		ju.Selector = 0
	} else if err = value.Decode(&ju.V1); err == nil {
		ju.Selector = 1
	} else {
		return err
	}
	return nil
}

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
	panic(fmt.Sprintf("v is not convertable neither to %T nor to %T", zero0, zero1))
}

func UnmarhalRawsUnion2(union Union2[json.RawMessage, yaml.Node], target any) error {
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

type Union3[T0, T1, T2 any] struct {
	V0       T0
	V1       T1
	V2       T2
	Selector uint8
}

func (ju *Union3[T0, T1, T2]) MarshalJSON() ([]byte, error) {
	switch ju.Selector {
	case 0:
		return json.Marshal(ju.V0)
	case 1:
		return json.Marshal(ju.V1)
	case 2:
		return json.Marshal(ju.V2)
	default:
		panic(fmt.Sprintf("Selector can be 0, 1 or 2 only, got %d", ju.Selector))
	}
}

func (ju *Union3[T0, T1, T2]) MarshalYAML() (any, error) {
	switch ju.Selector {
	case 0:
		return ju.V0, nil
	case 1:
		return ju.V1, nil
	case 2:
		return ju.V2, nil
	default:
		panic(fmt.Sprintf("Selector can be 0, 1 or 2 only, got %d", ju.Selector))
	}
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
