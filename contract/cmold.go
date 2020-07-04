package contract

import (
	"errors"

	"kortho/contract/mixed"
	"kortho/contract/motor"
)

const (
	S_VAR  = 0x01
	S_SCS  = 0x02
	S_TMP  = 0x03
	S_CNT  = 0x04
	S_ARG  = 0x05
	S_FUNC = 0x06
)

const (
	KEY_OFF  = 16
	VAL_OFF  = 24
	KEY_MASK = 0xFF0000
	VAL_MASK = 0xFF000000
)

const (
	G_VAR = iota
	G_FUNC
	G_ARGS
	G_BODY
	G_BLK
	G_STMT
	G_CSTMT
	G_JSTMT
	G_ISTMT
	G_SSTMT
	G_ESTMT
	G_EXPR
	G_ASGN
	G_LGOR
	G_LGAND
	G_INOR
	G_EXOR
	G_AND
	G_EQ
	G_CMP
	G_SHF
	G_ADD
	G_MUL
	G_UNARY
	G_SIZEOF
	G_PTF
	G_PRM
	G_PARA
)

type diste struct {
	name     uint32
	attr     uint32
	info     uint32
	size     uint32
	value    uint32
	extra    uint32
	address  uint64
	raddress uint64
}

type contractGenerator struct {
	attr *contractAttr
	e    motor.Fate
}

type contractFunc struct {
	idx   int
	rtv   uint32
	ops   []uint64
	loads []uint64
	regs  []*contractSymbol
	args  []*contractSymbol
	calls []*contractAttrCall
}

type contractAttr struct {
	st *contractSattr
	it *contractIattr
}

type contractAttrCall struct {
	name string
	args []*contractSymbol
}

type contractDataHeader struct {
	typ     uint32
	length  uint64
	offset0 uint64
	offset1 uint64
	offset2 uint64
}

type contractSattr struct {
	strs   []byte
	ops    []uint64
	syms   []*diste
	data   []*contractDataHeader
	sMap   map[string]uint32
	funcs  map[string]*contractFunc
	bConst map[bool]*contractSymbol
	cConst map[byte]*contractSymbol
	iConst map[int64]*contractSymbol
	uConst map[uint64]*contractSymbol
	sConst map[string]*contractSymbol
	fConst map[float64]*contractSymbol
}

type contractIattr struct {
	ops   []uint64
	cFunc *contractFunc
	kSym  []*contractSymbol
	mSym  []*contractSymbol
	cSym  *contractSymbol
	args  [][]*contractSymbol
}

type contractGenFunc (func(*contractGenerator, *contractNode) error)

var rootRegistry map[int]contractGenFunc

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

func (a *diste) achieveAddr() uint64 {
	switch a.info {
	case motor.RAM_VAR, motor.CONSTANT:
		return a.address
	default:
		return a.raddress
	}
}

func (a *diste) Show() ([]byte, error) {
	buf := []byte{}
	buf = append(buf, mixed.E32func(a.name)...)
	buf = append(buf, mixed.E32func(a.attr)...)
	buf = append(buf, mixed.E32func(a.info)...)
	buf = append(buf, mixed.E32func(a.size)...)
	buf = append(buf, mixed.E32func(a.value)...)
	buf = append(buf, mixed.E32func(a.extra)...)
	buf = append(buf, mixed.E64func(a.address)...)
	buf = append(buf, mixed.E64func(a.raddress)...)
	return buf, nil
}

func (a *diste) Read(b []byte) ([]byte, error) {
	if len(b) < 40 {
		return nil, errors.New("Fate Exectue Symbol: Illegal slice length")
	}
	a.name, _ = mixed.D32func(b[:4])
	a.attr, _ = mixed.D32func(b[4:8])
	a.info, _ = mixed.D32func(b[8:12])
	a.size, _ = mixed.D32func(b[12:16])
	a.value, _ = mixed.D32func(b[16:20])
	a.extra, _ = mixed.D32func(b[20:24])
	a.address, _ = mixed.D64func(b[24:32])
	a.raddress, _ = mixed.D64func(b[32:40])
	return b[40:], nil
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
	if len(b) < motor.DATA_HEADER_SIZE {
		return nil, errors.New("Fate Data Header Read: Illegal Length")
	}
	a.typ, _ = mixed.D32func(b[:4])
	a.length, _ = mixed.D64func(b[4:12])
	a.offset0, _ = mixed.D64func(b[12:20])
	a.offset1, _ = mixed.D64func(b[20:28])
	a.offset2, _ = mixed.D64func(b[28:36])
	return b[motor.DATA_HEADER_SIZE:], nil
}
