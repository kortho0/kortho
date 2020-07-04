package types

import (
	"crypto/ed25519"
	"crypto/rand"
	"kortho/util"
)

type Wallet struct {
	PrivateKey []byte
	Address    string
}

func NewWallet() *Wallet {
	publicKey, privateKey, _ := GenKeyPair()
	wallet := &Wallet{
		PrivateKey: privateKey,
		Address:    PublicKeyToAddress(publicKey),
	}
	return wallet
}

func GenKeyPair() ([]byte, []byte, error) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	return []byte(publicKey), []byte(privateKey), nil
}

func AddressToPublicKey(address string) []byte {
	return util.Base58Decode(address[AddrPrefixSize:])
}
