package motor

import (
	"errors"
	"fmt"
	"math/big"

	"kortho/contract/mixed"
	"kortho/contract/virtul"
	"kortho/contract/virtul/db"
)

func recentEngineStack() *contractEngineStack {
	return &contractEngineStack{
		sp:    0,
		size:  DEFAULT_STACK_SIZE,
		stack: make([]uint64, DEFAULT_STACK_SIZE),
	}
}

func recentEngineRegisters() *contractEngineRegisters {
	return &contractEngineRegisters{
		pc:     0,
		eflags: 0,
		gpRegs: make([]uint64, DEFAULT_REGISTERS_NUMBER),
	}
}

func recentEngineProgramme(gas int32, ops []uint64, sym *fsceSym, syms []*fsceSym) *contractEngineProgramme {
	return &contractEngineProgramme{
		pow:  gas,
		ops:  ops,
		sym:  sym,
		syms: syms,
		load: []uint64{},
	}
}

func (e *contractEngine) argumentAlloc(data []byte) (*contractData, error) {
	if len(data) < 5 {
		return nil, errors.New("Fate Argument Alloc: Illegal Length")
	}
	addr, err := e.virtul.Alloc(virtul.RAM, DATA_HEADER_SIZE)
	if err != nil {
		return nil, fmt.Errorf("Fate Argument Alloc: %v", err)
	}
	typ, _ := mixed.D32func(data[:4])
	arg, err := e.recentVar(int(typ), addr)
	if err != nil {
		e.virtul.Free(addr)
		return nil, fmt.Errorf("Fate Argument Alloc: %v", err)
	}
	switch typ {
	case BOOL, CHAR:
		arg.header.offset0 = uint64(data[4])
		if err = e.setData(arg); err != nil {
			e.remove(arg)
			return nil, fmt.Errorf("Fate Argument Alloc: %v", err)
		}
		return arg, nil
	case STRING:
		if err = e.setString(arg, data[4:]); err != nil {
			e.remove(arg)
			return nil, fmt.Errorf("Fate Argument Alloc: %v", err)
		}
		return arg, nil
	case INT8, UINT8, INT16, UINT16, INT32, UINT32, INT64, UINT64:
		v := recent(big.Int)
		if _, ok := v.SetString(string(data[4:]), 0); !ok {
			e.remove(arg)
			return nil, fmt.Errorf("Fate Argument Alloc: Illegal Value")
		}
		vbuf := fillByte([]byte(v.String()))
		arg.header.offset0, _ = mixed.D64func(vbuf[:8])
		arg.header.offset1, _ = mixed.D64func(vbuf[8:16])
		arg.header.offset2, _ = mixed.D64func(vbuf[16:24])
		if err = e.setData(arg); err != nil {
			e.remove(arg)
			return nil, fmt.Errorf("Fate Argument Alloc: %v", err)
		}
		return arg, nil
	case FLOAT32, FLOAT64:
		v := recent(big.Float)
		if _, ok := v.SetString(string(data[4:])); !ok {
			e.remove(arg)
			return nil, fmt.Errorf("Fate Argument Alloc: Illegal Value")
		}
		vbuf := fillByte([]byte(v.String()))
		arg.header.offset0, _ = mixed.D64func(vbuf[:8])
		arg.header.offset1, _ = mixed.D64func(vbuf[8:16])
		arg.header.offset2, _ = mixed.D64func(vbuf[16:24])
		if err = e.setData(arg); err != nil {
			e.remove(arg)
			return nil, fmt.Errorf("Fate Argument Alloc: %v", err)
		}
		return arg, nil
	default:
		e.remove(arg)
		return nil, errors.New("Fate Argument Init: Unsupport type")
	}
}

func (e *contractEngine) argumentInit(args [][]byte) error {
	for i, j := len(args)-1, 0; i >= j; i-- {
		arg, err := e.argumentAlloc(args[i])
		if err != nil {
			return err
		}
		e.regs.gpRegs[0] = arg.address
		rsc1(e, 0)
	}
	e.regs.gpRegs[0] = 0
	rsc1(e, 0)
	return nil
}

func NewSc(name string) (*contractEngine, error) {
	if db, err := db.New(name); err == nil {
		if _, err := db.GetExecute(); err == nil {
			return nil, fmt.Errorf("Smart Contract '%s' Exist", name)
		}
	} else {
		return nil, err
	}
	v, err := virtul.New([]byte{}, name)
	if err != nil {
		return nil, err
	}
	v.SetPage(virtul.RAM_PAGE_COUNT, append(virtul.Sentry, make([]byte, int(virtul.PAGE_SIZE)-len(virtul.Sentry))...))
	return &contractEngine{
		virtul: v,
	}, nil
}

func New(gas int32, name, funcName string, args [][]byte) (*contractEngine, error) {
	db, err := db.New(name)
	if err != nil {
		return nil, fmt.Errorf("Fate Engine New: %v", err)
	}
	fsceData, err := db.GetExecute()
	if err != nil {
		return nil, fmt.Errorf("Fate Engine New: %v", err)
	}
	ops, syms, data, err := loadFsce(fsceData)
	if err != nil {
		return nil, fmt.Errorf("Fate Engine New: %v", err)
	}
	db.Close()
	for _, sym := range syms {
		if sym.attr == FSCE_TEXT && sym.name == funcName {
			virtul, err := virtul.New(data, name)
			if err != nil {
				return nil, fmt.Errorf("Fate Engine New: %v", err)
			}
			e := &contractEngine{
				virtul: virtul,
				stk:    recentEngineStack(),
				regs:   recentEngineRegisters(),
				prog:   recentEngineProgramme(gas, ops, sym, syms),
			}
			if err := e.argumentInit(args); err != nil {
				return nil, err
			}
			return e, nil
		}
	}
	return nil, fmt.Errorf("contract Exectue: cannot find function %s", funcName)
}

func (e *contractEngine) Dup(gas int32, funcName string, args [][]byte) error {
	for _, sym := range e.prog.syms {
		if sym.attr == FSCE_TEXT && sym.name == funcName {
			e.prog.pow = gas
			e.prog.sym = sym
			return e.argumentInit(args)
		}
	}
	return fmt.Errorf("Fate Exectue: Cannot Find Function %s", funcName)
}

func (e *contractEngine) Run() (string, error) {
	e.off = append([]int{}, int(e.prog.sym.size))
	for i, j := int(e.prog.sym.value), len(e.prog.ops); i < j; {
		op := e.prog.ops[i]
		opCode := achieveOpCode(op)
		opCons := achieveOpCons(opCode)
		if e.prog.pow -= opCons; e.prog.pow < 0 {
			return "", errors.New("Fate Engine: Out of Power")
		}
		e.regs.pc = i
		if opCode >= len(rscRegistry) {
			return "", errors.New("Fate Engine: Unsupport operation code")
		}
		if err := rscRegistry[opCode](e, op); err != nil {
			return "", err
		}
		i = e.regs.pc
		if opCode == RET && e.stk.sp == 0 {
			break
		}
	}
	d, err := e.achieveData(achieveBCR(e.regs.gpRegs[0]))
	if err != nil {
		return "", fmt.Errorf("Fate Engine: %v", err)
	}
	r, err := e.achieveVisualData(d)
	if err != nil {
		return "", fmt.Errorf("Fate Engine: %v", err)
	}
	return r, nil
}

func (e *contractEngine) mapping() error {
	for _, v := range e.prog.load {
		for _, sym := range e.prog.syms {
			if sym.attr == FSCE_DATA && sym.address == v {
				src, err := e.achieveData(sym.raddress)
				if err != nil {
					return fmt.Errorf("Fate Mapping: %v", err)
				}
				dst, err := e.achieveData(sym.address)
				if err != nil {
					return fmt.Errorf("Fate Mapping: %v", err)
				}
				if err = e.move(dst, src); err != nil {
					return fmt.Errorf("Fate Mapping: %v", err)
				}
				break
			}
		}
	}
	return nil
}

func (e *contractEngine) Update() error {
	if err := e.mapping(); err != nil {
		return err
	}
	if err := e.virtul.Flush(); err != nil {
		return err
	}
	e.prog.load = []uint64{}
	return nil
}

func (e *contractEngine) findFunc(address uint64) *fsceSym {
	for _, sym := range e.prog.syms {
		if sym.attr == FSCE_TEXT && sym.value == uint32(address) {
			return sym
		}
	}
	return nil
}
