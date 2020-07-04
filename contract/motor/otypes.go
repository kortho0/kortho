package motor

const ( // 符号位
	SIGN_BIT = 0x8000
	DATA_BIT = 0x7FFF
)

const (
	EQ = 0x01
	LE = 0x02
	GR = 0x04
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
	ELEM
)

type contractEU (func(*contractEngine, uint64) error)

var cons = []int{
	1, 32, 32, 32, 32, // NOP, PUSH, POP, CALL, RET
	16, 16, 16, 16, // JMP, JZ, JB, JA
	8, 64, 64, 32, // LOAD, MOVE, CMP, TIME
	4, 4, 8, 8, 8, // ADD, SUB, MUL, DIV, MOD
	4, 4, 4, 4, // SHL, SHR, NEG, NOT
	4, 4, 4, 4, // OR, AND, XOR, SIZEOF
	8, 16, 32, // CUT, INDEX, CONCAT
	32, 64, 32, // FIND, INSERT, DELETE
	512, 1024, // SM3, SM2
	64,     // TMP
	16, 16, // JAE, JBE
	32, // ELEM
}

var rscRegistry = []contractEU{
	rsc0,  // NOP
	rsc1,  // PUSH
	rsc2,  // POP
	rsc3,  // CALL
	rsc4,  // RET
	rsc5,  // JMP
	rsc6,  // JZ
	rsc7,  // JB
	rsc8,  // JA
	rsc9,  // LOAD
	rsc10, // MOVE
	rsc11, // CMP
	rsc12, // TIME
	rsc13, // ADD
	rsc14, // SUB
	rsc15, // MUL
	rsc16, // DIV
	rsc17, // MOD
	rsc18, // SHL
	rsc19, // SHR
	rsc20, // NEG
	rsc21, // NOT
	rsc22, // OR
	rsc23, // AND
	rsc24, // XOR
	rsc25, // SIZEOF
	rsc26, // CUT
	rsc27, // INDEX
	rsc28, // CONCAT
	rsc29, // FIND
	rsc30, // INSERT
	rsc31, // DELETE
	rsc32, // SM3
	rsc33, // SM2
	rsc34, // TMP
	rsc35, // JAE
	rsc36, // JBE
	rsc37, // ELEM
}

var codeName = []string{
	"NOP", "PUSH", "POP", "CALL", "RET",
	"JMP", "JZ", "JB", "JA",
	"LOAD", "MOVE", "CMP", "TIME",
	"ADD", "SUB", "MUL", "DIV", "MOD",
	"SHL", "SHR", "NEG", "NOT",
	"OR", "AND", "XOR", "SIZEOF",
	"CUT", "INDEX", "CONCAT",
	"FIND", "INSERT", "DELETE",
	"SM3", "SM2",
	"TMP",
	"JAE", "JBE",
	"ELEM",
}
