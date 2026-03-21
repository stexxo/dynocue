// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package utils

import (
	"fmt"
	"reflect"
	"strconv"
)

// SetFieldByTag updates a field in a struct based on its tag and value.
// It searches for a field with the given tagName and value, and sets it to newVal.
func SetFieldByTag(obj any, tagName string, tagValue string, newVal string) error {
	v := reflect.ValueOf(obj)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("obj must be a pointer to a struct")
	}

	v = v.Elem()
	t := v.Type()

	for i := range t.NumField() {
		f := t.Field(i)
		tag := f.Tag.Get(tagName)
		if tag == tagValue {
			fieldVal := v.Field(i)
			switch fieldVal.Kind() {
			case reflect.String:
				fieldVal.SetString(newVal)
				return nil
			case reflect.Float64:
				val, err := strconv.ParseFloat(newVal, 64)
				if err != nil {
					return fmt.Errorf("failed to parse float: %w", err)
				}
				fieldVal.SetFloat(val)
				return nil
			default:
				return fmt.Errorf("field type %s is not supported", fieldVal.Kind())
			}
		}
	}

	return fmt.Errorf("field with %s tag %s not found", tagName, tagValue)
}
