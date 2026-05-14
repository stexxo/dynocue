// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package util

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// DeepCopyStruct creates a deep copy of a struct, including nested slices.
func DeepCopyStruct[T any](input *T) *T {
	if input == nil {
		return nil
	}

	v := reflect.ValueOf(input)
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		// If not a struct, just return a copy of the value if possible,
		// but this function is intended for structs.
		res := *input
		return &res
	}

	clone := reflect.New(v.Type()).Elem()
	deepCopyValue(v, clone)

	res := clone.Addr().Interface().(*T)
	return res
}

func deepCopyValue(src, dst reflect.Value) {
	switch src.Kind() {
	case reflect.Slice:
		if src.IsNil() {
			return
		}
		dst.Set(reflect.MakeSlice(src.Type(), src.Len(), src.Cap()))
		for i := 0; i < src.Len(); i++ {
			deepCopyValue(src.Index(i), dst.Index(i))
		}
	case reflect.Struct:
		for i := 0; i < src.NumField(); i++ {
			if dst.Field(i).CanSet() {
				deepCopyValue(src.Field(i), dst.Field(i))
			}
		}
	case reflect.Ptr:
		if src.IsNil() {
			return
		}
		dst.Set(reflect.New(src.Type().Elem()))
		deepCopyValue(src.Elem(), dst.Elem())
	case reflect.Map:
		if src.IsNil() {
			return
		}
		dst.Set(reflect.MakeMap(src.Type()))
		for _, key := range src.MapKeys() {
			val := src.MapIndex(key)
			newVal := reflect.New(val.Type()).Elem()
			deepCopyValue(val, newVal)
			dst.SetMapIndex(key, newVal)
		}
	default:
		dst.Set(src)
	}
}

// UpdateStructByTag updates a field in a struct identified by a tag value.
// It supports pointers and pointers to pointers for the data argument.
func UpdateStructByTag(tag string, tagValue string, value interface{}, data interface{}) error {
	v := reflect.ValueOf(data)

	// Unpack pointers until we find something else
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return errors.New("data is a nil pointer")
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return errors.New("data is not a struct")
	}

	if !v.CanSet() {
		return errors.New("cannot set fields on data (ensure data was passed as a pointer)")
	}

	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tv := field.Tag.Get(tag)
		if tv == "" {
			continue
		}

		// Handle tag options like omitempty
		actualTagValue := strings.Split(tv, ",")[0]

		if actualTagValue == tagValue {
			structField := v.Field(i)
			if !structField.CanSet() {
				return errors.New("field is not settable (might be unexported)")
			}

			val := reflect.ValueOf(value)
			if !val.Type().AssignableTo(structField.Type()) {
				if val.Type().ConvertibleTo(structField.Type()) {
					val = val.Convert(structField.Type())
				} else {
					return fmt.Errorf("inputType mismatch: cannot set %s with %s", structField.Type().String(), val.Type().String())
				}
			}

			structField.Set(val)
			return nil
		}
	}

	return fmt.Errorf("field with tag %s=%q not found", tag, tagValue)
}
