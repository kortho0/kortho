package motor

import (
	"fmt"

	"kortho/contract/virtul/db"
)

type asmInfo struct {
	idx  int
	name string
}

var vars []*asmInfo

var funcs []*asmInfo

func printOp(op uint64) {
	code := achieveOpCode(op)
	fmt.Printf("\t%s\t", codeName[code])
	if len(codeName[code]) < 4 {
		fmt.Printf("\t")
	}
	switch code {
	case CALL:
		for _, v := range funcs {
			if v.idx == int(achieveBCR(op)) {
				fmt.Printf("%s\n", v.name)
				return
			}
		}
	case LOAD:
		for _, v := range vars {
			if v.idx == int(achieveBCR(op)) {
				fmt.Printf("%v\t%s\n", achieveA(op), v.name)
				return
			}
		}
		fmt.Printf("%v\t%x\n", achieveA(op), achieveBCR(op))
	case JA, JB, JZ, JMP, JAE, JBE:
		if achieveA(op)&SIGN_BIT != 0 {
			fmt.Printf("-%v\t%v\t%v\n", achieveA(op)&DATA_BIT, achieveB(op), achieveC(op))
		} else {
			fmt.Printf("%v\t%v\t%v\n", achieveA(op), achieveB(op), achieveC(op))
		}
	default:
		fmt.Printf("%v\t%v\t%v\n", achieveA(op), achieveB(op), achieveC(op))
	}
}

func Asm(name string) error {
	var addr int

	db, err := db.New(name)
	if err != nil {
		return fmt.Errorf("Fate Engine New: %v", err)
	}
	fsceData, err := db.GetExecute()
	if err != nil {
		return fmt.Errorf("Fate Engine New: %v", err)
	}
	ops, syms, _, err := loadFsce(fsceData)
	if err != nil {
		return fmt.Errorf("Fate Engine New: %v", err)
	}
	db.Close()
	for _, sym := range syms {
		if sym.attr == FSCE_TEXT {
			fmt.Printf("Func %s:\t%v\t%x\n", sym.name, sym.size, sym.value)
			funcs = append(funcs, &asmInfo{
				name: sym.name,
				idx:  int(sym.value),
			})
		}
		if sym.attr == FSCE_DATA {
			switch sym.info {
			case RAM_VAR:
				addr = int(sym.address)
				fmt.Printf("Data %s:\n\t%s\t%x\t%x\n", sym.name, "ram var", sym.address, sym.raddress)
			case CONSTANT:
				addr = int(sym.address)
				fmt.Printf("Data %s:\n\t%s\t%x\t%x\n", sym.name, "constant", sym.address, sym.raddress)
			case FLASH_VAR:
				addr = int(sym.raddress)
				fmt.Printf("Data %s:\n\t%s\t%x\t%x\n", sym.name, "flash var", sym.address, sym.raddress)
			}
			if sym.info != CONSTANT {
				vars = append(vars, &asmInfo{
					idx:  addr,
					name: sym.name,
				})
			}
		}
	}
	for i, j := 0, len(ops); i < j; i++ {
		for _, v := range funcs {
			if i == v.idx {
				fmt.Printf("%s:\n", v.name)
				break
			}
		}
		printOp(ops[i])
	}
	return nil
}
