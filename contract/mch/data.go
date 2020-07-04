package main

import (
	"errors"
	"fmt"
	"math"

	"kortho/contract/mixed"
	"kortho/contract/virtul"
)

func achieveMask(a uint64) uint64 {
	switch a {
	case 1:
		return 0xFF
	case 2:
		return 0xFFFF
	case 4:
		return 0xFFFFFFFF
	case 8:
		return 0xFFFFFFFFFFFFFFFF
	}
	return 0
}

func (e *contractEngine) achieveData(address uint64) (*contractData, error) {
	h := contractDataHeader{}
	if hData, err := e.virtul.GetRawData(address, DATA_HEADER_SIZE); err != nil {
		return nil, err
	} else if _, err = h.Read(hData); err != nil {
		return nil, err
	}
	return &contractData{
		header:  &h,
		address: address,
	}, nil
}

func (e *contractEngine) setData(a *contractData) error {
	hData, _ := a.header.Show()
	return e.virtul.SetRawData(a.address, DATA_HEADER_SIZE, hData)
}

func (e *contractEngine) achieveVisualData(a *contractData) (string, error) {
	switch a.header.typ {
	case BOOL:
		switch a.header.offset {
		case 1: // true
			return fmt.Sprintf("bool: true"), nil
		default:
			return fmt.Sprintf("bool: false"), nil
		}
	case CHAR:
		return fmt.Sprintf("char: %c", byte(a.header.offset&0xFF)), nil
	case INT8:
		return fmt.Sprintf("int8: %v", int8(a.header.offset&0xFF)), nil
	case UINT8:
		return fmt.Sprintf("uint8: %v", uint8(a.header.offset&0xFF)), nil
	case INT16:
		return fmt.Sprintf("int16: %v", int16(a.header.offset&0xFFFF)), nil
	case UINT16:
		return fmt.Sprintf("uint16: %v", uint16(a.header.offset&0xFFFF)), nil
	case INT32:
		return fmt.Sprintf("int32: %v", int32(a.header.offset&0xFFFFFFFF)), nil
	case UINT32:
		return fmt.Sprintf("uint32: %v", uint32(a.header.offset&0xFFFFFFFF)), nil
	case INT64:
		return fmt.Sprintf("int64: %v", int64(a.header.offset)), nil
	case UINT64:
		return fmt.Sprintf("uint64: %v", a.header.offset), nil
	case FLOAT32:
		f := math.Float32frombits(uint32(a.header.offset))
		return fmt.Sprintf("float32: %v", f), nil
	case FLOAT64:
		f := math.Float64frombits(a.header.offset)
		return fmt.Sprintf("float64: %v", f), nil
	case STRING:
		if s, err := e.achieveString(a); err == nil {
			return fmt.Sprintf("string: %s", s), nil
		} else {
			return "", err
		}
	case CONST_INT:
		return fmt.Sprintf("const int: %v", int64(a.header.offset)), nil
	case CONST_BOOL:
		switch a.header.offset {
		case 1: // true
			return fmt.Sprintf("const bool: true"), nil
		default:
			return fmt.Sprintf("const bool: false"), nil
		}
	case CONST_CHAR:
		return fmt.Sprintf("const char: %c", byte(a.header.offset&0xFF)), nil
	case CONST_FLOAT:
		f := math.Float64frombits(a.header.offset)
		return fmt.Sprintf("const float: %v", f), nil
	case CONST_STRING:
		if s, err := e.achieveString(a); err == nil {
			return fmt.Sprintf("const string: %s", s), nil
		} else {
			return "", err
		}
	default:
		return "", errors.New("Fate Visual Data: Unsupport type")
	}
}

func (e *contractEngine) achieveBinaryData(a *contractData) ([]byte, error) {
	buf := []byte{}
	switch a.header.typ {
	case INT16, UINT16:
		buf = append(buf, mixed.E16func(uint16(a.header.offset&0xFFFF))...)
	case INT32, UINT32, FLOAT32:
		buf = append(buf, mixed.E32func(uint32(a.header.offset&0xFFFFFFFF))...)
	case INT64, UINT64, FLOAT64, CONST_INT, CONST_FLOAT:
		buf = append(buf, mixed.E64func(a.header.offset)...)
	case STRING, CONST_STRING:
		s, err := e.achieveString(a)
		if err != nil {
			return nil, fmt.Errorf("Fate Get Binary Data: %v", err)
		}
		buf = append(buf, s...)
	case MAP:
		m, err := e.achieveMap(a)
		if err != nil {
			return nil, fmt.Errorf("Fate Get Binary Data: %v", err)
		}
		buf = append(buf, m...)
	case BOOL, CHAR, INT8, UINT8, CONST_BOOL, CONST_CHAR:
		buf = append(buf, mixed.E8func(uint8(a.header.offset&0xFF))...)
	default:
		return nil, errors.New("Fate Get Binary Data: Unsupport type")
	}
	return buf, nil
}

func (e *contractEngine) sizeof(a *contractData) uint64 {
	if a.header.typ != STRING && a.header.typ != CONST_STRING {
		return a.header.length
	}
	return strlen(a.header.length)
}

func (e *contractEngine) recentVar(typ int, address uint64) (*contractData, error) {
	switch typ {
	case BOOL, CHAR, INT8, INT16, INT32, INT64, UINT8, UINT16, UINT32, UINT64,
		FLOAT32, FLOAT64, CONST_INT, CONST_CHAR, CONST_BOOL, CONST_FLOAT:
		a := &contractData{
			address: address,
			header: &contractDataHeader{
				offset: 0,
				typ:    uint32(typ),
				length: uint64(ltb[typ]),
			},
		}
		e.setData(a)
		return a, nil
	case STRING, CONST_STRING:
		addr, err := e.virtul.Alloc(virtul.StoreType(address), DEFAULT_STRING_LENGTH*DATA_HEADER_SIZE)
		if err != nil {
			return nil, fmt.Errorf("Fate New Variable: %v", err)
		}
		if err := e.virtul.SetRawData(addr, DEFAULT_STRING_LENGTH*DATA_HEADER_SIZE, dcl); err != nil {
			return nil, fmt.Errorf("Fate New Variable: %v", err)
		}
		a := &contractData{
			address: address,
			header: &contractDataHeader{
				offset: addr,
				typ:    uint32(typ),
				length: uint64(ltb[typ]) | uint64(DEFAULT_STRING_LENGTH)<<32,
			},
		}
		e.setData(a)
		return a, nil
	case MAP:
		return nil, errors.New("Fate New Variable: Don't support Map")
	default:
		return nil, errors.New("Fate New Variable: Unsupport type")
	}
}

func (e *contractEngine) typeCmp(a, b *contractData) (int, error) {
	if r := a.header.typ - b.header.typ; r == 0 {
		if a.header.typ == MAP {
			if a.header.length == 0 || b.header.length == 0 {
				return -1, errors.New("Fate Type Cmp: Empty Map")
			}
			ka, va, _ := e.findByIndex(a, 0)
			kb, vb, _ := e.findByIndex(b, 0)
			if rk, err := e.typeCmp(ka, kb); err != nil {
				return -1, fmt.Errorf("Fate Type Cmp: %v", err)
			} else if rk != 0 {
				return rk, nil
			}
			if rv, err := e.typeCmp(va, vb); err != nil {
				return -1, fmt.Errorf("Fate Type Cmp: %v", err)
			} else if rv != 0 {
				return rv, nil
			}
			return 0, nil
		}
		return 0, nil
	} else {
		return int(r), nil
	}
}

func (e *contractEngine) remove(a *contractData) error {
	switch a.header.typ {
	case MAP:
		if err := e.mapRemove(a); err != nil {
			return fmt.Errorf("Fate Remove: %v", err)
		}
	case STRING, CONST_STRING:
		if err := e.stringRemove(a); err != nil {
			return fmt.Errorf("Fate Remove: %v", err)
		}
	case BOOL, CHAR, INT8, INT16, INT32, INT64, UINT8, UINT16, UINT32, UINT64,
		FLOAT32, FLOAT64, CONST_INT, CONST_CHAR, CONST_BOOL, CONST_FLOAT:
	default:
		return errors.New("Fate Remove: Unsupport type")
	}
	return e.virtul.Free(a.address)
}

func (e *contractEngine) dup(a *contractData, typ int) (*contractData, error) {
	addr, err := e.virtul.Alloc(typ, DATA_HEADER_SIZE)
	if err != nil {
		return nil, fmt.Errorf("Fate Dup: %v", err)
	}
	switch a.header.typ {
	case MAP:
		if bk, err := e.achieveBucket(a); err == nil {
			if b, err := e.recentMap(int(bk.header.ktyp), int(bk.header.vtyp), addr); err == nil {
				if err = e.move(b, a); err != nil {
					e.virtul.Free(addr)
					return nil, fmt.Errorf("Fate Dup: %v", err)
				}
				e.setData(b)
				return b, nil
			}
			e.virtul.Free(addr)
			return nil, fmt.Errorf("Fate Dup: %v", err)
		}
		e.virtul.Free(addr)
		return nil, fmt.Errorf("Fate Dup: %v", err)
	case BOOL, CHAR, INT8, INT16, INT32, INT64, UINT8, UINT16, UINT32, UINT64,
		FLOAT32, FLOAT64, STRING:
		if b, err := e.recentVar(int(a.header.typ), addr); err == nil {
			if err = e.move(b, a); err != nil {
				e.virtul.Free(addr)
				return nil, fmt.Errorf("Fate Dup: %v", err)
			}
			e.setData(b)
			return b, nil
		}
		e.virtul.Free(addr)
		return nil, fmt.Errorf("Fate Dup: %v", err)
	default:
		return nil, errors.New("Fate Dup: Unsupport type")
	}
}

func (e *contractEngine) cmp(a, b *contractData) (int, error) {
	switch a.header.typ {
	case MAP:
		return e.mapCmp(a, b)
	case STRING, CONST_STRING:
		return e.stringCmp(a, b)
	case BOOL, CONST_BOOL:
		if b.header.typ != BOOL && b.header.typ != CONST_BOOL {
			return 0, errors.New("Fate Compare: Type Error")
		}
		x := a.header.offset & 0xFF
		y := b.header.offset & 0xFF
		switch {
		case x > y:
			return 1, nil
		case x == y:
			return 0, nil
		default:
			return -1, nil
		}
	case CHAR, CONST_CHAR:
		if b.header.typ != CHAR && b.header.typ != CONST_CHAR {
			return 0, errors.New("Fate Compare: Type Error")
		}
		x := byte(a.header.offset & 0xFF)
		y := byte(b.header.offset & 0xFF)
		switch {
		case x > y:
			return 1, nil
		case x == y:
			return 0, nil
		default:
			return -1, nil
		}
	case FLOAT32:
		if b.header.typ != a.header.typ && b.header.typ != CONST_FLOAT {
			return 0, errors.New("Fate Compare: Type Error")
		}
		x := math.Float32frombits(uint32(a.header.offset & 0xFFFFFFFF))
		y := math.Float32frombits(uint32(b.header.offset & 0xFFFFFFFF))
		switch {
		case x > y:
			return 1, nil
		case x == y:
			return 0, nil
		default:
			return -1, nil
		}
	case FLOAT64:
		if b.header.typ != a.header.typ && b.header.typ != CONST_FLOAT {
			return 0, errors.New("Fate Compare: Type Error")
		}
		x := math.Float64frombits(a.header.offset)
		y := math.Float64frombits(b.header.offset)
		switch {
		case x > y:
			return 1, nil
		case x == y:
			return 0, nil
		default:
			return -1, nil
		}
	case CONST_FLOAT:
		switch b.header.typ {
		case FLOAT32:
			x := math.Float32frombits(uint32(a.header.offset & 0xFFFFFFFF))
			y := math.Float32frombits(uint32(b.header.offset & 0xFFFFFFFF))
			switch {
			case x > y:
				return 1, nil
			case x == y:
				return 0, nil
			default:
				return -1, nil
			}
		case FLOAT64, CONST_FLOAT:
			x := math.Float64frombits(a.header.offset)
			y := math.Float64frombits(b.header.offset)
			switch {
			case x > y:
				return 1, nil
			case x == y:
				return 0, nil
			default:
				return -1, nil
			}
		default:
			return 0, errors.New("Fate Compare: Type Error")
		}
	case INT8:
		if b.header.typ != a.header.typ && b.header.typ != CONST_INT {
			return 0, errors.New("Fate Compare: Type Error")
		}
		x := int8(a.header.offset & achieveMask(a.header.length))
		y := int8(b.header.offset & achieveMask(a.header.length))
		switch {
		case x > y:
			return 1, nil
		case x == y:
			return 0, nil
		default:
			return -1, nil
		}
	case INT16:
		if b.header.typ != a.header.typ && b.header.typ != CONST_INT {
			return 0, errors.New("Fate Compare: Type Error")
		}
		x := int16(a.header.offset & achieveMask(a.header.length))
		y := int16(b.header.offset & achieveMask(a.header.length))
		switch {
		case x > y:
			return 1, nil
		case x == y:
			return 0, nil
		default:
			return -1, nil
		}
	case INT32:
		if b.header.typ != a.header.typ && b.header.typ != CONST_INT {
			return 0, errors.New("Fate Compare: Type Error")
		}
		x := int32(a.header.offset & achieveMask(a.header.length))
		y := int32(b.header.offset & achieveMask(a.header.length))
		switch {
		case x > y:
			return 1, nil
		case x == y:
			return 0, nil
		default:
			return -1, nil
		}
	case INT64:
		if b.header.typ != a.header.typ && b.header.typ != CONST_INT {
			return 0, errors.New("Fate Compare: Type Error")
		}
		x := int64(a.header.offset & achieveMask(a.header.length))
		y := int64(b.header.offset & achieveMask(a.header.length))
		switch {
		case x > y:
			return 1, nil
		case x == y:
			return 0, nil
		default:
			return -1, nil
		}
	case UINT8:
		if b.header.typ != a.header.typ && b.header.typ != CONST_INT {
			return 0, errors.New("Fate Compare: Type Error")
		}
		x := uint8(a.header.offset & achieveMask(a.header.length))
		y := uint8(b.header.offset & achieveMask(a.header.length))
		switch {
		case x > y:
			return 1, nil
		case x == y:
			return 0, nil
		default:
			return -1, nil
		}
	case UINT16:
		if b.header.typ != a.header.typ && b.header.typ != CONST_INT {
			return 0, errors.New("Fate Compare: Type Error")
		}
		x := uint16(a.header.offset & achieveMask(a.header.length))
		y := uint16(b.header.offset & achieveMask(a.header.length))
		switch {
		case x > y:
			return 1, nil
		case x == y:
			return 0, nil
		default:
			return -1, nil
		}
	case UINT32:
		if b.header.typ != a.header.typ && b.header.typ != CONST_INT {
			return 0, errors.New("Fate Compare: Type Error")
		}
		x := uint32(a.header.offset & achieveMask(a.header.length))
		y := uint32(b.header.offset & achieveMask(a.header.length))
		switch {
		case x > y:
			return 1, nil
		case x == y:
			return 0, nil
		default:
			return -1, nil
		}
	case UINT64:
		if b.header.typ != a.header.typ && b.header.typ != CONST_INT {
			return 0, errors.New("Fate Compare: Type Error")
		}
		x := uint64(a.header.offset & achieveMask(a.header.length))
		y := uint64(b.header.offset & achieveMask(a.header.length))
		switch {
		case x > y:
			return 1, nil
		case x == y:
			return 0, nil
		default:
			return -1, nil
		}
	case CONST_INT:
		switch b.header.typ {
		case INT8:
			x := int8(a.header.offset & achieveMask(b.header.length))
			y := int8(b.header.offset & achieveMask(b.header.length))
			switch {
			case x > y:
				return 1, nil
			case x == y:
				return 0, nil
			default:
				return -1, nil
			}
		case INT16:
			x := int16(a.header.offset & achieveMask(b.header.length))
			y := int16(b.header.offset & achieveMask(b.header.length))
			switch {
			case x > y:
				return 1, nil
			case x == y:
				return 0, nil
			default:
				return -1, nil
			}
		case INT32:
			x := int32(a.header.offset & achieveMask(b.header.length))
			y := int32(b.header.offset & achieveMask(b.header.length))
			switch {
			case x > y:
				return 1, nil
			case x == y:
				return 0, nil
			default:
				return -1, nil
			}
		case INT64:
			x := int64(a.header.offset & achieveMask(b.header.length))
			y := int64(b.header.offset & achieveMask(b.header.length))
			switch {
			case x > y:
				return 1, nil
			case x == y:
				return 0, nil
			default:
				return -1, nil
			}
		case UINT8:
			x := uint8(a.header.offset & achieveMask(b.header.length))
			y := uint8(b.header.offset & achieveMask(b.header.length))
			switch {
			case x > y:
				return 1, nil
			case x == y:
				return 0, nil
			default:
				return -1, nil
			}
		case UINT16:
			x := uint16(a.header.offset & achieveMask(b.header.length))
			y := uint16(b.header.offset & achieveMask(b.header.length))
			switch {
			case x > y:
				return 1, nil
			case x == y:
				return 0, nil
			default:
				return -1, nil
			}
		case UINT32:
			x := uint32(a.header.offset & achieveMask(b.header.length))
			y := uint32(b.header.offset & achieveMask(b.header.length))
			switch {
			case x > y:
				return 1, nil
			case x == y:
				return 0, nil
			default:
				return -1, nil
			}
		case UINT64:
			x := uint64(a.header.offset & achieveMask(b.header.length))
			y := uint64(b.header.offset & achieveMask(b.header.length))
			switch {
			case x > y:
				return 1, nil
			case x == y:
				return 0, nil
			default:
				return -1, nil
			}
		default:
			return 0, errors.New("Fate Compare: Type Error")
		}
	default:
		return 0, errors.New("Fate Compare: Unsupport type")
	}
}

// a = b
func (e *contractEngine) moveInit(a, b *contractData) error {
	switch a.header.typ {
	case MAP:
		return e.mapCopy(a, b)
	case STRING, CONST_STRING:
		return e.stringCopy(a, b)
	case BOOL, CONST_BOOL, CHAR, CONST_CHAR, FLOAT32, FLOAT64, CONST_FLOAT,
		INT8, INT16, INT32, INT64, UINT8, UINT16, UINT32, UINT64, CONST_INT:
		if b.header.typ != a.header.typ {
			return errors.New("Fate Move: Type Error")
		}
	default:
		return errors.New("Fate Move: Unsupport type")
	}
	a.header.offset = b.header.offset & achieveMask(b.header.length)
	return e.setData(a)
}

func (e *contractEngine) move(a, b *contractData) error {
	switch a.header.typ {
	case MAP:
		return e.mapCopy(a, b)
	case STRING:
		return e.stringCopy(a, b)
	case BOOL:
		if b.header.typ != a.header.typ && b.header.typ != CONST_BOOL {
			return errors.New("Fate Move: Type Error")
		}
	case CHAR:
		if b.header.typ != a.header.typ && b.header.typ != CONST_CHAR {
			return errors.New("Fate Move: Type Error")
		}
	case FLOAT32, FLOAT64:
		if b.header.typ != a.header.typ && b.header.typ != CONST_FLOAT {
			return errors.New("Fate Move: Type Error")
		}
	case INT8, INT16, INT32, INT64, UINT8, UINT16, UINT32, UINT64:
		if b.header.typ != a.header.typ && b.header.typ != CONST_INT {
			return errors.New("Fate Move: Type Error")
		}
	default:
		return errors.New("Fate Move: Unsupport type")
	}
	a.header.offset = b.header.offset & achieveMask(b.header.length)
	return e.setData(a)
}
