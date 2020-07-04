package types

import (
	"crypto/ed25519"
	"kortho/util"

	"errors"
)

const (
	AddrPrefix     = "kto"
	AddrPrefixSize = len(AddrPrefix)
	AddressSize    = 47
)

type Address [AddressSize]byte

func (a *Address) Bytes() []byte {
	return a[:]
}

func BytesToAddress(from []byte) (*Address, error) {
	var addr Address
	copy(addr[:], from[:])
	if !addr.Verify() {
		return nil, errors.New("This byte slice is not a address type")
	}
	return &addr, nil
}

func (a *Address) ToPublicKey() []byte {
	return util.Base58Decode(a.String()[AddrPrefixSize:])
}

func PublicKeyToAddress(publicKey []byte) string {
	pubStr := util.Base58EncodeToString(publicKey)
	return AddrPrefix + pubStr
}

func (a *Address) IsNil() bool {
	for _, char := range a {
		if char == '\x00' {
			return true
		}
	}
	return false
}

func (a *Address) String() string {
	return string(a[:])
}

func StringToAddress(str string) (*Address, error) {
	var addr Address
	copy(addr[:], []byte(str)[:])
	if !addr.Verify() {
		return nil, errors.New("This string is not a address type")
	}
	return &addr, nil
}

func (a *Address) Verify() bool {
	if a.String()[:AddrPrefixSize] != AddrPrefix {
		return false
	} else if len(a.ToPublicKey()) != ed25519.PublicKeySize {
		return false
	}
	return true
}
