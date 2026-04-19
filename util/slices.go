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

func (cl *OrderedArray[T]) Add(in float64) *T {
	var index int
	var number float64

	if in == 0 {
		if cl.Len() == 0 {
			number = 0
			index = 0
		} else {
			c := cl.Data[cl.Len()-1]
			number = math.Floor(c.Num()) + 1
			index = cl.Len()
		}
	} else {

		i, found := slices.BinarySearchFunc(cl.Data, number, func(a T, b float64) int {
			return cmp.Compare(a.Num(), b)
		})
		if found {
			return nil
		}
		index = i
		number = in
	}

	cPtr := new(T)
	(*cPtr).SetNum(number)
	cl.Data = slices.Insert(cl.Data, index, *cPtr)
	return cPtr
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
