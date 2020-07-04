package main

import (
	"fmt"
	"io/ioutil"
	"log"

	"kortho/contract/mixed"

	base58 "github.com/jbenet/go-base58"
	"github.com/tjfoc/gmsm/sm3"
)

var e *contractEngine

func createOp(o, a, b, c int) []byte {
	op := uint64(o)<<56 | uint64(a)<<40 | uint64(b)<<16 | uint64(c)
	return mixed.E64func(op)
}

func createOp1(o, a int, bcr uint64) []byte {
	op := uint64(o)<<56 | uint64(a)<<40 | bcr
	return mixed.E64func(op)
}

func create_strs() []byte {
	strs := []byte{0x00}

	strs = append(strs, []byte("owner")...)
	strs = append(strs, byte(0x00))

	strs = append(strs, []byte("account")...)
	strs = append(strs, byte(0x00))

	strs = append(strs, []byte("init")...)
	strs = append(strs, byte(0x00))

	strs = append(strs, []byte("addAccount")...)
	strs = append(strs, byte(0x00))

	strs = append(strs, []byte("removeAccount")...)
	strs = append(strs, byte(0x00))

	strs = append(strs, []byte("transfer")...)
	strs = append(strs, byte(0x00))

	strs = append(strs, []byte("queryAccount")...)
	strs = append(strs, byte(0x00))

	strs = append(strs, []byte("sm3Hash")...)
	strs = append(strs, byte(0x00))

	strs = append(strs, []byte("sm2Verify")...)
	strs = append(strs, byte(0x00))

	strs = append(strs, []byte("append")...)
	strs = append(strs, byte(0x00))

	strs = append(strs, []byte("msg")...)
	strs = append(strs, byte(0x00))

	strs = append(strs, []byte("h")...)
	strs = append(strs, byte(0x00))

	strs = append(strs, []byte("r")...)
	strs = append(strs, byte(0x00))
	strs = append(strs, []byte("c")...)
	strs = append(strs, byte(0x00))

	strs = append(strs, []byte("i")...)
	strs = append(strs, byte(0x00))

	strs = append(strs, []byte("j")...)
	strs = append(strs, byte(0x00))

	return strs
}

func create_text() []byte {
	text := []byte{}

	fmt.Printf("init = %d, %x\n", len(text)/8, len(text)/8)
	text = append(text, createOp(POP, 0, 0, 0)...)
	text = append(text, createOp(POP, 2, 0, 0)...)
	text = append(text, createOp(POP, 3, 0, 0)...)
	text = append(text, createOp(PUSH, 0, 0, 0)...)
	text = append(text, createOp1(LOAD, 4, 0x14)...)
	text = append(text, createOp1(LOAD, 5, 0x28)...)
	text = append(text, createOp1(LOAD, 6, 0x3c)...)
	text = append(text, createOp1(LOAD, 7, 0x50)...)
	text = append(text, createOp(MOVE, 4, 3, 0)...)
	text = append(text, createOp(INSERT, 5, 2, 7)...)
	text = append(text, createOp(RET, 6, 0, 0)...)

	fmt.Printf("addAccount = %d, %x\n", len(text)/8, len(text)/8)

	text = append(text, createOp(POP, 0, 0, 0)...)
	text = append(text, createOp(POP, 2, 0, 0)...)
	text = append(text, createOp(POP, 3, 0, 0)...)
	text = append(text, createOp(PUSH, 0, 0, 0)...)
	text = append(text, createOp1(LOAD, 4, 0x64)...)
	text = append(text, createOp1(LOAD, 5, 0x14)...)
	text = append(text, createOp1(LOAD, 6, 0x28)...)
	text = append(text, createOp1(LOAD, 7, 0x78)...)
	text = append(text, createOp1(LOAD, 8, 0x3c)...)
	text = append(text, createOp1(LOAD, 9, 0x8c)...)
	text = append(text, createOp(TMP, 10, STRING, 0)...)
	text = append(text, createOp(TIME, 10, 0, 0)...)
	text = append(text, createOp(PUSH, 10, 0, 0)...)
	text = append(text, createOp(PUSH, 2, 0, 0)...)
	text = append(text, createOp1(CALL, 0, 0x94)...)
	text = append(text, createOp(MOVE, 4, 0, 0)...)
	text = append(text, createOp(PUSH, 5, 0, 0)...)
	text = append(text, createOp(PUSH, 4, 0, 0)...)
	text = append(text, createOp(PUSH, 3, 0, 0)...)
	text = append(text, createOp1(CALL, 0, 0x8b)...)
	text = append(text, createOp(CMP, 0, 8, 0)...)
	text = append(text, createOp(JZ, 2, 0, 0)...)
	text = append(text, createOp(RET, 7, 0, 0)...)
	text = append(text, createOp(INSERT, 6, 2, 9)...)
	text = append(text, createOp(RET, 8, 0, 0)...)

	fmt.Printf("removeAccount = %d, %x\n", len(text)/8, len(text)/8)

	text = append(text, createOp(POP, 0, 0, 0)...)
	text = append(text, createOp(POP, 2, 0, 0)...)
	text = append(text, createOp(POP, 3, 0, 0)...)
	text = append(text, createOp(PUSH, 0, 0, 0)...)
	text = append(text, createOp1(LOAD, 4, 0xa0)...)
	text = append(text, createOp1(LOAD, 5, 0x14)...)
	text = append(text, createOp1(LOAD, 6, 0x28)...)
	text = append(text, createOp1(LOAD, 7, 0x78)...)
	text = append(text, createOp1(LOAD, 8, 0x3c)...)
	text = append(text, createOp(TMP, 9, STRING, 0)...)
	text = append(text, createOp(TIME, 9, 0, 0)...)
	text = append(text, createOp(PUSH, 9, 0, 0)...)
	text = append(text, createOp(PUSH, 2, 0, 0)...)
	text = append(text, createOp1(CALL, 0, 0x94)...)
	text = append(text, createOp(MOVE, 4, 0, 0)...)
	text = append(text, createOp(PUSH, 5, 0, 0)...)
	text = append(text, createOp(PUSH, 4, 0, 0)...)
	text = append(text, createOp(PUSH, 3, 0, 0)...)
	text = append(text, createOp1(CALL, 0, 0x8b)...)
	text = append(text, createOp(CMP, 0, 8, 0)...)
	text = append(text, createOp(JZ, 2, 0, 0)...)
	text = append(text, createOp(RET, 7, 0, 0)...)
	text = append(text, createOp(DELETE, 6, 2, 0)...)
	text = append(text, createOp(RET, 8, 0, 0)...)

	fmt.Printf("queryAccount = %d, %x\n", len(text)/8, len(text)/8)

	text = append(text, createOp(POP, 0, 0, 0)...)
	text = append(text, createOp(POP, 2, 0, 0)...)
	text = append(text, createOp(POP, 3, 0, 0)...)      // Pop(reg_n + 1)
	text = append(text, createOp(POP, 4, 0, 0)...)      // Pop(reg_n + 2)
	text = append(text, createOp(PUSH, 0, 0, 0)...)     // Push(reg_0)
	text = append(text, createOp1(LOAD, 5, 0xb4)...)    // Load(reg_n + 3, address of msg)
	text = append(text, createOp1(LOAD, 6, 0x28)...)    // Load(reg_n + 4, address of account)
	text = append(text, createOp1(LOAD, 7, 0xc8)...)    // Load(reg_n + 5, address of -1)
	text = append(text, createOp1(LOAD, 8, 0x3c)...)    // Load(reg_n + 6, address of true)
	text = append(text, createOp(PUSH, 4, 0, 0)...)     // Push(reg_n + 2)
	text = append(text, createOp1(CALL, 0, 0x84)...)    // Call(Hash)
	text = append(text, createOp(CMP, 0, 2, 0)...)      // Cmp(reg_0, reg_n)
	text = append(text, createOp(JZ, 2, 0, 0)...)       // Jz(+2)
	text = append(text, createOp(RET, 7, 0, 0)...)      // Ret(reg_n + 5)
	text = append(text, createOp(TMP, 9, STRING, 0)...) // Tmp(reg_n + 7, STRING)
	text = append(text, createOp(TIME, 9, 0, 0)...)     // Time(reg_n + 7)
	text = append(text, createOp(PUSH, 9, 0, 0)...)     // Push(reg_n + 7)
	text = append(text, createOp(PUSH, 2, 0, 0)...)     // Push(reg_n)
	text = append(text, createOp1(CALL, 0, 0x94)...)    // Call(append)
	text = append(text, createOp(MOVE, 5, 0, 0)...)     // Move(reg_n + 3, reg_0)
	text = append(text, createOp(PUSH, 4, 0, 0)...)     // Push(reg_n + 2)
	text = append(text, createOp(PUSH, 5, 0, 0)...)     // Push(reg_n + 3)
	text = append(text, createOp(PUSH, 3, 0, 0)...)     // Push(reg_n + 1)
	text = append(text, createOp1(CALL, 0, 0x8b)...)    // Call(verify)
	text = append(text, createOp(CMP, 0, 8, 0)...)      // Cmp(reg_0, reg_n + 6)
	text = append(text, createOp(JZ, 2, 0, 0)...)       // Jz(+2)
	text = append(text, createOp(RET, 7, 0, 0)...)      // Ret(reg_n + 5)
	text = append(text, createOp(FIND, 10, 6, 2)...)    // Find(reg_n + 8, reg_n + 4, reg_n)
	text = append(text, createOp(RET, 10, 0, 0)...)     // Ret(reg_n + 8)

	fmt.Printf("transfer = %d, %x\n", len(text)/8, len(text)/8)

	text = append(text, createOp(POP, 0, 0, 0)...)       // Pop(reg_0)
	text = append(text, createOp(POP, 2, 0, 0)...)       // Pop(reg_n)
	text = append(text, createOp(POP, 3, 0, 0)...)       // Pop(reg_n + 1)
	text = append(text, createOp(POP, 4, 0, 0)...)       // Pop(reg_n + 2)
	text = append(text, createOp(POP, 5, 0, 0)...)       // Pop(reg_n + 3)
	text = append(text, createOp(POP, 6, 0, 0)...)       // Pop(reg_n + 4)
	text = append(text, createOp(PUSH, 0, 0, 0)...)      // Push(reg_0)
	text = append(text, createOp1(LOAD, 7, 0xdc)...)     // Load(reg_n + 5, address of msg)
	text = append(text, createOp1(LOAD, 8, 0x28)...)     // Load(reg_n + 6, address of account in RAM)
	text = append(text, createOp1(LOAD, 9, 0x3c)...)     // Load(reg_n + 7, address of true)
	text = append(text, createOp1(LOAD, 10, 0x78)...)    // Load(reg_n + 8, address of false)
	text = append(text, createOp1(LOAD, 11, 0x8c)...)    // Load(reg_n + 9, address of zero)
	text = append(text, createOp(PUSH, 5, 0, 0)...)      // Push(reg_n + 3)
	text = append(text, createOp1(CALL, 0, 0x84)...)     // Call(Hash)
	text = append(text, createOp(CMP, 0, 2, 0)...)       // Cmp(reg_0, reg_n)
	text = append(text, createOp(JZ, 2, 0, 0)...)        // Jz(+2)
	text = append(text, createOp(RET, 10, 0, 0)...)      // Ret(reg_n + 8)
	text = append(text, createOp(TMP, 12, STRING, 0)...) // Tmp(reg_n + 10, STRING)
	text = append(text, createOp(TIME, 12, 0, 0)...)     // Time(reg_n + 10)
	text = append(text, createOp(PUSH, 12, 0, 0)...)     // Push(reg_n + 10)
	text = append(text, createOp(PUSH, 3, 0, 0)...)      // Push(reg_n + 1)
	text = append(text, createOp1(CALL, 0, 0x94)...)     // Call(append)
	text = append(text, createOp(PUSH, 0, 0, 0)...)      // Push(reg_0)
	text = append(text, createOp(PUSH, 2, 0, 0)...)      // Push(reg_n)
	text = append(text, createOp1(CALL, 0, 0x94)...)     // Call(append)
	text = append(text, createOp(MOVE, 7, 0, 0)...)      // Move(reg_n + 5, reg_0)
	text = append(text, createOp(PUSH, 5, 0, 0)...)      // Push(reg_n + 3)
	text = append(text, createOp(PUSH, 7, 0, 0)...)      // Push(reg_n + 5)
	text = append(text, createOp(PUSH, 4, 0, 0)...)      // Push(reg_n + 2)
	text = append(text, createOp1(CALL, 0, 0x8b)...)     // Call(Verify)
	text = append(text, createOp(CMP, 0, 9, 0)...)       // Cmp(reg_0, reg_n + 7)
	text = append(text, createOp(JZ, 2, 0, 0)...)        // Jz(+2)
	text = append(text, createOp(RET, 10, 0, 0)...)      // Ret(reg_n + 8)
	text = append(text, createOp(FIND, 12, 8, 2)...)     // Find(reg_n + 10, reg_n + 6, reg_n)
	text = append(text, createOp(TMP, 13, INT32, 0)...)  // Tmp(reg_n + 11, INT32)
	text = append(text, createOp(SUB, 13, 12, 6)...)     // Sub(reg_n + 11, reg_n + 10, reg_n + 4)
	text = append(text, createOp(CMP, 13, 11, 0)...)     // Cmp(reg_n + 11, reg_n + 9)
	text = append(text, createOp(JB, 5, 0, 0)...)        // Jb(+5)
	text = append(text, createOp(FIND, 14, 8, 3)...)     // Find(reg_n + 12, reg_n + 6, reg_n + 1)
	text = append(text, createOp(ADD, 14, 14, 6)...)     // Add(reg_n + 12, reg_n + 12, reg_n + 4)
	text = append(text, createOp(SUB, 12, 12, 6)...)     // Sub(reg_n + 10, reg_n + 10, reg_n + 4)
	text = append(text, createOp(RET, 9, 0, 0)...)       // Ret(reg_n + 7)
	text = append(text, createOp(RET, 10, 0, 0)...)      // Ret(reg_n + 8)

	fmt.Printf("sm3Hash = %d, %x\n", len(text)/8, len(text)/8)

	text = append(text, createOp(POP, 0, 0, 0)...)   //  Pop(reg_0)
	text = append(text, createOp(POP, 2, 0, 0)...)   // Pop(reg_n)
	text = append(text, createOp(PUSH, 0, 0, 0)...)  // Push(reg_0)
	text = append(text, createOp1(LOAD, 3, 0xf0)...) // Load(reg_n + 1, address of h)
	text = append(text, createOp(SM3, 2, 0, 0)...)   // SM3(reg_n)
	text = append(text, createOp(MOVE, 3, 0, 0)...)  // Move(reg_n + 1, reg_0)
	text = append(text, createOp(RET, 3, 0, 0)...)   // Ret(reg_n + 1)

	fmt.Printf("sm2Verify = %d, %x\n", len(text)/8, len(text)/8)

	text = append(text, createOp(POP, 0, 0, 0)...)    // Pop(reg_0)
	text = append(text, createOp(POP, 2, 0, 0)...)    // Pop(reg_n)
	text = append(text, createOp(POP, 3, 0, 0)...)    // Pop(reg_n + 1)
	text = append(text, createOp(POP, 4, 0, 0)...)    // Pop(reg_n + 2)
	text = append(text, createOp(PUSH, 0, 0, 0)...)   // Push(reg_0)
	text = append(text, createOp1(LOAD, 5, 0x104)...) // Load(reg_n + 3, address of r)
	text = append(text, createOp(SM2, 2, 3, 4)...)    // Sm2(reg_n, reg_n + 1, reg_n + 2)
	text = append(text, createOp(MOVE, 5, 0, 0)...)   // Move(reg_n + 3, reg_0)
	text = append(text, createOp(RET, 5, 0, 0)...)    // Ret(reg_n + 3)

	fmt.Printf("append = %d, %x\n", len(text)/8, len(text)/8)

	text = append(text, createOp(POP, 0, 0, 0)...)          // Pop(reg_0)
	text = append(text, createOp(POP, 2, 0, 0)...)          // Pop(reg_n)
	text = append(text, createOp(POP, 3, 0, 0)...)          // Pop(reg_n + 1)
	text = append(text, createOp(PUSH, 0, 0, 0)...)         // Push(reg_0)
	text = append(text, createOp1(LOAD, 4, 0x118)...)       // Load(reg_n + 2, address of c)
	text = append(text, createOp1(LOAD, 5, 0x12c)...)       // Load(reg_n + 3, address of i)
	text = append(text, createOp1(LOAD, 6, 0x140)...)       // Load(reg_n + 4, address of j)
	text = append(text, createOp1(LOAD, 7, 0x8c)...)        // Load(reg_n + 5, address of zero)
	text = append(text, createOp1(LOAD, 8, 0x154)...)       // Load(reg_n + 6, address of one)
	text = append(text, createOp(MOVE, 5, 7, 0)...)         // Move(reg_n + 3, reg_n + 5)
	text = append(text, createOp(SIZEOF, 6, 2, 0)...)       // Sizeof(reg_n + 4, reg_n)
	text = append(text, createOp(CMP, 5, 6, 0)...)          // Cmp(reg_n + 3, reg_n + 4)
	text = append(text, createOp(JAE, 5, 0, 0)...)          // Jae(+5)
	text = append(text, createOp(INDEX, 9, 2, 5)...)        // Index(reg_n + 7, reg_n, reg_n + 3)
	text = append(text, createOp(CONCAT, 4, 9, 0)...)       // Concat(reg_n + 2, reg_n + 7)
	text = append(text, createOp(ADD, 5, 5, 8)...)          // Add(reg_n + 3, reg_n + 3, reg_n + 6)
	text = append(text, createOp(JMP, 5|SIGN_BIT, 0, 0)...) // JMP(-5)
	text = append(text, createOp(MOVE, 5, 7, 0)...)         // Move(reg_n + 3, reg_n + 5)
	text = append(text, createOp(SIZEOF, 6, 3, 0)...)       // Sizeof(reg_n + 4, reg_n + 1)
	text = append(text, createOp(CMP, 5, 6, 0)...)          // Cmp(reg_n + 3, reg_n + 4)
	text = append(text, createOp(JAE, 5, 0, 0)...)          // Jae(+5)
	text = append(text, createOp(INDEX, 10, 3, 5)...)       // Index(reg_n + 8, reg_n + 1, reg_n + 3)
	text = append(text, createOp(CONCAT, 4, 10, 0)...)      // Concat(reg_n + 2, reg_n + 8)
	text = append(text, createOp(ADD, 5, 5, 8)...)          // Add(reg_n + 3, reg_n + 3, reg_n + 6)
	text = append(text, createOp(JMP, 5|SIGN_BIT, 0, 0)...) // JMP(-5)
	text = append(text, createOp(RET, 4, 0, 0)...)          // Ret(reg_n + 2)
	return text
}

func header_data(typ uint32, length, offset uint64) []byte {
	h := &contractDataHeader{
		typ:    typ,
		length: length,
		offset: offset,
	}
	hData, _ := h.Show()
	return hData
}

func create_data() []byte {
	buf := []byte{}
	buf = append(buf, header_data(0x00, 0, 0)...)
	buf = append(buf, header_data(0x0B, 0, 0)...) // owner, 0x14
	buf = append(buf, header_data(0x00, 0, 0)...) // account, 0x28
	buf = append(buf, header_data(0x10, 0, 0)...) // true, 0x3c
	buf = append(buf, header_data(0x0E, 0, 0)...) // 10000, 0x50
	buf = append(buf, header_data(0x0B, 0, 0)...) // msg of addAccount, 0x64
	buf = append(buf, header_data(0x10, 0, 0)...) // false, 0x78
	buf = append(buf, header_data(0x0E, 0, 0)...) // 0, 0x8c
	buf = append(buf, header_data(0x0B, 0, 0)...) // msg of removeAccount, 0xa0
	buf = append(buf, header_data(0x0B, 0, 0)...) // msg of queryAccount, 0xb4
	buf = append(buf, header_data(0x0E, 0, 0)...) // -1, 0xc8
	buf = append(buf, header_data(0x0B, 0, 0)...) // msg of transfer, 0xdc
	buf = append(buf, header_data(0x0B, 0, 0)...) // h of sm3Hash, 0xf0
	buf = append(buf, header_data(0x01, 0, 0)...) // r of sm2Verify, 0x104
	buf = append(buf, header_data(0x0B, 0, 0)...) // c of append, 0x118
	buf = append(buf, header_data(0x0A, 0, 0)...) // i of append, 0x12c
	buf = append(buf, header_data(0x0A, 0, 0)...) // j of append, 0x140
	buf = append(buf, header_data(0x0E, 0, 0)...) // 1, 0x154
	return buf
}

func sym_data(name, attr, info, size, value, extra uint32, address, raddress uint64) []byte {
	s := &diste{
		name:     name,
		attr:     attr,
		info:     info,
		size:     size,
		value:    value,
		extra:    extra,
		address:  address,
		raddress: raddress,
	}
	sData, _ := s.Show()
	return sData
}

func create_syms() []byte {
	buf := []byte{}
	buf = append(buf, sym_data(15, FSCE_TEXT, 0, 6, 0, 0, 0, 0)...)     // init, 6
	buf = append(buf, sym_data(20, FSCE_TEXT, 0, 9, 0x0B, 0, 0, 0)...)  // addAccount, 9
	buf = append(buf, sym_data(31, FSCE_TEXT, 0, 8, 0x24, 0, 0, 0)...)  // removeAccount, 8
	buf = append(buf, sym_data(54, FSCE_TEXT, 0, 9, 0x3c, 0, 0, 0)...)  // queryAccount, 9
	buf = append(buf, sym_data(45, FSCE_TEXT, 0, 13, 0x59, 0, 0, 0)...) // transfer, 13
	buf = append(buf, sym_data(67, FSCE_TEXT, 0, 2, 0x84, 0, 0, 0)...)  // sm3Hash, 2
	buf = append(buf, sym_data(75, FSCE_TEXT, 0, 4, 0x8b, 0, 0, 0)...)  // sm2Verify, 4
	buf = append(buf, sym_data(85, FSCE_TEXT, 0, 9, 0x94, 0, 0, 0)...)  // append, 9

	if addr, err := e.virtul.Alloc(FLASH, DATA_HEADER_SIZE); err == nil {
		if _, err = e.recentVar(STRING, addr); err == nil {
			fmt.Printf("owner: %x\n", addr)
			buf = append(buf, sym_data(1, FSCE_DATA, FLASH_VAR, 0, 0, 0, addr, 0x14)...) // owner
		} else {
			log.Fatal(err)
		}
	} else {
		log.Fatal(err)
	}

	if addr, err := e.virtul.Alloc(FLASH, DATA_HEADER_SIZE); err == nil {
		if _, err := e.recentMap(STRING, INT32, addr); err == nil {
			fmt.Printf("account: %x\n", addr)
			buf = append(buf, sym_data(7, FSCE_DATA, FLASH_VAR, 0, 0, 0, addr, 0x28)...) // account
		} else {
			log.Fatal(err)
		}
	} else {
		log.Fatal(err)
	}

	if addr, err := e.virtul.Alloc(FLASH, DATA_HEADER_SIZE); err == nil {
		if _, err = e.recentVar(STRING, addr); err == nil {
			fmt.Printf("msg of addAccount: %x\n", addr)
			buf = append(buf, sym_data(92, FSCE_DATA, RAM_VAR, 0, 0, 0, 0x64, addr)...) // msg of addAccount
		} else {
			log.Fatal(err)
		}
	} else {
		log.Fatal(err)
	}

	if addr, err := e.virtul.Alloc(FLASH, DATA_HEADER_SIZE); err == nil {
		if _, err = e.recentVar(STRING, addr); err == nil {
			fmt.Printf("msg of removeAccount: %x\n", addr)
			buf = append(buf, sym_data(92, FSCE_DATA, RAM_VAR, 0, 0, 0, 0xa0, addr)...) // msg of removeAccount
		} else {
			log.Fatal(err)
		}
	} else {
		log.Fatal(err)
	}

	if addr, err := e.virtul.Alloc(FLASH, DATA_HEADER_SIZE); err == nil {
		if _, err = e.recentVar(STRING, addr); err == nil {
			fmt.Printf("msg of queryAccount: %x\n", addr)
			buf = append(buf, sym_data(92, FSCE_DATA, RAM_VAR, 0, 0, 0, 0xb4, addr)...) // msg of queryAccount
		} else {
			log.Fatal(err)
		}
	} else {
		log.Fatal(err)
	}

	if addr, err := e.virtul.Alloc(FLASH, DATA_HEADER_SIZE); err == nil {
		if _, err = e.recentVar(STRING, addr); err == nil {
			fmt.Printf("msg of transfer: %x\n", addr)
			buf = append(buf, sym_data(92, FSCE_DATA, RAM_VAR, 0, 0, 0, 0xdc, addr)...) // msg of transfer
		} else {
			log.Fatal(err)
		}
	} else {
		log.Fatal(err)
	}

	if addr, err := e.virtul.Alloc(FLASH, DATA_HEADER_SIZE); err == nil {
		if _, err = e.recentVar(STRING, addr); err == nil {
			fmt.Printf("h: %x\n", addr)
			buf = append(buf, sym_data(96, FSCE_DATA, RAM_VAR, 0, 0, 0, 0xf0, addr)...) // h
		} else {
			log.Fatal(err)
		}
	} else {
		log.Fatal(err)
	}

	if addr, err := e.virtul.Alloc(FLASH, DATA_HEADER_SIZE); err == nil {
		if _, err = e.recentVar(BOOL, addr); err == nil {
			fmt.Printf("r: %x\n", addr)
			buf = append(buf, sym_data(98, FSCE_DATA, RAM_VAR, 0, 0, 0, 0x104, addr)...) // r
		} else {
			log.Fatal(err)
		}
	} else {
		log.Fatal(err)
	}

	if addr, err := e.virtul.Alloc(FLASH, DATA_HEADER_SIZE); err == nil {
		if _, err = e.recentVar(STRING, addr); err == nil {
			fmt.Printf("c: %x\n", addr)
			buf = append(buf, sym_data(100, FSCE_DATA, RAM_VAR, 0, 0, 0, 0x118, addr)...) // c
		} else {
			log.Fatal(err)
		}
	} else {
		log.Fatal(err)

	}

	if addr, err := e.virtul.Alloc(FLASH, DATA_HEADER_SIZE); err == nil {
		if _, err = e.recentVar(UINT64, addr); err == nil {
			fmt.Printf("i: %x\n", addr)
			buf = append(buf, sym_data(102, FSCE_DATA, RAM_VAR, 0, 0, 0, 0x12c, addr)...) // i
		} else {
			log.Fatal(err)
		}
	} else {
		log.Fatal(err)
	}

	if addr, err := e.virtul.Alloc(FLASH, DATA_HEADER_SIZE); err == nil {
		if _, err = e.recentVar(UINT64, addr); err == nil {
			fmt.Printf("j: %x\n", addr)
			buf = append(buf, sym_data(104, FSCE_DATA, RAM_VAR, 0, 0, 0, 0x140, addr)...) // j
		} else {
			log.Fatal(err)
		}
	} else {
		log.Fatal(err)
	}

	if addr, err := e.virtul.Alloc(FLASH, DATA_HEADER_SIZE); err == nil {
		if d, err := e.recentVar(CONST_BOOL, addr); err == nil {
			fmt.Printf("true: %x\n", addr)
			d.header.offset = 1
			e.setData(d)
			buf = append(buf, sym_data(0, FSCE_DATA, CONSTANT, 0, 0, 0, 0x3c, addr)...) // true
		} else {
			log.Fatal(err)
		}
	} else {
		log.Fatal(err)
	}

	if addr, err := e.virtul.Alloc(FLASH, DATA_HEADER_SIZE); err == nil {
		if d, err := e.recentVar(CONST_INT, addr); err == nil {
			fmt.Printf("10000: %x\n", addr)
			d.header.offset = 10000
			e.setData(d)
			buf = append(buf, sym_data(0, FSCE_DATA, CONSTANT, 0, 0, 0, 0x50, addr)...) // 10000
		} else {
			log.Fatal(err)
		}
	} else {
		log.Fatal(err)
	}

	if addr, err := e.virtul.Alloc(FLASH, DATA_HEADER_SIZE); err == nil {
		if _, err = e.recentVar(CONST_BOOL, addr); err == nil {
			fmt.Printf("false: %x\n", addr)
			buf = append(buf, sym_data(0, FSCE_DATA, CONSTANT, 0, 0, 0, 0x78, addr)...) // false
		} else {
			log.Fatal(err)
		}
	} else {
		log.Fatal(err)
	}

	if addr, err := e.virtul.Alloc(FLASH, DATA_HEADER_SIZE); err == nil {
		if _, err = e.recentVar(CONST_INT, addr); err == nil {
			fmt.Printf("0: %x\n", addr)
			buf = append(buf, sym_data(0, FSCE_DATA, CONSTANT, 0, 0, 0, 0x8c, addr)...) // zero
		} else {
			log.Fatal(err)
		}
	} else {
		log.Fatal(err)
	}

	if addr, err := e.virtul.Alloc(FLASH, DATA_HEADER_SIZE); err == nil {
		if d, err := e.recentVar(CONST_INT, addr); err == nil {
			fmt.Printf("-1: %x\n", addr)
			a := int64(-1)
			d.header.offset = uint64(a)
			e.setData(d)
			buf = append(buf, sym_data(0, FSCE_DATA, CONSTANT, 0, 0, 0, 0xc8, addr)...) // -1
		} else {
			log.Fatal(err)
		}
	} else {
		log.Fatal(err)
	}

	if addr, err := e.virtul.Alloc(FLASH, DATA_HEADER_SIZE); err == nil {
		if d, err := e.recentVar(CONST_INT, addr); err == nil {
			fmt.Printf("1: %x\n", addr)
			d.header.offset = 1
			e.setData(d)
			buf = append(buf, sym_data(0, FSCE_DATA, CONSTANT, 0, 0, 0, 0x154, addr)...) // 1
		} else {
			log.Fatal(err)
		}
	} else {
		log.Fatal(err)
	}

	return buf
}

/*
func recentEngineMemory0(db store.Store) *contractEngineStore {
	return &contractEngineStore{
		es: &contractEngineEndStore{db},
		ms: &contractEngineMemStore{
			start: 1024,
			pages: make([][]byte, 1024),
		},
	}
}

func recentEngineProgramme0(code []byte) *contractEngineProgramme {
	h := sm3.New()
	h.Write(code)
	name := h.Sum(nil)
	address := base58.Encode(name)
	fmt.Printf("address = %s\n", address)
	return &contractEngineProgramme{
		name: name,
	}
}
*/

func codeAddress(code []byte) string {
	h := sm3.New()
	h.Write(code)
	name := h.Sum(nil)
	return base58.Encode(name)
}

func main() {
	code, _ := ioutil.ReadFile("train.ft")
	name := codeAddress(code)
	e = recentEngine(name)

	msg := []byte{}

	strs := create_strs()
	strs_length := mixed.E32func(uint32(len(strs)))

	text := create_text()
	text_length := mixed.E32func(uint32(len(text)))

	syms := create_syms()
	syms_length := mixed.E32func(uint32(len(syms)))

	data := create_data()
	data_length := mixed.E32func(uint32(len(data)))

	msg = append(msg, FSCE_MAGIC...)
	msg = append(msg, strs_length...)
	msg = append(msg, text_length...)
	msg = append(msg, syms_length...)
	msg = append(msg, data_length...)
	msg = append(msg, strs...)
	msg = append(msg, text...)
	msg = append(msg, syms...)
	msg = append(msg, data...)

	e.virtul.Flush()

	e.virtul.Close()

	e.virtul.SetExecute(msg)

	/*
		db, _ := store.NewDb("ft.db")

		e = &contractEngine{
			prog: recentEngineProgramme0(code),
			mem:  recentEngineMemory0(db),
		}

		db.Set(append(e.prog.name, mixed.E32func(0)...), append(Sentry, make([]byte, int(SEG_LIMIT)-len(Sentry))...))

		msg := []byte{}

		strs := create_strs()
		strs_length := mixed.E32func(uint32(len(strs)))

		text := create_text()
		text_length := mixed.E32func(uint32(len(text)))

		syms := create_syms()
		syms_length := mixed.E32func(uint32(len(syms)))

		data := create_data()
		data_length := mixed.E32func(uint32(len(data)))

		msg = append(msg, FSCE_MAGIC...)
		msg = append(msg, strs_length...)
		msg = append(msg, text_length...)
		msg = append(msg, syms_length...)
		msg = append(msg, data_length...)
		msg = append(msg, strs...)
		msg = append(msg, text...)
		msg = append(msg, syms...)
		msg = append(msg, data...)

		db.Set(e.prog.name, msg)

		//	ioutil.WriteFile("ft.ext", msg, os.FileMode(0777))
	*/
}
