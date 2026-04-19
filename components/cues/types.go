// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cues

import (
	"cmp"
	"slices"
)

type CueingModel struct {
	CueLists CueLists       `json:"cueLists"`
	Settings CueingSettings `json:"settings"`
}

type CueingSettings struct{}

type CueLists []CueList

func (cl *CueLists) Len() int {
	return len(*cl)
}

func (cl *CueLists) add(c CueList) {
	i, _ := slices.BinarySearchFunc(*cl, c.Number, func(a CueList, b float64) int {
		return cmp.Compare(a.Number, b)
	})
	*cl = slices.Insert(*cl, i, c)
}

func (cl *CueLists) remove(c CueList) {
	i, _ := slices.BinarySearchFunc(*cl, c.Number, func(a CueList, b float64) int {
		return cmp.Compare(a.Number, b)
	})
	*cl = slices.Delete(*cl, i, i+1)
}

func (cl *CueLists) getByNumber(number float64) CueList {
	i, _ := slices.BinarySearchFunc(*cl, number, func(a CueList, b float64) int {
		return cmp.Compare(a.Number, b)
	})
	if i >= 0 && i < len(*cl) {
		return (*cl)[i]
	}
	return CueList{}
}

type CueList struct {
	Settings CueListSettings `json:"settings"`
	Number   float64         `json:"number"`
	Label    string          `json:"label"`
	Cues     []string        `json:"cues"`
}

type CueListSettings struct {
}
