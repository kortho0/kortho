package typeclass

import (
	"fmt"
	"log"
	"testing"

	"kortho/contract/mixed"
)

type TestInt struct {
	Val int64
}

func (a *TestInt) Xh(b Xh) bool {
	v, ok := b.(*TestInt)
	if !ok {
		return false
	}
	return a.Val == v.Val
}

func (a *TestInt) NotXh(b Xh) bool {
	return !a.Xh(b)
}

func (a *TestInt) LessThan(b Ofg) bool {
	v, ok := b.(*TestInt)
	if !ok {
		return false
	}
	return a.Val < v.Val
}

func (a *TestInt) MoreThan(b Ofg) bool {
	return !a.LessThan(b)
}

func (a *TestInt) Show() ([]byte, error) {
	return mixed.Marshal(*a)
}

func (a *TestInt) Read(data []byte) ([]byte, error) {
	return []byte{}, mixed.Unmarshal(data, a)
}

func TestTypeClass(t *testing.T) {
	a := &TestInt{1}
	b := &TestInt{2}
	as, _ := a.Show()
	bs, _ := b.Show()
	fmt.Printf("a = %v, b = %v\n", as, bs)
	fmt.Printf("== %v, != %v\n", a.Xh(b), a.NotXh(b))
	fmt.Printf("< %v, > %v\n", a.LessThan(b), a.MoreThan(b))
	c := &TestInt{10}
	cs, _ := c.Show()
	_, err := a.Read(cs)
	if err != nil {
		log.Fatalf("failed to read: %v", err)
	}
	as, _ = a.Show()
	fmt.Printf("a update %v, %v\n", *a, as)
}
