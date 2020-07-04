package motor

import (
	"errors"
	"fmt"
)

func strlen(a uint64) uint64 {
	return a & 0xFFFFFFFF
}

func strcap(a uint64) uint64 {
	return (a >> 32) & 0xFFFFFFFF
}

func (e *contractEngine) recentString(address uint64, str []byte) (*contractData, error) {
	a, err := e.recentVar(STRING, address)
	if err != nil {
		return nil, fmt.Errorf("Fate New String: %v", err)
	}
	return a, e.setString(a, str)
}

func (e *contractEngine) cut(a *contractData, length uint64) error {
	if length > strlen(a.header.length) {
		return errors.New("Fate Index: String too Short")
	}
	a.header.length = length | strcap(a.header.length)<<32
	return e.setData(a)
}

func (e *contractEngine) index(a *contractData, index uint64) (*contractData, error) {
	if index >= strlen(a.header.length) {
		return nil, errors.New("Fate Index: Cross Broder")
	}
	h := contractDataHeader{}
	addr := a.header.offset0 + index*DATA_HEADER_SIZE
	if hData, err := e.virtul.GetRawData(addr, DATA_HEADER_SIZE); err != nil {
		return nil, fmt.Errorf("Fate Index: %v", err)
	} else if _, err = h.Read(hData); err != nil {
		return nil, fmt.Errorf("Fate Index: %v", err)
	}
	return &contractData{
		header:  &h,
		address: addr,
	}, nil
}

func (e *contractEngine) concat(a, b *contractData) error {
	lth := strlen(a.header.length)
	cap := strcap(a.header.length)
	if lth >= cap {
		if offset, err := e.virtul.Realloc(a.header.offset0, cap*DATA_HEADER_SIZE, (cap+DEFAULT_STRING_LENGTH)*DATA_HEADER_SIZE); err == nil {
			if err = e.virtul.SetRawData(offset+lth*DATA_HEADER_SIZE, DEFAULT_STRING_LENGTH*DATA_HEADER_SIZE, dcl); err != nil {
				return fmt.Errorf("Fate Append: %v", err)
			}
			a.header.offset0 = offset
			cap = cap + DEFAULT_STRING_LENGTH
			if err = e.setData(a); err != nil {
				return fmt.Errorf("Fate Append: %v", err)
			}
		} else {
			return fmt.Errorf("Fate Append: %v", err)
		}
	}
	a.header.length = (lth + 1) | cap<<32
	if c, err := e.index(a, lth); err != nil {
		return fmt.Errorf("Fate Append: %v", err)
	} else if err = e.move(c, b); err != nil {
		return fmt.Errorf("Fate Append: %v", err)
	}
	return e.setData(a)
}

func (e *contractEngine) achieveString(a *contractData) ([]byte, error) {
	s := []byte{}
	for i, j := uint64(0), strlen(a.header.length); i < j; i++ {
		if c, err := e.index(a, i); err != nil {
			return nil, fmt.Errorf("Fate Get String: %v", err)
		} else {
			s = append(s, byte(c.header.offset0&0xFF))
		}
	}
	return s, nil
}

// a = s
func (e *contractEngine) setString(a *contractData, s []byte) error {
	a.header.length = 0 | strcap(a.header.length)<<32
	for _, v := range s {
		if err := e.concat(a, &contractData{
			address: 0,
			header: &contractDataHeader{
				typ:     CHAR,
				offset0: uint64(v),
				length:  uint64(ltb[CHAR]),
			},
		}); err != nil {
			return fmt.Errorf("Fate Set String: %v", err)
		}
	}
	return e.setData(a)
}

func (e *contractEngine) stringCmp(a, b *contractData) (int, error) {
	la := strlen(a.header.length)
	lb := strlen(b.header.length)
	if r := la - lb; r != 0 {
		return int(r), nil
	}
	for i := uint64(0); i < la; i++ {
		c0, err := e.index(a, i)
		if err != nil {
			return 0, fmt.Errorf("Fate String Compare: %v", err)
		}
		c1, err := e.index(b, i)
		if err != nil {
			return 0, fmt.Errorf("Fate String Compare: %v", err)
		}
		if r, _ := e.cmp(c0, c1); r != 0 {
			return r, nil
		}
	}
	return 0, nil
}

func (e *contractEngine) stringRemove(a *contractData) error {
	return e.virtul.Free(a.header.offset0)
}

func (e *contractEngine) stringCopy(a, b *contractData) error {
	i := uint64(0)
	j := strlen(a.header.length)
	k := strlen(b.header.length)
	if j > k {
		j = k
		if err := e.cut(a, j); err != nil {
			return fmt.Errorf("Fate String Copy: %v", err)
		}
	}
	for i < j {
		src, err := e.index(b, i)
		if err != nil {
			return fmt.Errorf("Fate String Copy: %v", err)
		}
		dst, err := e.index(a, i)
		if err != nil {
			return fmt.Errorf("Fate String Copy: %v", err)
		}
		if err := e.move(dst, src); err != nil {
			return fmt.Errorf("Fate String Copy: %v", err)
		}
		i++
	}
	for i < k {
		c, err := e.index(b, i)
		if err != nil {
			return fmt.Errorf("Fate String Copy: %v", err)
		}
		if err := e.concat(a, c); err != nil {
			return fmt.Errorf("Fate String Copy: %v", err)
		}
		i++
	}
	return nil
}
