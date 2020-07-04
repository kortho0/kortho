package motor

import (
	"errors"
	"math/big"

	"kortho/contract/mixed"
)

const (
	UNDEFINED = iota
	CONSTANT
	RAM_VAR
	FLASH_VAR
)

const (
	MAP = iota
	BOOL
	CHAR
	INT8
	INT16
	INT32
	INT64
	UINT8
	UINT16
	UINT32
	UINT64
	STRING
	FLOAT32
	FLOAT64
	CONST_INT
	CONST_CHAR
	CONST_BOOL
	CONST_FLOAT
	CONST_STRING
)

const (
	DATA_HEADER_SIZE      = 36
	DEFAULT_STRING_LENGTH = 10
)

type FateData interface {
	Update(Fate, interface{}) error
}

type contractDataHeader struct {
	typ     uint32
	length  uint64
	offset0 uint64
	offset1 uint64
	offset2 uint64
}

type contractData struct {
	address uint64 // header所在的偏移地址
	header  *contractDataHeader
}

var MaxInt8 *big.Int
var MinInt8 *big.Int
var MaxInt16 *big.Int
var MinInt16 *big.Int
var MaxInt32 *big.Int
var MinInt32 *big.Int
var MaxInt64 *big.Int
var MinInt64 *big.Int
var MaxFloat32 *big.Float
var MinFloat32 *big.Float
var MaxFloat64 *big.Float
var MinFloat64 *big.Float

var MaxUint8 *big.Int
var MaxUint16 *big.Int
var MaxUint32 *big.Int
var MaxUint64 *big.Int

var MaxInt *big.Int
var MaxFloat *big.Float

var ltb = []int{
	0, 1, 1, // MAP, BOOL, CHAR
	1, 2, 4, 8, // INT8, INT16, INT32, INT64
	1, 2, 4, 8, // UINT8, UINT16, UINT32, UINT64
	0,    // STRING
	4, 8, // FLAOT32, FLOAT64,
	8, 1, 1, 8, // CONST_INT, CONST_CHAR, CONST_BOOL, CONST_FLOAT
	0, // CONST_STRING
}

var dcl = []byte{ // default char List
	2, 0, 0, 0,
	1, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,

	2, 0, 0, 0,
	1, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,

	2, 0, 0, 0,
	1, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,

	2, 0, 0, 0,
	1, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,

	2, 0, 0, 0,
	1, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,

	2, 0, 0, 0,
	1, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,

	2, 0, 0, 0,
	1, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,

	2, 0, 0, 0,
	1, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,

	2, 0, 0, 0,
	1, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,

	2, 0, 0, 0,
	1, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,
}

func init() {
	MaxInt8, _ = recent(big.Int).SetString("127", 0)
	MinInt8, _ = recent(big.Int).SetString("-128", 0)
	MaxInt16, _ = recent(big.Int).SetString("32767", 0)
	MinInt16, _ = recent(big.Int).SetString("-32768", 0)
	MaxInt32, _ = recent(big.Int).SetString("2147483647", 0)
	MinInt32, _ = recent(big.Int).SetString("-2147483648", 0)
	MaxInt64, _ = recent(big.Int).SetString("9223372036854775807", 0)
	MinInt64, _ = recent(big.Int).SetString("-9223372036854775808", 0)
	MaxFloat32, _ = recent(big.Float).SetString("+3.4E+38")
	MinFloat32, _ = recent(big.Float).SetString("-3.4E+38")
	MaxFloat64, _ = recent(big.Float).SetString("+1.7E+308")
	MinFloat64, _ = recent(big.Float).SetString("-1.7E+308")

	MaxUint8, _ = recent(big.Int).SetString("255", 0)
	MaxUint16, _ = recent(big.Int).SetString("65535", 0)
	MaxUint32, _ = recent(big.Int).SetString("4294967295", 0)
	MaxUint64, _ = recent(big.Int).SetString("18446744073709551615", 0)

	MaxInt, _ = recent(big.Int).SetString("18446744073709551615", 0)
	MaxFloat, _ = recent(big.Float).SetString("1.7976931348623157e+308")
}

// 24 bytes
func zeroByteSlice() []byte {
	return []byte{
		0, 0, 0, 0,
		0, 0, 0, 0,
		0, 0, 0, 0,
		0, 0, 0, 0,
		0, 0, 0, 0,
		0, 0, 0, 0,
	}
}

func effByte(a []byte) []byte {
	for i, j := 0, len(a); i < j; i++ {
		switch a[i] {
		case 0:
			break
		default:
			return a[i:]
		}
	}
	return a
}

func fillByte(a []byte) []byte {
	if n := len(a); n < 24 {
		return append(zeroByteSlice()[:24-n], a...)
	}
	return a
}

func (a *contractData) Update(f Fate, b interface{}) error {
	e, _ := f.(*contractEngine)
	switch b.(type) {
	case byte:
		v, _ := b.(byte)
		a.header.offset0 = uint64(v)
		return e.setData(a)
	case uint64:
		v, _ := b.(uint64)
		a.header.offset0 = v
		return e.setData(a)
	case string:
		v, _ := b.(string)
		e, _ := f.(*contractEngine)
		return e.setString(a, []byte(v))
	case *big.Int:
		v, _ := b.(*big.Int)
		vbuf := fillByte([]byte(v.String()))
		a.header.offset0, _ = mixed.D64func(vbuf[:8])
		a.header.offset1, _ = mixed.D64func(vbuf[8:16])
		a.header.offset2, _ = mixed.D64func(vbuf[16:24])
		return e.setData(a)
	case *big.Float:
		v, _ := b.(*big.Float)
		vbuf := fillByte([]byte(v.String()))
		a.header.offset0, _ = mixed.D64func(vbuf[:8])
		a.header.offset1, _ = mixed.D64func(vbuf[8:16])
		a.header.offset2, _ = mixed.D64func(vbuf[16:24])
		return e.setData(a)
	default:
		return errors.New("Unsupport Type")
	}
}

func (a *contractDataHeader) Show() ([]byte, error) {
	buf := []byte{}
	buf = append(buf, mixed.E32func(a.typ)...)
	buf = append(buf, mixed.E64func(a.length)...)
	buf = append(buf, mixed.E64func(a.offset0)...)
	buf = append(buf, mixed.E64func(a.offset1)...)
	buf = append(buf, mixed.E64func(a.offset2)...)
	return buf, nil
}

func (a *contractDataHeader) Read(b []byte) ([]byte, error) {
	if len(b) < DATA_HEADER_SIZE {
		return nil, errors.New("Fate Data Header Read: Illegal Length")
	}
	a.typ, _ = mixed.D32func(b[:4])
	a.length, _ = mixed.D64func(b[4:12])
	a.offset0, _ = mixed.D64func(b[12:20])
	a.offset1, _ = mixed.D64func(b[20:28])
	a.offset2, _ = mixed.D64func(b[28:36])
	return b[DATA_HEADER_SIZE:], nil
}

var typName = []string{
	"MAP",
	"BOOL",
	"CHAR",
	"INT8",
	"INT16",
	"INT32",
	"INT64",
	"UINT8",
	"UINT16",
	"UINT32",
	"UINT64",
	"STRING",
	"FLOAT32",
	"FLOAT64",
	"CONST_INT",
	"CONST_CHAR",
	"CONST_BOOL",
	"CONST_FLOAT",
	"CONST_STRING",
}
