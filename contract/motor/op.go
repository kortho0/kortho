package motor

import (
	"errors"
	"fmt"
	"math"
	"math/big"

	"kortho/contract/mixed"

	"kortho/contract/virtul"

	base58 "github.com/jbenet/go-base58"
	"github.com/tjfoc/gmsm/sm2"
	"github.com/tjfoc/gmsm/sm3"
)

func achieveOpCode(o uint64) int {
	return int((o >> 56) & 0xFF)
}

func achieveA(o uint64) int {
	return int((o >> 40) & 0xFFFF)
}

func achieveB(o uint64) int {
	return int((o >> 16) & 0xFFFF)
}

func achieveC(o uint64) int {
	return int(o & 0xFFFF)
}

func achieveRegType(r int) int {
	return (r >> 14) & 0x3
}

func achieveBCR(o uint64) uint64 {
	return o & 0xFFFFFFFFFF
}

func achieveReg(a, b int) int {
	switch a {
	case 0:
		return 0
	case 1:
		return 1
	default:
		return a + b
	}
}

func achieveOpCons(op int) int32 {
	if op >= len(cons) {
		return math.MaxInt32
	}
	return int32(cons[op])
}

func rsc0(e *contractEngine, op uint64) error {
	e.regs.pc = e.regs.pc + 1
	return nil
}

func rsc1(e *contractEngine, op uint64) error {
	r := e.regs.gpRegs[achieveReg(achieveA(op), e.regs.offset)]
	switch {
	case e.stk.sp < e.stk.size:
		e.stk.stack[e.stk.sp] = achieveBCR(r)
	default:
		e.stk.size++
		e.stk.stack = append(e.stk.stack, achieveBCR(r))
	}
	e.stk.sp++
	e.regs.pc = e.regs.pc + 1
	return nil
}

func rsc2(e *contractEngine, op uint64) error {
	if e.stk.sp > 0 {
		e.stk.sp--
		e.regs.gpRegs[achieveReg(achieveA(op), e.regs.offset)] = e.stk.stack[e.stk.sp]
		e.regs.pc = e.regs.pc + 1
		return nil
	}
	return errors.New("Empty Stack")
}

func rsc3(e *contractEngine, op uint64) error {
	addr := achieveBCR(op)
	e.regs.gpRegs[1] = uint64(e.regs.pc + 1)
	rsc1(e, uint64(1)<<40)
	e.regs.pc = int(addr)
	if sym := e.findFunc(addr); sym != nil {
		e.regs.offset += e.off[len(e.off)-1]
		e.off = append(e.off, int(sym.size))
	} else {
		return errors.New("Fate Run: Cannot Find function")
	}
	return nil
}

func rsc4(e *contractEngine, op uint64) error {
	rsc2(e, uint64(1)<<40)
	e.regs.pc = int(achieveBCR(e.regs.gpRegs[1]))
	e.regs.gpRegs[0] = achieveBCR(e.regs.gpRegs[achieveReg(achieveA(op), e.regs.offset)])
	if da, err := e.achieveData(e.regs.gpRegs[0]); err == nil {
		if db, err := e.tmp(0, int(da.header.typ)); err == nil {
			if err = e.move(db, da); err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		return err
	}
	if len(e.off) > 1 {
		e.regs.offset -= e.off[len(e.off)-2]
		e.off = e.off[:len(e.off)-1]
	}
	return nil
}

func rsc5(e *contractEngine, op uint64) error {
	if a := achieveA(op); a&SIGN_BIT != 0 {
		e.regs.pc = e.regs.pc - (a & DATA_BIT)
	} else {
		e.regs.pc = e.regs.pc + (a & DATA_BIT)
	}
	return nil
}

func rsc6(e *contractEngine, op uint64) error {
	if e.regs.eflags&EQ != 0 {
		if a := achieveA(op); a&SIGN_BIT != 0 {
			e.regs.pc = e.regs.pc - (a & DATA_BIT)
		} else {
			e.regs.pc = e.regs.pc + (a & DATA_BIT)
		}
	} else {
		e.regs.pc = e.regs.pc + 1
	}
	return nil
}

func rsc7(e *contractEngine, op uint64) error {
	if e.regs.eflags&LE != 0 {
		if a := achieveA(op); a&SIGN_BIT != 0 {
			e.regs.pc = e.regs.pc - (a & DATA_BIT)
		} else {
			e.regs.pc = e.regs.pc + (a & DATA_BIT)
		}
	} else {
		e.regs.pc = e.regs.pc + 1
	}
	return nil
}

func rsc8(e *contractEngine, op uint64) error {
	if e.regs.eflags&GR != 0 {
		if a := achieveA(op); a&SIGN_BIT != 0 {
			e.regs.pc = e.regs.pc - (a & DATA_BIT)
		} else {
			e.regs.pc = e.regs.pc + (a & DATA_BIT)
		}
	} else {
		e.regs.pc = e.regs.pc + 1
	}
	return nil
}

func rsc9(e *contractEngine, op uint64) error {
	address := achieveBCR(op)
	for _, sym := range e.prog.syms {
		if sym.attr == FSCE_DATA && (sym.address == address || sym.raddress == address) {
			raddress := sym.address
			if raddress == address {
				raddress = sym.raddress
			}
			src, err := e.achieveData(raddress)
			if err != nil {
				return fmt.Errorf("Fate ReMap: %v", err)
			}
			dst, err := e.achieveData(address)
			if err != nil {
				return fmt.Errorf("Fate ReMap: %v", err)
			}
			if sym.info == CONSTANT && (dst.header.offset0 != 0 || dst.header.offset1 != 0 || dst.header.offset2 != 0) {
				break
			}
			if sym.info == RAM_VAR {
				if dst.header.offset0 != 0 || dst.header.offset1 != 0 || dst.header.offset2 != 0 {
					switch dst.header.typ {
					case MAP:
						e.mapRemove(dst)
					case STRING, CONST_STRING:
						e.stringRemove(dst)
					}
					dst.header.offset0 = 0
					dst.header.offset1 = 0
					dst.header.offset1 = 0
				}
			}
			if sym.info == FLASH_VAR {
				copy := true
				for _, v := range e.prog.load {
					if v == raddress {
						copy = false
						break
					}
				}
				if !copy {
					break
				}
				e.prog.load = append(e.prog.load, raddress)
			}
			switch src.header.typ {
			case MAP:
				rbk, err := e.achieveBucket(src)
				if err != nil {
					return fmt.Errorf("Fate ReMap: %v", err)
				}
				addr, err := e.virtul.Alloc(virtul.StoreType(address), BUCKET_HEADER_SIZE+uint64(rbk.header.size)*8)
				if err != nil {
					return fmt.Errorf("Fate ReMap: %v", err)
				}
				bk := &contractBucket{
					slots: make([]uint64, rbk.header.size),
					header: &contractBucketHeader{
						size: rbk.header.size,
						ktyp: rbk.header.ktyp,
						vtyp: rbk.header.vtyp,
					},
				}
				dst.header.offset0 = addr
				dst.header.length = uint64(ltb[MAP])
				e.setBucket(dst, bk)
				e.setData(dst)
			case STRING:
				addr, err := e.virtul.Alloc(virtul.StoreType(address), DEFAULT_STRING_LENGTH*DATA_HEADER_SIZE)
				if err != nil {
					return fmt.Errorf("Fate Remap: %v", err)
				}
				if err := e.virtul.SetRawData(addr, DEFAULT_STRING_LENGTH*DATA_HEADER_SIZE, dcl); err != nil {
					return fmt.Errorf("Fate Remap: %v", err)
				}
				dst.header.offset0 = addr
				dst.header.length = uint64(ltb[STRING]) | uint64(DEFAULT_STRING_LENGTH)<<32
				e.setData(dst)
			default:
				dst.header.length = uint64(ltb[src.header.typ])
				e.setData(dst)
			}
			if err = e.move(dst, src); err != nil {
				return fmt.Errorf("Fate ReMap: %v", err)
			}
			break
		}
	}
	a := achieveReg(achieveA(op), e.regs.offset)
	switch achieveRegType(a) {
	case ROUTE:
	case GENERAL:
		e.regs.gpRegs[a] = address
	default:
		return errors.New("Fate LOAD: Illegal Register Type")
	}
	e.regs.pc = e.regs.pc + 1
	return nil
}

func rsc10(e *contractEngine, op uint64) error {
	a := achieveReg(achieveA(op), e.regs.offset)
	b := achieveReg(achieveB(op), e.regs.offset)
	switch achieveRegType(a) {
	case ROUTE:
		e.regs.pc = e.regs.pc + 1
		return nil
	case GENERAL:
		da, err := e.achieveData(e.regs.gpRegs[a])
		if err != nil {
			return fmt.Errorf("Fate Move: %v", err)
		}
		db, err := e.achieveData(e.regs.gpRegs[b])
		if err != nil {
			return fmt.Errorf("Fate Move: %v", err)
		}
		if err := e.move(da, db); err != nil {
			return fmt.Errorf("Fate Move: %v", err)
		}
		e.regs.pc = e.regs.pc + 1
		return nil
	default:
		return errors.New("Fate MOVE: Illegal Register Type")
	}
}

func rsc11(e *contractEngine, op uint64) error {
	a := achieveReg(achieveA(op), e.regs.offset)
	b := achieveReg(achieveB(op), e.regs.offset)
	da, err := e.achieveData(e.regs.gpRegs[a])
	if err != nil {
		return fmt.Errorf("Fate CMP: %v", err)
	}
	db, err := e.achieveData(e.regs.gpRegs[b])
	if err != nil {
		return fmt.Errorf("Fate CMP: %v", err)
	}
	e.regs.eflags = 0
	if r, err := e.cmp(da, db); err == nil {
		switch {
		case r < 0:
			e.regs.eflags |= LE
		case r > 0:
			e.regs.eflags |= GR
		case r == 0:
			e.regs.eflags |= EQ
		}
		e.regs.pc = e.regs.pc + 1
		return nil
	} else {
		return fmt.Errorf("Fate CMP: %v", err)
	}
}

func rsc12(e *contractEngine, op uint64) error {
	da, err := e.achieveData(e.regs.gpRegs[achieveReg(achieveA(op), e.regs.offset)])
	if err != nil {
		return fmt.Errorf("Fate Time: %v", err)
	}

	if err := e.setString(da, []byte("1")); err != nil {
		return fmt.Errorf("Fate Time: %v", err)
	}
	e.regs.pc = e.regs.pc + 1
	return nil
}

func rsc13(e *contractEngine, op uint64) error {
	a := achieveReg(achieveA(op), e.regs.offset)
	b := achieveReg(achieveB(op), e.regs.offset)
	c := achieveReg(achieveC(op), e.regs.offset)
	defer func() {
		e.regs.pc = e.regs.pc + 1
	}()
	da, err := e.achieveData(e.regs.gpRegs[a])
	if err != nil {
		return fmt.Errorf("Fate ADD: %v", err)
	}
	db, err := e.achieveData(e.regs.gpRegs[b])
	if err != nil {
		return fmt.Errorf("Fate ADD: %v", err)
	}
	dc, err := e.achieveData(e.regs.gpRegs[c])
	if err != nil {
		return fmt.Errorf("Fate ADD: %v", err)
	}
	switch da.header.typ {
	case CHAR:
		if db.header.typ != da.header.typ && db.header.typ != CONST_CHAR {
			return errors.New("Fate ADD: Type Error")
		}
		if dc.header.typ != da.header.typ && dc.header.typ != CONST_CHAR {
			return errors.New("Fate ADD: Type Error")
		}
		vb := byte(db.header.offset0)
		vc := byte(dc.header.offset0)
		return da.Update(e, byte(vb+vc))
	case CONST_CHAR:
		if db.header.typ != da.header.typ && db.header.typ != CHAR {
			return errors.New("Fate ADD: Type Error")
		}
		if dc.header.typ != da.header.typ && dc.header.typ != CHAR {
			return errors.New("Fate ADD: Type Error")
		}
		vb := byte(db.header.offset0)
		vc := byte(dc.header.offset0)
		return da.Update(e, byte(vb+vc))
	case CONST_FLOAT:
		if typ := db.header.typ; typ != da.header.typ && typ != FLOAT32 && typ != FLOAT64 {
			return errors.New("Fate ADD: Type Error")
		}
		if typ := dc.header.typ; typ != da.header.typ && typ != FLOAT32 && typ != FLOAT64 {
			return errors.New("Fate ADD: Type Error")
		}
		if db.header.typ != CONST_FLOAT && dc.header.typ != CONST_FLOAT &&
			dc.header.typ != db.header.typ {
			return errors.New("Fate ADD: Type Error")
		}
		va := recent(big.Float)
		vb := recent(big.Float)
		vc := recent(big.Float)
		if _, ok := vb.SetString(string(effByte(append(mixed.E64func(db.header.offset0),
			append(mixed.E64func(db.header.offset1),
				mixed.E64func(db.header.offset2)...)...)))); !ok {
			err = errors.New("Illegal Value")
			return fmt.Errorf("Fate ADD: %v\n", err)
		}
		if _, ok := vc.SetString(string(effByte(append(mixed.E64func(dc.header.offset0),
			append(mixed.E64func(dc.header.offset1),
				mixed.E64func(dc.header.offset2)...)...)))); !ok {
			err = errors.New("Illegal Value")
			return fmt.Errorf("Fate ADD: %v\n", err)
		}
		va.Add(vb, vc)
		if va.Cmp(MinFloat64) < 0 || va.Cmp(MaxFloat64) > 0 {
			return errors.New("Fate ADD: Overflows")
		}
		return da.Update(e, va)
	case FLOAT32, FLOAT64:
		if db.header.typ != da.header.typ && db.header.typ != CONST_FLOAT {
			return errors.New("Fate ADD: Type Error")
		}
		if dc.header.typ != da.header.typ && dc.header.typ != CONST_FLOAT {
			return errors.New("Fate ADD: Type Error")
		}
		va := recent(big.Float)
		vb := recent(big.Float)
		vc := recent(big.Float)
		if _, ok := vb.SetString(string(effByte(append(mixed.E64func(db.header.offset0),
			append(mixed.E64func(db.header.offset1),
				mixed.E64func(db.header.offset2)...)...)))); !ok {
			err = errors.New("Illegal Value")
			return fmt.Errorf("Fate ADD: %v\n", err)
		}
		if _, ok := vc.SetString(string(effByte(append(mixed.E64func(dc.header.offset0),
			append(mixed.E64func(dc.header.offset1),
				mixed.E64func(dc.header.offset2)...)...)))); !ok {
			err = errors.New("Illegal Value")
			return fmt.Errorf("Fate ADD: %v\n", err)
		}
		va.Add(vb, vc)
		switch da.header.length {
		case 4:
			if va.Cmp(MinFloat32) < 0 || va.Cmp(MaxFloat32) > 0 {
				return errors.New("Fate ADD: Overflows")
			}
		case 8:
			if va.Cmp(MinFloat64) < 0 || va.Cmp(MaxFloat64) > 0 {
				return errors.New("Fate ADD: Overflows")
			}
		}
		return da.Update(e, va)
	case CONST_INT:
		if typ := db.header.typ; typ != da.header.typ && typ != INT8 && typ != INT16 &&
			typ != INT32 && typ != INT64 && typ != UINT8 && typ != UINT16 &&
			typ != UINT32 && typ != UINT64 {
			return errors.New("Fate ADD: Type Error")
		}
		if typ := dc.header.typ; typ != da.header.typ && typ != INT8 && typ != INT16 &&
			typ != INT32 && typ != INT64 && typ != UINT8 && typ != UINT16 &&
			typ != UINT32 && typ != UINT64 {
			return errors.New("Fate ADD: Type Error")
		}
		if db.header.typ != CONST_INT && dc.header.typ != CONST_INT &&
			dc.header.typ != db.header.typ {
			return errors.New("Fate ADD: Type Error")
		}
		va := recent(big.Int)
		vb := recent(big.Int)
		vc := recent(big.Int)
		if _, ok := vb.SetString(string(effByte(append(mixed.E64func(db.header.offset0),
			append(mixed.E64func(db.header.offset1),
				mixed.E64func(db.header.offset2)...)...))), 0); !ok {
			err = errors.New("Illegal Value")
			return fmt.Errorf("Fate ADD: %v\n", err)
		}
		if _, ok := vc.SetString(string(effByte(append(mixed.E64func(dc.header.offset0),
			append(mixed.E64func(dc.header.offset1),
				mixed.E64func(dc.header.offset2)...)...))), 0); !ok {
			err = errors.New("Illegal Value")
			return fmt.Errorf("Fate ADD: %v\n", err)
		}
		va.Add(vb, vc)
		switch {
		case va.IsInt64():
			if va.Cmp(MinInt64) < 0 || va.Cmp(MaxInt64) > 0 {
				return errors.New("Fate ADD: Overflows")
			}
		default:
			if va.Cmp(MaxUint64) > 0 {
				return errors.New("Fate ADD: Overflows")
			}
		}
		return da.Update(e, va)
	case INT8, INT16, INT32, INT64, UINT8, UINT16, UINT32, UINT64:
		if db.header.typ != da.header.typ && db.header.typ != CONST_INT {
			return errors.New("Fate ADD: Type Error")
		}
		if dc.header.typ != da.header.typ && dc.header.typ != CONST_INT {
			return errors.New("Fate ADD: Type Error")
		}
		va := recent(big.Int)
		vb := recent(big.Int)
		vc := recent(big.Int)
		if _, ok := vb.SetString(string(effByte(append(mixed.E64func(db.header.offset0),
			append(mixed.E64func(db.header.offset1),
				mixed.E64func(db.header.offset2)...)...))), 0); !ok {
			err = errors.New("Illegal Value")
			return fmt.Errorf("Fate ADD: %v\n", err)
		}
		if _, ok := vc.SetString(string(effByte(append(mixed.E64func(dc.header.offset0),
			append(mixed.E64func(dc.header.offset1),
				mixed.E64func(dc.header.offset2)...)...))), 0); !ok {
			err = errors.New("Illegal Value")
			return fmt.Errorf("Fate ADD: %v\n", err)
		}
		va.Add(vb, vc)
		switch da.header.typ {
		case INT8, INT16, INT32, INT64:
			switch da.header.length {
			case 1:
				if va.Cmp(MinInt8) < 0 || va.Cmp(MaxInt8) > 0 {
					return errors.New("Fate ADD: Overflows")
				}
			case 2:
				if va.Cmp(MinInt16) < 0 || va.Cmp(MaxInt16) > 0 {
					return errors.New("Fate ADD: Overflows")
				}
			case 4:
				if va.Cmp(MinInt32) < 0 || va.Cmp(MaxInt32) > 0 {
					return errors.New("Fate ADD: Overflows")
				}
			case 8:
				if va.Cmp(MinInt64) < 0 || va.Cmp(MaxInt64) > 0 {
					return errors.New("Fate ADD: Overflows")
				}
			}
		default:
			switch da.header.length {
			case 1:
				if va.Cmp(MaxUint8) > 0 {
					return errors.New("Fate ADD: Overflows")
				}
			case 2:
				if va.Cmp(MaxUint16) > 0 {
					return errors.New("Fate ADD: Overflows")
				}
			case 4:
				if va.Cmp(MaxUint32) > 0 {
					return errors.New("Fate ADD: Overflows")
				}
			case 8:
				if va.Cmp(MaxUint64) > 0 {
					return errors.New("Fate ADD: Overflows")
				}
			}
		}
		return da.Update(e, va)
	default:
		return errors.New("Fate ADD: Unsupport type")
	}
}

// SUB  --->  a = b - c
func rsc14(e *contractEngine, op uint64) error {
	a := achieveReg(achieveA(op), e.regs.offset)
	b := achieveReg(achieveB(op), e.regs.offset)
	c := achieveReg(achieveC(op), e.regs.offset)
	defer func() {
		e.regs.pc = e.regs.pc + 1
	}()
	da, err := e.achieveData(e.regs.gpRegs[a])
	if err != nil {
		return fmt.Errorf("Fate SUB: %v", err)
	}
	db, err := e.achieveData(e.regs.gpRegs[b])
	if err != nil {
		return fmt.Errorf("Fate SUB: %v", err)
	}
	dc, err := e.achieveData(e.regs.gpRegs[c])
	if err != nil {
		return fmt.Errorf("Fate SUB: %v", err)
	}
	switch da.header.typ {
	case CHAR:
		if db.header.typ != da.header.typ && db.header.typ != CONST_CHAR {
			return errors.New("Fate SUB: Type Error")
		}
		if dc.header.typ != da.header.typ && dc.header.typ != CONST_CHAR {
			return errors.New("Fate SUB: Type Error")
		}
		vb := byte(db.header.offset0)
		vc := byte(dc.header.offset0)
		return da.Update(e, byte(vb-vc))
	case CONST_CHAR:
		if db.header.typ != da.header.typ && db.header.typ != CHAR {
			return errors.New("Fate SUB: Type Error")
		}
		if dc.header.typ != da.header.typ && dc.header.typ != CHAR {
			return errors.New("Fate SUB: Type Error")
		}
		vb := byte(db.header.offset0)
		vc := byte(dc.header.offset0)
		return da.Update(e, byte(vb-vc))
	case CONST_FLOAT:
		if typ := db.header.typ; typ != da.header.typ && typ != FLOAT32 && typ != FLOAT64 {
			return errors.New("Fate SUB: Type Error")
		}
		if typ := dc.header.typ; typ != da.header.typ && typ != FLOAT32 && typ != FLOAT64 {
			return errors.New("Fate SUB: Type Error")
		}
		if db.header.typ != CONST_FLOAT && dc.header.typ != CONST_FLOAT &&
			dc.header.typ != db.header.typ {
			return errors.New("Fate SUB: Type Error")
		}
		va := recent(big.Float)
		vb := recent(big.Float)
		vc := recent(big.Float)
		if _, ok := vb.SetString(string(effByte(append(mixed.E64func(db.header.offset0),
			append(mixed.E64func(db.header.offset1),
				mixed.E64func(db.header.offset2)...)...)))); !ok {
			err = errors.New("Illegal Value")
			return fmt.Errorf("Fate SUB: %v\n", err)
		}
		if _, ok := vc.SetString(string(effByte(append(mixed.E64func(dc.header.offset0),
			append(mixed.E64func(dc.header.offset1),
				mixed.E64func(dc.header.offset2)...)...)))); !ok {
			err = errors.New("Illegal Value")
			return fmt.Errorf("Fate SUB: %v\n", err)
		}
		va.Sub(vb, vc)
		if va.Cmp(MinFloat64) < 0 || va.Cmp(MaxFloat64) > 0 {
			return errors.New("Fate SUB: Overflows")
		}
		return da.Update(e, va)
	case FLOAT32, FLOAT64:
		if db.header.typ != da.header.typ && db.header.typ != CONST_FLOAT {
			return errors.New("Fate SUB: Type Error")
		}
		if dc.header.typ != da.header.typ && dc.header.typ != CONST_FLOAT {
			return errors.New("Fate SUB: Type Error")
		}
		va := recent(big.Float)
		vb := recent(big.Float)
		vc := recent(big.Float)
		if _, ok := vb.SetString(string(effByte(append(mixed.E64func(db.header.offset0),
			append(mixed.E64func(db.header.offset1),
				mixed.E64func(db.header.offset2)...)...)))); !ok {
			err = errors.New("Illegal Value")
			return fmt.Errorf("Fate SUB: %v\n", err)
		}
		if _, ok := vc.SetString(string(effByte(append(mixed.E64func(dc.header.offset0),
			append(mixed.E64func(dc.header.offset1),
				mixed.E64func(dc.header.offset2)...)...)))); !ok {
			err = errors.New("Illegal Value")
			return fmt.Errorf("Fate SUB: %v\n", err)
		}
		va.Sub(vb, vc)
		switch da.header.length {
		case 4:
			if va.Cmp(MinFloat32) < 0 || va.Cmp(MaxFloat32) > 0 {
				return errors.New("Fate SUB: Overflows")
			}
		case 8:
			if va.Cmp(MinFloat64) < 0 || va.Cmp(MaxFloat64) > 0 {
				return errors.New("Fate SUB: Overflows")
			}
		}
		return da.Update(e, va)
	case CONST_INT:
		if typ := db.header.typ; typ != da.header.typ && typ != INT8 && typ != INT16 &&
			typ != INT32 && typ != INT64 && typ != UINT8 && typ != UINT16 &&
			typ != UINT32 && typ != UINT64 {
			return errors.New("Fate SUB: Type Error")
		}
		if typ := dc.header.typ; typ != da.header.typ && typ != INT8 && typ != INT16 &&
			typ != INT32 && typ != INT64 && typ != UINT8 && typ != UINT16 &&
			typ != UINT32 && typ != UINT64 {
			return errors.New("Fate SUB: Type Error")
		}
		if db.header.typ != CONST_INT && dc.header.typ != CONST_INT &&
			dc.header.typ != db.header.typ {
			return errors.New("Fate SUB: Type Error")
		}
		va := recent(big.Int)
		vb := recent(big.Int)
		vc := recent(big.Int)
		if _, ok := vb.SetString(string(effByte(append(mixed.E64func(db.header.offset0),
			append(mixed.E64func(db.header.offset1),
				mixed.E64func(db.header.offset2)...)...))), 0); !ok {
			err = errors.New("Illegal Value")
			return fmt.Errorf("Fate SUB: %v\n", err)
		}
		if _, ok := vc.SetString(string(effByte(append(mixed.E64func(dc.header.offset0),
			append(mixed.E64func(dc.header.offset1),
				mixed.E64func(dc.header.offset2)...)...))), 0); !ok {
			err = errors.New("Illegal Value")
			return fmt.Errorf("Fate SUB: %v\n", err)
		}
		va.Sub(vb, vc)
		switch {
		case va.IsInt64():
			if va.Cmp(MinInt64) < 0 || va.Cmp(MaxInt64) > 0 {
				return errors.New("Fate SUB: Overflows")
			}
		default:
			if va.Cmp(MaxUint64) > 0 {
				return errors.New("Fate SUB: Overflows")
			}
		}
		return da.Update(e, va)
	case INT8, INT16, INT32, INT64, UINT8, UINT16, UINT32, UINT64:
		if db.header.typ != da.header.typ && db.header.typ != CONST_INT {
			return errors.New("Fate SUB: Type Error")
		}
		if dc.header.typ != da.header.typ && dc.header.typ != CONST_INT {
			return errors.New("Fate SUB: Type Error")
		}
		va := recent(big.Int)
		vb := recent(big.Int)
		vc := recent(big.Int)
		if _, ok := vb.SetString(string(effByte(append(mixed.E64func(db.header.offset0),
			append(mixed.E64func(db.header.offset1),
				mixed.E64func(db.header.offset2)...)...))), 0); !ok {
			err = errors.New("Illegal Value")
			return fmt.Errorf("Fate SUB: %v\n", err)
		}
		if _, ok := vc.SetString(string(effByte(append(mixed.E64func(dc.header.offset0),
			append(mixed.E64func(dc.header.offset1),
				mixed.E64func(dc.header.offset2)...)...))), 0); !ok {
			err = errors.New("Illegal Value")
			return fmt.Errorf("Fate SUB: %v\n", err)
		}
		va.Sub(vb, vc)
		switch da.header.typ {
		case INT8, INT16, INT32, INT64:
			switch da.header.length {
			case 1:
				if va.Cmp(MinInt8) < 0 || va.Cmp(MaxInt8) > 0 {
					return errors.New("Fate SUB: Overflows")
				}
			case 2:
				if va.Cmp(MinInt16) < 0 || va.Cmp(MaxInt16) > 0 {
					return errors.New("Fate SUB: Overflows")
				}
			case 4:
				if va.Cmp(MinInt32) < 0 || va.Cmp(MaxInt32) > 0 {
					return errors.New("Fate SUB: Overflows")
				}
			case 8:
				if va.Cmp(MinInt64) < 0 || va.Cmp(MaxInt64) > 0 {
					return errors.New("Fate SUB: Overflows")
				}
			}
		default:
			switch da.header.length {
			case 1:
				if va.Cmp(MaxUint8) > 0 {
					return errors.New("Fate SUB: Overflows")
				}
			case 2:
				if va.Cmp(MaxUint16) > 0 {
					return errors.New("Fate SUB: Overflows")
				}
			case 4:
				if va.Cmp(MaxUint32) > 0 {
					return errors.New("Fate SUB: Overflows")
				}
			case 8:
				if va.Cmp(MaxUint64) > 0 {
					return errors.New("Fate SUB: Overflows")
				}
			}
		}

		return da.Update(e, va)
	default:
		return errors.New("Fate SUB: Unsupport type")
	}
}

// MUL  --->  a = b * c
func rsc15(e *contractEngine, op uint64) error {
	a := achieveReg(achieveA(op), e.regs.offset)
	b := achieveReg(achieveB(op), e.regs.offset)
	c := achieveReg(achieveC(op), e.regs.offset)
	defer func() {
		e.regs.pc = e.regs.pc + 1
	}()
	da, err := e.achieveData(e.regs.gpRegs[a])
	if err != nil {
		return fmt.Errorf("Fate MUL: %v", err)
	}
	db, err := e.achieveData(e.regs.gpRegs[b])
	if err != nil {
		return fmt.Errorf("Fate MUL: %v", err)
	}
	dc, err := e.achieveData(e.regs.gpRegs[c])
	if err != nil {
		return fmt.Errorf("Fate MUL: %v", err)
	}
	switch da.header.typ {
	case CHAR:
		if db.header.typ != da.header.typ && db.header.typ != CONST_CHAR {
			return errors.New("Fate MUL: Type Error")
		}
		if dc.header.typ != da.header.typ && dc.header.typ != CONST_CHAR {
			return errors.New("Fate MUL: Type Error")
		}
		vb := byte(db.header.offset0)
		vc := byte(dc.header.offset0)
		return da.Update(e, byte(vb*vc))
	case CONST_CHAR:
		if db.header.typ != da.header.typ && db.header.typ != CHAR {
			return errors.New("Fate MUL: Type Error")
		}
		if dc.header.typ != da.header.typ && dc.header.typ != CHAR {
			return errors.New("Fate MUL: Type Error")
		}
		vb := byte(db.header.offset0)
		vc := byte(dc.header.offset0)
		return da.Update(e, byte(vb*vc))
	case CONST_FLOAT:
		if typ := db.header.typ; typ != da.header.typ && typ != FLOAT32 && typ != FLOAT64 {
			return errors.New("Fate MUL: Type Error")
		}
		if typ := dc.header.typ; typ != da.header.typ && typ != FLOAT32 && typ != FLOAT64 {
			return errors.New("Fate MUL: Type Error")
		}
		if db.header.typ != CONST_FLOAT && dc.header.typ != CONST_FLOAT &&
			dc.header.typ != db.header.typ {
			return errors.New("Fate MUL: Type Error")
		}
		va := recent(big.Float)
		vb := recent(big.Float)
		vc := recent(big.Float)
		if _, ok := vb.SetString(string(effByte(append(mixed.E64func(db.header.offset0),
			append(mixed.E64func(db.header.offset1),
				mixed.E64func(db.header.offset2)...)...)))); !ok {
			err = errors.New("Illegal Value")
			return fmt.Errorf("Fate MUL: %v\n", err)
		}
		if _, ok := vc.SetString(string(effByte(append(mixed.E64func(dc.header.offset0),
			append(mixed.E64func(dc.header.offset1),
				mixed.E64func(dc.header.offset2)...)...)))); !ok {
			err = errors.New("Illegal Value")
			return fmt.Errorf("Fate MUL: %v\n", err)
		}
		va.Mul(vb, vc)
		if va.Cmp(MinFloat64) < 0 || va.Cmp(MaxFloat64) > 0 {
			return errors.New("Fate MUL: Overflows")
		}
		return da.Update(e, va)
	case FLOAT32, FLOAT64:
		if db.header.typ != da.header.typ && db.header.typ != CONST_FLOAT {
			return errors.New("Fate MUL: Type Error")
		}
		if dc.header.typ != da.header.typ && dc.header.typ != CONST_FLOAT {
			return errors.New("Fate MUL: Type Error")
		}
		va := recent(big.Float)
		vb := recent(big.Float)
		vc := recent(big.Float)
		if _, ok := vb.SetString(string(effByte(append(mixed.E64func(db.header.offset0),
			append(mixed.E64func(db.header.offset1),
				mixed.E64func(db.header.offset2)...)...)))); !ok {
			err = errors.New("Illegal Value")
			return fmt.Errorf("Fate MUL: %v\n", err)
		}
		if _, ok := vc.SetString(string(effByte(append(mixed.E64func(dc.header.offset0),
			append(mixed.E64func(dc.header.offset1),
				mixed.E64func(dc.header.offset2)...)...)))); !ok {
			err = errors.New("Illegal Value")
			return fmt.Errorf("Fate MUL: %v\n", err)
		}
		va.Mul(vb, vc)
		switch da.header.length {
		case 4:
			if va.Cmp(MinFloat32) < 0 || va.Cmp(MaxFloat32) > 0 {
				return errors.New("Fate MUL: Overflows")
			}
		case 8:
			if va.Cmp(MinFloat64) < 0 || va.Cmp(MaxFloat64) > 0 {
				return errors.New("Fate MUL: Overflows")
			}
		}
		return da.Update(e, va)
	case CONST_INT:
		if typ := db.header.typ; typ != da.header.typ && typ != INT8 && typ != INT16 &&
			typ != INT32 && typ != INT64 && typ != UINT8 && typ != UINT16 &&
			typ != UINT32 && typ != UINT64 {
			return errors.New("Fate MUL: Type Error")
		}
		if typ := dc.header.typ; typ != da.header.typ && typ != INT8 && typ != INT16 &&
			typ != INT32 && typ != INT64 && typ != UINT8 && typ != UINT16 &&
			typ != UINT32 && typ != UINT64 {
			return errors.New("Fate MUL: Type Error")
		}
		if db.header.typ != CONST_INT && dc.header.typ != CONST_INT &&
			dc.header.typ != db.header.typ {
			return errors.New("Fate MUL: Type Error")
		}
		va := recent(big.Int)
		vb := recent(big.Int)
		vc := recent(big.Int)
		if _, ok := vb.SetString(string(effByte(append(mixed.E64func(db.header.offset0),
			append(mixed.E64func(db.header.offset1),
				mixed.E64func(db.header.offset2)...)...))), 0); !ok {
			err = errors.New("Illegal Value")
			return fmt.Errorf("Fate MUL: %v\n", err)
		}
		if _, ok := vc.SetString(string(effByte(append(mixed.E64func(dc.header.offset0),
			append(mixed.E64func(dc.header.offset1),
				mixed.E64func(dc.header.offset2)...)...))), 0); !ok {
			err = errors.New("Illegal Value")
			return fmt.Errorf("Fate MUL: %v\n", err)
		}
		va.Mul(vb, vc)
		switch {
		case va.IsInt64():
			if va.Cmp(MinInt64) < 0 || va.Cmp(MaxInt64) > 0 {
				return errors.New("Fate MUL: Overflows")
			}
		default:
			if va.Cmp(MaxUint64) > 0 {
				return errors.New("Fate MUL: Overflows")
			}
		}
		return da.Update(e, va)
	case INT8, INT16, INT32, INT64, UINT8, UINT16, UINT32, UINT64:
		if db.header.typ != da.header.typ && db.header.typ != CONST_INT {
			return errors.New("Fate MUL: Type Error")
		}
		if dc.header.typ != da.header.typ && dc.header.typ != CONST_INT {
			return errors.New("Fate MUL: Type Error")
		}
		va := recent(big.Int)
		vb := recent(big.Int)
		vc := recent(big.Int)
		if _, ok := vb.SetString(string(effByte(append(mixed.E64func(db.header.offset0),
			append(mixed.E64func(db.header.offset1),
				mixed.E64func(db.header.offset2)...)...))), 0); !ok {
			err = errors.New("Illegal Value")
			return fmt.Errorf("Fate MUL: %v\n", err)
		}
		if _, ok := vc.SetString(string(effByte(append(mixed.E64func(dc.header.offset0),
			append(mixed.E64func(dc.header.offset1),
				mixed.E64func(dc.header.offset2)...)...))), 0); !ok {
			err = errors.New("Illegal Value")
			return fmt.Errorf("Fate MUL: %v\n", err)
		}
		va.Mul(vb, vc)
		switch da.header.typ {
		case INT8, INT16, INT32, INT64:
			switch da.header.length {
			case 1:
				if va.Cmp(MinInt8) < 0 || va.Cmp(MaxInt8) > 0 {
					return errors.New("Fate MUL: Overflows")
				}
			case 2:
				if va.Cmp(MinInt16) < 0 || va.Cmp(MaxInt16) > 0 {
					return errors.New("Fate MUL: Overflows")
				}
			case 4:
				if va.Cmp(MinInt32) < 0 || va.Cmp(MaxInt32) > 0 {
					return errors.New("Fate MUL: Overflows")
				}
			case 8:
				if va.Cmp(MinInt64) < 0 || va.Cmp(MaxInt64) > 0 {
					return errors.New("Fate MUL: Overflows")
				}
			}
		default:
			switch da.header.length {
			case 1:
				if va.Cmp(MaxUint8) > 0 {
					return errors.New("Fate MUL: Overflows")
				}
			case 2:
				if va.Cmp(MaxUint16) > 0 {
					return errors.New("Fate MUL: Overflows")
				}
			case 4:
				if va.Cmp(MaxUint32) > 0 {
					return errors.New("Fate MUL: Overflows")
				}
			case 8:
				if va.Cmp(MaxUint64) > 0 {
					return errors.New("Fate MUL: Overflows")
				}
			}
		}
		return da.Update(e, va)
	default:
		return errors.New("Fate MUL: Unsupport type")
	}
}

func rsc16(e *contractEngine, op uint64) error {
	a := achieveReg(achieveA(op), e.regs.offset)
	b := achieveReg(achieveB(op), e.regs.offset)
	c := achieveReg(achieveC(op), e.regs.offset)
	defer func() {
		e.regs.pc = e.regs.pc + 1
	}()
	da, err := e.achieveData(e.regs.gpRegs[a])
	if err != nil {
		return fmt.Errorf("Fate DIV: %v", err)
	}
	db, err := e.achieveData(e.regs.gpRegs[b])
	if err != nil {
		return fmt.Errorf("Fate DIV: %v", err)
	}
	dc, err := e.achieveData(e.regs.gpRegs[c])
	if err != nil {
		return fmt.Errorf("Fate DIV: %v", err)
	}
	switch da.header.typ {
	case CHAR:
		if dc.header.offset0 == 0 {
			return errors.New("Fate DIV: Divide Zero")
		}
		if db.header.typ != da.header.typ && db.header.typ != CONST_CHAR {
			return errors.New("Fate DIV: Type Error")
		}
		if dc.header.typ != da.header.typ && dc.header.typ != CONST_CHAR {
			return errors.New("Fate DIV: Type Error")
		}
		vb := byte(db.header.offset0)
		vc := byte(dc.header.offset0)
		return da.Update(e, byte(vb/vc))
	case CONST_CHAR:
		if dc.header.offset0 == 0 {
			return errors.New("Fate DIV: Divide Zero")
		}
		if db.header.typ != da.header.typ && db.header.typ != CHAR {
			return errors.New("Fate DIV: Type Error")
		}
		if dc.header.typ != da.header.typ && dc.header.typ != CHAR {
			return errors.New("Fate DIV: Type Error")
		}
		vb := byte(db.header.offset0)
		vc := byte(dc.header.offset0)
		return da.Update(e, byte(vb/vc))
	case CONST_FLOAT:
		if dc.header.offset0 == 48 {
			return errors.New("Fate DIV: Divide Zero")
		}
		if typ := db.header.typ; typ != da.header.typ && typ != FLOAT32 && typ != FLOAT64 {
			return errors.New("Fate DIV: Type Error")
		}
		if typ := dc.header.typ; typ != da.header.typ && typ != FLOAT32 && typ != FLOAT64 {
			return errors.New("Fate DIV: Type Error")
		}
		if db.header.typ != CONST_FLOAT && dc.header.typ != CONST_FLOAT &&
			dc.header.typ != db.header.typ {
			return errors.New("Fate DIV: Type Error")
		}
		va := recent(big.Float)
		vb := recent(big.Float)
		vc := recent(big.Float)
		if _, ok := vb.SetString(string(effByte(append(mixed.E64func(db.header.offset0),
			append(mixed.E64func(db.header.offset1),
				mixed.E64func(db.header.offset2)...)...)))); !ok {
			err = errors.New("Illegal Value")
			return fmt.Errorf("Fate DIV: %v\n", err)
		}
		if _, ok := vc.SetString(string(effByte(append(mixed.E64func(dc.header.offset0),
			append(mixed.E64func(dc.header.offset1),
				mixed.E64func(dc.header.offset2)...)...)))); !ok {
			err = errors.New("Illegal Value")
			return fmt.Errorf("Fate DIV: %v\n", err)
		}
		va.Quo(vb, vc)
		if va.Cmp(MinFloat64) < 0 || va.Cmp(MaxFloat64) > 0 {
			return errors.New("Fate DIV: Overflows")
		}
		return da.Update(e, va)
	case FLOAT32, FLOAT64:
		if dc.header.offset0 == 48 {
			return errors.New("Fate DIV: Divide Zero")
		}
		if db.header.typ != da.header.typ && db.header.typ != CONST_FLOAT {
			return errors.New("Fate DIV: Type Error")
		}
		if dc.header.typ != da.header.typ && dc.header.typ != CONST_FLOAT {
			return errors.New("Fate DIV: Type Error")
		}
		va := recent(big.Float)
		vb := recent(big.Float)
		vc := recent(big.Float)
		if _, ok := vb.SetString(string(effByte(append(mixed.E64func(db.header.offset0),
			append(mixed.E64func(db.header.offset1),
				mixed.E64func(db.header.offset2)...)...)))); !ok {
			err = errors.New("Illegal Value")
			return fmt.Errorf("Fate DIV: %v\n", err)
		}
		if _, ok := vc.SetString(string(effByte(append(mixed.E64func(dc.header.offset0),
			append(mixed.E64func(dc.header.offset1),
				mixed.E64func(dc.header.offset2)...)...)))); !ok {
			err = errors.New("Illegal Value")
			return fmt.Errorf("Fate DIV: %v\n", err)
		}
		va.Quo(vb, vc)
		switch da.header.length {
		case 4:
			if va.Cmp(MinFloat32) < 0 || va.Cmp(MaxFloat32) > 0 {
				return errors.New("Fate DIV: Overflows")
			}
		case 8:
			if va.Cmp(MinFloat64) < 0 || va.Cmp(MaxFloat64) > 0 {
				return errors.New("Fate DIV: Overflows")
			}
		}
		return da.Update(e, va)
	case CONST_INT:
		if dc.header.offset0 == 48 {
			return errors.New("Fate DIV: Divide Zero")
		}
		if typ := db.header.typ; typ != da.header.typ && typ != INT8 && typ != INT16 &&
			typ != INT32 && typ != INT64 && typ != UINT8 && typ != UINT16 &&
			typ != UINT32 && typ != UINT64 {
			return errors.New("Fate DIV: Type Error")
		}
		if typ := dc.header.typ; typ != da.header.typ && typ != INT8 && typ != INT16 &&
			typ != INT32 && typ != INT64 && typ != UINT8 && typ != UINT16 &&
			typ != UINT32 && typ != UINT64 {
			return errors.New("Fate DIV: Type Error")
		}
		if db.header.typ != CONST_INT && dc.header.typ != CONST_INT &&
			dc.header.typ != db.header.typ {
			return errors.New("Fate DIV: Type Error")
		}
		va := recent(big.Int)
		vb := recent(big.Int)
		vc := recent(big.Int)
		if _, ok := vb.SetString(string(effByte(append(mixed.E64func(db.header.offset0),
			append(mixed.E64func(db.header.offset1),
				mixed.E64func(db.header.offset2)...)...))), 0); !ok {
			err = errors.New("Illegal Value")
			return fmt.Errorf("Fate DIV: %v\n", err)
		}
		if _, ok := vc.SetString(string(effByte(append(mixed.E64func(dc.header.offset0),
			append(mixed.E64func(dc.header.offset1),
				mixed.E64func(dc.header.offset2)...)...))), 0); !ok {
			err = errors.New("Illegal Value")
			return fmt.Errorf("Fate DIV: %v\n", err)
		}
		va.Div(vb, vc)
		switch {
		case va.IsInt64():
			if va.Cmp(MinInt64) < 0 || va.Cmp(MaxInt64) > 0 {
				return errors.New("Fate DIV: Overflows")
			}
		default:
			if va.Cmp(MaxUint64) > 0 {
				return errors.New("Fate DIV: Overflows")
			}
		}
		return da.Update(e, va)
	case INT8, INT16, INT32, INT64, UINT8, UINT16, UINT32, UINT64:
		if dc.header.offset0 == 48 {
			return errors.New("Fate DIV: Divide Zero")
		}
		if db.header.typ != da.header.typ && db.header.typ != CONST_INT {
			return errors.New("Fate DIV: Type Error")
		}
		if dc.header.typ != da.header.typ && dc.header.typ != CONST_INT {
			return errors.New("Fate DIV: Type Error")
		}
		va := recent(big.Int)
		vb := recent(big.Int)
		vc := recent(big.Int)
		if _, ok := vb.SetString(string(effByte(append(mixed.E64func(db.header.offset0),
			append(mixed.E64func(db.header.offset1),
				mixed.E64func(db.header.offset2)...)...))), 0); !ok {
			err = errors.New("Illegal Value")
			return fmt.Errorf("Fate DIV: %v\n", err)
		}
		if _, ok := vc.SetString(string(effByte(append(mixed.E64func(dc.header.offset0),
			append(mixed.E64func(dc.header.offset1),
				mixed.E64func(dc.header.offset2)...)...))), 0); !ok {
			err = errors.New("Illegal Value")
			return fmt.Errorf("Fate DIV: %v\n", err)
		}
		va.Div(vb, vc)
		switch da.header.typ {
		case INT8, INT16, INT32, INT64:
			switch da.header.length {
			case 1:
				if va.Cmp(MinInt8) < 0 || va.Cmp(MaxInt8) > 0 {
					return errors.New("Fate DIV: Overflows")
				}
			case 2:
				if va.Cmp(MinInt16) < 0 || va.Cmp(MaxInt16) > 0 {
					return errors.New("Fate DIV: Overflows")
				}
			case 4:
				if va.Cmp(MinInt32) < 0 || va.Cmp(MaxInt32) > 0 {
					return errors.New("Fate DIV: Overflows")
				}
			case 8:
				if va.Cmp(MinInt64) < 0 || va.Cmp(MaxInt64) > 0 {
					return errors.New("Fate DIV: Overflows")
				}
			}
		default:
			switch da.header.length {
			case 1:
				if va.Cmp(MaxUint8) > 0 {
					return errors.New("Fate DIV: Overflows")
				}
			case 2:
				if va.Cmp(MaxUint16) > 0 {
					return errors.New("Fate DIV: Overflows")
				}
			case 4:
				if va.Cmp(MaxUint32) > 0 {
					return errors.New("Fate DIV: Overflows")
				}
			case 8:
				if va.Cmp(MaxUint64) > 0 {
					return errors.New("Fate DIV: Overflows")
				}
			}
		}
		return da.Update(e, va)
	default:
		return errors.New("Fate DIV: Unsupport type")
	}
}

// MOD  --->  a = b % c
func rsc17(e *contractEngine, op uint64) error {
	a := achieveReg(achieveA(op), e.regs.offset)
	b := achieveReg(achieveB(op), e.regs.offset)
	c := achieveReg(achieveC(op), e.regs.offset)
	defer func() {
		e.regs.pc = e.regs.pc + 1
	}()
	da, err := e.achieveData(e.regs.gpRegs[a])
	if err != nil {
		return fmt.Errorf("Fate MOD: %v", err)
	}
	db, err := e.achieveData(e.regs.gpRegs[b])
	if err != nil {
		return fmt.Errorf("Fate MOD: %v", err)
	}
	dc, err := e.achieveData(e.regs.gpRegs[c])
	if err != nil {
		return fmt.Errorf("Fate MOD: %v", err)
	}
	switch da.header.typ {
	case CHAR:
		if dc.header.offset0 == 0 {
			return errors.New("Fate MOD: Divide Zero")
		}
		if db.header.typ != da.header.typ && db.header.typ != CONST_CHAR {
			return errors.New("Fate MOD: Type Error")
		}
		if dc.header.typ != da.header.typ && dc.header.typ != CONST_CHAR {
			return errors.New("Fate MOD: Type Error")
		}
		vb := byte(db.header.offset0)
		vc := byte(dc.header.offset0)
		return da.Update(e, byte(vb%vc))
	case CONST_CHAR:
		if dc.header.offset0 == 0 {
			return errors.New("Fate MOD: Divide Zero")
		}
		if db.header.typ != da.header.typ && db.header.typ != CHAR {
			return errors.New("Fate MOD: Type Error")
		}
		if dc.header.typ != da.header.typ && dc.header.typ != CHAR {
			return errors.New("Fate MOD: Type Error")
		}
		vb := byte(db.header.offset0)
		vc := byte(dc.header.offset0)
		return da.Update(e, byte(vb%vc))
	case CONST_INT:
		if dc.header.offset0 == 48 {
			return errors.New("Fate MOD: Divide Zero")
		}
		if typ := db.header.typ; typ != da.header.typ && typ != INT8 && typ != INT16 &&
			typ != INT32 && typ != INT64 && typ != UINT8 && typ != UINT16 &&
			typ != UINT32 && typ != UINT64 {
			return errors.New("Fate MOD: Type Error")
		}
		if typ := dc.header.typ; typ != da.header.typ && typ != INT8 && typ != INT16 &&
			typ != INT32 && typ != INT64 && typ != UINT8 && typ != UINT16 &&
			typ != UINT32 && typ != UINT64 {
			return errors.New("Fate MOD: Type Error")
		}
		if db.header.typ != CONST_INT && dc.header.typ != CONST_INT &&
			dc.header.typ != db.header.typ {
			return errors.New("Fate MOD: Type Error")
		}
		va := recent(big.Int)
		vb := recent(big.Int)
		vc := recent(big.Int)
		if _, ok := vb.SetString(string(effByte(append(mixed.E64func(db.header.offset0),
			append(mixed.E64func(db.header.offset1),
				mixed.E64func(db.header.offset2)...)...))), 0); !ok {
			err = errors.New("Illegal Value")
			return fmt.Errorf("Fate MOD: %v\n", err)
		}
		if _, ok := vc.SetString(string(effByte(append(mixed.E64func(dc.header.offset0),
			append(mixed.E64func(dc.header.offset1),
				mixed.E64func(dc.header.offset2)...)...))), 0); !ok {
			err = errors.New("Illegal Value")
			return fmt.Errorf("Fate MOD: %v\n", err)
		}
		va.Mod(vb, vc)
		switch {
		case va.IsInt64():
			if va.Cmp(MinInt64) < 0 || va.Cmp(MaxInt64) > 0 {
				return errors.New("Fate MOD: Overflows")
			}
		default:
			if va.Cmp(MaxUint64) > 0 {
				return errors.New("Fate MOD: Overflows")
			}
		}
		return da.Update(e, va)
	case INT8, INT16, INT32, INT64, UINT8, UINT16, UINT32, UINT64:
		if dc.header.offset0 == 48 {
			return errors.New("Fate MOD: Divide Zero")
		}
		if db.header.typ != da.header.typ && db.header.typ != CONST_INT {
			return errors.New("Fate MOD: Type Error")
		}
		if dc.header.typ != da.header.typ && dc.header.typ != CONST_INT {
			return errors.New("Fate MOD: Type Error")
		}
		va := recent(big.Int)
		vb := recent(big.Int)
		vc := recent(big.Int)
		if _, ok := vb.SetString(string(effByte(append(mixed.E64func(db.header.offset0),
			append(mixed.E64func(db.header.offset1),
				mixed.E64func(db.header.offset2)...)...))), 0); !ok {
			err = errors.New("Illegal Value")
			return fmt.Errorf("Fate MOD: %v\n", err)
		}
		if _, ok := vc.SetString(string(effByte(append(mixed.E64func(dc.header.offset0),
			append(mixed.E64func(dc.header.offset1),
				mixed.E64func(dc.header.offset2)...)...))), 0); !ok {
			err = errors.New("Illegal Value")
			return fmt.Errorf("Fate MOD: %v\n", err)
		}
		va.Mod(vb, vc)
		switch da.header.typ {
		case INT8, INT16, INT32, INT64:
			switch da.header.length {
			case 1:
				if va.Cmp(MinInt8) < 0 || va.Cmp(MaxInt8) > 0 {
					return errors.New("Fate MOD: Overflows")
				}
			case 2:
				if va.Cmp(MinInt16) < 0 || va.Cmp(MaxInt16) > 0 {
					return errors.New("Fate MOD: Overflows")
				}
			case 4:
				if va.Cmp(MinInt32) < 0 || va.Cmp(MaxInt32) > 0 {
					return errors.New("Fate MOD: Overflows")
				}
			case 8:
				if va.Cmp(MinInt64) < 0 || va.Cmp(MaxInt64) > 0 {
					return errors.New("Fate MOD: Overflows")
				}
			}
		default:
			switch da.header.length {
			case 1:
				if va.Cmp(MaxUint8) > 0 {
					return errors.New("Fate MOD: Overflows")
				}
			case 2:
				if va.Cmp(MaxUint16) > 0 {
					return errors.New("Fate MOD: Overflows")
				}
			case 4:
				if va.Cmp(MaxUint32) > 0 {
					return errors.New("Fate MOD: Overflows")
				}
			case 8:
				if va.Cmp(MaxUint64) > 0 {
					return errors.New("Fate MOD: Overflows")
				}
			}
		}
		return da.Update(e, va)
	default:
		return errors.New("Fate MOD: Unsupport type")
	}
}

// SHL  --->  a = b << c
func rsc18(e *contractEngine, op uint64) error {
	a := achieveReg(achieveA(op), e.regs.offset)
	b := achieveReg(achieveB(op), e.regs.offset)
	c := achieveReg(achieveC(op), e.regs.offset)
	defer func() {
		e.regs.pc = e.regs.pc + 1
	}()
	da, err := e.achieveData(e.regs.gpRegs[a])
	if err != nil {
		return fmt.Errorf("Fate SHL: %v", err)
	}
	db, err := e.achieveData(e.regs.gpRegs[b])
	if err != nil {
		return fmt.Errorf("Fate SHL: %v", err)
	}
	dc, err := e.achieveData(e.regs.gpRegs[c])
	if err != nil {
		return fmt.Errorf("Fate SHL: %v", err)
	}
	switch da.header.typ {
	case CHAR:
		if db.header.typ != da.header.typ && db.header.typ != CONST_CHAR {
			return errors.New("Fate SHL: Type Error")
		}
		if dc.header.typ != da.header.typ && dc.header.typ != CONST_CHAR {
			return errors.New("Fate SHL: Type Error")
		}
		vb := byte(db.header.offset0)
		vc := byte(dc.header.offset0)
		return da.Update(e, byte(vb<<vc))
	case CONST_CHAR:
		if db.header.typ != da.header.typ && db.header.typ != CHAR {
			return errors.New("Fate SHL: Type Error")
		}
		if dc.header.typ != da.header.typ && dc.header.typ != CHAR {
			return errors.New("Fate SHL: Type Error")
		}
		vb := byte(db.header.offset0)
		vc := byte(dc.header.offset0)
		return da.Update(e, byte(vb<<vc))
	case CONST_INT:
		if typ := db.header.typ; typ != da.header.typ && typ != INT8 && typ != INT16 &&
			typ != INT32 && typ != INT64 && typ != UINT8 && typ != UINT16 &&
			typ != UINT32 && typ != UINT64 {
			return errors.New("Fate SHL: Type Error")
		}
		if typ := dc.header.typ; typ != da.header.typ && typ != INT8 && typ != INT16 &&
			typ != INT32 && typ != INT64 && typ != UINT8 && typ != UINT16 &&
			typ != UINT32 && typ != UINT64 {
			return errors.New("Fate SHL: Type Error")
		}
		if db.header.typ != CONST_INT && dc.header.typ != CONST_INT &&
			dc.header.typ != db.header.typ {
			return errors.New("Fate SHL: Type Error")
		}
		va := recent(big.Int)
		vb := recent(big.Int)
		vc := recent(big.Int)
		if _, ok := vb.SetString(string(effByte(append(mixed.E64func(db.header.offset0),
			append(mixed.E64func(db.header.offset1),
				mixed.E64func(db.header.offset2)...)...))), 0); !ok {
			err = errors.New("Illegal Value")
			return fmt.Errorf("Fate SHL: %v\n", err)
		}
		if _, ok := vc.SetString(string(effByte(append(mixed.E64func(dc.header.offset0),
			append(mixed.E64func(dc.header.offset1),
				mixed.E64func(dc.header.offset2)...)...))), 0); !ok {
			err = errors.New("Illegal Value")
			return fmt.Errorf("Fate SHL: %v\n", err)
		}
		va.Lsh(vb, uint(vc.Uint64()))
		switch {
		case va.IsInt64():
			if va.Cmp(MinInt64) < 0 || va.Cmp(MaxInt64) > 0 {
				return errors.New("Fate SHL: Overflows")
			}
		default:
			if va.Cmp(MaxUint64) > 0 {
				return errors.New("Fate SHL: Overflows")
			}
		}
		return da.Update(e, va)
	case INT8, INT16, INT32, INT64, UINT8, UINT16, UINT32, UINT64:
		if db.header.typ != da.header.typ && db.header.typ != CONST_INT {
			return errors.New("Fate SHL: Type Error")
		}
		if dc.header.typ != CONST_INT && dc.header.typ != UINT8 && dc.header.typ != UINT16 &&
			dc.header.typ != UINT32 && dc.header.typ != UINT64 {
			return errors.New("Fate SHL: Type Error")
		}
		va := recent(big.Int)
		vb := recent(big.Int)
		vc := recent(big.Int)
		if _, ok := vb.SetString(string(effByte(append(mixed.E64func(db.header.offset0),
			append(mixed.E64func(db.header.offset1),
				mixed.E64func(db.header.offset2)...)...))), 0); !ok {
			err = errors.New("Illegal Value")
			return fmt.Errorf("Fate SHL: %v\n", err)
		}
		if _, ok := vc.SetString(string(effByte(append(mixed.E64func(dc.header.offset0),
			append(mixed.E64func(dc.header.offset1),
				mixed.E64func(dc.header.offset2)...)...))), 0); !ok {
			err = errors.New("Illegal Value")
			return fmt.Errorf("Fate SHL: %v\n", err)
		}
		va.Lsh(vb, uint(vc.Uint64()))
		switch da.header.typ {
		case INT8, INT16, INT32, INT64:
			switch da.header.length {
			case 1:
				if va.Cmp(MinInt8) < 0 || va.Cmp(MaxInt8) > 0 {
					return errors.New("Fate SHL: Overflows")
				}
			case 2:
				if va.Cmp(MinInt16) < 0 || va.Cmp(MaxInt16) > 0 {
					return errors.New("Fate SHL: Overflows")
				}
			case 4:
				if va.Cmp(MinInt32) < 0 || va.Cmp(MaxInt32) > 0 {
					return errors.New("Fate SHL: Overflows")
				}
			case 8:
				if va.Cmp(MinInt64) < 0 || va.Cmp(MaxInt64) > 0 {
					return errors.New("Fate SHL: Overflows")
				}
			}
		default:
			switch da.header.length {
			case 1:
				if va.Cmp(MaxUint8) > 0 {
					return errors.New("Fate SHL: Overflows")
				}
			case 2:
				if va.Cmp(MaxUint16) > 0 {
					return errors.New("Fate SHL: Overflows")
				}
			case 4:
				if va.Cmp(MaxUint32) > 0 {
					return errors.New("Fate SHL: Overflows")
				}
			case 8:
				if va.Cmp(MaxUint64) > 0 {
					return errors.New("Fate SHL: Overflows")
				}
			}
		}
		return da.Update(e, va)
	default:
		return errors.New("Fate SHL: Unsupport type")
	}
}

// SHR  --->  a = b >> c
func rsc19(e *contractEngine, op uint64) error {
	a := achieveReg(achieveA(op), e.regs.offset)
	b := achieveReg(achieveB(op), e.regs.offset)
	c := achieveReg(achieveC(op), e.regs.offset)
	defer func() {
		e.regs.pc = e.regs.pc + 1
	}()
	da, err := e.achieveData(e.regs.gpRegs[a])
	if err != nil {
		return fmt.Errorf("Fate SHR: %v", err)
	}
	db, err := e.achieveData(e.regs.gpRegs[b])
	if err != nil {
		return fmt.Errorf("Fate SHR: %v", err)
	}
	dc, err := e.achieveData(e.regs.gpRegs[c])
	if err != nil {
		return fmt.Errorf("Fate SHR: %v", err)
	}
	switch da.header.typ {
	case CHAR:
		if db.header.typ != da.header.typ && db.header.typ != CONST_CHAR {
			return errors.New("Fate SHR: Type Error")
		}
		if dc.header.typ != da.header.typ && dc.header.typ != CONST_CHAR {
			return errors.New("Fate SHR: Type Error")
		}
		vb := byte(db.header.offset0)
		vc := byte(dc.header.offset0)
		return da.Update(e, byte(vb>>vc))
	case CONST_CHAR:
		if db.header.typ != da.header.typ && db.header.typ != CHAR {
			return errors.New("Fate SHR: Type Error")
		}
		if dc.header.typ != da.header.typ && dc.header.typ != CHAR {
			return errors.New("Fate SHR: Type Error")
		}
		vb := byte(db.header.offset0)
		vc := byte(dc.header.offset0)
		return da.Update(e, byte(vb>>vc))
	case CONST_INT:
		if typ := db.header.typ; typ != da.header.typ && typ != INT8 && typ != INT16 &&
			typ != INT32 && typ != INT64 && typ != UINT8 && typ != UINT16 &&
			typ != UINT32 && typ != UINT64 {
			return errors.New("Fate SHR: Type Error")
		}
		if typ := dc.header.typ; typ != da.header.typ && typ != INT8 && typ != INT16 &&
			typ != INT32 && typ != INT64 && typ != UINT8 && typ != UINT16 &&
			typ != UINT32 && typ != UINT64 {
			return errors.New("Fate SHR: Type Error")
		}
		if db.header.typ != CONST_INT && dc.header.typ != CONST_INT &&
			dc.header.typ != db.header.typ {
			return errors.New("Fate SHR: Type Error")
		}
		va := recent(big.Int)
		vb := recent(big.Int)
		vc := recent(big.Int)
		if _, ok := vb.SetString(string(effByte(append(mixed.E64func(db.header.offset0),
			append(mixed.E64func(db.header.offset1),
				mixed.E64func(db.header.offset2)...)...))), 0); !ok {
			err = errors.New("Illegal Value")
			return fmt.Errorf("Fate SHR: %v\n", err)
		}
		if _, ok := vc.SetString(string(effByte(append(mixed.E64func(dc.header.offset0),
			append(mixed.E64func(dc.header.offset1),
				mixed.E64func(dc.header.offset2)...)...))), 0); !ok {
			err = errors.New("Illegal Value")
			return fmt.Errorf("Fate SHR: %v\n", err)
		}
		va.Rsh(vb, uint(vc.Uint64()))
		switch {
		case va.IsInt64():
			if va.Cmp(MinInt64) < 0 || va.Cmp(MaxInt64) > 0 {
				return errors.New("Fate SHR: Overflows")
			}
		default:
			if va.Cmp(MaxUint64) > 0 {
				return errors.New("Fate SHR: Overflows")
			}
		}
		return da.Update(e, va)
	case INT8, INT16, INT32, INT64, UINT8, UINT16, UINT32, UINT64:
		if db.header.typ != da.header.typ && db.header.typ != CONST_INT {
			return errors.New("Fate SHR: Type Error")
		}
		if dc.header.typ != CONST_INT && dc.header.typ != UINT8 && dc.header.typ != UINT16 &&
			dc.header.typ != UINT32 && dc.header.typ != UINT64 {
			return errors.New("Fate SHR: Type Error")
		}
		va := recent(big.Int)
		vb := recent(big.Int)
		vc := recent(big.Int)
		if _, ok := vb.SetString(string(effByte(append(mixed.E64func(db.header.offset0),
			append(mixed.E64func(db.header.offset1),
				mixed.E64func(db.header.offset2)...)...))), 0); !ok {
			err = errors.New("Illegal Value")
			return fmt.Errorf("Fate SHR: %v\n", err)
		}
		if _, ok := vc.SetString(string(effByte(append(mixed.E64func(dc.header.offset0),
			append(mixed.E64func(dc.header.offset1),
				mixed.E64func(dc.header.offset2)...)...))), 0); !ok {
			err = errors.New("Illegal Value")
			return fmt.Errorf("Fate SHR: %v\n", err)
		}
		va.Rsh(vb, uint(vc.Uint64()))
		switch da.header.typ {
		case INT8, INT16, INT32, INT64:
			switch da.header.length {
			case 1:
				if va.Cmp(MinInt8) < 0 || va.Cmp(MaxInt8) > 0 {
					return errors.New("Fate SHR: Overflows")
				}
			case 2:
				if va.Cmp(MinInt16) < 0 || va.Cmp(MaxInt16) > 0 {
					return errors.New("Fate SHR: Overflows")
				}
			case 4:
				if va.Cmp(MinInt32) < 0 || va.Cmp(MaxInt32) > 0 {
					return errors.New("Fate SHR: Overflows")
				}
			case 8:
				if va.Cmp(MinInt64) < 0 || va.Cmp(MaxInt64) > 0 {
					return errors.New("Fate SHR: Overflows")
				}
			}
		default:
			switch da.header.length {
			case 1:
				if va.Cmp(MaxUint8) > 0 {
					return errors.New("Fate SHR: Overflows")
				}
			case 2:
				if va.Cmp(MaxUint16) > 0 {
					return errors.New("Fate SHR: Overflows")
				}
			case 4:
				if va.Cmp(MaxUint32) > 0 {
					return errors.New("Fate SHR: Overflows")
				}
			case 8:
				if va.Cmp(MaxUint64) > 0 {
					return errors.New("Fate SHR: Overflows")
				}
			}
		}
		return da.Update(e, va)
	default:
		return errors.New("Fate SHR: Unsupport type")
	}
}

// NEG  --->  a = -a
func rsc20(e *contractEngine, op uint64) error {
	a := achieveReg(achieveA(op), e.regs.offset)
	defer func() {
		e.regs.pc = e.regs.pc + 1
	}()
	da, err := e.achieveData(e.regs.gpRegs[a])
	if err != nil {
		return fmt.Errorf("Fate NEG: %v", err)
	}
	switch da.header.typ {
	case CONST_INT:
		va := recent(big.Int)
		if _, ok := va.SetString(string(effByte(append(mixed.E64func(da.header.offset0),
			append(mixed.E64func(da.header.offset1),
				mixed.E64func(da.header.offset2)...)...))), 0); !ok {
			return fmt.Errorf("Fate NEG: Illegal Value\n")
		}
		va.Neg(va)
		switch {
		case va.IsInt64():
			if va.Cmp(MinInt64) < 0 || va.Cmp(MaxInt64) > 0 {
				return errors.New("Fate NEG: Overflows")
			}
		default:
			if va.Cmp(MaxUint64) > 0 {
				return errors.New("Fate NEG: Overflows")
			}
		}
		return da.Update(e, va)
	case INT8, INT16, INT32, INT64:
		va := recent(big.Int)
		if _, ok := va.SetString(string(effByte(append(mixed.E64func(da.header.offset0),
			append(mixed.E64func(da.header.offset1),
				mixed.E64func(da.header.offset2)...)...))), 0); !ok {
			return fmt.Errorf("Fate NEG: Illegal Value\n")
		}
		va.Neg(va)
		switch da.header.length {
		case 1:
			if va.Cmp(MinInt8) < 0 || va.Cmp(MaxInt8) > 0 {
				return errors.New("Fate NEG: Overflows")
			}
		case 2:
			if va.Cmp(MinInt16) < 0 || va.Cmp(MaxInt16) > 0 {
				return errors.New("Fate NEG: Overflows")
			}
		case 4:
			if va.Cmp(MinInt32) < 0 || va.Cmp(MaxInt32) > 0 {
				return errors.New("Fate NEG: Overflows")
			}
		case 8:
			if va.Cmp(MinInt64) < 0 || va.Cmp(MaxInt64) > 0 {
				return errors.New("Fate NEG: Overflows")
			}
		}
		return da.Update(e, va)
	case FLOAT32, FLOAT64, CONST_FLOAT:
		va := recent(big.Float)
		if _, ok := va.SetString(string(effByte(append(mixed.E64func(da.header.offset0),
			append(mixed.E64func(da.header.offset1),
				mixed.E64func(da.header.offset2)...)...)))); !ok {
			return fmt.Errorf("Fate NEG: Illegal Value\n")
		}
		va.Neg(va)
		switch da.header.length {
		case 4:
			if va.Cmp(MinFloat32) < 0 || va.Cmp(MaxFloat32) > 0 {
				return errors.New("Fate NEG: Overflows")
			}
		case 8:
			if va.Cmp(MinFloat64) < 0 || va.Cmp(MaxFloat64) > 0 {
				return errors.New("Fate NEG: Overflows")
			}
		}
		return da.Update(e, va)
	default:
		return errors.New("Fate NEG: Unsupport type")
	}
}

// NOT  --->  a = ~a
func rsc21(e *contractEngine, op uint64) error {
	a := achieveReg(achieveA(op), e.regs.offset)
	defer func() {
		e.regs.pc = e.regs.pc + 1
	}()
	da, err := e.achieveData(e.regs.gpRegs[a])
	if err != nil {
		return fmt.Errorf("Fate NOT: %v", err)
	}
	switch da.header.typ {
	case CONST_INT:
		va := recent(big.Int)
		if _, ok := va.SetString(string(effByte(append(mixed.E64func(da.header.offset0),
			append(mixed.E64func(da.header.offset1),
				mixed.E64func(da.header.offset2)...)...))), 0); !ok {
			return fmt.Errorf("Fate NOT: Illegal Value\n")
		}
		va.Not(va)
		switch {
		case va.IsInt64():
			if va.Cmp(MinInt64) < 0 || va.Cmp(MaxInt64) > 0 {
				return errors.New("Fate NOT: Overflows")
			}
		default:
			if va.Cmp(MaxUint64) > 0 {
				return errors.New("Fate NOT: Overflows")
			}
		}
		return da.Update(e, va)
	case INT8, INT16, INT32, INT64:
		va := recent(big.Int)
		if _, ok := va.SetString(string(effByte(append(mixed.E64func(da.header.offset0),
			append(mixed.E64func(da.header.offset1),
				mixed.E64func(da.header.offset2)...)...))), 0); !ok {
			return fmt.Errorf("Fate NOT: Illegal Value\n")
		}
		va.Not(va)
		return da.Update(e, va)
	case UINT8, UINT16, UINT32, UINT64:
		va := recent(big.Int)
		if _, ok := va.SetString(string(effByte(append(mixed.E64func(da.header.offset0),
			append(mixed.E64func(da.header.offset1),
				mixed.E64func(da.header.offset2)...)...))), 0); !ok {
			return fmt.Errorf("Fate NOT: Illegal Value\n")
		}
		switch da.header.length {
		case 1:
			va.Xor(va, MaxUint8)
		case 2:
			va.Xor(va, MaxUint16)
		case 4:
			va.Xor(va, MaxUint32)
		case 8:
			va.Xor(va, MaxUint64)
		}
		return da.Update(e, va)
	default:
		return errors.New("Fate NOT: Unsupport type")
	}
}

// OR  --->  a = b | c
func rsc22(e *contractEngine, op uint64) error {
	a := achieveReg(achieveA(op), e.regs.offset)
	b := achieveReg(achieveB(op), e.regs.offset)
	c := achieveReg(achieveC(op), e.regs.offset)
	defer func() {
		e.regs.pc = e.regs.pc + 1
	}()
	da, err := e.achieveData(e.regs.gpRegs[a])
	if err != nil {
		return fmt.Errorf("Fate OR: %v", err)
	}
	db, err := e.achieveData(e.regs.gpRegs[b])
	if err != nil {
		return fmt.Errorf("Fate OR: %v", err)
	}
	dc, err := e.achieveData(e.regs.gpRegs[c])
	if err != nil {
		return fmt.Errorf("Fate OR: %v", err)
	}
	switch da.header.typ {
	case CHAR:
		if db.header.typ != da.header.typ && db.header.typ != CONST_CHAR {
			return errors.New("Fate OR: Type Error")
		}
		if dc.header.typ != da.header.typ && dc.header.typ != CONST_CHAR {
			return errors.New("Fate OR: Type Error")
		}
		vb := byte(db.header.offset0)
		vc := byte(dc.header.offset0)
		return da.Update(e, byte(vb|vc))
	case CONST_CHAR:
		if db.header.typ != da.header.typ && db.header.typ != CHAR {
			return errors.New("Fate OR: Type Error")
		}
		if dc.header.typ != da.header.typ && dc.header.typ != CHAR {
			return errors.New("Fate OR: Type Error")
		}
		vb := byte(db.header.offset0)
		vc := byte(dc.header.offset0)
		return da.Update(e, byte(vb|vc))
	case CONST_INT:
		if typ := db.header.typ; typ != da.header.typ && typ != INT8 && typ != INT16 &&
			typ != INT32 && typ != INT64 && typ != UINT8 && typ != UINT16 &&
			typ != UINT32 && typ != UINT64 {
			return errors.New("Fate OR: Type Error")
		}
		if typ := dc.header.typ; typ != da.header.typ && typ != INT8 && typ != INT16 &&
			typ != INT32 && typ != INT64 && typ != UINT8 && typ != UINT16 &&
			typ != UINT32 && typ != UINT64 {
			return errors.New("Fate OR: Type Error")
		}
		if db.header.typ != CONST_INT && dc.header.typ != CONST_INT &&
			dc.header.typ != db.header.typ {
			return errors.New("Fate OR: Type Error")
		}
		va := recent(big.Int)
		vb := recent(big.Int)
		vc := recent(big.Int)
		if _, ok := vb.SetString(string(effByte(append(mixed.E64func(db.header.offset0),
			append(mixed.E64func(db.header.offset1),
				mixed.E64func(db.header.offset2)...)...))), 0); !ok {
			err = errors.New("Illegal Value")
			return fmt.Errorf("Fate OR: %v\n", err)
		}
		if _, ok := vc.SetString(string(effByte(append(mixed.E64func(dc.header.offset0),
			append(mixed.E64func(dc.header.offset1),
				mixed.E64func(dc.header.offset2)...)...))), 0); !ok {
			err = errors.New("Illegal Value")
			return fmt.Errorf("Fate OR: %v\n", err)
		}
		va.Or(vb, vc)
		switch {
		case va.IsInt64():
			if va.Cmp(MinInt64) < 0 || va.Cmp(MaxInt64) > 0 {
				return errors.New("Fate NOT: Overflows")
			}
		default:
			if va.Cmp(MaxUint64) > 0 {
				return errors.New("Fate NOT: Overflows")
			}
		}
		return da.Update(e, va)
	case INT8, INT16, INT32, INT64, UINT8, UINT16, UINT32, UINT64:
		if db.header.typ != da.header.typ && db.header.typ != CONST_INT {
			return errors.New("Fate OR: Type Error")
		}
		if dc.header.typ != da.header.typ && dc.header.typ != CONST_INT {
			return errors.New("Fate OR: Type Error")
		}
		va := recent(big.Int)
		vb := recent(big.Int)
		vc := recent(big.Int)
		if _, ok := vb.SetString(string(effByte(append(mixed.E64func(db.header.offset0),
			append(mixed.E64func(db.header.offset1),
				mixed.E64func(db.header.offset2)...)...))), 0); !ok {
			err = errors.New("Illegal Value")
			return fmt.Errorf("Fate OR: %v\n", err)
		}
		if _, ok := vc.SetString(string(effByte(append(mixed.E64func(dc.header.offset0),
			append(mixed.E64func(dc.header.offset1),
				mixed.E64func(dc.header.offset2)...)...))), 0); !ok {
			err = errors.New("Illegal Value")
			return fmt.Errorf("Fate OR: %v\n", err)
		}
		va.Or(vb, vc)
		switch da.header.typ {
		case INT8, INT16, INT32, INT64:
			switch da.header.length {
			case 1:
				if va.Cmp(MinInt8) < 0 || va.Cmp(MaxInt8) > 0 {
					return errors.New("Fate OR: Overflows")
				}
			case 2:
				if va.Cmp(MinInt16) < 0 || va.Cmp(MaxInt16) > 0 {
					return errors.New("Fate OR: Overflows")
				}
			case 4:
				if va.Cmp(MinInt32) < 0 || va.Cmp(MaxInt32) > 0 {
					return errors.New("Fate OR: Overflows")
				}
			case 8:
				if va.Cmp(MinInt64) < 0 || va.Cmp(MaxInt64) > 0 {
					return errors.New("Fate OR: Overflows")
				}
			}
		default:
			switch da.header.length {
			case 1:
				if va.Cmp(MaxUint8) > 0 {
					return errors.New("Fate OR: Overflows")
				}
			case 2:
				if va.Cmp(MaxUint16) > 0 {
					return errors.New("Fate OR: Overflows")
				}
			case 4:
				if va.Cmp(MaxUint32) > 0 {
					return errors.New("Fate OR: Overflows")
				}
			case 8:
				if va.Cmp(MaxUint64) > 0 {
					return errors.New("Fate OR: Overflows")
				}
			}
		}
		return da.Update(e, va)
	default:
		return errors.New("Fate OR: Unsupport type")
	}
}

// AND  --->  a = b & c
func rsc23(e *contractEngine, op uint64) error {
	a := achieveReg(achieveA(op), e.regs.offset)
	b := achieveReg(achieveB(op), e.regs.offset)
	c := achieveReg(achieveC(op), e.regs.offset)
	defer func() {
		e.regs.pc = e.regs.pc + 1
	}()
	da, err := e.achieveData(e.regs.gpRegs[a])
	if err != nil {
		return fmt.Errorf("Fate AND: %v", err)
	}
	db, err := e.achieveData(e.regs.gpRegs[b])
	if err != nil {
		return fmt.Errorf("Fate AND: %v", err)
	}
	dc, err := e.achieveData(e.regs.gpRegs[c])
	if err != nil {
		return fmt.Errorf("Fate AND: %v", err)
	}
	switch da.header.typ {
	case CHAR:
		if db.header.typ != da.header.typ && db.header.typ != CONST_CHAR {
			return errors.New("Fate AND: Type Error")
		}
		if dc.header.typ != da.header.typ && dc.header.typ != CONST_CHAR {
			return errors.New("Fate AND: Type Error")
		}
		vb := byte(db.header.offset0)
		vc := byte(dc.header.offset0)
		return da.Update(e, byte(vb&vc))
	case CONST_CHAR:
		if db.header.typ != da.header.typ && db.header.typ != CHAR {
			return errors.New("Fate AND: Type Error")
		}
		if dc.header.typ != da.header.typ && dc.header.typ != CHAR {
			return errors.New("Fate AND: Type Error")
		}
		vb := byte(db.header.offset0)
		vc := byte(dc.header.offset0)
		return da.Update(e, byte(vb&vc))
	case CONST_INT:
		if typ := db.header.typ; typ != da.header.typ && typ != INT8 && typ != INT16 &&
			typ != INT32 && typ != INT64 && typ != UINT8 && typ != UINT16 &&
			typ != UINT32 && typ != UINT64 {
			return errors.New("Fate AND: Type Error")
		}
		if typ := dc.header.typ; typ != da.header.typ && typ != INT8 && typ != INT16 &&
			typ != INT32 && typ != INT64 && typ != UINT8 && typ != UINT16 &&
			typ != UINT32 && typ != UINT64 {
			return errors.New("Fate AND: Type Error")
		}
		if db.header.typ != CONST_INT && dc.header.typ != CONST_INT &&
			dc.header.typ != db.header.typ {
			return errors.New("Fate AND: Type Error")
		}
		va := recent(big.Int)
		vb := recent(big.Int)
		vc := recent(big.Int)
		if _, ok := vb.SetString(string(effByte(append(mixed.E64func(db.header.offset0),
			append(mixed.E64func(db.header.offset1),
				mixed.E64func(db.header.offset2)...)...))), 0); !ok {
			err = errors.New("Illegal Value")
			return fmt.Errorf("Fate AND: %v\n", err)
		}
		if _, ok := vc.SetString(string(effByte(append(mixed.E64func(dc.header.offset0),
			append(mixed.E64func(dc.header.offset1),
				mixed.E64func(dc.header.offset2)...)...))), 0); !ok {
			err = errors.New("Illegal Value")
			return fmt.Errorf("Fate AND: %v\n", err)
		}
		va.And(vb, vc)
		switch {
		case va.IsInt64():
			if va.Cmp(MinInt64) < 0 || va.Cmp(MaxInt64) > 0 {
				return errors.New("Fate AND: Overflows")
			}
		default:
			if va.Cmp(MaxUint64) > 0 {
				return errors.New("Fate AND: Overflows")
			}
		}
		return da.Update(e, va)
	case INT8, INT16, INT32, INT64, UINT8, UINT16, UINT32, UINT64:
		if db.header.typ != da.header.typ && db.header.typ != CONST_INT {
			return errors.New("Fate AND: Type Error")
		}
		if dc.header.typ != da.header.typ && dc.header.typ != CONST_INT {
			return errors.New("Fate AND: Type Error")
		}
		va := recent(big.Int)
		vb := recent(big.Int)
		vc := recent(big.Int)
		if _, ok := vb.SetString(string(effByte(append(mixed.E64func(db.header.offset0),
			append(mixed.E64func(db.header.offset1),
				mixed.E64func(db.header.offset2)...)...))), 0); !ok {
			err = errors.New("Illegal Value")
			return fmt.Errorf("Fate AND: %v\n", err)
		}
		if _, ok := vc.SetString(string(effByte(append(mixed.E64func(dc.header.offset0),
			append(mixed.E64func(dc.header.offset1),
				mixed.E64func(dc.header.offset2)...)...))), 0); !ok {
			err = errors.New("Illegal Value")
			return fmt.Errorf("Fate AND: %v\n", err)
		}
		va.And(vb, vc)
		switch da.header.typ {
		case INT8, INT16, INT32, INT64:
			switch da.header.length {
			case 1:
				if va.Cmp(MinInt8) < 0 || va.Cmp(MaxInt8) > 0 {
					return errors.New("Fate AND: Overflows")
				}
			case 2:
				if va.Cmp(MinInt16) < 0 || va.Cmp(MaxInt16) > 0 {
					return errors.New("Fate AND: Overflows")
				}
			case 4:
				if va.Cmp(MinInt32) < 0 || va.Cmp(MaxInt32) > 0 {
					return errors.New("Fate AND: Overflows")
				}
			case 8:
				if va.Cmp(MinInt64) < 0 || va.Cmp(MaxInt64) > 0 {
					return errors.New("Fate AND: Overflows")
				}
			}
		default:
			switch da.header.length {
			case 1:
				if va.Cmp(MaxUint8) > 0 {
					return errors.New("Fate AND: Overflows")
				}
			case 2:
				if va.Cmp(MaxUint16) > 0 {
					return errors.New("Fate AND: Overflows")
				}
			case 4:
				if va.Cmp(MaxUint32) > 0 {
					return errors.New("Fate AND: Overflows")
				}
			case 8:
				if va.Cmp(MaxUint64) > 0 {
					return errors.New("Fate AND: Overflows")
				}
			}
		}
		return da.Update(e, va)
	default:
		return errors.New("Fate AND: Unsupport type")
	}
}

// XOR  --->  a = b ^ c
func rsc24(e *contractEngine, op uint64) error {
	a := achieveReg(achieveA(op), e.regs.offset)
	b := achieveReg(achieveB(op), e.regs.offset)
	c := achieveReg(achieveC(op), e.regs.offset)
	defer func() {
		e.regs.pc = e.regs.pc + 1
	}()
	da, err := e.achieveData(e.regs.gpRegs[a])
	if err != nil {
		return fmt.Errorf("Fate XOR: %v", err)
	}
	db, err := e.achieveData(e.regs.gpRegs[b])
	if err != nil {
		return fmt.Errorf("Fate XOR: %v", err)
	}
	dc, err := e.achieveData(e.regs.gpRegs[c])
	if err != nil {
		return fmt.Errorf("Fate XOR: %v", err)
	}
	switch da.header.typ {
	case CHAR:
		if db.header.typ != da.header.typ && db.header.typ != CONST_CHAR {
			return errors.New("Fate XOR: Type Error")
		}
		if dc.header.typ != da.header.typ && dc.header.typ != CONST_CHAR {
			return errors.New("Fate XOR: Type Error")
		}
		vb := byte(db.header.offset0)
		vc := byte(dc.header.offset0)
		return da.Update(e, byte(vb^vc))
	case CONST_CHAR:
		if db.header.typ != da.header.typ && db.header.typ != CHAR {
			return errors.New("Fate XOR: Type Error")
		}
		if dc.header.typ != da.header.typ && dc.header.typ != CHAR {
			return errors.New("Fate XOR: Type Error")
		}
		vb := byte(db.header.offset0)
		vc := byte(dc.header.offset0)
		return da.Update(e, byte(vb^vc))
	case CONST_INT:
		if typ := db.header.typ; typ != da.header.typ && typ != INT8 && typ != INT16 &&
			typ != INT32 && typ != INT64 && typ != UINT8 && typ != UINT16 &&
			typ != UINT32 && typ != UINT64 {
			return errors.New("Fate XOR: Type Error")
		}
		if typ := dc.header.typ; typ != da.header.typ && typ != INT8 && typ != INT16 &&
			typ != INT32 && typ != INT64 && typ != UINT8 && typ != UINT16 &&
			typ != UINT32 && typ != UINT64 {
			return errors.New("Fate XOR: Type Error")
		}
		if db.header.typ != CONST_INT && dc.header.typ != CONST_INT &&
			dc.header.typ != db.header.typ {
			return errors.New("Fate XOR: Type Error")
		}
		va := recent(big.Int)
		vb := recent(big.Int)
		vc := recent(big.Int)
		if _, ok := vb.SetString(string(effByte(append(mixed.E64func(db.header.offset0),
			append(mixed.E64func(db.header.offset1),
				mixed.E64func(db.header.offset2)...)...))), 0); !ok {
			err = errors.New("Illegal Value")
			return fmt.Errorf("Fate XOR: %v\n", err)
		}
		if _, ok := vc.SetString(string(effByte(append(mixed.E64func(dc.header.offset0),
			append(mixed.E64func(dc.header.offset1),
				mixed.E64func(dc.header.offset2)...)...))), 0); !ok {
			err = errors.New("Illegal Value")
			return fmt.Errorf("Fate XOR: %v\n", err)
		}
		va.Xor(vb, vc)
		switch {
		case va.IsInt64():
			if va.Cmp(MinInt64) < 0 || va.Cmp(MaxInt64) > 0 {
				return errors.New("Fate XOR: Overflows")
			}
		default:
			if va.Cmp(MaxUint64) > 0 {
				return errors.New("Fate XOR: Overflows")
			}
		}
		return da.Update(e, va)
	case INT8, INT16, INT32, INT64, UINT8, UINT16, UINT32, UINT64:
		if db.header.typ != da.header.typ && db.header.typ != CONST_INT {
			return errors.New("Fate XOR: Type Error")
		}
		if dc.header.typ != da.header.typ && dc.header.typ != CONST_INT {
			return errors.New("Fate XOR: Type Error")
		}
		va := recent(big.Int)
		vb := recent(big.Int)
		vc := recent(big.Int)
		if _, ok := vb.SetString(string(effByte(append(mixed.E64func(db.header.offset0),
			append(mixed.E64func(db.header.offset1),
				mixed.E64func(db.header.offset2)...)...))), 0); !ok {
			err = errors.New("Illegal Value")
			return fmt.Errorf("Fate XOR: %v\n", err)
		}
		if _, ok := vc.SetString(string(effByte(append(mixed.E64func(dc.header.offset0),
			append(mixed.E64func(dc.header.offset1),
				mixed.E64func(dc.header.offset2)...)...))), 0); !ok {
			err = errors.New("Illegal Value")
			return fmt.Errorf("Fate XOR: %v\n", err)
		}
		va.Xor(vb, vc)
		switch da.header.typ {
		case INT8, INT16, INT32, INT64:
			switch da.header.length {
			case 1:
				if va.Cmp(MinInt8) < 0 || va.Cmp(MaxInt8) > 0 {
					return errors.New("Fate XOR: Overflows")
				}
			case 2:
				if va.Cmp(MinInt16) < 0 || va.Cmp(MaxInt16) > 0 {
					return errors.New("Fate XOR: Overflows")
				}
			case 4:
				if va.Cmp(MinInt32) < 0 || va.Cmp(MaxInt32) > 0 {
					return errors.New("Fate XOR: Overflows")
				}
			case 8:
				if va.Cmp(MinInt64) < 0 || va.Cmp(MaxInt64) > 0 {
					return errors.New("Fate XOR: Overflows")
				}
			}
		default:
			switch da.header.length {
			case 1:
				if va.Cmp(MaxUint8) > 0 {
					return errors.New("Fate XOR: Overflows")
				}
			case 2:
				if va.Cmp(MaxUint16) > 0 {
					return errors.New("Fate XOR: Overflows")
				}
			case 4:
				if va.Cmp(MaxUint32) > 0 {
					return errors.New("Fate XOR: Overflows")
				}
			case 8:
				if va.Cmp(MaxUint64) > 0 {
					return errors.New("Fate XOR: Overflows")
				}
			}
		}
		return da.Update(e, va)
	default:
		return errors.New("Fate XOR: Unsupport type")
	}
}

// SIZEOF  --->  a = sizeof(b)
func rsc25(e *contractEngine, op uint64) error {
	a := achieveReg(achieveA(op), e.regs.offset)
	b := achieveReg(achieveB(op), e.regs.offset)
	defer func() {
		e.regs.pc = e.regs.pc + 1
	}()
	da, err := e.achieveData(e.regs.gpRegs[a])
	if err != nil {
		return fmt.Errorf("Fate SIZEOF: %v", err)
	}
	db, err := e.achieveData(e.regs.gpRegs[b])
	if err != nil {
		return fmt.Errorf("Fate SIZEOF: %v", err)
	}
	switch da.header.typ {
	case INT8, INT16, INT32, INT64,
		UINT8, UINT16, UINT32, UINT64, CONST_INT:
		va := recent(big.Int).SetInt64(int64(e.sizeof(db)))
		return da.Update(e, va)
	case FLOAT32, FLOAT64, CONST_FLOAT:
		va := recent(big.Float).SetFloat64(float64(e.sizeof(db)))
		return da.Update(e, va)
	default:
		return errors.New("Fate SIZEOF: Unsupport type")
	}
}

func rsc26(e *contractEngine, op uint64) error {
	a := achieveReg(achieveA(op), e.regs.offset)
	b := achieveReg(achieveB(op), e.regs.offset)
	defer func() {
		e.regs.pc = e.regs.pc + 1
	}()
	da, err := e.achieveData(e.regs.gpRegs[a])
	if err != nil {
		return fmt.Errorf("Fate CUT: %v", err)
	}
	db, err := e.achieveData(e.regs.gpRegs[b])
	if err != nil {
		return fmt.Errorf("Fate CUT: %v", err)
	}
	switch db.header.typ {
	case INT8, INT16, INT32, INT64,
		UINT8, UINT16, UINT32, UINT64, CONST_INT:
	default:
		return errors.New("Fate CUT: Unsupport type")
	}
	if da.header.typ != STRING && da.header.typ != CONST_STRING {
		return errors.New("Fate CUT: Unsupport type")
	}
	v := recent(big.Int)
	if _, ok := v.SetString(string(effByte(append(mixed.E64func(db.header.offset0),
		append(mixed.E64func(db.header.offset1),
			mixed.E64func(db.header.offset2)...)...))), 0); !ok {
		return errors.New("Fate CUT: Illegal Length")
	}
	return e.cut(da, v.Uint64())
}

func rsc27(e *contractEngine, op uint64) error {
	a := achieveReg(achieveA(op), e.regs.offset)
	b := achieveReg(achieveB(op), e.regs.offset)
	c := achieveReg(achieveC(op), e.regs.offset)
	defer func() {
		e.regs.pc = e.regs.pc + 1
	}()
	db, err := e.achieveData(e.regs.gpRegs[b])
	if err != nil {
		return fmt.Errorf("Fate INDEX: %v", err)
	}
	dc, err := e.achieveData(e.regs.gpRegs[c])
	if err != nil {
		return fmt.Errorf("Fate INDEX: %v", err)
	}
	if db.header.typ != STRING && db.header.typ != CONST_STRING {
		return errors.New("Fate INDEX: Unsupport type")
	}
	switch dc.header.typ {
	case INT8, INT16, INT32, INT64,
		UINT8, UINT16, UINT32, UINT64, CONST_INT:
	default:
		return errors.New("Fate INDEX: Unsupport type")
	}
	v := recent(big.Int)
	if _, ok := v.SetString(string(effByte(append(mixed.E64func(dc.header.offset0),
		append(mixed.E64func(dc.header.offset1),
			mixed.E64func(dc.header.offset2)...)...))), 0); !ok {
		return errors.New("Fate INDEX: Illegal Length")
	}
	if da, err := e.index(db, v.Uint64()); err == nil {
		e.regs.gpRegs[a] = da.address
		return nil
	} else {
		return fmt.Errorf("Fate INDEX: %v", err)
	}
}

func rsc28(e *contractEngine, op uint64) error {
	a := achieveReg(achieveA(op), e.regs.offset)
	b := achieveReg(achieveB(op), e.regs.offset)
	defer func() {
		e.regs.pc = e.regs.pc + 1
	}()
	da, err := e.achieveData(e.regs.gpRegs[a])
	if err != nil {
		return fmt.Errorf("Fate CONCAT: %v", err)
	}
	db, err := e.achieveData(e.regs.gpRegs[b])
	if err != nil {
		return fmt.Errorf("Fate CONCAT: %v", err)
	}
	if da.header.typ != STRING && db.header.typ != CHAR && db.header.typ != CONST_CHAR {
		return errors.New("Fate CONCAT: Unsupport type")
	}
	return e.concat(da, db)
}

func rsc29(e *contractEngine, op uint64) error {
	a := achieveReg(achieveA(op), e.regs.offset)
	b := achieveReg(achieveB(op), e.regs.offset)
	c := achieveReg(achieveC(op), e.regs.offset)
	defer func() {
		e.regs.pc = e.regs.pc + 1
	}()
	db, err := e.achieveData(e.regs.gpRegs[b])
	if err != nil {
		return fmt.Errorf("Fate FIND: %v", err)
	}
	dc, err := e.achieveData(e.regs.gpRegs[c])
	if err != nil {
		return fmt.Errorf("Fate FIND: %v", err)
	}
	if db.header.typ != MAP {
		return errors.New("Fate Find: Unsupport type")
	}
	if da, err := e.find(db, dc); err == nil {
		e.regs.gpRegs[a] = da.address
		return nil
	} else {
		return fmt.Errorf("Fate FIND: %v", err)
	}
}

func rsc37(e *contractEngine, op uint64) error {
	a := achieveReg(achieveA(op), e.regs.offset)
	b := achieveReg(achieveB(op), e.regs.offset)
	c := achieveReg(achieveC(op), e.regs.offset)
	defer func() {
		e.regs.pc = e.regs.pc + 1
	}()
	da, err := e.achieveData(e.regs.gpRegs[a])
	if err != nil {
		return fmt.Errorf("Fate ELEM: %v", err)
	}
	db, err := e.achieveData(e.regs.gpRegs[b])
	if err != nil {
		return fmt.Errorf("Fate ELEM: %v", err)
	}
	dc, err := e.achieveData(e.regs.gpRegs[c])
	if err != nil {
		return fmt.Errorf("Fate ELEM: %v", err)
	}
	if db.header.typ != MAP {
		return errors.New("Fate ELEM: Unsupport type")
	}
	if da.header.typ != BOOL && da.header.typ != CONST_BOOL {
		return errors.New("Fate ELEM: Unsupport type")
	}
	if _, err := e.find(db, dc); err == nil {
		da.Update(e, uint64(1)) // true
		e.regs.gpRegs[a] = da.address
		return nil
	} else {
		da.Update(e, uint64(0)) // false
		return nil
	}
}

func rsc30(e *contractEngine, op uint64) error {
	a := achieveReg(achieveA(op), e.regs.offset)
	b := achieveReg(achieveB(op), e.regs.offset)
	c := achieveReg(achieveC(op), e.regs.offset)
	defer func() {
		e.regs.pc = e.regs.pc + 1
	}()
	da, err := e.achieveData(e.regs.gpRegs[a])
	if err != nil {
		return fmt.Errorf("contract INSERT: %v", err)
	}
	db, err := e.achieveData(e.regs.gpRegs[b])
	if err != nil {
		return fmt.Errorf("contract INSERT: %v", err)
	}
	dc, err := e.achieveData(e.regs.gpRegs[c])
	if err != nil {
		return fmt.Errorf("Fate INSERT: %v", err)
	}
	if da.header.typ != MAP {
		return errors.New("Fate INSERT: Unsupport type")
	}
	if v, err := e.find(da, db); err != nil {
		return e.insert(da, db, dc)
	} else {
		return e.move(v, dc)
	}
}

// DELETE  --->  remove a[b]
func rsc31(e *contractEngine, op uint64) error {
	a := achieveReg(achieveA(op), e.regs.offset)
	b := achieveReg(achieveB(op), e.regs.offset)
	defer func() {
		e.regs.pc = e.regs.pc + 1
	}()
	da, err := e.achieveData(e.regs.gpRegs[a])
	if err != nil {
		return fmt.Errorf("Fate Delete: %v", err)
	}
	db, err := e.achieveData(e.regs.gpRegs[b])
	if err != nil {
		return fmt.Errorf("Fate Delete: %v", err)
	}
	if da.header.typ != MAP {
		return errors.New("Fate DELETE: Only Support MAP")
	}
	return e.rubout(da, db)
}

// SM3  --->  0 = SM3(a)
func rsc32(e *contractEngine, op uint64) error {
	a := achieveReg(achieveA(op), e.regs.offset)
	defer func() {
		e.regs.pc = e.regs.pc + 1
	}()
	da, err := e.achieveData(e.regs.gpRegs[a])
	if err != nil {
		return fmt.Errorf("Fate SM3: %v", err)
	}
	if da.header.typ != STRING && da.header.typ != CONST_STRING {
		return errors.New("Fate SM3: Only Support STRING")
	}
	if db, err := e.tmp(0, STRING); err == nil {
		if s, err := e.achieveString(da); err == nil {
			h := sm3.New()
			h.Write(base58.Decode(string(s)))
			hData := base58.Encode(h.Sum(nil))
			return e.setString(db, []byte(hData))
		} else {
			return fmt.Errorf("Fate SM3: %v", err)
		}
	} else {
		return fmt.Errorf("Fate SM3: %v", err)
	}
}

func rsc33(e *contractEngine, op uint64) error {
	a := achieveReg(achieveA(op), e.regs.offset)
	b := achieveReg(achieveB(op), e.regs.offset)
	c := achieveReg(achieveC(op), e.regs.offset)
	defer func() {
		e.regs.pc = e.regs.pc + 1
	}()
	if dd, err := e.tmp(0, BOOL); err == nil {
		da, err := e.achieveData(e.regs.gpRegs[a])
		if err != nil {
			return fmt.Errorf("Fate SM2: %v", err)
		}
		db, err := e.achieveData(e.regs.gpRegs[b])
		if err != nil {
			return fmt.Errorf("Fate SM2: %v", err)
		}
		dc, err := e.achieveData(e.regs.gpRegs[c])
		if err != nil {
			return fmt.Errorf("Fate SM2: %v", err)
		}
		if da.header.typ != STRING && da.header.typ != CONST_STRING && db.header.typ != STRING &&
			db.header.typ != CONST_STRING && dc.header.typ != STRING && dc.header.typ != CONST_STRING {
			return errors.New("Fate SM2: Unsupport type")
		}
		sa, err := e.achieveString(da)
		if err != nil {
			return fmt.Errorf("Fate SM2: %v", err)
		}
		sb, err := e.achieveString(db)
		if err != nil {
			return fmt.Errorf("Fate SM2: %v", err)
		}
		sc, err := e.achieveString(dc)
		if err != nil {
			return fmt.Errorf("Fate SM2: %v", err)
		}
		data := sb
		sign := base58.Decode(string(sa))
		pubKeyData := base58.Decode(string(sc))
		if len(pubKeyData) != 33 {
			return errors.New("Fate SM2: Illegal Length of PubKey")
		}
		r, s, err := sm2.SignDataToSignDigit(sign)
		if err != nil {
			return fmt.Errorf("Fate SM2: %v", err)
		}
		pubKey := sm2.Decompress(pubKeyData)
		ok := sm2.Verify(pubKey, data, r, s)
		switch ok {
		case true:
			dd.header.offset0 = 1
		case false:
			dd.header.offset0 = 0
		}
		return e.setData(dd)
	} else {
		return fmt.Errorf("Fate SM2: %v", err)
	}
}

// TMP
func rsc34(e *contractEngine, op uint64) error {
	typ := achieveB(op)
	a := achieveReg(achieveA(op), e.regs.offset)
	defer func() {
		e.regs.pc = e.regs.pc + 1
	}()
	_, err := e.tmp(a, typ)
	return err
}

func rsc35(e *contractEngine, op uint64) error {
	if e.regs.eflags&EQ != 0 || e.regs.eflags&GR != 0 {
		if a := achieveA(op); a&SIGN_BIT != 0 {
			e.regs.pc = e.regs.pc - (a & DATA_BIT)
		} else {
			e.regs.pc = e.regs.pc + (a & DATA_BIT)
		}
	} else {
		e.regs.pc = e.regs.pc + 1
	}
	return nil
}

func rsc36(e *contractEngine, op uint64) error {
	if e.regs.eflags&EQ != 0 || e.regs.eflags&LE != 0 {
		if a := achieveA(op); a&SIGN_BIT != 0 {
			e.regs.pc = e.regs.pc - (a & DATA_BIT)
		} else {
			e.regs.pc = e.regs.pc + (a & DATA_BIT)
		}
	} else {
		e.regs.pc = e.regs.pc + 1
	}
	return nil
}

func (e *contractEngine) tmp(a, typ int) (*contractData, error) {
	if a >= len(e.regs.gpRegs) {
		return nil, errors.New("Fate TMP: Cross Border")
	}
	addr, err := e.virtul.Alloc(virtul.RAM, DATA_HEADER_SIZE)
	if err != nil {
		return nil, fmt.Errorf("Fate TMP: %v", err)
	}
	if da, err := e.recentVar(typ, addr); err == nil {
		if err = e.setData(da); err != nil {
			e.virtul.Free(addr)
			return nil, fmt.Errorf("Fate TMP: %v", err)
		}
		e.regs.gpRegs[a] = addr
		return da, nil
	} else {
		e.virtul.Free(addr)
		return nil, fmt.Errorf("Fate TMP: %v", err)
	}
}
