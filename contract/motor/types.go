package motor

import (
	"errors"

	"kortho/contract/virtul"
)

const (
	DEFAULT_STACK_SIZE       = 1024
	DEFAULT_REGISTERS_NUMBER = 1024
	REGISTERS_NUMBER_LIMIT   = 16384 // 2 ^ 14
)

const (
	GENERAL = iota
	ROUTE
)

const (
	TCP = iota
)

type Fate interface {
	Update() error
	Memory() uint64
	Run() (string, error)
	SetData(FateData) error
	Alloc(int, uint64) (uint64, error)
	Dup(int32, string, [][]byte) error
	NewVar(int, uint64) (FateData, error)
	NewMap(int, int, uint64) (FateData, error)

	Close() error
	Flush() error
	SetExecute([]byte) error
}

type contractEngineStack struct {
	sp    int
	size  int
	stack []uint64
}

/*
type contractEngineRoute struct {
	typ  int
	ctrl interface{}
}
*/

type contractEngineRegisters struct {
	pc     int
	offset int
	eflags uint64
	gpRegs []uint64 // rooteral-purpose registers
	//	routeReg []*contractEngineRoute
}

type contractEngineProgramme struct {
	pow  int32
	ops  []uint64
	load []uint64
	sym  *fsceSym
	syms []*fsceSym
}

type contractEngine struct {
	off    []int
	virtul virtul.FateVM
	stk    *contractEngineStack
	regs   *contractEngineRegisters
	prog   *contractEngineProgramme
}

func (e *contractEngine) Memory() uint64 {
	return e.virtul.Size()
}

func (e *contractEngine) NewVar(typ int, address uint64) (FateData, error) {
	return e.recentVar(typ, address)
}

func (e *contractEngine) NewMap(ktyp, vtyp int, address uint64) (FateData, error) {
	return e.recentMap(ktyp, vtyp, address)
}

func (e *contractEngine) Alloc(typ int, size uint64) (uint64, error) {
	return e.virtul.Alloc(typ, size)
}

func (e *contractEngine) SetData(a FateData) error {
	if d, ok := a.(*contractData); ok {
		return e.setData(d)
	}
	return errors.New("Unsupport Data Type")
}

func (e *contractEngine) Close() error {
	return e.virtul.Close()
}

func (e *contractEngine) Flush() error {
	return e.virtul.Flush()
}

func (e *contractEngine) SetExecute(a []byte) error {
	return e.virtul.SetExecute(a)
}
