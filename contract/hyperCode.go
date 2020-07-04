package contract

import (
	"errors"
	"fmt"
	"math/big"

	"kortho/contract/mixed"
	"kortho/contract/motor"
	"kortho/contract/virtul"
)

func init() {
	rootRegistry = map[int]contractGenFunc{
		G_VAR:    root0,
		G_FUNC:   root1,
		G_ARGS:   root2,
		G_BODY:   root3,
		G_BLK:    root4,
		G_STMT:   root5,
		G_CSTMT:  root6,
		G_JSTMT:  root7,
		G_ISTMT:  root8,
		G_SSTMT:  root9,
		G_ESTMT:  root10,
		G_EXPR:   root11,
		G_ASGN:   root12,
		G_LGOR:   root13,
		G_LGAND:  root14,
		G_INOR:   root15,
		G_EXOR:   root16,
		G_AND:    root17,
		G_EQ:     root18,
		G_CMP:    root19,
		G_SHF:    root20,
		G_ADD:    root21,
		G_MUL:    root22,
		G_UNARY:  root23,
		G_SIZEOF: root24,
		G_PTF:    root25,
		G_PRM:    root26,
		G_PARA:   root27,
	}
}

func achieveOp(o uint64) int {
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

func createOp(o, a, b, c int) uint64 {
	return uint64(o)<<56 | uint64(a)<<40 | uint64(b)<<16 | uint64(c)
}

func createOp1(o, a int, bcr uint64) uint64 {
	return uint64(o)<<56 | uint64(a)<<40 | bcr
}

func isLeft(typ uint32) bool {
	switch typ & TYP_MASK {
	case S_VAR, S_SCS:
		return true
	default:
		return false
	}
}

func isFunc(typ uint32) bool {
	if typ&TYP_MASK == S_FUNC {
		return true
	}
	return false
}

func isMap(typ uint32) bool {
	if typ&TYP_MASK == S_VAR && (typ&SUB_MASK)>>SUB_OFF == O_MAP {
		return true
	}
	return false
}

func isInt(typ uint32) bool {
	switch (typ & SUB_MASK) >> SUB_OFF {
	case motor.CONST_INT:
		return true
	case motor.INT8, motor.INT16, motor.INT32, motor.INT64:
		return true
	case motor.UINT8, motor.UINT16, motor.UINT32, motor.UINT64:
		return true
	default:
		return false
	}
}

func isChar(typ uint32) bool {
	switch (typ & SUB_MASK) >> SUB_OFF {
	case motor.CHAR, motor.CONST_CHAR:
		return true
	default:
		return false
	}
}

func isBool(typ uint32) bool {
	switch (typ & SUB_MASK) >> SUB_OFF {
	case motor.BOOL, motor.CONST_BOOL:
		return true
	default:
		return false
	}
}

func isFloat(typ uint32) bool {
	switch (typ & SUB_MASK) >> SUB_OFF {
	case motor.FLOAT32, motor.FLOAT64, motor.CONST_FLOAT:
		return true
	default:
		return false
	}
}

func isString(typ uint32) bool {
	switch (typ & SUB_MASK) >> SUB_OFF {
	case motor.STRING, motor.CONST_STRING:
		return true
	default:
		return false
	}
}

func isMatch(x, y uint32) bool {
	xt := (x & SUB_MASK) >> SUB_OFF
	yt := (y & SUB_MASK) >> SUB_OFF
	switch xt {
	case motor.BOOL:
		if yt != xt && yt != motor.CONST_BOOL {
			return false
		}
	case motor.CHAR:
		if yt != xt && yt != motor.CONST_CHAR {
			return false
		}
	case motor.STRING:
		if yt != xt && yt != motor.CONST_STRING {
			return false
		}
	case motor.FLOAT32, motor.FLOAT64:
		if yt != xt && yt != motor.CONST_FLOAT {
			return false
		}
	case motor.INT8, motor.INT16, motor.INT32, motor.INT64, motor.UINT8, motor.UINT16, motor.UINT32, motor.UINT64:
		if yt != xt && yt != motor.CONST_INT {
			return false
		}
	case motor.CONST_INT:
		return isInt(y)
	case motor.CONST_BOOL:
		return isBool(y)
	case motor.CONST_CHAR:
		return isChar(y)
	case motor.CONST_FLOAT:
		return isFloat(y)
	case motor.CONST_STRING:
		return isString(y)
	default:
		return false
	}
	return true
}

func isKeyMatch(x, y uint32) bool {
	xt := (x & KEY_MASK) >> KEY_OFF
	yt := (y & SUB_MASK) >> SUB_OFF
	switch xt {
	case motor.BOOL:
		if yt != xt && yt != motor.CONST_BOOL {
			return false
		}
	case motor.CHAR:
		if yt != xt && yt != motor.CONST_CHAR {
			return false
		}
	case motor.STRING:
		if yt != xt && yt != motor.CONST_STRING {
			return false
		}
	case motor.FLOAT32, motor.FLOAT64:
		if yt != xt && yt != motor.CONST_FLOAT {
			return false
		}
	case motor.INT8, motor.INT16, motor.INT32, motor.INT64, motor.UINT8, motor.UINT16, motor.UINT32, motor.UINT64:
		if yt != xt && yt != motor.CONST_INT {
			return false
		}
	default:
		return false
	}
	return true
}

func isRtv(x, y uint32) bool {
	typ := (x & SUB_MASK) >> SUB_OFF
	switch rt := (y & SUB_MASK) >> SUB_OFF; rt {
	case motor.CONST_INT:
		return isInt(x)
	case motor.CONST_FLOAT:
		return isFloat(x)
	case motor.BOOL, motor.CONST_BOOL:
		return isBool(x)
	case motor.CHAR, motor.CONST_CHAR:
		return isChar(x)
	case motor.STRING, motor.CONST_STRING:
		return isString(x)
	case motor.FLOAT32, motor.FLOAT64:
		return rt == typ || typ == motor.CONST_FLOAT
	case motor.INT8, motor.INT16, motor.INT32, motor.INT64,
		motor.UINT8, motor.UINT16, motor.UINT32, motor.UINT64:
		return rt == typ || typ == motor.CONST_INT
	default:
		return false
	}
}

// Binary Operator
func checkType2(x, y uint32, typ int) (int, error) {
	if !isMap(x) && !isMap(y) && !isMatch(x, y) {
		return 0, errors.New("Type Don't Match")
	}
	switch typ & TYP_MASK {
	case O_EQ, O_NE, O_ASSIGN:
		switch {
		case isMap(x) && isMap(y):
			return motor.MAP, nil
		case isInt(x) && isInt(y):
			return motor.CONST_INT, nil
		case isBool(x) && isBool(y):
			return motor.CONST_BOOL, nil
		case isChar(x) && isChar(y):
			return motor.CONST_CHAR, nil
		case isFloat(x) && isFloat(y):
			return motor.CONST_FLOAT, nil
		case isString(x) && isString(y):
			return motor.CONST_STRING, nil
		default:
			return 0, errors.New("ASSIGN: Type Don't Match")
		}
	case O_LOR, O_LAND:
		switch {
		case isBool(x) && isBool(y):
			return motor.CONST_BOOL, nil
		default:
			return 0, errors.New("Type Don't Match")
		}
	case O_OR, O_MOD, O_AND, O_XOR, O_LEFT, O_RIGHT:
		switch {
		case isInt(x) && isInt(y):
			return motor.CONST_INT, nil
		case isChar(x) && isChar(y):
			return motor.CONST_CHAR, nil
		default:
			return 0, errors.New("Type Don't Match")
		}
	case O_ADD, O_SUB, O_MUL, O_DIV, O_LT, O_GT, O_LE, O_GE:
		switch {
		case isInt(x) && isInt(y):
			return motor.CONST_INT, nil
		case isChar(x) && isChar(y):
			return motor.CONST_CHAR, nil
		case isFloat(x) && isFloat(y):
			return motor.CONST_FLOAT, nil
		default:
			return 0, errors.New("Type Don't Match")
		}
	default:
		return 0, errors.New("Type Don't Match")
	}
}

func checkArg(a, b []*contractSymbol, name string) error {
	if len(a) != len(b) {
		return fmt.Errorf("Number of Arguments Not Match In Call To '%s'", name)
	}
	for i, j := 0, len(a); i < j; i++ {
		if (a[i].typ&SUB_MASK)>>SUB_OFF == 0xFF {
			continue
		}
		typ := (b[i].typ & SUB_MASK) >> SUB_OFF
		if a[i].typ&SUB_MASK != b[i].typ&SUB_MASK {
			switch {
			case isInt(a[i].typ) && typ == motor.CONST_INT:
			case isChar(a[i].typ) && typ == motor.CONST_CHAR:
			case isBool(a[i].typ) && typ == motor.CONST_BOOL:
			case isFloat(a[i].typ) && typ == motor.CONST_FLOAT:
			case isString(a[i].typ) && typ == motor.CONST_STRING:
			default:
				return fmt.Errorf("Type of Arguments Not Match In Call To '%s'", name)
			}
		}
	}
	return nil
}

func recentSym(name, attr, info, size, value, extra uint32, address, raddress uint64) *diste {
	return &diste{
		name:     name,
		attr:     attr,
		info:     info,
		size:     size,
		value:    value,
		extra:    extra,
		address:  address,
		raddress: raddress,
	}
}

func (a *contractIattr) dump() []uint64 {
	ops := a.ops
	a.ops = []uint64{}
	return ops
}

func (a *contractSymbol) allocReg(it *contractIattr) {
	if a.reg != 0 {
		return
	}
	a.reg = len(it.cFunc.regs) + 2
	it.cFunc.regs = append(it.cFunc.regs, a)
	switch a.typ & TYP_MASK {
	case S_TMP:
		it.ops = append(it.ops, createOp(motor.TMP, a.reg, int((a.typ&SUB_MASK)>>SUB_OFF), 0))
	case S_ARG:
		it.ops = append(it.ops, createOp(motor.POP, a.reg, 0, 0))
	case S_SCS:
		defer func() {
			it.mSym = it.mSym[1:]
			it.kSym = it.kSym[1:]
		}()
		it.ops = append(it.ops, createOp(motor.FIND,
			a.reg, it.mSym[0].reg, it.kSym[0].reg))
	case S_FUNC:
	case S_VAR, S_CNT:
		it.cFunc.loads = append(it.cFunc.loads, createOp1(motor.LOAD, a.reg, a.fst.achieveAddr()))
	}
}

func (fg *contractGenerator) allocRet(rtv int, fi *contractFunc) error {
	var r *contractSymbol

	switch rtv {
	case motor.BOOL:
		r, _ = fg.attr.st.bConst[true]
	case motor.CHAR:
		r, _ = fg.attr.st.cConst[byte(0)]
	case motor.STRING:
		r, _ = fg.attr.st.sConst[""]
	case motor.FLOAT32, motor.FLOAT64:
		r, _ = fg.attr.st.fConst[0.0]
	case motor.INT8, motor.INT16, motor.INT32, motor.INT64,
		motor.UINT8, motor.UINT16, motor.UINT32, motor.UINT64:
		r, _ = fg.attr.st.iConst[0]
	default:
		return errors.New("Unsupport Type of Return Value")
	}
	r.reg = len(fg.attr.it.cFunc.regs) + 2
	fg.attr.it.cFunc.regs = append(fg.attr.it.cFunc.regs, r)
	fi.ops = append(fi.ops, createOp1(motor.LOAD, r.reg, r.fst.achieveAddr()))
	fi.ops = append(fi.ops, createOp(motor.RET, r.reg, 0, 0))
	return nil
}

func (st *contractSattr) addName(name string) {
	st.strs = append(st.strs, []byte(name)...)
	st.strs = append(st.strs, byte(0x00))
}

func (st *contractSattr) addFunc(fi *contractFunc, name string) {
	var ok bool
	var idx uint32

	fi.idx = len(st.ops)
	st.ops = append(st.ops, fi.ops...)
	defer func() {
		if !ok {
			st.addName(name)
			st.sMap[name] = idx
		}
	}()
	if idx, ok = st.sMap[name]; !ok {
		idx = uint32(len(st.strs))
	}
	st.funcs[name] = fi
	st.syms = append(st.syms, recentSym(idx, motor.FSCE_TEXT, 0, uint32(len(fi.regs)), uint32(fi.idx), 0, 0, 0))
	for i, j := 0, len(fi.regs); i < j; i++ {
		fi.regs[i].reg = 0
	}
}

func (a *contractFunc) backfill(fg *contractGenerator) error {
	i := 0
	j := len(a.ops)
	for _, v := range a.calls {
		f, ok := fg.attr.st.funcs[v.name]
		if !ok {
			return fmt.Errorf("Cannot Find Function '%s'", v.name)
		}
		if err := checkArg(f.args, v.args, v.name); err != nil {
			return err
		}
		for ; i < j; i++ {
			if achieveOp(a.ops[i]) == motor.CALL {
				fg.attr.st.ops[a.idx+i] = createOp1(motor.CALL, 0, uint64(f.idx))
				goto OUT
			}
		}
		return fmt.Errorf("Cannot Find Call For Function '%s'", v.name)
	OUT:
		i++
	}
	return nil
}

// it's too low
func (fg *contractGenerator) loadLibrary() error {
	var err error
	var addr uint64
	var addrs []uint64

	fg.attr.st.strs = []byte{0x00}
	fg.attr.st.data = append(fg.attr.st.data, &contractDataHeader{}) // reserved data
	if addr, err = fg.recentLibVar(motor.BOOL, virtul.RAM, "r"); err != nil {
		return err
	}
	addrs = append(addrs, addr)
	if addr, err = fg.recentLibVar(motor.UINT64, virtul.RAM, "i"); err != nil {
		return err
	}
	addrs = append(addrs, addr)
	if addr, err = fg.recentLibVar(motor.UINT64, virtul.RAM, "j"); err != nil {
		return err
	}
	addrs = append(addrs, addr)
	if addr, err = fg.recentLibVar(motor.STRING, virtul.RAM, "c"); err != nil {
		return err
	}
	addrs = append(addrs, addr)
	if addr, err = fg.recentLibVar(motor.STRING, virtul.RAM, "h"); err != nil {
		return err
	}
	addrs = append(addrs, addr)
	if addr, err = fg.recentLibVar(motor.STRING, virtul.RAM, "t"); err != nil {
		return err
	}
	addrs = append(addrs, addr)
	if addr, err = fg.recentLibVar(motor.BOOL, virtul.RAM, "s"); err != nil {
		return err
	}
	addrs = append(addrs, addr)
	if err = fg.loadConstant(); err != nil {
		return err
	}
	if err = fg.loadLibFunc(addrs); err != nil {
		return err
	}
	return nil
}

func (fg *contractGenerator) loadLibFunc(addrs []uint64) error {
	{
		ops := []uint64{}
		ops = append(ops, createOp(motor.POP, 0, 0, 0))
		ops = append(ops, createOp(motor.PUSH, 0, 0, 0))
		ops = append(ops, createOp1(motor.LOAD, 2, addrs[5])) // address of t
		ops = append(ops, createOp(motor.TIME, 2, 0, 0))
		ops = append(ops, createOp(motor.RET, 2, 0, 0))
		args := []*contractSymbol{}
		regs := []*contractSymbol{}
		regs = append(regs, &contractSymbol{})
		fg.attr.st.addFunc(&contractFunc{
			ops:  ops,
			args: args,
			regs: regs,
			idx:  len(ops),
			rtv:  motor.STRING,
		}, "time")
	}
	{
		ops := []uint64{}
		b0, _ := fg.attr.st.bConst[true]
		ops = append(ops, createOp(motor.POP, 0, 0, 0))
		ops = append(ops, createOp(motor.POP, 2, 0, 0)) // map
		ops = append(ops, createOp(motor.POP, 3, 0, 0)) // key
		ops = append(ops, createOp(motor.PUSH, 0, 0, 0))
		ops = append(ops, createOp1(motor.LOAD, 4, b0.fst.achieveAddr()))
		ops = append(ops, createOp(motor.DELETE, 2, 3, 0))
		ops = append(ops, createOp(motor.RET, 4, 0, 0))
		args := []*contractSymbol{}
		args = append(args, &contractSymbol{
			typ: motor.MAP << SUB_OFF,
		})
		args = append(args, &contractSymbol{
			typ: 0xFF << SUB_OFF,
		})
		regs := []*contractSymbol{}
		regs = append(regs, &contractSymbol{})
		regs = append(regs, &contractSymbol{})
		fg.attr.st.addFunc(&contractFunc{
			ops:  ops,
			args: args,
			regs: regs,
			idx:  len(ops),
			rtv:  motor.BOOL,
		}, "delete")
	}
	{
		ops := []uint64{}
		ops = append(ops, createOp(motor.POP, 0, 0, 0))
		ops = append(ops, createOp(motor.POP, 2, 0, 0)) // map
		ops = append(ops, createOp(motor.POP, 3, 0, 0)) // key
		ops = append(ops, createOp(motor.PUSH, 0, 0, 0))
		ops = append(ops, createOp1(motor.LOAD, 4, addrs[6])) // address of s
		ops = append(ops, createOp(motor.ELEM, 4, 2, 3))
		ops = append(ops, createOp(motor.RET, 4, 0, 0))
		args := []*contractSymbol{}
		args = append(args, &contractSymbol{
			typ: motor.MAP << SUB_OFF,
		})
		args = append(args, &contractSymbol{
			typ: 0xFF << SUB_OFF,
		})
		regs := []*contractSymbol{}
		regs = append(regs, &contractSymbol{})
		regs = append(regs, &contractSymbol{})
		fg.attr.st.addFunc(&contractFunc{
			ops:  ops,
			args: args,
			regs: regs,
			idx:  len(ops),
			rtv:  motor.BOOL,
		}, "elem")
	}
	{
		ops := []uint64{}
		ops = append(ops, createOp(motor.POP, 0, 0, 0))
		ops = append(ops, createOp(motor.POP, 2, 0, 0))
		ops = append(ops, createOp(motor.PUSH, 0, 0, 0))
		ops = append(ops, createOp1(motor.LOAD, 3, addrs[4])) // address of h
		ops = append(ops, createOp(motor.SM3, 2, 0, 0))
		ops = append(ops, createOp(motor.MOVE, 3, 0, 0))
		ops = append(ops, createOp(motor.RET, 3, 0, 0))
		args := []*contractSymbol{}
		args = append(args, &contractSymbol{
			typ: motor.STRING << SUB_OFF,
		})
		regs := []*contractSymbol{}
		regs = append(regs, &contractSymbol{})
		regs = append(regs, &contractSymbol{})
		fg.attr.st.addFunc(&contractFunc{
			ops:  ops,
			args: args,
			regs: regs,
			idx:  len(ops),
			rtv:  motor.STRING,
		}, "sm3Hash")
	}
	{
		ops := []uint64{}
		ops = append(ops, createOp(motor.POP, 0, 0, 0))
		ops = append(ops, createOp(motor.POP, 2, 0, 0))
		ops = append(ops, createOp(motor.POP, 3, 0, 0))
		ops = append(ops, createOp(motor.POP, 4, 0, 0))
		ops = append(ops, createOp(motor.PUSH, 0, 0, 0))
		ops = append(ops, createOp1(motor.LOAD, 5, addrs[0])) // address of r
		ops = append(ops, createOp(motor.SM2, 2, 3, 4))
		ops = append(ops, createOp(motor.MOVE, 5, 0, 0))
		ops = append(ops, createOp(motor.RET, 5, 0, 0))
		args := []*contractSymbol{}
		args = append(args, &contractSymbol{
			typ: motor.STRING << SUB_OFF,
		})
		args = append(args, &contractSymbol{
			typ: motor.STRING << SUB_OFF,
		})
		args = append(args, &contractSymbol{
			typ: motor.STRING << SUB_OFF,
		})
		regs := []*contractSymbol{}
		regs = append(regs, &contractSymbol{})
		regs = append(regs, &contractSymbol{})
		regs = append(regs, &contractSymbol{})
		regs = append(regs, &contractSymbol{})
		fg.attr.st.addFunc(&contractFunc{
			ops:  ops,
			args: args,
			regs: regs,
			idx:  len(ops),
			rtv:  motor.BOOL,
		}, "sm2Verify")
	}
	{
		ops := []uint64{}
		one, _ := fg.attr.st.iConst[1]
		zero, _ := fg.attr.st.iConst[0]
		ops = append(ops, createOp(motor.POP, 0, 0, 0))
		ops = append(ops, createOp(motor.POP, 2, 0, 0))
		ops = append(ops, createOp(motor.POP, 3, 0, 0))
		ops = append(ops, createOp(motor.PUSH, 0, 0, 0))
		ops = append(ops, createOp1(motor.LOAD, 4, addrs[3]))               // address of c
		ops = append(ops, createOp1(motor.LOAD, 5, addrs[1]))               // address of i
		ops = append(ops, createOp1(motor.LOAD, 6, addrs[2]))               // address of j
		ops = append(ops, createOp1(motor.LOAD, 7, zero.fst.achieveAddr())) // address of zero
		ops = append(ops, createOp1(motor.LOAD, 8, one.fst.achieveAddr()))  // address of one
		ops = append(ops, createOp(motor.MOVE, 5, 7, 0))                    // Move(reg_n + 3, reg_n + 5)
		ops = append(ops, createOp(motor.SIZEOF, 6, 2, 0))                  // Sizeof(reg_n + 4, reg_n)
		ops = append(ops, createOp(motor.CMP, 5, 6, 0))                     // Cmp(reg_n + 3, reg_n + 4)
		ops = append(ops, createOp(motor.JAE, 5, 0, 0))                     // Jae(+5)
		ops = append(ops, createOp(motor.INDEX, 9, 2, 5))                   // Index(reg_n + 7, reg_n, reg_n + 3)
		ops = append(ops, createOp(motor.CONCAT, 4, 9, 0))                  // Concat(reg_n + 2, reg_n + 7)
		ops = append(ops, createOp(motor.ADD, 5, 5, 8))                     // Add(reg_n + 3, reg_n + 3, reg_n + 6)
		ops = append(ops, createOp(motor.JMP, 5|motor.SIGN_BIT, 0, 0))      // JMP(-5)
		ops = append(ops, createOp(motor.MOVE, 5, 7, 0))                    // Move(reg_n + 3, reg_n + 5)
		ops = append(ops, createOp(motor.SIZEOF, 6, 3, 0))                  // Sizeof(reg_n + 4, reg_n + 1)
		ops = append(ops, createOp(motor.CMP, 5, 6, 0))                     // Cmp(reg_n + 3, reg_n + 4)
		ops = append(ops, createOp(motor.JAE, 5, 0, 0))                     // Jae(+5)
		ops = append(ops, createOp(motor.INDEX, 10, 3, 5))                  // Index(reg_n + 8, reg_n + 1, reg_n + 3)
		ops = append(ops, createOp(motor.CONCAT, 4, 10, 0))                 // Concat(reg_n + 2, reg_n + 8)
		ops = append(ops, createOp(motor.ADD, 5, 5, 8))                     // Add(reg_n + 3, reg_n + 3, reg_n + 6)
		ops = append(ops, createOp(motor.JMP, 5|motor.SIGN_BIT, 0, 0))      // JMP(-5)
		ops = append(ops, createOp(motor.RET, 4, 0, 0))                     // Ret(reg_n + 2)
		args := []*contractSymbol{}
		args = append(args, &contractSymbol{
			typ: motor.STRING << SUB_OFF,
		})
		args = append(args, &contractSymbol{
			typ: motor.STRING << SUB_OFF,
		})
		regs := []*contractSymbol{}
		regs = append(regs, &contractSymbol{})
		regs = append(regs, &contractSymbol{})
		regs = append(regs, &contractSymbol{})
		regs = append(regs, &contractSymbol{})
		regs = append(regs, &contractSymbol{})
		regs = append(regs, &contractSymbol{})
		regs = append(regs, &contractSymbol{})
		regs = append(regs, &contractSymbol{})
		regs = append(regs, &contractSymbol{})
		fg.attr.st.addFunc(&contractFunc{
			ops:  ops,
			args: args,
			regs: regs,
			idx:  len(ops),
			rtv:  motor.STRING,
		}, "append")
	}
	return nil
}

func (fg *contractGenerator) loadConstant() error {
	var err error
	var addr uint64
	var data motor.FateData

	{
		if addr, err = fg.e.Alloc(virtul.FLASH, motor.DATA_HEADER_SIZE); err != nil { // true
			return err
		}
		if data, err = fg.e.NewVar(motor.CONST_BOOL, addr); err != nil {
			return err
		}
		data.Update(fg.e, uint64(1))
		raddr := uint64(len(fg.attr.st.data)) * motor.DATA_HEADER_SIZE
		sym := recentSym(0, motor.FSCE_DATA, motor.CONSTANT, 0, 0, 0, uint64(len(fg.attr.st.data))*motor.DATA_HEADER_SIZE, addr)
		fg.attr.st.syms = append(fg.attr.st.syms, sym)
		fg.attr.st.data = append(fg.attr.st.data, &contractDataHeader{
			typ: motor.CONST_BOOL,
		})
		fg.e.SetData(data)
		fg.attr.st.bConst[true] = &contractSymbol{
			fst: sym,
			typ: S_CNT | motor.CONST_BOOL<<SUB_OFF,
		}
		fmt.Printf("true: %x\n", raddr)
	}
	{
		if addr, err = fg.e.Alloc(virtul.FLASH, motor.DATA_HEADER_SIZE); err != nil { // false
			return err
		}
		if _, err = fg.e.NewVar(motor.CONST_BOOL, addr); err != nil {
			return err
		}
		raddr := uint64(len(fg.attr.st.data)) * motor.DATA_HEADER_SIZE
		sym := recentSym(0, motor.FSCE_DATA, motor.CONSTANT, 0, 0, 0, uint64(len(fg.attr.st.data))*motor.DATA_HEADER_SIZE, addr)
		fg.attr.st.syms = append(fg.attr.st.syms, sym)
		fg.attr.st.data = append(fg.attr.st.data, &contractDataHeader{
			typ: motor.CONST_BOOL,
		})
		fg.e.SetData(data)
		fg.attr.st.bConst[false] = &contractSymbol{
			fst: sym,
			typ: S_CNT | motor.CONST_BOOL<<SUB_OFF,
		}
		fmt.Printf("flase: %x\n", raddr)
	}
	if err = fg.recentIconst(recent(big.Int).SetInt64(0)); err != nil {
		return err
	}
	if err = fg.recentIconst(recent(big.Int).SetInt64(1)); err != nil {
		return err
	}
	if err = fg.recentIconst(recent(big.Int).SetInt64(-1)); err != nil {
		return err
	}
	if err = fg.recentCconst(byte(0)); err != nil {
		return err
	}
	if err = fg.recentCconst(byte(1)); err != nil {
		return err
	}
	if err = fg.recentFconst(recent(big.Float).SetFloat64(0.0)); err != nil {
		return err
	}
	if err = fg.recentFconst(recent(big.Float).SetFloat64(1.0)); err != nil {
		return err
	}
	if err = fg.recentSconst(""); err != nil {
		return nil
	}
	return nil
}

func (fg *contractGenerator) recentCconst(a byte) error {
	var err error
	var addr uint64
	var data motor.FateData

	if addr, err = fg.e.Alloc(virtul.FLASH, motor.DATA_HEADER_SIZE); err != nil {
		return err
	}
	if data, err = fg.e.NewVar(motor.CONST_CHAR, addr); err != nil {
		return err
	}
	data.Update(fg.e, a)
	raddr := uint64(len(fg.attr.st.data)) * motor.DATA_HEADER_SIZE
	sym := recentSym(0, motor.FSCE_DATA, motor.CONSTANT, 0, 0, 0, uint64(len(fg.attr.st.data))*motor.DATA_HEADER_SIZE, addr)
	fg.attr.st.syms = append(fg.attr.st.syms, sym)
	fg.attr.st.data = append(fg.attr.st.data, &contractDataHeader{
		typ: motor.CONST_CHAR,
	})
	fg.attr.st.cConst[a] = &contractSymbol{
		fst: sym,
		typ: S_CNT | motor.CONST_CHAR<<SUB_OFF,
	}
	fmt.Printf("%v:\t%x\n", a, raddr)
	return nil
}

func (fg *contractGenerator) recentIconst(a *big.Int) error {
	var err error
	var addr uint64
	var data motor.FateData

	if addr, err = fg.e.Alloc(virtul.FLASH, motor.DATA_HEADER_SIZE); err != nil {
		return err
	}
	if data, err = fg.e.NewVar(motor.CONST_INT, addr); err != nil {
		return err
	}
	data.Update(fg.e, a)
	raddr := uint64(len(fg.attr.st.data)) * motor.DATA_HEADER_SIZE
	sym := recentSym(0, motor.FSCE_DATA, motor.CONSTANT, 0, 0, 0, uint64(len(fg.attr.st.data))*motor.DATA_HEADER_SIZE, addr)
	fg.attr.st.syms = append(fg.attr.st.syms, sym)
	fg.attr.st.data = append(fg.attr.st.data, &contractDataHeader{
		typ: motor.CONST_INT,
	})
	switch {
	case a.IsInt64():
		fg.attr.st.iConst[a.Int64()] = &contractSymbol{
			fst: sym,
			typ: S_CNT | motor.CONST_INT<<SUB_OFF,
		}
		fmt.Printf("%v:\t%x\n", a.Int64(), raddr)
	default:
		fg.attr.st.uConst[a.Uint64()] = &contractSymbol{
			fst: sym,
			typ: S_CNT | motor.CONST_INT<<SUB_OFF,
		}
		fmt.Printf("%v:\t%x\n", a.Uint64(), raddr)
	}
	return nil
}

func (fg *contractGenerator) recentFconst(a *big.Float) error {
	var err error
	var addr uint64
	var data motor.FateData

	if addr, err = fg.e.Alloc(virtul.FLASH, motor.DATA_HEADER_SIZE); err != nil {
		return err
	}
	if data, err = fg.e.NewVar(motor.CONST_FLOAT, addr); err != nil {
		return err
	}
	data.Update(fg.e, a)
	raddr := uint64(len(fg.attr.st.data)) * motor.DATA_HEADER_SIZE
	sym := recentSym(0, motor.FSCE_DATA, motor.CONSTANT, 0, 0, 0, uint64(len(fg.attr.st.data))*motor.DATA_HEADER_SIZE, addr)
	fg.attr.st.syms = append(fg.attr.st.syms, sym)
	fg.attr.st.data = append(fg.attr.st.data, &contractDataHeader{
		typ: motor.CONST_FLOAT,
	})
	v, _ := a.Float64()
	fg.attr.st.fConst[v] = &contractSymbol{
		fst: sym,
		typ: S_CNT | motor.CONST_FLOAT<<SUB_OFF,
	}
	fmt.Printf("%f:\t%x\n", v, raddr)
	return nil
}

func (fg *contractGenerator) recentSconst(a string) error {
	var err error
	var addr uint64
	var data motor.FateData

	if addr, err = fg.e.Alloc(virtul.FLASH, motor.DATA_HEADER_SIZE); err != nil {
		return err
	}
	if data, err = fg.e.NewVar(motor.CONST_STRING, addr); err != nil {
		return err
	}
	data.Update(fg.e, a)
	raddr := uint64(len(fg.attr.st.data)) * motor.DATA_HEADER_SIZE
	sym := recentSym(0, motor.FSCE_DATA, motor.CONSTANT, 0, 0, 0, uint64(len(fg.attr.st.data))*motor.DATA_HEADER_SIZE, addr)
	fg.attr.st.syms = append(fg.attr.st.syms, sym)
	fg.attr.st.data = append(fg.attr.st.data, &contractDataHeader{
		typ: motor.CONST_STRING,
	})
	fg.attr.st.sConst[a] = &contractSymbol{
		fst: sym,
		typ: S_CNT | motor.CONST_STRING<<SUB_OFF,
	}
	fmt.Printf("%s:\t%x\n", a, raddr)
	return nil
}

func (fg *contractGenerator) recentLibVar(typ, class int, name string) (uint64, error) {
	var err error
	var idx uint32
	var addr, rAddr uint64

	ok := true
	defer func() {
		if !ok {
			fg.attr.st.addName(name)
			fg.attr.st.sMap[name] = idx
		}
	}()
	if addr, err = fg.e.Alloc(virtul.FLASH, motor.DATA_HEADER_SIZE); err != nil {
		return 0, err
	}
	if _, err = fg.e.NewVar(typ, addr); err != nil {
		return 0, err
	}
	if idx, ok = fg.attr.st.sMap[name]; !ok {
		idx = uint32(len(fg.attr.st.strs))
	}
	sym := &contractSymbol{}
	rAddr = uint64(len(fg.attr.st.data)) * motor.DATA_HEADER_SIZE
	switch class {
	case virtul.RAM:
		sym.fst = recentSym(idx, motor.FSCE_DATA, motor.RAM_VAR, 0, 0, 0, rAddr, addr)
	case virtul.FLASH:
		sym.fst = recentSym(idx, motor.FSCE_DATA, motor.FLASH_VAR, 0, 0, 0, addr, rAddr)
	}
	fg.attr.st.syms = append(fg.attr.st.syms, sym.fst)
	fg.attr.st.data = append(fg.attr.st.data, &contractDataHeader{
		typ: uint32(typ),
	})
	return rAddr, nil
}

func (fg *contractGenerator) recentVar(typ, class int, name string, nameList *contractNameList) error {
	var s int
	var err error
	var idx uint32
	var addr uint64
	var sym *contractSymbol

	if s, sym = nameList.lookUp(name); s != EXT_CUR {
		return fmt.Errorf("NewVar: Cannot Find Symbol: %s", name)
	}
	ok := true
	defer func() {
		if !ok {
			fg.attr.st.addName(name)
			fg.attr.st.sMap[name] = idx
		}
	}()
	if addr, err = fg.e.Alloc(virtul.FLASH, motor.DATA_HEADER_SIZE); err != nil {
		return err
	}
	if _, err = fg.e.NewVar(typ, addr); err != nil {
		return err
	}
	if idx, ok = fg.attr.st.sMap[name]; !ok {
		idx = uint32(len(fg.attr.st.strs))
	}
	switch class {
	case virtul.RAM:
		sym.fst = recentSym(idx, motor.FSCE_DATA, motor.RAM_VAR, 0, 0, 0, uint64(len(fg.attr.st.data))*motor.DATA_HEADER_SIZE, addr)
	case virtul.FLASH:
		sym.fst = recentSym(idx, motor.FSCE_DATA, motor.FLASH_VAR, 0, 0, 0, addr, uint64(len(fg.attr.st.data))*motor.DATA_HEADER_SIZE)
	}
	sym.typ = uint32(S_VAR | typ<<SUB_OFF)
	fg.attr.st.syms = append(fg.attr.st.syms, sym.fst)
	fg.attr.st.data = append(fg.attr.st.data, &contractDataHeader{
		typ: uint32(typ),
	})
	return nil
}

func (fg *contractGenerator) recentMap(ktyp, vtyp, class int, name string, nameList *contractNameList) error {
	var s int
	var err error
	var idx uint32
	var addr uint64
	var sym *contractSymbol

	if s, sym = nameList.lookUp(name); s != EXT_CUR {
		return fmt.Errorf("NewMAP: Cannot Find Symbol: %s", name)
	}
	ok := true
	defer func() {
		if !ok {
			fg.attr.st.addName(name)
			fg.attr.st.sMap[name] = idx
		}
	}()
	if addr, err = fg.e.Alloc(virtul.FLASH, motor.DATA_HEADER_SIZE); err != nil {
		return err
	}
	if _, err = fg.e.NewMap(ktyp, vtyp, addr); err != nil {
		return err
	}
	if idx, ok = fg.attr.st.sMap[name]; !ok {
		idx = uint32(len(fg.attr.st.strs))
	}
	switch class {
	case virtul.RAM:
		sym.fst = recentSym(idx, motor.FSCE_DATA, motor.RAM_VAR, 0, 0, 0, uint64(len(fg.attr.st.data))*motor.DATA_HEADER_SIZE, addr)
	case virtul.FLASH:
		sym.fst = recentSym(idx, motor.FSCE_DATA, motor.FLASH_VAR, 0, 0, 0, addr, uint64(len(fg.attr.st.data))*motor.DATA_HEADER_SIZE)
	}
	sym.typ = uint32(S_VAR | O_MAP<<SUB_OFF | ktyp<<KEY_OFF | vtyp<<VAL_OFF)
	fg.attr.st.syms = append(fg.attr.st.syms, sym.fst)
	fg.attr.st.data = append(fg.attr.st.data, &contractDataHeader{
		typ: motor.MAP,
	})
	return nil
}

// VariableDefinition
func root0(fg *contractGenerator, node *contractNode) error {
	class := virtul.RAM
	if (node.middle.op&SUB_MASK)>>SUB_OFF == O_FLASH {
		class = virtul.FLASH
	}
	name, _ := node.value.(string)
	switch (node.op & SUB_MASK) >> SUB_OFF {
	case O_MAP:
		ktyp, _ := node.left.value.(int)
		vtyp, _ := node.right.value.(int)
		return fg.recentMap(ktyp, vtyp, class, name, node.nameList)
	}
	typ, _ := node.right.value.(int)
	return fg.recentVar(typ, class, name, node.nameList)
}

// FunctionDefinition
func root1(fg *contractGenerator, node *contractNode) error {
	name, _ := node.value.(string)
	{
		fmt.Printf("FUNC %s:\n", name)
	}
	rtv, _ := node.right.value.(int)
	fi := &contractFunc{
		ops:   []uint64{},
		loads: []uint64{},
		rtv:   uint32(rtv),
		regs:  []*contractSymbol{},
		args:  []*contractSymbol{},
		calls: []*contractAttrCall{},
	}
	fg.attr.it.cFunc = fi
	fg.attr.it.ops = []uint64{}
	if err := rootRegistry[G_ARGS](fg, node.left); err != nil {
		return err
	}
	fi.ops = append(fi.ops, fg.attr.it.dump()...)
	if err := rootRegistry[G_BODY](fg, node.middle); err != nil {
		return err
	}
	fi.ops = append(fi.ops, fi.loads...)
	fi.ops = append(fi.ops, fg.attr.it.dump()...)
	if err := fg.allocRet(rtv, fi); err != nil {
		return err
	}
	fg.attr.st.addFunc(fi, name)
	return nil
}

// ParameterList
func root2(fg *contractGenerator, node *contractNode) error {
	fg.attr.it.ops = append(fg.attr.it.ops, createOp(motor.POP, 0, 0, 0))
	for np := node; np != nil; np = np.middle {
		typ, _ := np.right.value.(int)
		name, _ := np.left.value.(string)
		if s, sym := np.nameList.lookUp(name); s != NOT_EXT {
			sym.typ = uint32(S_ARG | typ<<SUB_OFF)
			sym.allocReg(fg.attr.it)
			fg.attr.it.cFunc.args = append(fg.attr.it.cFunc.args, sym)
		} else {
			return fmt.Errorf("Cannot Find Symbol: %s\n", name)
		}
	}
	fg.attr.it.ops = append(fg.attr.it.ops, createOp(motor.PUSH, 0, 0, 0))
	return nil
}

// Func Body
func root3(fg *contractGenerator, node *contractNode) error {
	return rootRegistry[G_BLK](fg, node.left)
}

// BlockItem
func root4(fg *contractGenerator, node *contractNode) error {
	var err error

	for np := node; np != nil && np.left != nil; np = np.middle {
		switch np.left.op & TYP_MASK {
		case O_VAR:
			err = rootRegistry[G_VAR](fg, np.left)
		default:
			err = rootRegistry[G_STMT](fg, np.left)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// Statement
func root5(fg *contractGenerator, node *contractNode) error {
	switch node.op & TYP_MASK {
	case O_IF:
		return rootRegistry[G_SSTMT](fg, node)
	case O_CPDS:
		return rootRegistry[G_CSTMT](fg, node)
	case O_WHILE:
		return rootRegistry[G_ISTMT](fg, node)
	case O_RTN, O_CONT, O_BREAK:
		return rootRegistry[G_JSTMT](fg, node)
	default:
		return rootRegistry[G_ESTMT](fg, node)
	}
}

// CompoundStatement
func root6(fg *contractGenerator, node *contractNode) error {
	return rootRegistry[G_BLK](fg, node.left)
}

// JumpStatement
func root7(fg *contractGenerator, node *contractNode) error {
	switch node.op & TYP_MASK {
	case O_RTN:
		if err := rootRegistry[G_EXPR](fg, node.left); err != nil {
			return err
		}
		x := fg.attr.it.cSym
		if !isRtv(x.typ, fg.attr.it.cFunc.rtv<<SUB_OFF) {
			return fmt.Errorf("Cannot Use type '%s' As type '%s' In Return Argument",
				typName[(x.typ&SUB_MASK)>>SUB_OFF], typName[fg.attr.it.cFunc.rtv])
		}
		x.allocReg(fg.attr.it)
		fg.attr.it.ops = append(fg.attr.it.ops, createOp(motor.RET, x.reg, 0, 0))
	case O_CONT:
		fg.attr.it.ops = append(fg.attr.it.ops, createOp(motor.JMP, 1, 0, 1))
	case O_BREAK:
		fg.attr.it.ops = append(fg.attr.it.ops, createOp(motor.JMP, 1, 0, 2))
	}
	return nil
}

// IterationStatement
func root8(fg *contractGenerator, node *contractNode) error {
	var ops, cops, lops []uint64

	ops = append(ops, fg.attr.it.dump()...)
	if err := rootRegistry[G_EXPR](fg, node.middle); err != nil {
		return err
	}
	x := fg.attr.it.cSym
	if !isBool(x.typ) {
		return fmt.Errorf("Unsupport Type '%s' used as While Condition", typName[(x.typ&SUB_MASK)>>SUB_OFF])
	}
	cops = append(cops, fg.attr.it.dump()...)
	if err := rootRegistry[G_STMT](fg, node.left); err != nil {
		return err
	}
	lops = append(lops, fg.attr.it.dump()...)
	for i, j := 0, len(lops); i < j; i++ {
		if achieveOp(lops[i]) == motor.JMP && achieveC(lops[i]) != 0 {
			switch achieveC(lops[i]) {
			case 1: // CONTINUE
				lops[i] = createOp(motor.JMP, (i+len(cops)+1)|motor.SIGN_BIT, 0, 0)
			case 2: // BREAK
				lops[i] = createOp(motor.JMP, (j - i + 1), 0, 0)
			}
		}
	}
	y, _ := fg.attr.st.bConst[false]
	x.allocReg(fg.attr.it)
	y.allocReg(fg.attr.it)
	fg.attr.it.ops = append(fg.attr.it.ops, ops...)
	fg.attr.it.ops = append(fg.attr.it.ops, cops...)
	fg.attr.it.ops = append(fg.attr.it.ops, createOp(motor.CMP, x.reg, y.reg, 0))
	fg.attr.it.ops = append(fg.attr.it.ops, createOp(motor.JZ, len(lops)+2, 0, 0))
	fg.attr.it.ops = append(fg.attr.it.ops, lops...)
	fg.attr.it.ops = append(fg.attr.it.ops, createOp(motor.JMP, (len(lops)+len(cops)+2)|motor.SIGN_BIT, 0, 0))
	fg.attr.it.ops = append(fg.attr.it.ops, createOp(motor.NOP, 0, 0, 0))
	return nil
}

// SelectionStatement
func root9(fg *contractGenerator, node *contractNode) error {
	var cops, lops, rops []uint64

	if err := rootRegistry[G_EXPR](fg, node.middle); err != nil {
		return err
	}
	x := fg.attr.it.cSym
	if !isBool(x.typ) {
		return fmt.Errorf("Unsupport Type '%s' used as If Condition", typName[(x.typ&SUB_MASK)>>SUB_OFF])
	}
	cops = append(cops, fg.attr.it.dump()...)
	if err := rootRegistry[G_STMT](fg, node.left); err != nil {
		return err
	}
	lops = append(lops, fg.attr.it.dump()...)
	if node.right != nil {
		if err := rootRegistry[G_STMT](fg, node.right); err != nil {
			return err
		}
		rops = append(rops, fg.attr.it.dump()...)
	}
	y, _ := fg.attr.st.bConst[false]
	x.allocReg(fg.attr.it)
	y.allocReg(fg.attr.it)
	fg.attr.it.ops = append(fg.attr.it.ops, cops...)
	fg.attr.it.ops = append(fg.attr.it.ops, createOp(motor.CMP, x.reg, y.reg, 0))
	fg.attr.it.ops = append(fg.attr.it.ops, createOp(motor.JZ, len(lops)+2, 0, 0))
	fg.attr.it.ops = append(fg.attr.it.ops, lops...)
	fg.attr.it.ops = append(fg.attr.it.ops, createOp(motor.JMP, len(rops)+1, 0, 0))
	fg.attr.it.ops = append(fg.attr.it.ops, rops...)
	fg.attr.it.ops = append(fg.attr.it.ops, createOp(motor.NOP, 0, 0, 0))
	return nil
}

// ExpressionStatement
func root10(fg *contractGenerator, node *contractNode) error {
	return rootRegistry[G_EXPR](fg, node.left)
}

// Expression
func root11(fg *contractGenerator, node *contractNode) error {
	for np := node; np != nil; np = np.middle {
		if np.op&TYP_MASK != O_ELIST {
			return rootRegistry[G_ASGN](fg, np)
		}
		if err := rootRegistry[G_ASGN](fg, np.left); err != nil {
			return err
		}
	}
	return nil
}

// AssignmentExpression
func root12(fg *contractGenerator, node *contractNode) error {
	if node.op&TYP_MASK != O_ASSIGN {
		return rootRegistry[G_LGOR](fg, node)
	}
	if err := rootRegistry[G_UNARY](fg, node.left); err != nil {
		return err
	}
	x := fg.attr.it.cSym
	if err := rootRegistry[G_ASGN](fg, node.right); err != nil {
		return err
	}
	y := fg.attr.it.cSym
	if !isLeft(x.typ) {
		return errors.New("Not Left Value")
	}
	_, err := checkType2(x.typ, y.typ, node.op)
	if err != nil {
		return err
	}
	if (x.typ&TYP_MASK) == S_SCS && x.reg == 0 {
		defer func() {
			fg.attr.it.mSym = fg.attr.it.mSym[1:]
			fg.attr.it.kSym = fg.attr.it.kSym[1:]
		}()
		y.allocReg(fg.attr.it)
		fg.attr.it.ops = append(fg.attr.it.ops, createOp(motor.INSERT,
			fg.attr.it.mSym[0].reg, fg.attr.it.kSym[0].reg, y.reg))
	} else {
		x.allocReg(fg.attr.it)
		y.allocReg(fg.attr.it)
		fg.attr.it.ops = append(fg.attr.it.ops, createOp(motor.MOVE, x.reg, y.reg, 0))
	}
	return nil
}

// LogicalOrExpression
func root13(fg *contractGenerator, node *contractNode) error {
	if node.op&TYP_MASK != O_LOR {
		return rootRegistry[G_LGAND](fg, node)
	}
	if err := rootRegistry[G_LGAND](fg, node.left); err != nil {
		return err
	}
	x := fg.attr.it.cSym
	if err := rootRegistry[G_LGOR](fg, node.right); err != nil {
		return err
	}
	y := fg.attr.it.cSym
	_, err := checkType2(x.typ, y.typ, node.op)
	if err != nil {
		return err
	}
	z := &contractSymbol{
		typ: S_TMP | motor.CONST_BOOL<<SUB_OFF,
	}
	b0, _ := fg.attr.st.bConst[true]
	b1, _ := fg.attr.st.bConst[false]
	fg.attr.it.cSym = z
	x.allocReg(fg.attr.it)
	y.allocReg(fg.attr.it)
	z.allocReg(fg.attr.it)
	b0.allocReg(fg.attr.it)
	b1.allocReg(fg.attr.it)
	fg.attr.it.ops = append(fg.attr.it.ops, createOp(motor.CMP, x.reg, b0.reg, 0))
	fg.attr.it.ops = append(fg.attr.it.ops, createOp(motor.JZ, 5, 0, 0))
	fg.attr.it.ops = append(fg.attr.it.ops, createOp(motor.CMP, y.reg, b0.reg, 0))
	fg.attr.it.ops = append(fg.attr.it.ops, createOp(motor.JZ, 3, 0, 0))
	fg.attr.it.ops = append(fg.attr.it.ops, createOp(motor.MOVE, z.reg, b1.reg, 0)) // false
	fg.attr.it.ops = append(fg.attr.it.ops, createOp(motor.JMP, 2, 0, 0))
	fg.attr.it.ops = append(fg.attr.it.ops, createOp(motor.MOVE, z.reg, b0.reg, 0)) // true
	fg.attr.it.ops = append(fg.attr.it.ops, createOp(motor.NOP, 0, 0, 0))
	return nil
}

// LogicalAndExpression
func root14(fg *contractGenerator, node *contractNode) error {
	if node.op&TYP_MASK != O_LAND {
		return rootRegistry[G_INOR](fg, node)
	}
	if err := rootRegistry[G_INOR](fg, node.left); err != nil {
		return err
	}
	x := fg.attr.it.cSym
	if err := rootRegistry[G_LGAND](fg, node.right); err != nil {
		return err
	}
	y := fg.attr.it.cSym
	_, err := checkType2(x.typ, y.typ, node.op)
	if err != nil {
		return err
	}
	z := &contractSymbol{
		typ: S_TMP | motor.CONST_BOOL<<SUB_OFF,
	}
	b0, _ := fg.attr.st.bConst[true]
	b1, _ := fg.attr.st.bConst[false]
	fg.attr.it.cSym = z
	x.allocReg(fg.attr.it)
	y.allocReg(fg.attr.it)
	z.allocReg(fg.attr.it)
	b0.allocReg(fg.attr.it)
	b1.allocReg(fg.attr.it)
	fg.attr.it.ops = append(fg.attr.it.ops, createOp(motor.CMP, x.reg, b0.reg, 0))
	fg.attr.it.ops = append(fg.attr.it.ops, createOp(motor.JZ, 3, 0, 0))
	fg.attr.it.ops = append(fg.attr.it.ops, createOp(motor.MOVE, z.reg, b1.reg, 0)) // false
	fg.attr.it.ops = append(fg.attr.it.ops, createOp(motor.JMP, 6, 0, 0))
	fg.attr.it.ops = append(fg.attr.it.ops, createOp(motor.CMP, y.reg, b0.reg, 0))
	fg.attr.it.ops = append(fg.attr.it.ops, createOp(motor.JZ, 3, 0, 0))
	fg.attr.it.ops = append(fg.attr.it.ops, createOp(motor.MOVE, z.reg, b1.reg, 0)) // false
	fg.attr.it.ops = append(fg.attr.it.ops, createOp(motor.JMP, 2, 0, 0))
	fg.attr.it.ops = append(fg.attr.it.ops, createOp(motor.MOVE, z.reg, b0.reg, 0)) // true
	fg.attr.it.ops = append(fg.attr.it.ops, createOp(motor.NOP, 0, 0, 0))
	return nil
}

// InclusiveOrExpression
func root15(fg *contractGenerator, node *contractNode) error {
	if node.op&TYP_MASK != O_OR {
		return rootRegistry[G_EXOR](fg, node)
	}
	if err := rootRegistry[G_EXOR](fg, node.left); err != nil {
		return err
	}
	x := fg.attr.it.cSym
	if err := rootRegistry[G_INOR](fg, node.right); err != nil {
		return err
	}
	y := fg.attr.it.cSym
	typ, err := checkType2(x.typ, y.typ, node.op)
	if err != nil {
		return err
	}
	z := &contractSymbol{
		typ: uint32(S_TMP | typ<<SUB_OFF),
	}
	fg.attr.it.cSym = z
	x.allocReg(fg.attr.it)
	y.allocReg(fg.attr.it)
	z.allocReg(fg.attr.it)
	fg.attr.it.ops = append(fg.attr.it.ops, createOp(motor.OR, z.reg, x.reg, y.reg))
	return nil
}

// ExclusiveOrExpression
func root16(fg *contractGenerator, node *contractNode) error {
	if node.op&TYP_MASK != O_XOR {
		return rootRegistry[G_AND](fg, node)
	}
	if err := rootRegistry[G_AND](fg, node.left); err != nil {
		return err
	}
	x := fg.attr.it.cSym
	if err := rootRegistry[G_EXOR](fg, node.right); err != nil {
		return err
	}
	y := fg.attr.it.cSym
	typ, err := checkType2(x.typ, y.typ, node.op)
	if err != nil {
		return err
	}
	z := &contractSymbol{
		typ: uint32(S_TMP | typ<<SUB_OFF),
	}
	fg.attr.it.cSym = z
	x.allocReg(fg.attr.it)
	y.allocReg(fg.attr.it)
	z.allocReg(fg.attr.it)
	fg.attr.it.ops = append(fg.attr.it.ops, createOp(motor.XOR, z.reg, x.reg, y.reg))
	return nil
}

// AndExpression
func root17(fg *contractGenerator, node *contractNode) error {
	if node.op&TYP_MASK != O_AND {
		return rootRegistry[G_EQ](fg, node)
	}
	if err := rootRegistry[G_EQ](fg, node.left); err != nil {
		return err
	}
	x := fg.attr.it.cSym
	if err := rootRegistry[G_AND](fg, node.right); err != nil {
		return err
	}
	y := fg.attr.it.cSym
	typ, err := checkType2(x.typ, y.typ, node.op)
	if err != nil {
		return err
	}
	z := &contractSymbol{
		typ: uint32(S_TMP | typ<<SUB_OFF),
	}
	fg.attr.it.cSym = z
	x.allocReg(fg.attr.it)
	y.allocReg(fg.attr.it)
	z.allocReg(fg.attr.it)
	fg.attr.it.ops = append(fg.attr.it.ops, createOp(motor.AND, z.reg, x.reg, y.reg))
	return nil
}

// XhualityExpression
func root18(fg *contractGenerator, node *contractNode) error {
	if typ := node.op & TYP_MASK; typ != O_EQ && typ != O_NE {
		return rootRegistry[G_CMP](fg, node)
	}
	if err := rootRegistry[G_CMP](fg, node.left); err != nil {
		return err
	}
	x := fg.attr.it.cSym
	if err := rootRegistry[G_EQ](fg, node.right); err != nil {
		return err
	}
	y := fg.attr.it.cSym
	_, err := checkType2(x.typ, y.typ, node.op)
	if err != nil {
		return err
	}
	z := &contractSymbol{
		typ: uint32(S_TMP | motor.CONST_BOOL<<SUB_OFF),
	}
	b0, _ := fg.attr.st.bConst[true]
	b1, _ := fg.attr.st.bConst[false]
	fg.attr.it.cSym = z
	x.allocReg(fg.attr.it)
	y.allocReg(fg.attr.it)
	z.allocReg(fg.attr.it)
	b0.allocReg(fg.attr.it)
	b1.allocReg(fg.attr.it)
	fg.attr.it.ops = append(fg.attr.it.ops, createOp(motor.CMP, x.reg, y.reg, 0))
	fg.attr.it.ops = append(fg.attr.it.ops, createOp(motor.JZ, 3, 0, 0))
	switch node.op & TYP_MASK {
	case O_EQ:
		fg.attr.it.ops = append(fg.attr.it.ops, createOp(motor.MOVE, z.reg, b1.reg, 0))
		fg.attr.it.ops = append(fg.attr.it.ops, createOp(motor.JMP, 2, 0, 0))
		fg.attr.it.ops = append(fg.attr.it.ops, createOp(motor.MOVE, z.reg, b0.reg, 0))
	case O_NE:
		fg.attr.it.ops = append(fg.attr.it.ops, createOp(motor.MOVE, z.reg, b0.reg, 0))
		fg.attr.it.ops = append(fg.attr.it.ops, createOp(motor.JMP, 2, 0, 0))
		fg.attr.it.ops = append(fg.attr.it.ops, createOp(motor.MOVE, z.reg, b1.reg, 0))
	}
	fg.attr.it.ops = append(fg.attr.it.ops, createOp(motor.NOP, 0, 0, 0))
	return nil
}

// RelationalExpression
func root19(fg *contractGenerator, node *contractNode) error {
	if typ := node.op & TYP_MASK; typ != O_LT && typ != O_GT && typ != O_LE && typ != O_GE {
		return rootRegistry[G_SHF](fg, node)
	}
	if err := rootRegistry[G_SHF](fg, node.left); err != nil {
		return err
	}
	x := fg.attr.it.cSym
	if err := rootRegistry[G_CMP](fg, node.right); err != nil {
		return err
	}
	y := fg.attr.it.cSym
	_, err := checkType2(x.typ, y.typ, node.op)
	if err != nil {
		return err
	}
	z := &contractSymbol{
		typ: uint32(S_TMP | motor.CONST_BOOL<<SUB_OFF),
	}
	b0, _ := fg.attr.st.bConst[true]
	b1, _ := fg.attr.st.bConst[false]
	fg.attr.it.cSym = z
	x.allocReg(fg.attr.it)
	y.allocReg(fg.attr.it)
	z.allocReg(fg.attr.it)
	b0.allocReg(fg.attr.it)
	b1.allocReg(fg.attr.it)
	fg.attr.it.ops = append(fg.attr.it.ops, createOp(motor.CMP, x.reg, y.reg, 0))
	switch node.op & TYP_MASK {
	case O_LT:
		fg.attr.it.ops = append(fg.attr.it.ops, createOp(motor.JB, 3, 0, 0))
	case O_GT:
		fg.attr.it.ops = append(fg.attr.it.ops, createOp(motor.JA, 3, 0, 0))
	case O_LE:
		fg.attr.it.ops = append(fg.attr.it.ops, createOp(motor.JBE, 3, 0, 0))
	case O_GE:
		fg.attr.it.ops = append(fg.attr.it.ops, createOp(motor.JAE, 3, 0, 0))
	}
	fg.attr.it.ops = append(fg.attr.it.ops, createOp(motor.MOVE, z.reg, b1.reg, 0))
	fg.attr.it.ops = append(fg.attr.it.ops, createOp(motor.JMP, 2, 0, 0))
	fg.attr.it.ops = append(fg.attr.it.ops, createOp(motor.MOVE, z.reg, b0.reg, 0))
	fg.attr.it.ops = append(fg.attr.it.ops, createOp(motor.NOP, 0, 0, 0))
	return nil
}

// ShiftExpression
func root20(fg *contractGenerator, node *contractNode) error {
	if typ := node.op & TYP_MASK; typ != O_LEFT && typ != O_RIGHT {
		return rootRegistry[G_ADD](fg, node)
	}
	if err := rootRegistry[G_ADD](fg, node.left); err != nil {
		return err
	}
	x := fg.attr.it.cSym
	if err := rootRegistry[G_SHF](fg, node.right); err != nil {
		return err
	}
	y := fg.attr.it.cSym
	typ, err := checkType2(x.typ, y.typ, node.op)
	if err != nil {
		return err
	}
	z := &contractSymbol{
		typ: uint32(S_TMP | typ<<SUB_OFF),
	}
	fg.attr.it.cSym = z
	x.allocReg(fg.attr.it)
	y.allocReg(fg.attr.it)
	z.allocReg(fg.attr.it)
	switch node.op & TYP_MASK {
	case O_LEFT:
		fg.attr.it.ops = append(fg.attr.it.ops, createOp(motor.SHL, z.reg, x.reg, y.reg))
	case O_RIGHT:
		fg.attr.it.ops = append(fg.attr.it.ops, createOp(motor.SHR, z.reg, x.reg, y.reg))
	}
	return nil
}

// AdditiveExpression
func root21(fg *contractGenerator, node *contractNode) error {
	if typ := node.op & TYP_MASK; typ != O_ADD && typ != O_SUB {
		return rootRegistry[G_MUL](fg, node)
	}
	if err := rootRegistry[G_MUL](fg, node.left); err != nil {
		return err
	}
	x := fg.attr.it.cSym
	if err := rootRegistry[G_ADD](fg, node.right); err != nil {
		return err
	}
	y := fg.attr.it.cSym
	typ, err := checkType2(x.typ, y.typ, node.op)
	if err != nil {
		return err
	}
	z := &contractSymbol{
		typ: uint32(S_TMP | typ<<SUB_OFF),
	}
	fg.attr.it.cSym = z
	x.allocReg(fg.attr.it)
	y.allocReg(fg.attr.it)
	z.allocReg(fg.attr.it)
	switch node.op & TYP_MASK {
	case O_ADD:
		fg.attr.it.ops = append(fg.attr.it.ops, createOp(motor.ADD, z.reg, x.reg, y.reg))
	case O_SUB:
		fg.attr.it.ops = append(fg.attr.it.ops, createOp(motor.SUB, z.reg, x.reg, y.reg))
	}
	return nil
}

// MultiplicativeExpression
func root22(fg *contractGenerator, node *contractNode) error {
	if typ := node.op & TYP_MASK; typ != O_MUL && typ != O_DIV && typ != O_MOD {
		return rootRegistry[G_UNARY](fg, node)
	}
	if err := rootRegistry[G_UNARY](fg, node.left); err != nil {
		return err
	}
	x := fg.attr.it.cSym
	if err := rootRegistry[G_MUL](fg, node.right); err != nil {
		return err
	}
	y := fg.attr.it.cSym
	typ, err := checkType2(x.typ, y.typ, node.op)
	if err != nil {
		return err
	}
	z := &contractSymbol{
		typ: uint32(S_TMP | typ<<SUB_OFF),
	}
	fg.attr.it.cSym = z
	x.allocReg(fg.attr.it)
	y.allocReg(fg.attr.it)
	z.allocReg(fg.attr.it)
	switch node.op & TYP_MASK {
	case O_MUL:
		fg.attr.it.ops = append(fg.attr.it.ops, createOp(motor.MUL, z.reg, x.reg, y.reg))
	case O_DIV:
		fg.attr.it.ops = append(fg.attr.it.ops, createOp(motor.DIV, z.reg, x.reg, y.reg))
	case O_MOD:
		fg.attr.it.ops = append(fg.attr.it.ops, createOp(motor.MOD, z.reg, x.reg, y.reg))
	}
	return nil
}

// UnaryExpression
func root23(fg *contractGenerator, node *contractNode) error {
	switch node.op & TYP_MASK {
	case O_UNARY:
		if err := rootRegistry[G_UNARY](fg, node.left); err != nil {
			return err
		}
		x := fg.attr.it.cSym
		if isMap(x.typ) || isFunc(x.typ) {
			return fmt.Errorf("Unsupport Type '%s' Use Unary Operator", typName[(x.typ&SUB_MASK)>>SUB_OFF])
		}
		op, _ := node.value.(int)
		switch op {
		case NOT:
			if !isInt(x.typ) && !isChar(x.typ) {
				return fmt.Errorf("Unsupport Type '%s' Use ^ Operator", typName[(x.typ&SUB_MASK)>>SUB_OFF])
			}
			z := &contractSymbol{
				typ: uint32(S_TMP | (x.typ & SUB_MASK)),
			}
			fg.attr.it.cSym = z
			x.allocReg(fg.attr.it)
			z.allocReg(fg.attr.it)
			fg.attr.it.ops = append(fg.attr.it.ops, createOp(motor.MOVE, z.reg, x.reg, 0))
			fg.attr.it.ops = append(fg.attr.it.ops, createOp(motor.NOT, z.reg, z.reg, 0))
			return nil
		case ADD:
			if !isInt(x.typ) && !isChar(x.typ) && !isFloat(x.typ) {
				return fmt.Errorf("Unsupport Type '%s' Use + Unary Operator", typName[(x.typ&SUB_MASK)>>SUB_OFF])
			}
			fg.attr.it.cSym = x
			x.allocReg(fg.attr.it)
			return nil
		case SUB:
			if !isInt(x.typ) && !isChar(x.typ) && !isFloat(x.typ) {
				return fmt.Errorf("Unsupport Type '%s' Use - Unary Operator", typName[(x.typ&SUB_MASK)>>SUB_OFF])
			}
			z := &contractSymbol{
				typ: uint32(S_TMP | (x.typ & SUB_MASK)),
			}
			fg.attr.it.cSym = z
			x.allocReg(fg.attr.it)
			z.allocReg(fg.attr.it)
			fg.attr.it.ops = append(fg.attr.it.ops, createOp(motor.MOVE, z.reg, x.reg, 0))
			fg.attr.it.ops = append(fg.attr.it.ops, createOp(motor.NEG, z.reg, z.reg, 0))
			return nil
		}
	case O_SIZEOF:
		return rootRegistry[G_SIZEOF](fg, node.left)
	default:
		return rootRegistry[G_PTF](fg, node)
	}
	return nil
}

// SIZEOF
func root24(fg *contractGenerator, node *contractNode) error {
	if err := rootRegistry[G_UNARY](fg, node); err != nil {
		return err
	}
	x := fg.attr.it.cSym
	z := &contractSymbol{
		typ: uint32(S_TMP | motor.CONST_INT<<SUB_OFF),
	}
	fg.attr.it.cSym = z
	x.allocReg(fg.attr.it)
	z.allocReg(fg.attr.it)
	fg.attr.it.ops = append(fg.attr.it.ops, createOp(motor.SIZEOF, z.reg, x.reg, 0))
	return nil
}

// PostfixExpression
func root25(fg *contractGenerator, node *contractNode) error {
	switch node.op & TYP_MASK {
	case O_INC:
		if err := rootRegistry[G_PRM](fg, node.left); err != nil {
			return err
		}
		x := fg.attr.it.cSym
		if !isLeft(x.typ) {
			return errors.New("Not Left Value")
		}
		var y *contractSymbol
		switch {
		case isInt(x.typ):
			y, _ = fg.attr.st.iConst[1]
		case isChar(x.typ):
			y, _ = fg.attr.st.cConst[byte(1)]
		case isFloat(x.typ):
			y, _ = fg.attr.st.fConst[1.0]
		default:
			return fmt.Errorf("Unsupport Type '%s' Use INC", typName[(x.typ&SUB_MASK)>>SUB_OFF])
		}
		fg.attr.it.cSym = x
		x.allocReg(fg.attr.it)
		y.allocReg(fg.attr.it)
		fg.attr.it.ops = append(fg.attr.it.ops, createOp(motor.ADD, x.reg, x.reg, y.reg))
		return nil
	case O_DEC:
		if err := rootRegistry[G_PRM](fg, node.left); err != nil {
			return err
		}
		x := fg.attr.it.cSym
		if !isLeft(x.typ) {
			return errors.New("Not Left Value")
		}
		var y *contractSymbol
		switch {
		case isInt(x.typ):
			y, _ = fg.attr.st.iConst[1]
		case isChar(x.typ):
			y, _ = fg.attr.st.cConst[byte(1)]
		case isFloat(x.typ):
			y, _ = fg.attr.st.fConst[1.0]
		default:
			return fmt.Errorf("Unsupport Type '%s' Use DEC", typName[(x.typ&SUB_MASK)>>SUB_OFF])
		}
		fg.attr.it.cSym = x
		x.allocReg(fg.attr.it)
		y.allocReg(fg.attr.it)
		fg.attr.it.ops = append(fg.attr.it.ops, createOp(motor.SUB, x.reg, x.reg, y.reg))
		return nil
	case O_MACC:
		if err := rootRegistry[G_PRM](fg, node.left); err != nil {
			return err
		}
		x := fg.attr.it.cSym
		if !isMap(x.typ) {
			return fmt.Errorf("Unsupport Type '%s' Used As Map", typName[(x.typ&SUB_MASK)>>SUB_OFF])
		}
		if err := rootRegistry[G_EXPR](fg, node.right); err != nil {
			return err
		}
		y := fg.attr.it.cSym
		if !isKeyMatch(x.typ, y.typ) {
			return errors.New("Map Access: Type Don't Match")
		}
		z := &contractSymbol{
			typ: uint32(S_SCS | ((x.typ&VAL_MASK)>>VAL_OFF)<<SUB_OFF),
		}
		fg.attr.it.cSym = z
		fg.attr.it.mSym = append(fg.attr.it.mSym, x)
		fg.attr.it.kSym = append(fg.attr.it.kSym, y)
		x.allocReg(fg.attr.it)
		y.allocReg(fg.attr.it)
		return nil
	case O_CALL:
		if err := rootRegistry[G_PRM](fg, node.left); err != nil {
			return err
		}
		x := fg.attr.it.cSym
		if !isFunc(x.typ) {
			return errors.New("Expect Function")
		}
		name, ok := node.left.value.(string)
		if !ok {
			return errors.New("Expect Function Name")
		}
		f, ok := fg.attr.st.funcs[name]
		if !ok {
			return fmt.Errorf("Please Define Function '%s' First", name)
		}
		fg.attr.it.args = append(fg.attr.it.args, []*contractSymbol{})
		if node.right != nil {
			if err := rootRegistry[G_PARA](fg, node.right); err != nil {
				return err
			}
		}
		z := &contractSymbol{
			typ: uint32(S_TMP | f.rtv<<SUB_OFF),
		}
		fg.attr.it.cSym = z
		z.allocReg(fg.attr.it)
		call := &contractAttrCall{
			name: name,
			args: fg.attr.it.args[len(fg.attr.it.args)-1],
		}
		fg.attr.it.args = fg.attr.it.args[:len(fg.attr.it.args)-1]
		fg.attr.it.cFunc.calls = append(fg.attr.it.cFunc.calls, call)
		fg.attr.it.ops = append(fg.attr.it.ops, createOp(motor.CALL, 0, 0, 0))
		fg.attr.it.ops = append(fg.attr.it.ops, createOp(motor.MOVE, z.reg, 0, 0))
		return nil
	default:
		return rootRegistry[G_PRM](fg, node)
	}
}

// PrimaryExpression
func root26(fg *contractGenerator, node *contractNode) error {
	switch node.op & TYP_MASK {
	case O_NAME:
		name, _ := node.value.(string)
		if s, sym := node.nameList.lookUp(name); s != NOT_EXT {
			fg.attr.it.cSym = sym
		} else {
			return fmt.Errorf("Cannot Find Symbol: %s\n", name)
		}
		return nil
	case O_TUPLE:
		return rootRegistry[G_EXPR](fg, node.left)
	case O_CONST:
		var ok bool
		var sym *contractSymbol
		switch (node.op & SUB_MASK) >> SUB_OFF {
		case O_INT:
			v, _ := node.value.(*big.Int)
			switch {
			case v.IsInt64():
				if sym, ok = fg.attr.st.iConst[v.Int64()]; !ok {
					if err := fg.recentIconst(v); err != nil {
						return err
					}
				}
				sym, _ = fg.attr.st.iConst[v.Int64()]
			default:
				if sym, ok = fg.attr.st.uConst[v.Uint64()]; !ok {
					if err := fg.recentIconst(v); err != nil {
						return err
					}
				}
				sym, _ = fg.attr.st.uConst[v.Uint64()]
			}
		case O_CHAR:
			v, _ := node.value.(byte)
			if sym, ok = fg.attr.st.cConst[v]; !ok {
				if err := fg.recentCconst(v); err != nil {
					return err
				}
			}
			sym, _ = fg.attr.st.cConst[v]
		case O_BOOL:
			v, _ := node.value.(bool)
			sym, _ = fg.attr.st.bConst[v]
		case O_FLOAT:
			v, _ := node.value.(*big.Float)
			f, _ := v.Float64()
			if sym, ok = fg.attr.st.fConst[f]; !ok {
				if err := fg.recentFconst(v); err != nil {
					return err
				}
			}
			sym, _ = fg.attr.st.fConst[f]
		case O_STRING:
			v, _ := node.value.(string)
			if sym, ok = fg.attr.st.sConst[v]; !ok {
				if err := fg.recentSconst(v); err != nil {
					return err
				}
			}
			sym, _ = fg.attr.st.sConst[v]
		}
		fg.attr.it.cSym = sym
		return nil
	}
	return nil
}

// ParameterList
func root27(fg *contractGenerator, node *contractNode) error {
	var err error
	var typ uint32
	var n *contractNode

	if node.op&TYP_MASK != O_ELIST {
		n = node
		goto OUT
	}
	if err = rootRegistry[G_PARA](fg, node.middle); err != nil {
		return err
	}
	n = node.left
OUT:
	if err = rootRegistry[G_ASGN](fg, n); err != nil {
		return err
	}
	fg.attr.it.cSym.allocReg(fg.attr.it)
	{
		fmt.Printf("Param  = %p, %v, reg = %v\n", n, n.value, fg.attr.it.cSym.reg)
	}
	switch {
	case isMap(fg.attr.it.cSym.typ):
		typ = motor.MAP << SUB_OFF
	default:
		typ = fg.attr.it.cSym.typ
	}
	fg.attr.it.args[len(fg.attr.it.args)-1] = append([]*contractSymbol{&contractSymbol{
		typ: typ,
	}}, fg.attr.it.args[len(fg.attr.it.args)-1]...)
	fg.attr.it.ops = append(fg.attr.it.ops, createOp(motor.PUSH, fg.attr.it.cSym.reg, 0, 0))
	return nil
}

func NewGenerator(name string) *contractGenerator {
	e, err := motor.NewSc(name)
	if err != nil {
		return nil
	}
	return &contractGenerator{
		e: e,
		attr: &contractAttr{
			st: &contractSattr{
				strs:   []byte{},
				ops:    []uint64{},
				syms:   []*diste{},
				data:   []*contractDataHeader{},
				sMap:   make(map[string]uint32),
				funcs:  make(map[string]*contractFunc),
				bConst: make(map[bool]*contractSymbol),
				cConst: make(map[byte]*contractSymbol),
				iConst: make(map[int64]*contractSymbol),
				uConst: make(map[uint64]*contractSymbol),
				sConst: make(map[string]*contractSymbol),
				fConst: make(map[float64]*contractSymbol),
			},
			it: &contractIattr{},
		},
	}
}

func (fg *contractGenerator) Generate(nodes []*contractNode) error {
	var err error

	if err = fg.loadLibrary(); err != nil {
		return err
	}
	for i, j := 0, len(nodes); i < j; i++ {
		switch nodes[i].op & TYP_MASK {
		case O_VAR:
			err = rootRegistry[G_VAR](fg, nodes[i])
		case O_FUNC:
			err = rootRegistry[G_FUNC](fg, nodes[i])
		default:
			err = errors.New("Unsupport Type")
		}
		if err != nil {
			return err
		}
	}
	for _, v := range fg.attr.st.funcs {
		if err = v.backfill(fg); err != nil {
			return err
		}
	}

	strs := fg.attr.st.strs
	strs_length := mixed.E32func(uint32(len(strs)))

	text := []byte{}
	for i, j := 0, len(fg.attr.st.ops); i < j; i++ {
		text = append(text, mixed.E64func(fg.attr.st.ops[i])...)
	}
	text_length := mixed.E32func(uint32(len(text)))

	syms := []byte{}
	for i, j := 0, len(fg.attr.st.syms); i < j; i++ {
		sData, _ := fg.attr.st.syms[i].Show()
		syms = append(syms, sData...)
	}
	syms_length := mixed.E32func(uint32(len(syms)))

	data := []byte{}
	for i, j := 0, len(fg.attr.st.data); i < j; i++ {
		dData, _ := fg.attr.st.data[i].Show()
		data = append(data, dData...)
	}
	data_length := mixed.E32func(uint32(len(data)))

	msg := []byte{}
	msg = append(msg, motor.FSCE_MAGIC...)
	msg = append(msg, strs_length...)
	msg = append(msg, text_length...)
	msg = append(msg, syms_length...)
	msg = append(msg, data_length...)
	msg = append(msg, strs...)
	msg = append(msg, text...)
	msg = append(msg, syms...)
	msg = append(msg, data...)
	fg.e.Flush()
	fg.e.Close()
	fg.e.SetExecute(msg)
	return nil
}
