package motor

import (
	"errors"

	"kortho/contract/mixed"
)

const (
	FSCE_UNDF = iota
	FSCE_TEXT
	FSCE_DATA
)

var FSCE_MAGIC = []byte{70, 65, 84, 69}
var FSCE_MAGIC_NUMBER = uint32(0x45544146)

type fsce struct {
	magic uint32
	strs  uint32
	text  uint32
	syms  uint32
	data  uint32
}

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

type fsceSym struct {
	attr     uint32
	size     uint32
	info     uint32
	value    uint32
	address  uint64
	raddress uint64
	name     string
}

func (a *fsce) Show() ([]byte, error) {
	buf := []byte{}
	buf = append(buf, mixed.E32func(a.magic)...)
	buf = append(buf, mixed.E32func(a.strs)...)
	buf = append(buf, mixed.E32func(a.text)...)
	buf = append(buf, mixed.E32func(a.syms)...)
	buf = append(buf, mixed.E32func(a.data)...)
	return buf, nil
}

func (a *fsce) Read(b []byte) ([]byte, error) {
	if len(b) < 20 {
		return nil, errors.New("Fate Exectue Header: Illegal slice length")
	}
	a.magic, _ = mixed.D32func(b[:4])
	a.strs, _ = mixed.D32func(b[4:8])
	a.text, _ = mixed.D32func(b[8:12])
	a.syms, _ = mixed.D32func(b[12:16])
	a.data, _ = mixed.D32func(b[16:20])
	return b[20:], nil
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
