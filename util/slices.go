package util

import (
	"cmp"
	"math"
	"slices"
)

type Number interface {
	Num() float64
	SetNum(float64)
}

type OrderedArray[T Number] struct {
	Data []T `json:"data"`
}

func (cl *OrderedArray[T]) Len() int {
	return len(cl.Data)
}

func (cl *OrderedArray[T]) Add(in T) bool {
	var index int

	if in.Num() == 0 {
		if cl.Len() == 0 {
			in.SetNum(1)
			index = 0
		} else {
			c := cl.Data[cl.Len()-1]
			in.SetNum(math.Floor(c.Num()) + 1)
			index = cl.Len()
		}
	} else {
		i, found := slices.BinarySearchFunc(cl.Data, in.Num(), func(a T, b float64) int {
			return cmp.Compare(a.Num(), b)
		})
		if found {
			return false
		}
		index = i
	}

	cl.Data = slices.Insert(cl.Data, index, in)
	return true
}

func (cl *OrderedArray[T]) Remove(number float64) {
	i, found := slices.BinarySearchFunc(cl.Data, number, func(a T, b float64) int {
		return cmp.Compare(a.Num(), b)
	})
	if !found {
		return
	}
	cl.Data = slices.Delete(cl.Data, i, i+1)
}

func (cl *OrderedArray[T]) Get(number float64) *T {
	i, _ := slices.BinarySearchFunc(cl.Data, number, func(a T, b float64) int {
		return cmp.Compare(a.Num(), b)
	})
	if i >= 0 && i < cl.Len() {
		return &cl.Data[i]
	}

	return nil
}
