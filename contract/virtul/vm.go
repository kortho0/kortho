package virtul

import (
	"errors"
	"fmt"
	"sync"

	"kortho/contract/virtul/db"
)

func pde(a uint64) uint32 {
	return uint32((a >> PDE_OFF) & PDE_MASK)
}

func recentPM(reserved uint64, db db.DB) *contractPM {
	pages := []*contractPage{}
	for i := 0; i < PAGE_COUNT; i++ {
		pages = append(pages, &contractPage{
			flag: 0,
			page: []byte{},
		})
	}
	return &contractPM{
		db:       db,
		pages:    pages,
		reserved: reserved,
	}
}

func (pm *contractPM) achievePage(pn uint32) ([]byte, error) {
	if pm.pages[int(pn)].flag&P == 0 {
		if pn >= RAM_PAGE_COUNT {
			if pg, _ := pm.db.Get([]byte(fmt.Sprint(pn))); len(pg) != 0 {
				pm.pages[int(pn)].page = pg
			} else {
				pm.pages[int(pn)].page = make([]byte, PAGE_SIZE)
			}
		} else {
			pm.pages[int(pn)].page = make([]byte, PAGE_SIZE)
		}
		pm.pc++
		pm.pages[int(pn)].flag = P
	}
	pm.pages[int(pn)].flag |= A
	return pm.pages[int(pn)].page, nil
}

func (pm *contractPM) setPage(pn uint32, page []byte) error {
	if pm.pages[int(pn)].flag&P == 0 {
		pm.pc++
		pm.pages[int(pn)].flag = P
	}
	pm.pages[int(pn)].flag |= D
	pm.pages[int(pn)].page = page
	return nil
}

func StoreType(a uint64) int {
	if a < RAM_LIMIT {
		return RAM
	}
	if a > FLASH_LIMIT {
		return -1
	}
	return FLASH
}

func New(reserved []byte, name string) (*contractVM, error) {
	db, err := db.New(name)
	if err != nil {
		return nil, err
	}
	buf := []byte{}
	buf = append(buf, reserved...)
	reservedLen := uint64(len(reserved))
	h1 := &contractMemHeader{
		size: 0,
		next: reservedLen + MEM_HEADER_SIZE,
	}
	h2 := &contractMemHeader{
		next: 0,
		size: 4*GB - reservedLen - 2*MEM_HEADER_SIZE,
	}
	h1Data, _ := h1.Show()
	buf = append(buf, h1Data...)
	h2Data, _ := h2.Show()
	buf = append(buf, h2Data...)
	pm := recentPM(reservedLen, db)
	for i := 0; len(buf) > 0; i++ {
		if j := len(buf); j < int(PAGE_SIZE) {
			pm.setPage(uint32(i), append(buf, make([]byte, int(PAGE_SIZE-uint64(j)))...))
			break
		}
		pm.setPage(uint32(i), buf[:PAGE_SIZE])
		buf = buf[PAGE_SIZE:]
	}
	return &contractVM{
		pm: pm,
	}, nil
}

func (pm *contractPM) flush() error {
	var wg sync.WaitGroup
	for i := 1; i < PAGE_COUNT; i++ {
		if pm.pages[i].flag&D != 0 {
			wg.Add(1)
			go func(pn int) {
				if pn >= RAM_PAGE_COUNT {
					pm.db.Set([]byte(fmt.Sprint(pn)), pm.pages[pn].page)
				}
				pm.pc--
				pm.pages[pn].flag = 0
				pm.pages[pn].page = []byte{}
				wg.Done()
			}(i)
		}
	}
	wg.Wait()
	return nil
}

func (pm *contractPM) achieveRawData(address, length uint64) ([]byte, error) {
	buf := []byte{}
	for i, j, k := uint32(0), pde(((address&PAGE_MASK + length + PAGE_MASK) & ^uint64(PAGE_MASK))), pde(address); i < j; i++ {
		pg, _ := pm.achievePage(i + k)
		q, r := uint64(0), PAGE_SIZE
		switch i {
		case 0:
			q = address & PAGE_MASK
			if i == j-1 {
				r = q + length
			}
		case j - 1:
			r = length
		}
		buf = append(buf, pg[int(q):int(r)]...)
		length -= uint64(len(pg[int(q):int(r)]))
	}
	return buf, nil
}

func (pm *contractPM) setRawData(address, length uint64, data []byte) error {
	if length == 0 || len(data) != int(length) {
		return nil
	}
	o := 0
	for i, j, k := uint32(0), pde(((address&PAGE_MASK + length + PAGE_MASK) & ^uint64(PAGE_MASK))), pde(address); i < j; i++ {
		pg, _ := pm.achievePage(i + k)
		q, r := uint64(0), PAGE_SIZE
		switch i {
		case 0:
			q = address & PAGE_MASK
			if i == j-1 {
				r = q + length
			}
		case j - 1:
			r = length
		}
		length -= r - q
		for q < r {
			pg[int(q)] = data[o]
			o++
			q++
		}
		pm.setPage(i+k, pg)
	}
	return nil
}

func (pm *contractPM) alloc(typ int, size uint64) (uint64, error) {
	if size == 0 {
		return 0, nil
	}
	offset, prevOffset := pm.reserved, uint64(0)
	if typ == FLASH {
		offset = RAM_LIMIT
	}
	h, prevH := contractMemHeader{}, contractMemHeader{}
	for {
		prevH = h
		prevOffset = offset
		if hData, err := pm.achieveRawData(offset, MEM_HEADER_SIZE); err == nil {
			if _, err = h.Read(hData); err != nil {
				return 0, fmt.Errorf("Fate Alloc: %v", err)
			}
		} else {
			return 0, fmt.Errorf("Fate Alloc: %v", err)
		}
		switch {
		case h.size >= size && h.size <= size+MEM_HEADER_SIZE:
			prevH.next = h.next
			hData, _ := prevH.Show()
			pm.setRawData(prevOffset, MEM_HEADER_SIZE, hData)
			return offset + MEM_HEADER_SIZE, nil
		case h.size > size+MEM_HEADER_SIZE:
			h.size -= size + MEM_HEADER_SIZE
			hData, _ := h.Show()
			pm.setRawData(offset, MEM_HEADER_SIZE, hData)
			offset = offset + h.size + MEM_HEADER_SIZE
			recentH := contractMemHeader{
				next: 0,
				size: size,
			}
			hData, _ = recentH.Show()
			pm.setRawData(offset, MEM_HEADER_SIZE, hData)
			return offset + MEM_HEADER_SIZE, nil
		}
		if offset = h.next; offset == 0 {
			break
		}
	}
	return 0, errors.New("Fate Alloc: Out of Memory")
}

func (pm *contractPM) realloc(address, oldSize, recentSize uint64) (uint64, error) {
	h := contractMemHeader{}
	if hData, err := pm.achieveRawData(address-MEM_HEADER_SIZE, MEM_HEADER_SIZE); err == nil {
		if _, err = h.Read(hData); err != nil {
			return address, fmt.Errorf("Fate Realloc: %v", err)
		}
	} else {
		return address, fmt.Errorf("Fate Realloc: %v", err)
	}
	if h.size >= recentSize {
		return address, nil
	}
	data, err := pm.achieveRawData(address, oldSize)
	if err != nil {
		return address, fmt.Errorf("Fate Realloc: %v", err)
	}
	recentAddr, err := pm.alloc(StoreType(address), recentSize)
	if err != nil {
		return address, err
	}
	data = append(data, make([]byte, recentSize-oldSize)...)
	pm.setRawData(recentAddr, recentSize, data)
	pm.free(address)
	return recentAddr, nil
}

func (pm *contractPM) free(address uint64) error {
	offset := pm.reserved
	if address >= RAM_LIMIT {
		offset = RAM_LIMIT
	}
	h := contractMemHeader{}
	for {
		if hData, err := pm.achieveRawData(offset, MEM_HEADER_SIZE); err == nil {
			if _, err = h.Read(hData); err != nil {
				return fmt.Errorf("Fate Free: %v", err)
			}
		} else {
			return fmt.Errorf("Fate Free: %v", err)
		}
		if address >= offset+MEM_HEADER_SIZE+h.size && address < h.next {
			break
		}
		if offset = h.next; offset == 0 {
			return errors.New("Fate Free: Illegal Address")
		}
	}
	freeH := contractMemHeader{}
	if hData, err := pm.achieveRawData(address-MEM_HEADER_SIZE, MEM_HEADER_SIZE); err == nil {
		if _, err = freeH.Read(hData); err != nil {
			return fmt.Errorf("Fate Free: %v", err)
		}
	} else {
		return fmt.Errorf("Fate Free: %v", err)
	}
	if offset != pm.reserved && offset != RAM_LIMIT && offset+h.size+MEM_HEADER_SIZE == address-MEM_HEADER_SIZE {
		if address+freeH.size == h.next {
			nextH := contractMemHeader{}
			if hData, err := pm.achieveRawData(h.next, MEM_HEADER_SIZE); err == nil {
				if _, err = nextH.Read(hData); err != nil {
					return fmt.Errorf("Fate Free: %v", err)
				}
			} else {
				return fmt.Errorf("Fate Free: %v", err)
			}
			h.next = nextH.next
			h.size += MEM_HEADER_SIZE + freeH.size + nextH.size + MEM_HEADER_SIZE
		} else {
			h.size += MEM_HEADER_SIZE + freeH.size
		}
		hData, _ := h.Show()
		pm.setRawData(offset, MEM_HEADER_SIZE, hData)
	} else {
		if address+freeH.size == h.next {
			nextH := contractMemHeader{}
			if hData, err := pm.achieveRawData(h.next, MEM_HEADER_SIZE); err == nil {
				if _, err = nextH.Read(hData); err != nil {
					return fmt.Errorf("Fate Free: %v", err)
				}
			} else {
				return fmt.Errorf("Fate Free: %v", err)
			}
			freeH.next = nextH.next
			freeH.size += nextH.size + MEM_HEADER_SIZE
		} else {
			freeH.next = h.next
			h.next = address - MEM_HEADER_SIZE
		}
		hData, _ := h.Show()
		pm.setRawData(offset, MEM_HEADER_SIZE, hData)
		hData, _ = freeH.Show()
		pm.setRawData(address-MEM_HEADER_SIZE, MEM_HEADER_SIZE, hData)
	}
	return nil
}
