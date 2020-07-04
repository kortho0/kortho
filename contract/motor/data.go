package motor

import (
	"errors"
	"fmt"
	"math/big"

	"kortho/contract/mixed"
	"kortho/contract/virtul"
)

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
		switch a.header.offset0 {
		case 1: // true
			return fmt.Sprintf("bool: true"), nil
		default:
			return fmt.Sprintf("bool: false"), nil
		}
	case CHAR:
		return fmt.Sprintf("char: %c", byte(a.header.offset0&0xFF)), nil
	case INT8:
		v := recent(big.Int)
		if _, ok := v.SetString(string(effByte(append(mixed.E64func(a.header.offset0),
			append(mixed.E64func(a.header.offset1),
				mixed.E64func(a.header.offset2)...)...))), 0); !ok {
			return "", errors.New("Fate Visual Data: Illegal Value")
		}
		return fmt.Sprintf("int8: %s", v.String()), nil
	case UINT8:
		v := recent(big.Int)
		if _, ok := v.SetString(string(effByte(append(mixed.E64func(a.header.offset0),
			append(mixed.E64func(a.header.offset1),
				mixed.E64func(a.header.offset2)...)...))), 0); !ok {
			return "", errors.New("Fate Visual Data: Illegal Value")
		}
		return fmt.Sprintf("uint8: %s", v.String()), nil
	case INT16:
		v := recent(big.Int)
		if _, ok := v.SetString(string(effByte(append(mixed.E64func(a.header.offset0),
			append(mixed.E64func(a.header.offset1),
				mixed.E64func(a.header.offset2)...)...))), 0); !ok {
			return "", errors.New("Fate Visual Data: Illegal Value")
		}
		return fmt.Sprintf("int16: %s", v.String()), nil
	case UINT16:
		v := recent(big.Int)
		if _, ok := v.SetString(string(effByte(append(mixed.E64func(a.header.offset0),
			append(mixed.E64func(a.header.offset1),
				mixed.E64func(a.header.offset2)...)...))), 0); !ok {
			return "", errors.New("Fate Visual Data: Illegal Value")
		}
		return fmt.Sprintf("uint16: %s", v.String()), nil
	case INT32:
		v := recent(big.Int)
		if _, ok := v.SetString(string(effByte(append(mixed.E64func(a.header.offset0),
			append(mixed.E64func(a.header.offset1),
				mixed.E64func(a.header.offset2)...)...))), 0); !ok {
			return "", errors.New("Fate Visual Data: Illegal Value")
		}
		return fmt.Sprintf("int32: %s", v.String()), nil
	case UINT32:
		v := recent(big.Int)
		if _, ok := v.SetString(string(effByte(append(mixed.E64func(a.header.offset0),
			append(mixed.E64func(a.header.offset1),
				mixed.E64func(a.header.offset2)...)...))), 0); !ok {
			return "", errors.New("Fate Visual Data: Illegal Value")
		}
		return fmt.Sprintf("uint32: %s", v.String()), nil
	case INT64:
		v := recent(big.Int)
		if _, ok := v.SetString(string(effByte(append(mixed.E64func(a.header.offset0),
			append(mixed.E64func(a.header.offset1),
				mixed.E64func(a.header.offset2)...)...))), 0); !ok {
			return "", errors.New("Fate Visual Data: Illegal Value")
		}
		return fmt.Sprintf("int64: %s", v.String()), nil
	case UINT64:
		v := recent(big.Int)
		if _, ok := v.SetString(string(effByte(append(mixed.E64func(a.header.offset0),
			append(mixed.E64func(a.header.offset1),
				mixed.E64func(a.header.offset2)...)...))), 0); !ok {
			return "", errors.New("Fate Visual Data: Illegal Value")
		}
		return fmt.Sprintf("uint64: %s", v.String()), nil
	case FLOAT32:
		v := recent(big.Float)
		if _, ok := v.SetString(string(effByte(append(mixed.E64func(a.header.offset0),
			append(mixed.E64func(a.header.offset1),
				mixed.E64func(a.header.offset2)...)...)))); !ok {
			return "", errors.New("Fate Visual Data: Illegal Value")
		}
		return fmt.Sprintf("float32: %s", v.String()), nil
	case FLOAT64:
		v := recent(big.Float)
		if _, ok := v.SetString(string(effByte(append(mixed.E64func(a.header.offset0),
			append(mixed.E64func(a.header.offset1),
				mixed.E64func(a.header.offset2)...)...)))); !ok {
			return "", errors.New("Fate Visual Data: Illegal Value")
		}
		return fmt.Sprintf("float64: %s", v.String()), nil
	case STRING:
		if s, err := e.achieveString(a); err == nil {
			return fmt.Sprintf("string: %s", s), nil
		} else {
			return "", err
		}
	case CONST_INT:
		v := recent(big.Int)
		if _, ok := v.SetString(string(effByte(append(mixed.E64func(a.header.offset0),
			append(mixed.E64func(a.header.offset1),
				mixed.E64func(a.header.offset2)...)...))), 0); !ok {
			return "", errors.New("Fate Visual Data: Illegal Value")
		}
		return fmt.Sprintf("const int: %s", v.String()), nil
	case CONST_BOOL:
		switch a.header.offset0 {
		case 1: // true
			return fmt.Sprintf("const bool: true"), nil
		default:
			return fmt.Sprintf("const bool: false"), nil
		}
	case CONST_CHAR:
		return fmt.Sprintf("const char: %c", byte(a.header.offset0&0xFF)), nil
	case CONST_FLOAT:
		v := recent(big.Float)
		if _, ok := v.SetString(string(effByte(append(mixed.E64func(a.header.offset0),
			append(mixed.E64func(a.header.offset1),
				mixed.E64func(a.header.offset2)...)...)))); !ok {
			return "", errors.New("Fate Visual Data: Illegal Value")
		}
		return fmt.Sprintf("const float: %s", v.String()), nil
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
	switch a.header.typ {
	case MAP:
		m, err := e.achieveMap(a)
		if err != nil {
			return nil, fmt.Errorf("Fate Get Binary Data: %v", err)
		}
		return m, nil
	case CHAR, CONST_CHAR, BOOL, CONST_BOOL:
		return mixed.E64func(a.header.offset0), nil
	case STRING, CONST_STRING:
		s, err := e.achieveString(a)
		if err != nil {
			return nil, fmt.Errorf("Fate Get Binary Data: %v", err)
		}
		return s, nil
	case FLOAT32, FLOAT64, CONST_FLOAT, CONST_INT,
		INT8, UINT8, INT16, UINT16, INT32, UINT32, INT64, UINT64:
		return effByte(append(mixed.E64func(a.header.offset0),
			append(mixed.E64func(a.header.offset1),
				mixed.E64func(a.header.offset2)...)...)), nil
	default:
		return nil, errors.New("Fate Get Binary Data: Unsupport type")
	}
}

func (e *contractEngine) sizeof(a *contractData) uint64 {
	if a.header.typ != STRING && a.header.typ != CONST_STRING {
		return a.header.length
	}
	return strlen(a.header.length)
}

func (e *contractEngine) recentVar(typ int, address uint64) (*contractData, error) {
	switch typ {
	case BOOL, CONST_BOOL, CHAR, CONST_CHAR:
		a := &contractData{
			address: address,
			header: &contractDataHeader{
				typ:    uint32(typ),
				length: uint64(ltb[typ]),
			},
		}
		e.setData(a)
		return a, nil
	case INT8, INT16, INT32, INT64, UINT8, UINT16, UINT32, UINT64,
		FLOAT32, FLOAT64, CONST_INT, CONST_FLOAT:
		a := &contractData{
			address: address,
			header: &contractDataHeader{
				typ:     uint32(typ),
				length:  uint64(ltb[typ]),
				offset2: 3458764513820540928,
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
				offset0: addr,
				typ:     uint32(typ),
				length:  uint64(ltb[typ]) | uint64(DEFAULT_STRING_LENGTH)<<32,
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
			bka, err := e.achieveBucket(a)
			if err != nil {
				return -1, err
			}
			bkb, err := e.achieveBucket(b)
			if err != nil {
				return -1, err
			}
			if bka.header.ktyp != bkb.header.ktyp {
				return -1, errors.New("Fate Type Cmp: Key Type Not Match")
			}
			if bka.header.vtyp != bkb.header.vtyp {
				return -1, errors.New("Fate Type Cmp: Value Type Not Match")
			}
		}
		/* 暂时不考虑map嵌套的情况
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
		*/
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
		x := a.header.offset0
		y := b.header.offset0
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
		x := a.header.offset0
		y := b.header.offset0
		switch {
		case x > y:
			return 1, nil
		case x == y:
			return 0, nil
		default:
			return -1, nil
		}
	case CONST_FLOAT:
		if b.header.typ != FLOAT32 && b.header.typ != FLOAT64 && b.header.typ != CONST_FLOAT {
			return 0, errors.New("Fate Compare: Type Error")
		}
		x := recent(big.Float)
		y := recent(big.Float)
		if _, ok := x.SetString(string(effByte(append(mixed.E64func(a.header.offset0),
			append(mixed.E64func(a.header.offset1),
				mixed.E64func(a.header.offset2)...)...)))); !ok {
			return 0, fmt.Errorf("Fate Compare: Illegal Value")
		}
		if _, ok := y.SetString(string(effByte(append(mixed.E64func(b.header.offset0),
			append(mixed.E64func(b.header.offset1),
				mixed.E64func(b.header.offset2)...)...)))); !ok {
			return 0, fmt.Errorf("Fate Compare: Illegal Value")
		}
		return x.Cmp(y), nil
	case FLOAT32, FLOAT64:
		if b.header.typ != a.header.typ && b.header.typ != CONST_FLOAT {
			return 0, errors.New("Fate Compare: Type Error")
		}
		x := recent(big.Float)
		y := recent(big.Float)
		if _, ok := x.SetString(string(effByte(append(mixed.E64func(a.header.offset0),
			append(mixed.E64func(a.header.offset1),
				mixed.E64func(a.header.offset2)...)...)))); !ok {
			return 0, fmt.Errorf("Fate Compare: Illegal Value")
		}
		if _, ok := y.SetString(string(effByte(append(mixed.E64func(b.header.offset0),
			append(mixed.E64func(b.header.offset1),
				mixed.E64func(b.header.offset2)...)...)))); !ok {
			return 0, fmt.Errorf("Fate Compare: Illegal Value")
		}
		return x.Cmp(y), nil
	case CONST_INT:
		if typ := b.header.typ; typ != a.header.typ && typ != INT8 && typ != INT16 &&
			typ != INT32 && typ != INT64 && typ != UINT8 && typ != UINT16 &&
			typ != UINT32 && typ != UINT64 {
			return 0, errors.New("Fate Compare: Type Error")
		}
		x := recent(big.Int)
		y := recent(big.Int)
		if _, ok := x.SetString(string(effByte(append(mixed.E64func(a.header.offset0),
			append(mixed.E64func(a.header.offset1),
				mixed.E64func(a.header.offset2)...)...))), 0); !ok {
			return 0, fmt.Errorf("Fate Compare: Illegal Value")
		}
		if _, ok := y.SetString(string(effByte(append(mixed.E64func(b.header.offset0),
			append(mixed.E64func(b.header.offset1),
				mixed.E64func(b.header.offset2)...)...))), 0); !ok {
			return 0, fmt.Errorf("Fate Compare: Illegal Value")
		}
		return x.Cmp(y), nil
	case INT8, UINT8, INT16, UINT16, INT32, UINT32, INT64, UINT64:
		if b.header.typ != a.header.typ && b.header.typ != CONST_INT {
			return 0, errors.New("Fate Compare: Type Error")
		}
		x := recent(big.Int)
		y := recent(big.Int)
		if _, ok := x.SetString(string(effByte(append(mixed.E64func(a.header.offset0),
			append(mixed.E64func(a.header.offset1),
				mixed.E64func(a.header.offset2)...)...))), 0); !ok {
			return 0, fmt.Errorf("Fate Compare: Illegal Value")
		}
		if _, ok := y.SetString(string(effByte(append(mixed.E64func(b.header.offset0),
			append(mixed.E64func(b.header.offset1),
				mixed.E64func(b.header.offset2)...)...))), 0); !ok {
			return 0, fmt.Errorf("Fate Compare: Illegal Value")
		}
		return x.Cmp(y), nil
	default:
		return 0, errors.New("Fate Compare: Unsupport type")
	}
}

// a = b
func (e *contractEngine) move(a, b *contractData) error {
	switch a.header.typ {
	case MAP:
		return e.mapCopy(a, b)
	case STRING, CONST_STRING:
		return e.stringCopy(a, b)
	case BOOL:
		if b.header.typ != a.header.typ && b.header.typ != CONST_BOOL {
			return errors.New("Fate Move: Type Error")
		}
	case CHAR:
		if b.header.typ != a.header.typ && b.header.typ != CONST_CHAR {
			return errors.New("Fate Move: Type Error")
		}
	case CONST_BOOL:
		if b.header.typ != a.header.typ && b.header.typ != BOOL {
			return errors.New("Fate Move: Type Error")
		}
	case CONST_CHAR:
		if b.header.typ != a.header.typ && b.header.typ != CHAR {
			return errors.New("Fate Move: Type Error")
		}
	case CONST_FLOAT:
		if b.header.typ != a.header.typ && b.header.typ != FLOAT32 && b.header.typ != FLOAT64 {
			return errors.New("Fate Move: Type Error")
		}
		v := recent(big.Float)
		if _, ok := v.SetString(string(effByte(append(mixed.E64func(b.header.offset0),
			append(mixed.E64func(b.header.offset1),
				mixed.E64func(b.header.offset2)...)...)))); !ok {
			return fmt.Errorf("Fate Move: Illegal Value")
		}
		if v.Cmp(MinFloat32) < 0 || v.Cmp(MaxFloat64) > 0 {
			return errors.New("Fate MOVE: Overflows")
		}
	case FLOAT32, FLOAT64:
		if b.header.typ != a.header.typ && b.header.typ != CONST_FLOAT {
			return errors.New("Fate Move: Type Error")
		}
		v := recent(big.Float)
		if _, ok := v.SetString(string(effByte(append(mixed.E64func(b.header.offset0),
			append(mixed.E64func(b.header.offset1),
				mixed.E64func(b.header.offset2)...)...)))); !ok {
			return fmt.Errorf("Fate Move: Illegal Value")
		}
		switch a.header.length {
		case 4:
			if v.Cmp(MinFloat32) < 0 || v.Cmp(MaxFloat32) > 0 {
				return errors.New("Fate MOVE: Overflows")
			}
		case 8:
			if v.Cmp(MinFloat64) < 0 || v.Cmp(MaxFloat64) > 0 {
				return errors.New("Fate MOVE: Overflows")
			}
		}
	case CONST_INT:
		if typ := b.header.typ; typ != a.header.typ && typ != INT8 && typ != INT16 && typ != INT32 &&
			typ != INT64 && typ != UINT8 && typ != UINT16 && typ != UINT32 && typ != UINT64 {
			return errors.New("Fate Move: Type Error")
		}
		v := recent(big.Int)
		if _, ok := v.SetString(string(effByte(append(mixed.E64func(b.header.offset0),
			append(mixed.E64func(b.header.offset1),
				mixed.E64func(b.header.offset2)...)...))), 0); !ok {
			return fmt.Errorf("Fate Move: Illegal Value")
		}
		switch {
		case v.IsInt64():
			if v.Cmp(MinInt64) < 0 || v.Cmp(MaxInt64) > 0 {
				return errors.New("Fate MOVE: Overflows")
			}
		default:
			if v.Cmp(MaxUint64) > 0 {
				return errors.New("Fate MOVE: Overflows")
			}
		}
	case INT8, INT16, INT32, INT64, UINT8, UINT16, UINT32, UINT64:
		if b.header.typ != a.header.typ && b.header.typ != CONST_INT {
			return errors.New("Fate Move: Type Error")
		}
		v := recent(big.Int)
		if _, ok := v.SetString(string(effByte(append(mixed.E64func(b.header.offset0),
			append(mixed.E64func(b.header.offset1),
				mixed.E64func(b.header.offset2)...)...))), 0); !ok {
			return fmt.Errorf("Fate Move: Illegal Value")
		}
		switch a.header.typ {
		case INT8, INT16, INT32, INT64:
			switch a.header.length {
			case 1:
				if v.Cmp(MinInt8) < 0 || v.Cmp(MaxInt8) > 0 {
					return errors.New("Fate MOVE: Overflows")
				}
			case 2:
				if v.Cmp(MinInt16) < 0 || v.Cmp(MaxInt16) > 0 {
					return errors.New("Fate MOVE: Overflows")
				}
			case 4:
				if v.Cmp(MinInt32) < 0 || v.Cmp(MaxInt32) > 0 {
					return errors.New("Fate MOVE: Overflows")
				}
			case 8:
				if v.Cmp(MinInt64) < 0 || v.Cmp(MaxInt64) > 0 {
					return errors.New("Fate MOVE: Overflows")
				}
			}
		default:
			if v.IsInt64() && v.Sign() < 0 {
				return errors.New("Fate MOVE: Overflows")
			}
			switch a.header.length {
			case 1:
				if v.Cmp(MaxUint8) > 0 {
					return errors.New("Fate MOVE: Overflows")
				}
			case 2:
				if v.Cmp(MaxUint16) > 0 {
					return errors.New("Fate MOVE: Overflows")
				}
			case 4:
				if v.Cmp(MaxUint32) > 0 {
					return errors.New("Fate MOVE: Overflows")
				}
			case 8:
				if v.Cmp(MaxUint64) > 0 {
					return errors.New("Fate MOVE: Overflows")
				}
			}
		}
	default:
		return errors.New("Fate Move: Unsupport type")
	}
	a.header.offset0 = b.header.offset0
	a.header.offset1 = b.header.offset1
	a.header.offset2 = b.header.offset2
	return e.setData(a)
}
