package motor

import (
	"errors"
	"fmt"

	"kortho/contract/mixed"
	"kortho/contract/virtul"
)

func hash(a []byte) int {
	h := 0
	for _, v := range a {
		h += hlt[int(v)]
	}
	return h
}

func (e *contractEngine) achieveMap(a *contractData) ([]byte, error) {
	if data, err := e.virtul.GetRawData(a.header.offset0, BUCKET_HEADER_SIZE); err == nil {
		return data, nil
	} else {
		return nil, fmt.Errorf("Fate Get Map: %v", err)
	}
}

func (e *contractEngine) recentMap(ktyp, vtyp int, address uint64) (*contractData, error) {
	addr, err := e.virtul.Alloc(virtul.StoreType(address), DEFAULT_BUCKET_LENGTH)
	if err != nil {
		return nil, fmt.Errorf("Fate New Map: out of space %v", err)
	}
	bk := &contractBucket{
		slots: make([]uint64, phts[0]),
		header: &contractBucketHeader{
			ktyp: uint32(ktyp),
			vtyp: uint32(vtyp),
			size: uint32(phts[0]),
		},
	}
	a := &contractData{
		address: address,
		header: &contractDataHeader{
			typ:     MAP,
			offset0: addr,
			length:  uint64(ltb[MAP]),
		},
	}
	e.setBucket(a, bk)
	e.setData(a)
	return a, nil
}

func (e *contractEngine) achieveSlot(off uint64) (*contractSlot, error) {
	if off == 0 {
		return nil, errors.New("Fate Get Slot: Zero Offset")
	}
	s := contractSlot{}
	if sData, err := e.virtul.GetRawData(off, SLOT_SIZE); err != nil {
		return nil, fmt.Errorf("Fate Get Slot: %v", err)
	} else if _, err = s.Read(sData); err != nil {
		return nil, fmt.Errorf("Fate Get Slot: %v", err)
	}
	return &s, nil
}

func (e *contractEngine) setSlot(off uint64, s *contractSlot) error {
	sData, _ := s.Show()
	return e.virtul.SetRawData(off, SLOT_SIZE, sData)
}

func (e *contractEngine) achieveBucket(a *contractData) (*contractBucket, error) {
	if sData, err := e.virtul.GetRawData(a.header.offset0, 4); err == nil {
		bk := contractBucket{}
		size, _ := mixed.D32func(sData)
		if data, err := e.virtul.GetRawData(a.header.offset0, uint64(BUCKET_HEADER_SIZE+size*8)); err != nil {
			return nil, fmt.Errorf("Fate Get Bucket: %v", err)
		} else if _, err = bk.Read(data); err != nil {
			return nil, fmt.Errorf("Fate Get Bucket: %v", err)
		}
		return &bk, nil
	} else {
		return nil, fmt.Errorf("Fate Get Bucket: %v", err)
	}
}

func (e *contractEngine) setBucket(a *contractData, bk *contractBucket) error {
	data, _ := bk.Show()
	return e.virtul.SetRawData(a.header.offset0, uint64(BUCKET_HEADER_SIZE+bk.header.size*8), data)
}

func (e *contractEngine) find(a, k *contractData) (*contractData, error) {
	bk, err := e.achieveBucket(a)
	if err != nil {
		return nil, fmt.Errorf("Fate Map Find: %v", err)
	}
	kData, err := e.achieveBinaryData(k)
	if err != nil {
		return nil, fmt.Errorf("Fate Map Find: %v", err)
	}
	off := bk.slots[hash(kData)%int(bk.header.size)]
	for {
		if off == 0 {
			break
		}
		if c, err := e.achieveSlot(off); err == nil {
			if key, err := e.achieveData(c.koff); err == nil {
				r, err := e.cmp(k, key)
				if err != nil {
					return nil, fmt.Errorf("Fate Map Find: %v", err)
				}
				if r == 0 {
					return e.achieveData(c.voff)
				}
				off = c.next
			} else {
				return nil, fmt.Errorf("Fate Map Find: %v", err)
			}
		} else {
			return nil, fmt.Errorf("Fate Map Find: %v", err)
		}
	}
	return nil, errors.New("Fate Map Find: Cannot Find")
}

func (e *contractEngine) findByIndex(a *contractData, index uint64) (*contractData, *contractData, error) {
	if index >= a.header.length {
		return nil, nil, errors.New("Fate Map Find By Index: Cannot Find")
	}
	bk, err := e.achieveBucket(a)
	if err != nil {
		return nil, nil, fmt.Errorf("Fate Map Find By Index: %v", err)
	}
	cnt := uint64(0)
	for i := uint32(0); i < bk.header.size; i++ {
		off := bk.slots[int(i)]
		for {
			if off == 0 {
				break
			}
			if c, err := e.achieveSlot(off); err == nil {
				if cnt == index {
					k, err := e.achieveData(c.koff)
					if err != nil {
						return nil, nil, fmt.Errorf("Fate Map Find By Index: %v", err)
					}
					v, err := e.achieveData(c.voff)
					if err != nil {
						return nil, nil, fmt.Errorf("Fate Map Find By Index: %v", err)
					}
					return k, v, nil
				} else {
					off = c.next
				}
			} else {
				return nil, nil, fmt.Errorf("Fate Map Find By Index: %v", err)
			}
			cnt++
		}
	}
	return nil, nil, errors.New("Fate Map Find By Index: Cannot Find")
}

func (e *contractEngine) reHash(a *contractData, size uint32) error {
	var ktyp, vtyp uint32

	if bk, err := e.achieveBucket(a); err == nil {
		if size <= bk.header.size {
			return nil
		}
		ktyp = bk.header.ktyp
		vtyp = bk.header.vtyp
	} else {
		return fmt.Errorf("Fate Map ReHash: %v", err)
	}
	addr, err := e.virtul.Alloc(virtul.StoreType(a.header.offset0), uint64(size*8+DEFAULT_BUCKET_LENGTH))
	if err != nil {
		return fmt.Errorf("Fate Map ReHash: %v", err)
	}
	b := &contractData{
		address: a.address,
		header: &contractDataHeader{
			typ:     MAP,
			offset0: addr,
			length:  uint64(ltb[MAP]),
		},
	}
	bk := &contractBucket{
		header: &contractBucketHeader{
			size: size,
			ktyp: ktyp,
			vtyp: vtyp,
		},
		slots: make([]uint64, size),
	}
	if err = e.setBucket(b, bk); err != nil {
		e.virtul.Free(addr)
		return fmt.Errorf("Fate Map ReHash: %v", err)
	}
	for i, j := uint64(0), a.header.length; i < j; i++ {
		if k, v, err := e.findByIndex(a, i); err == nil {
			if err = e.insert(b, k, v); err != nil {
				e.mapRemove(b)
				return fmt.Errorf("Fate Map ReHash: %v", err)
			}
		} else {
			e.mapRemove(b)
			return fmt.Errorf("Fate Map ReHash: %v", err)
		}
	}
	e.mapRemove(a)
	e.setData(b)
	a.header = b.header
	return nil
}

func (e *contractEngine) mapRemove(a *contractData) error {
	bk, err := e.achieveBucket(a)
	if err != nil {
		return fmt.Errorf("Fate Map Clear: %v", err)
	}
	for i := uint32(0); i < bk.header.size; i++ {
		off := bk.slots[int(i)]
		for {
			if off == 0 {
				break
			}
			if c, err := e.achieveSlot(off); err == nil {
				if k, err := e.achieveData(c.koff); err == nil {
					e.remove(k)
				}
				if v, err := e.achieveData(c.voff); err == nil {
					e.remove(v)
				}
				e.virtul.Free(off)
				off = c.next
			} else {
				return fmt.Errorf("Fate Map Clear: %v", err)
			}
		}
	}
	return e.virtul.Free(a.header.offset0)
}

func (e *contractEngine) insert(a, k, v *contractData) error {
	if d, err := e.find(a, k); err != nil {
		bk, err := e.achieveBucket(a)
		if err != nil {
			return fmt.Errorf("Fate Map Insert: %v", err)
		}
		switch bk.header.ktyp {
		case BOOL:
			if k.header.typ != bk.header.ktyp && k.header.typ != CONST_BOOL {
				return errors.New("Fate Map Insert: key type don't match")
			}
		case CHAR:
			if k.header.typ != bk.header.ktyp && k.header.typ != CONST_CHAR {
				return errors.New("Fate Map Insert: key type don't match")
			}
		case STRING:
			if k.header.typ != bk.header.ktyp && k.header.typ != CONST_STRING {
				return errors.New("Fate Map Insert: key type don't match")
			}
		case FLOAT32, FLOAT64:
			if k.header.typ != bk.header.ktyp && k.header.typ != CONST_FLOAT {
				return errors.New("Fate Map Insert: key type don't match")
			}
		case INT8, INT16, INT32, INT64, UINT8, UINT16, UINT32, UINT64:
			if k.header.typ != bk.header.ktyp && k.header.typ != CONST_INT {
				return errors.New("Fate Map Insert: key type don't match")
			}
		default:
			return errors.New("Fate Map Insert: key type don't match")
		}
		switch bk.header.vtyp {
		case BOOL:
			if v.header.typ != bk.header.vtyp && v.header.typ != CONST_BOOL {
				return errors.New("Fate Map Insert: key type don't match")
			}
		case CHAR:
			if v.header.typ != bk.header.vtyp && v.header.typ != CONST_CHAR {
				return errors.New("Fate Map Insert: key type don't match")
			}
		case STRING:
			if v.header.typ != bk.header.vtyp && v.header.typ != CONST_STRING {
				return errors.New("Fate Map Insert: key type don't match")
			}
		case FLOAT32, FLOAT64:
			if v.header.typ != bk.header.vtyp && v.header.typ != CONST_FLOAT {
				return errors.New("Fate Map Insert: key type don't match")
			}
		case INT8, INT16, INT32, INT64, UINT8, UINT16, UINT32, UINT64:
			if v.header.typ != bk.header.vtyp && v.header.typ != CONST_INT {
				return errors.New("Fate Map Insert: key type don't match")
			}
		default:
			return errors.New("Fate Map Insert: key type don't match")
		}
		if dty := a.header.length / uint64(bk.header.size*DEFAULT_CHAIN_LENGTH); dty > 1 {
			size := bk.header.size
			for i, j := 0, len(phts); i < j; i++ {
				if phts[i] == int(bk.header.size) && i+1 < j {
					size = uint32(phts[i+1])
				}
			}
			if size != bk.header.size {
				if err := e.reHash(a, size); err != nil {
					return fmt.Errorf("Fate Map Insert: %v", err)
				}
				if bk, err = e.achieveBucket(a); err != nil {
					return fmt.Errorf("Fate Map Insert: %v", err)
				}
			}
		}
		addr, err := e.virtul.Alloc(virtul.StoreType(a.header.offset0), SLOT_SIZE)
		if err != nil {
			return fmt.Errorf("Fate Map Insert: %v", err)
		}
		ktyp := k.header.typ
		k.header.typ = bk.header.ktyp
		nk, err := e.dup(k, virtul.StoreType(a.header.offset0))
		k.header.typ = ktyp
		if err != nil {
			e.virtul.Free(addr)
			return fmt.Errorf("Fate Map Insert: %v", err)
		}
		vtyp := v.header.typ
		v.header.typ = bk.header.vtyp
		nv, err := e.dup(v, virtul.StoreType(a.header.offset0))
		v.header.typ = vtyp
		if err != nil {
			e.remove(nk)
			e.virtul.Free(addr)
			return fmt.Errorf("Fate Map Insert: %v", err)
		}
		recentC := &contractSlot{
			next: 0,
			koff: nk.address,
			voff: nv.address,
		}
		kData, err := e.achieveBinaryData(k)
		if err != nil {
			e.remove(nk)
			e.remove(nv)
			e.virtul.Free(addr)
			return fmt.Errorf("Fate Map Insert: %v", err)
		}
		if off := bk.slots[hash(kData)%int(bk.header.size)]; off != 0 {
			recentC.next = off
		}
		bk.slots[hash(kData)%int(bk.header.size)] = addr
		e.setSlot(addr, recentC)
		e.setBucket(a, bk)
		a.header.length++
		e.setData(a)
		e.setData(nk)
		e.setData(nv)
		return nil
	} else {
		return e.move(d, v)
	}
}

func (e *contractEngine) rubout(a, k *contractData) error {
	bk, err := e.achieveBucket(a)
	if err != nil {
		return fmt.Errorf("Fate Map Delete: %v", err)
	}
	kData, err := e.achieveBinaryData(k)
	if err != nil {
		return fmt.Errorf("Fate Map Delete: %v", err)
	}
	prevC := &contractSlot{}
	prevOff := uint64(0)
	off := bk.slots[hash(kData)%int(bk.header.size)]
	for {
		if off == 0 {
			break
		}
		if c, err := e.achieveSlot(off); err == nil {
			key, err := e.achieveData(c.koff)
			if err != nil {
				return fmt.Errorf("Fate Map Delete: %v", err)
			}
			r, err := e.cmp(k, key)
			if err != nil {
				return fmt.Errorf("Fate Map Delete: %v", err)
			}
			if r == 0 {
				value, err := e.achieveData(c.voff)
				if err != nil {
					return fmt.Errorf("Fate Map Delete: %v", err)
				}
				e.remove(key)
				e.remove(value)
				e.virtul.Free(off)
				if prevOff != 0 {
					prevC.next = c.next
					e.setSlot(prevOff, prevC)
				} else {
					bk.slots[hash(kData)%int(bk.header.size)] = c.next
					e.setBucket(a, bk)
				}
				a.header.length = a.header.length - 1
				e.setData(a)
				return nil
			} else {
				prevC = c
				prevOff = off
				off = c.next
			}
		} else {
			return fmt.Errorf("Fate Map Delete: %v", err)
		}
	}
	return errors.New("Fate Map Delete: Cannot find")
}

func (e *contractEngine) mapCmp(a, b *contractData) (int, error) {
	if r := a.header.length - b.header.length; r != 0 {
		return int(r), nil
	}
	if a.header.length == 0 {
		return e.typeCmp(a, b)
	}
	for i, j := uint64(0), a.header.length; i < j; i++ {
		ka, va, err := e.findByIndex(a, i)
		if err != nil {
			return -1, fmt.Errorf("Fate Map Cmp: %v", err)
		}
		if vb, err := e.find(b, ka); err == nil {
			r, err := e.cmp(va, vb)
			if err != nil {
				return -1, fmt.Errorf("Fate Map Cmp: %v", err)
			}
			if r != 0 {
				return r, nil
			}
		} else {
			return 1, nil
		}
	}
	return 0, nil
}

// a = b
func (e *contractEngine) mapCopy(a, b *contractData) error {
	bk, err := e.achieveBucket(a)
	if err != nil {
		return fmt.Errorf("Fate Map Copy: %v", err)
	}
	ktyp := bk.header.ktyp
	vtyp := bk.header.vtyp
	if _, err = e.typeCmp(a, b); err != nil {
		return err
	}
	address, err := e.virtul.Alloc(virtul.StoreType(a.address), DATA_HEADER_SIZE)
	if err != nil {
		return fmt.Errorf("Fate Map Copy: %v", err)
	}
	c, err := e.recentMap(int(ktyp), int(vtyp), address)
	if err != nil {
		e.virtul.Free(address)
		return fmt.Errorf("Fate Map Copy: %v", err)
	}
	for i, j := uint64(0), b.header.length; i < j; i++ {
		k, v, err := e.findByIndex(b, i)
		if err != nil {
			e.remove(c)
			return fmt.Errorf("Fate Map Copy: %v", err)
		}
		if err = e.insert(c, k, v); err != nil {
			e.remove(c)
			return fmt.Errorf("Fate Map Copy: %v", err)
		}
	}
	e.mapRemove(a)
	e.virtul.Free(c.address)
	a.header.length = c.header.length
	a.header.offset0 = c.header.offset0
	e.setData(a)
	return nil
}
