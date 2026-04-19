// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cues

import (
	"cmp"
	"math"
	"slices"
)

type CueingModel struct {
	CueLists CueLists       `msgpack:"cueLists" json:"cueLists"`
	Settings CueingSettings `msgpack:"settings" json:"settings"`
}

type CueingSettings struct{}

type CueLists []*CueList

func (cl *CueLists) Len() int {
	return len(*cl)
}

func (cl *CueLists) create(in float64, cueListType string) *CueList {
	var index int
	var number float64

	if in == 0 {
		if len(*cl) == 0 {
			number = 0
			index = 0
		} else {
			c := (*cl)[cl.Len()-1]
			number = math.Floor(c.Number) + 1
			index = cl.Len()
		}
	} else {

		i, found := slices.BinarySearchFunc(*cl, number, func(a *CueList, b float64) int {
			return cmp.Compare(a.Number, b)
		})
		if found {
			return nil
		}
		index = i
		number = in
	}

	c := &CueList{Number: number, CueListType: cueListType}
	*cl = slices.Insert(*cl, index, c)
	return c
}

func (cl *CueLists) remove(number float64) {
	i, found := slices.BinarySearchFunc(*cl, number, func(a *CueList, b float64) int {
		return cmp.Compare(a.Number, b)
	})
	if !found {
		return
	}
	*cl = slices.Delete(*cl, i, i+1)
}

func (cl *CueLists) getByNumber(number float64) *CueList {
	i, _ := slices.BinarySearchFunc(*cl, number, func(a *CueList, b float64) int {
		return cmp.Compare(a.Number, b)
	})
	if i >= 0 && i < len(*cl) {
		return (*cl)[i]
	}

	return &CueList{}
}

type CueList struct {
	Settings    CueListSettings `msgpack:"settings" json:"settings"`
	Number      float64         `msgpack:"number" json:"number"`
	Label       string          `msgpack:"label" json:"label"`
	Cues        []string        `msgpack:"cues" json:"cues"`
	CueListType string          `msgpack:"cueListType" json:"cueListType"`
}

type CueListSettings struct {
}
