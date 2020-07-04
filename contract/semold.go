package contract

const (
	EXT_CUR = iota
	EXT_UPP
	NOT_EXT
)

type contractSymbol struct {
	reg int
	typ uint32
	fst *diste
}

type contractNameList struct {
	up   *contractNameList
	down *contractNameList
	next *contractNameList
	syms map[string]*contractSymbol
}
