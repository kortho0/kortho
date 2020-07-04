package util

import (
	"crypto/rand"

	"golang.org/x/crypto/ed25519"
)

func GenPrivKey() ([]byte, []byte, error) {
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	return priv[32:], []byte(priv), nil
}
func PubtoAddr(pub []byte) string {

	addr := Encode(pub)
	return "Kto" + addr
}
