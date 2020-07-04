package main

import (
	"errors"

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
	DATA_HEADER_SIZE      = 20
	DEFAULT_STRING_LENGTH = 10
)

type contractDataHeader struct {
	typ    uint32
	length uint64
	offset uint64
}

type contractData struct {
	address uint64
	header  *contractDataHeader
}

var ltb = []int{
	0, 1, 1,
	1, 2, 4, 8,
	1, 2, 4, 8,
	0,
	4, 8,
	8, 1, 1, 8,
	0,
}

var dcl = []byte{
	2, 0, 0, 0,
	1, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,

	2, 0, 0, 0,
	1, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,

	2, 0, 0, 0,
	1, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,

	2, 0, 0, 0,
	1, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,

	2, 0, 0, 0,
	1, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,

	2, 0, 0, 0,
	1, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,

	2, 0, 0, 0,
	1, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,

	2, 0, 0, 0,
	1, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,

	2, 0, 0, 0,
	1, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,

	2, 0, 0, 0,
	1, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,
	0, 0, 0, 0,
}

func (a *contractDataHeader) Show() ([]byte, error) {
	buf := []byte{}
	buf = append(buf, mixed.E32func(a.typ)...)
	buf = append(buf, mixed.E64func(a.length)...)
	buf = append(buf, mixed.E64func(a.offset)...)
	return buf, nil
}

func (a *contractDataHeader) Read(b []byte) ([]byte, error) {
	if len(b) < DATA_HEADER_SIZE {
		return nil, errors.New("Fate Data Header Read: Illegal Length")
	}
	a.typ, _ = mixed.D32func(b[:4])
	a.length, _ = mixed.D64func(b[4:12])
	a.offset, _ = mixed.D64func(b[12:20])
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
