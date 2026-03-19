// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package utils

import (
	"fmt"
	"reflect"
)

// SetFieldByTag updates a field in a struct based on its tag and value.
// It searches for a field with the given tagName and value, and sets it to newVal.
// Only string fields are supported for now as per current requirements.
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
			if fieldVal.Kind() == reflect.String {
				fieldVal.SetString(newVal)
				return nil
			}
			return fmt.Errorf("field %s is not a string", tagValue)
		}
	}

	return fmt.Errorf("field with %s tag %s not found", tagName, tagValue)
}
