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
					return fmt.Errorf("type mismatch: cannot set %s with %s", structField.Type().String(), val.Type().String())
				}
			}

			structField.Set(val)
			return nil
		}
	}

	return fmt.Errorf("field with tag %s=%q not found", tag, tagValue)
}
