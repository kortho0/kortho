package merkle

import (
	"crypto/sha256"
	"fmt"
	"testing"
)

func TestMerkle(t *testing.T) {
	tree := New(sha256.New(), [][]byte{[]byte("0"), []byte("1"), []byte("2"), []byte("3"), []byte("4"), []byte("5")})
	hash := tree.GetMtHash()
	fmt.Printf("hash = %x\n", hash)
	fmt.Printf("VerifyNode = %v\n", tree.VerifyNode([]byte("6")))
}
