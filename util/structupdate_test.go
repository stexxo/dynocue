// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package util

import (
	"testing"
)

type TestStruct struct {
	Name  string  `json:"name"`
	Age   int     `json:"age,omitempty"`
	Score float64 `msgpack:"score"`
}

func TestUpdateStructByTag(t *testing.T) {
	s := TestStruct{Name: "Alice", Age: 30, Score: 95.5}

	t.Run("Update String Field", func(t *testing.T) {
		err := UpdateStructByTag("json", "name", "Bob", &s)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if s.Name != "Bob" {
			t.Errorf("expected Name to be Bob, got %s", s.Name)
		}
	})

	t.Run("Update Int Field with omitempty", func(t *testing.T) {
		err := UpdateStructByTag("json", "age", 35, &s)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if s.Age != 35 {
			t.Errorf("expected Age to be 35, got %d", s.Age)
		}
	})

	t.Run("Update with different tag", func(t *testing.T) {
		err := UpdateStructByTag("msgpack", "score", 99.9, &s)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if s.Score != 99.9 {
			t.Errorf("expected Score to be 99.9, got %f", s.Score)
		}
	})

	t.Run("Pointer to Pointer", func(t *testing.T) {
		ps := &s
		err := UpdateStructByTag("json", "name", "Charlie", &ps)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if s.Name != "Charlie" {
			t.Errorf("expected Name to be Charlie, got %s", s.Name)
		}
	})

	t.Run("Compatible Types (float64 to float32)", func(t *testing.T) {
		type CompatibleStruct struct {
			Ratio float32 `json:"ratio"`
		}
		cs := CompatibleStruct{Ratio: 0.5}
		err := UpdateStructByTag("json", "ratio", 0.75, &cs) // 0.75 is float64 by default
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if cs.Ratio != 0.75 {
			t.Errorf("expected Ratio to be 0.75, got %f", cs.Ratio)
		}
	})

	t.Run("Type Mismatch", func(t *testing.T) {
		err := UpdateStructByTag("json", "age", "forty", &s)
		if err == nil {
			t.Error("expected error due to type mismatch, got nil")
		}
	})

	t.Run("Not a Struct", func(t *testing.T) {
		x := 10
		err := UpdateStructByTag("json", "name", "Bob", &x)
		if err == nil {
			t.Error("expected error because data is not a struct, got nil")
		}
	})

	t.Run("Field Not Found", func(t *testing.T) {
		err := UpdateStructByTag("json", "missing", "value", &s)
		if err == nil {
			t.Error("expected error because field was not found, got nil")
		}
	})

	t.Run("Non-pointer data", func(t *testing.T) {
		err := UpdateStructByTag("json", "name", "Dave", s)
		if err == nil {
			t.Error("expected error because data is not a pointer, got nil")
		}
	})
}
