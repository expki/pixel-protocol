package config

import (
	"encoding/json"
)

// SingleOrSlice allows for a configuration field to be either a single value or a slice of values.
type SingleOrSlice[T any] []T

// UnmarshalJSON handles both single values and slices for the field.
func (s *SingleOrSlice[T]) UnmarshalJSON(data []byte) error {
	var single T
	if err := json.Unmarshal(data, &single); err == nil {
		*s = SingleOrSlice[T]{single}
		return nil
	}
	var slice []T
	if err := json.Unmarshal(data, &slice); err != nil {
		return err
	}
	*s = slice
	return nil
}

// MarshalJSON ensures that the field is marshaled correctly whether it's a single value or a slice.
func (s SingleOrSlice[T]) MarshalJSON() ([]byte, error) {
	if len(s) == 1 {
		return json.Marshal(s[0])
	}
	return json.Marshal([]T(s))
}
