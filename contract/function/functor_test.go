package functor

import (
	"errors"
	"fmt"
	"testing"
)

func mapTest(a interface{}) interface{} {
	v, ok := a.([]int)
	if !ok {
		return errors.New("unknown type")
	}
	j := len(v)
	r := make([]int, j, j)
	for i := 0; i < j; i++ {
		r[i] = v[i] * v[i]
	}
	return r
}

func foldlTest(a interface{}, b interface{}) interface{} {
	v, ok := a.([]int)
	if !ok {
		return errors.New("unknown type")
	}
	v1, ok := b.(int)
	if !ok {
		return errors.New("unknown type")
	}
	j := len(v)
	for i := 0; i < j; i++ {
		v1 += v[i]
	}
	return v1
}

func TestFunctor(t *testing.T) {
	numbers := []int{1, 2, 3}
	fmt.Printf("%v\n", Map(mapTest, numbers))
	fmt.Printf("%v\n", Foldl(foldlTest, numbers, 0))
	fmt.Printf("%v\n", Foldr(foldlTest, numbers, 0))
}
