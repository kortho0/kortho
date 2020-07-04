package main

import "kortho/contract/virtul"

const ( // 符号位
	SIGN_BIT = 0x8000
	DATA_BIT = 0x7FFF
)

const (
	RAM = iota
	FLASH
)

const (
	NOP = iota
	PUSH
	POP
	CALL
	RET
	JMP
	JZ
	JB
	JA
	LOAD
	MOVE
	CMP
	TIME
	ADD
	SUB
	MUL
	DIV
	MOD
	SHL
	SHR
	NEG
	NOT
	OR
	AND
	XOR
	SIZEOF
	CUT
	INDEX
	CONCAT
	FIND
	INSERT
	DELETE
	SM3
	SM2
	TMP
	JAE
	JBE
)

type contractEngine struct {
	virtul virtul.FateVM
}

func recentEngine(name string) *contractEngine {
	v, _ := virtul.New([]byte{}, name)
	v.SetPage(virtul.RAM_PAGE_COUNT, append(virtul.Sentry, make([]byte, int(virtul.PAGE_SIZE)-len(virtul.Sentry))...))
	return &contractEngine{v}
}
