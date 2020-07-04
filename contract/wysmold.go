package contract

import "io"

const (
	EOF = -1
	RET = -2
	ERR = -3
)

const (
	ROLLSIZE = 2
	BUFFSIZE = 1024
)

const (
	IDENTIFIER = iota
	INT_CONSTANT
	CHAR_CONSTANT
	BOOL_CONSTANT
	FLOAT_CONSTANT
	STRING_CONSTANT

	INC_OP
	DEC_OP
	LEFT_OP
	RIGHT_OP
	LE_OP
	GE_OP
	EQ_OP
	NE_OP
	AND_OP
	OR_OP
	MUL_ASSIGN
	DIV_ASSIGN
	MOD_ASSIGN
	ADD_ASSIGN
	SUB_ASSIGN
	LEFT_ASSIGN
	RIGHT_ASSIGN
	AND_ASSIGN
	XOR_ASSIGN
	OR_ASSIGN

	INT8
	INT16
	INT32
	INT64
	UINT8
	UINT16
	UINT32
	UINT64
	FLOAT32
	FLOAT64
	BOOL
	CHAR
	STRING

	LET
	SET
	FUNC

	IF
	ELSE
	WHILE
	CONTINUE
	BREAK
	RETURN

	SIZEOF


)

type contractLex struct {
	err        error
	fd         io.Reader
	curr       *contractWord 
	buffer     *contractLexBuffer
	rollBuffer *contractLexRollBuffer
}

type contractLexRollBuffer struct {
	cnt    int   
	buffer []int
}

type contractLexBuffer struct {
	pos    int
	lim    int
	buffer []byte
}

type contractWord struct {
	typ   int
	name  string
	value interface{}
}

var charTypeList = []int{
	ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR,
	ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR,
	ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR,
	ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR,
	ERR, ERR, ERR, ERR, ERR, MOD, AND, ERR,
	BRA, KET, MUL, ADD, COM, SUB, ERR, DIV,
	ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR,
	ERR, ERR, ERR, SEM, LT, ASSIGN, GT, ERR,
	ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR,
	ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR,
	ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR,
	ERR, ERR, ERR, SBR, ERR, SKT, XOR, ERR,
	ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR,
	ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR,
	ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR,
	ERR, ERR, ERR, CBR, OR, CKT, NOT, ERR,
	ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR,
	ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR,
	ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR,
	ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR,
	ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR,
	ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR,
	ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR,
	ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR,
	ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR,
	ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR,
	ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR,
	ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR,
	ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR,
	ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR,
	ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR,
	ERR, ERR, ERR, ERR, ERR, ERR, ERR, ERR,
}
