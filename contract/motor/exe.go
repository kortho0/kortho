package motor

import (
	"errors"
	"fmt"

	"kortho/contract/mixed"
)

func findString(s []byte, pos int) string {
	buf := []byte{}
	for i, j := pos, len(s); i < j; i++ {
		if s[i] == 0 {
			break
		}
		buf = append(buf, s[i])
	}
	return string(buf)
}

func loadFsce(a []byte) ([]uint64, []*fsceSym, []byte, error) {
	var err error

	fs := fsce{}
	if a, err = fs.Read(a); err != nil {
		return nil, nil, nil, fmt.Errorf("Fsce Load: %v", err)
	}
	if fs.magic != FSCE_MAGIC_NUMBER {
		return nil, nil, nil, errors.New("Fsce Load: Illegal MAGIC NUMBER")
	}

	if len(a) < int(fs.strs) {
		return nil, nil, nil, errors.New("Fsce Load: Illegal Length")
	}
	strSeg := a[:int(fs.strs)]
	a = a[int(fs.strs):]

	if len(a) < int(fs.text) {
		return nil, nil, nil, errors.New("Fsce Load: Illegal Length")
	}
	textSeg := a[:int(fs.text)]
	a = a[int(fs.text):]

	if len(a) < int(fs.syms) {
		return nil, nil, nil, errors.New("Fsce Load: Illegal Length")
	}
	symsSeg := a[:int(fs.syms)]
	a = a[int(fs.syms):]

	if len(a) < int(fs.data) {
		return nil, nil, nil, errors.New("Fsce Load: Illegal Length")
	}
	dataSeg := a[:int(fs.data)]
	a = a[int(fs.data):]

	ops := []uint64{}
	for len(textSeg) >= 8 {
		op, _ := mixed.D64func(textSeg[:8])
		ops = append(ops, op)
		textSeg = textSeg[8:]
	}

	syms := []*fsceSym{}
	for len(symsSeg) >= 40 {
		ste := diste{}
		if symsSeg, err = ste.Read(symsSeg); err != nil {
			return nil, nil, nil, fmt.Errorf("Fsce Load: %v\n", err)
		}
		syms = append(syms, &fsceSym{
			attr:     ste.attr,
			size:     ste.size,
			info:     ste.info,
			value:    ste.value,
			address:  ste.address,
			raddress: ste.raddress,
			name:     findString(strSeg, int(ste.name)),
		})
	}
	return ops, syms, dataSeg, nil
}
