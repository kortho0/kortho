package mixed

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func TestMiscellaneous(t *testing.T) {
	rand := rand.New(rand.NewSource(time.Now().Unix()))
	for i := 0; i < 1000000; i++ {
		a := rand.Uint32()
		b := E32func(a)
		c, _ := D32func(b)
		if a != c {
			fmt.Printf("%d: fail\n", i)
		}
	}
}
