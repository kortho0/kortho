package mixed

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"hash"
)

type uint8ToByteFunc (func(uint8) []byte)
type byteToUint8Func (func([]byte) (uint8, error))

type uint16ToByteFunc (func(uint16) []byte)
type byteToUint16Func (func([]byte) (uint16, error))

type uint32ToByteFunc (func(uint32) []byte)
type byteToUint32Func (func([]byte) (uint32, error))

type uint64ToByteFunc (func(uint64) []byte)
type byteToUint64Func (func([]byte) (uint64, error))

var E8func uint8ToByteFunc
var D8func byteToUint8Func
var E16func uint16ToByteFunc
var D16func byteToUint16Func
var E32func uint32ToByteFunc
var D32func byteToUint32Func
var E64func uint64ToByteFunc
var D64func byteToUint64Func

func Bytes2HexString(data []byte) string {
	return hex.EncodeToString(data)
}

func HexString2Bytes(s string) ([]byte, error) {
	return hex.DecodeString(s)
}

func IsEmptySclie(a []byte) bool {
	for _, v := range a {
		if v != 0 {
			return false
		}
	}
	return true
}

func Dup(a []byte) []byte {
	b := make([]byte, len(a))
	copy(b, a)
	return b
}

func Memmove(a, b []byte) bool {
	aLen, bLen := len(a), len(b)
	if aLen < bLen {
		return false
	}
	for i := 0; i < aLen; i++ {
		if i < bLen {
			a[i] = b[i]
		} else {
			a[i] = 0
		}
	}
	return true
}

func Marshal(a interface{}) ([]byte, error) {
	return json.Marshal(a)
}

func Unmarshal(data []byte, a interface{}) error {
	return json.Unmarshal(data, a)
}

func GenHash(h hash.Hash, data []byte) ([]byte, error) {
	h.Reset()
	h.Write(data)
	return h.Sum(nil), nil
}

func init() {
	buf := uint32ToBytes(1)
	E8func = uint8ToBytes
	D8func = bytesToUint8
	if buf[0] == 0 {
		E16func = bigEndianUint16ToBytes
		D16func = bigEndianBytesToUint16
		E32func = bigEndianUint32ToBytes
		D32func = bigEndianBytesToUint32
		E64func = bigEndianUint64ToBytes
		D64func = bigEndianBytesToUint64
	} else {
		E16func = uint16ToBytes
		D16func = bytesToUint16
		E32func = uint32ToBytes
		D32func = bytesToUint32
		E64func = uint64ToBytes
		D64func = bytesToUint64
	}
}

func bigEndianUint16ToBytes(a uint16) []byte {
	buf := make([]byte, 2)
	buf[1] = byte(a & 0xFF)
	buf[0] = byte((a >> 8) & 0xFF)
	return buf
}

func bigEndianBytesToUint16(a []byte) (uint16, error) {
	if len(a) != 2 {
		return 0, errors.New("bigEndianBytesToUint16: Illegal slice length")
	}
	b := uint16(0)
	for i, v := range a {
		b += uint16(v)
		if i > 0 {
			b <<= 8
		}
	}
	return b, nil
}

func bigEndianUint32ToBytes(a uint32) []byte {
	buf := make([]byte, 4)
	buf[3] = byte(a & 0xFF)
	buf[2] = byte((a >> 8) & 0xFF)
	buf[1] = byte((a >> 16) & 0xFF)
	buf[0] = byte((a >> 24) & 0xFF)
	return buf
}

func bigEndianBytesToUint32(a []byte) (uint32, error) {
	if len(a) != 4 {
		return 0, errors.New("bigEndianBytesToUint32: Illegal slice length")
	}
	b := uint32(0)
	for i, v := range a {
		b += uint32(v)
		if i > 0 {
			b <<= 8
		}
	}
	return b, nil
}

func bigEndianUint64ToBytes(a uint64) []byte {
	buf := make([]byte, 8)
	buf[7] = byte(a & 0xFF)
	buf[6] = byte((a >> 8) & 0xFF)
	buf[5] = byte((a >> 16) & 0xFF)
	buf[4] = byte((a >> 24) & 0xFF)
	buf[3] = byte((a >> 32) & 0xFF)
	buf[2] = byte((a >> 40) & 0xFF)
	buf[1] = byte((a >> 48) & 0xFF)
	buf[0] = byte((a >> 56) & 0xFF)
	return buf
}

func bigEndianBytesToUint64(a []byte) (uint64, error) {
	if len(a) != 8 {
		return 0, errors.New("bigEndianBytesToUint64: Illegal slice length")
	}
	b := uint64(0)
	for i, v := range a {
		b += uint64(v)
		if i > 0 {
			b <<= 8
		}
	}
	return b, nil
}

func uint8ToBytes(a uint8) []byte {
	buf := make([]byte, 1)
	buf[0] = a & 0xFF
	return buf
}

func bytesToUint8(a []byte) (uint8, error) {
	if len(a) != 1 {
		return 0, errors.New("BytesToUint8: Illegal slice length")
	}
	return uint8(a[0]), nil
}

func uint16ToBytes(a uint16) []byte {
	buf := make([]byte, 2)
	buf[0] = byte(a & 0xFF)
	buf[1] = byte((a >> 8) & 0xFF)
	return buf
}

func bytesToUint16(a []byte) (uint16, error) {
	if len(a) != 2 {
		return 0, errors.New("BytesToUint16: Illegal slice length")
	}
	b := uint16(0)
	for i, v := range a {
		b += uint16(v) << (8 * uint16(i))
	}
	return b, nil
}

func uint32ToBytes(a uint32) []byte {
	buf := make([]byte, 4)
	buf[0] = byte(a & 0xFF)
	buf[1] = byte((a >> 8) & 0xFF)
	buf[2] = byte((a >> 16) & 0xFF)
	buf[3] = byte((a >> 24) & 0xFF)
	return buf
}

func bytesToUint32(a []byte) (uint32, error) {
	if len(a) != 4 {
		return 0, errors.New("bytesToUint32: Illegal slice length")
	}
	b := uint32(0)
	for i, v := range a {
		b += uint32(v) << (8 * uint32(i))
	}
	return b, nil
}

func uint64ToBytes(a uint64) []byte {
	buf := make([]byte, 8)
	buf[0] = byte(a & 0xFF)
	buf[1] = byte((a >> 8) & 0xFF)
	buf[2] = byte((a >> 16) & 0xFF)
	buf[3] = byte((a >> 24) & 0xFF)
	buf[4] = byte((a >> 32) & 0xFF)
	buf[5] = byte((a >> 40) & 0xFF)
	buf[6] = byte((a >> 48) & 0xFF)
	buf[7] = byte((a >> 56) & 0xFF)
	return buf
}

func bytesToUint64(a []byte) (uint64, error) {
	if len(a) != 8 {
		return 0, errors.New("bytesToUint64: Illegal slice length")
	}
	b := uint64(0)
	for i, v := range a {
		b += uint64(v) << (8 * uint64(i))
	}
	return b, nil
}
