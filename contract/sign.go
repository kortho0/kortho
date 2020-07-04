package contract

import "fmt"

func recentNL(up *contractNameList) *contractNameList {
	return &contractNameList{
		up:   up,
		syms: make(map[string]*contractSymbol),
	}
}

func (fn *contractNameList) recentNL() *contractNameList {
	n := recentNL(fn)
	for fp := fn.down; fp != nil; fp = fp.next {
		if fp.next == nil {
			fp.next = n
			goto OUT
		}
	}
	fn.down = n
OUT:
	return n
}

func (fn *contractNameList) insert(name string, sym *contractSymbol) error {
	if _, ok := fn.syms[name]; ok {
		return fmt.Errorf("Symbol '%s' Exist", name)
	}
	fn.syms[name] = sym
	return nil
}

func (fn *contractNameList) lookUp(name string) (int, *contractSymbol) {
	if sym, ok := fn.syms[name]; ok {
		return EXT_CUR, sym
	}
	for fp := fn.up; fp != nil; fp = fp.up {
		if sym, ok := fp.syms[name]; ok {
			return EXT_UPP, sym
		}
	}
	return NOT_EXT, nil
}
