package virtul

import (
	"errors"

	"kortho/util/mixed"
	"kortho/util/storage/db"
)

const MEM_HEADER_SIZE = 16

const (
	P = 0x01
	A = 0x02
	D = 0x04
)

const (
	RAM = iota
	FLASH
)

const (
	KB = uint64(1024)
	MB = uint64(1024 * 1024)
	GB = uint64(1024 * 1024 * 1024)
)

const (
	RAM_LIMIT   = 4 * GB
	FLASH_LIMIT = 64 * GB
)

const (
	PDE_OFF        = 22
	PDE_MASK       = 0x3FFF
	RAM_PAGE_COUNT = 1024
	PAGE_COUNT     = 16384
	PAGE_SIZE      = 4 * MB
	PAGE_MASK      = 0x3FFFFF
)

type contractMemHeader struct {
	next uint64
	size uint64
}

type contractPage struct {
	flag uint8
	page []byte
}

// page manager
type contractPM struct {
	pc       int // page count
	db       db.DB
	reserved uint64
	pages    []*contractPage
}

type contractVM struct {
	pm *contractPM
}

type FateVM interface {
	Close() error
	Flush() error
	Size() uint64
	Free(uint64) error
	SetExecute([]byte) error
	GetExecute() ([]byte, error)
	SetPage(uint32, []byte) error
	GetPage(uint32) ([]byte, error)
	Alloc(int, uint64) (uint64, error)
	SetRawData(uint64, uint64, []byte) error
	GetRawData(uint64, uint64) ([]byte, error)
	Realloc(uint64, uint64, uint64) (uint64, error)
}

var Sentry = []byte{
	16, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, // next = 4G + 16, size = 0
	0, 0, 0, 0, 0, 0, 0, 0, 224, 255, 255, 255, 14, 0, 0, 0, // next = 0, size = 60G - 32
}

func (a *contractMemHeader) Show() ([]byte, error) {
	buf := []byte{}
	buf = append(buf, mixed.E64func(a.next)...)
	buf = append(buf, mixed.E64func(a.size)...)
	return buf, nil
}

func (a *contractMemHeader) Read(b []byte) ([]byte, error) {
	if len(b) < MEM_HEADER_SIZE {
		return nil, errors.New("Fate Memory Header Read: Illegal Length")
	}
	a.next, _ = mixed.D64func(b[:8])
	a.size, _ = mixed.D64func(b[8:16])
	return b[MEM_HEADER_SIZE:], nil
}

func (virtul *contractVM) Close() error {
	return virtul.pm.db.Close()
}

func (virtul *contractVM) Flush() error {
	return virtul.pm.flush()
}

func (virtul *contractVM) Size() uint64 {
	return uint64(virtul.pm.pc) * PAGE_SIZE
}

func (virtul *contractVM) Free(address uint64) error {
	return virtul.pm.free(address)
}

func (virtul *contractVM) Alloc(typ int, size uint64) (uint64, error) {
	return virtul.pm.alloc(typ, size)
}

func (virtul *contractVM) Realloc(address, oldSize, recentSize uint64) (uint64, error) {
	return virtul.pm.realloc(address, oldSize, recentSize)
}

func (virtul *contractVM) GetRawData(address, length uint64) ([]byte, error) {
	return virtul.pm.achieveRawData(address, length)
}

func (virtul *contractVM) SetRawData(address, length uint64, data []byte) error {
	return virtul.pm.setRawData(address, length, data)
}

func (virtul *contractVM) GetPage(pn uint32) ([]byte, error) {
	return virtul.pm.achievePage(pn)
}

func (virtul *contractVM) SetPage(pn uint32, page []byte) error {
	return virtul.pm.setPage(pn, page)
}

func (virtul *contractVM) GetExecute() ([]byte, error) {
	return virtul.pm.db.GetExecute()
}

func (virtul *contractVM) SetExecute(ft []byte) error {
	return virtul.pm.db.SetExecute(ft)
}
