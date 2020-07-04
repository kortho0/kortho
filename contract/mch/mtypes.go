package main

import (
	"errors"
	"fmt"

	"kortho/contract/mixed"
)

const (
	SLOT_SIZE             = 24
	BUCKET_HEADER_SIZE    = 12
	DEFAULT_CHAIN_LENGTH  = 4
	DEFAULT_BUCKET_LENGTH = 260
)

type contractBucketHeader struct {
	size uint32
	ktyp uint32
	vtyp uint32
}

type contractBucket struct {
	slots  []uint64
	header *contractBucketHeader
}

type contractSlot struct {
	next uint64
	koff uint64
	voff uint64
}

var hlt = []int{
	98, 6, 85, 150, 36, 23, 112, 164, 135, 207, 169, 5, 26, 64, 165, 219,
	61, 20, 68, 89, 130, 63, 52, 102, 24, 229, 132, 245, 80, 216, 195, 115,
	90, 168, 156, 203, 177, 120, 2, 190, 188, 7, 100, 185, 174, 243, 162, 10,
	237, 18, 253, 225, 8, 208, 172, 244, 255, 126, 101, 79, 145, 235, 228, 121,
	123, 251, 67, 250, 161, 0, 107, 97, 241, 111, 181, 82, 249, 33, 69, 55,
	59, 153, 29, 9, 213, 167, 84, 93, 30, 46, 94, 75, 151, 114, 73, 222,
	197, 96, 210, 45, 16, 227, 248, 202, 51, 152, 252, 125, 81, 206, 215, 186,
	39, 158, 178, 187, 131, 136, 1, 49, 50, 17, 141, 91, 47, 129, 60, 99,
	154, 35, 86, 171, 105, 34, 38, 200, 147, 58, 77, 118, 173, 246, 76, 254,
	133, 232, 196, 144, 198, 124, 53, 4, 108, 74, 223, 234, 134, 230, 157, 139,
	189, 205, 199, 128, 176, 19, 211, 236, 127, 192, 231, 70, 233, 88, 146, 44,
	183, 201, 22, 83, 13, 214, 116, 109, 159, 32, 95, 226, 140, 220, 57, 12,
	221, 31, 209, 182, 143, 92, 149, 184, 148, 62, 113, 65, 37, 27, 106, 166,
	3, 14, 204, 72, 21, 41, 56, 66, 28, 193, 40, 217, 25, 54, 179, 117,
	238, 87, 240, 155, 180, 170, 242, 212, 191, 163, 78, 218, 137, 194, 175, 110,
	43, 119, 224, 71, 122, 142, 42, 160, 104, 48, 247, 103, 15, 11, 138, 239,
}

var phts = []int{
	31, 71, 127, 233, 419, 811, 1597, 3001, 6067, 10007,
}

func (a *contractSlot) Show() ([]byte, error) {
	buf := []byte{}
	buf = append(buf, mixed.E64func(a.next)...)
	buf = append(buf, mixed.E64func(a.koff)...)
	buf = append(buf, mixed.E64func(a.voff)...)
	return buf, nil
}

func (a *contractSlot) Read(b []byte) ([]byte, error) {
	if len(b) < SLOT_SIZE {
		return nil, errors.New("Fate Slot Read: Illegal Length")
	}
	a.next, _ = mixed.D64func(b[:8])
	a.koff, _ = mixed.D64func(b[8:16])
	a.voff, _ = mixed.D64func(b[16:24])
	return b[SLOT_SIZE:], nil
}

func (a *contractBucketHeader) Show() ([]byte, error) {
	buf := []byte{}
	buf = append(buf, mixed.E32func(a.size)...)
	buf = append(buf, mixed.E32func(a.ktyp)...)
	buf = append(buf, mixed.E32func(a.vtyp)...)
	return buf, nil
}

func (a *contractBucketHeader) Read(b []byte) ([]byte, error) {
	if len(b) < BUCKET_HEADER_SIZE {
		return nil, errors.New("Fate Bucket Header Read: Illegal Length")
	}
	a.size, _ = mixed.D32func(b[:4])
	a.ktyp, _ = mixed.D32func(b[4:8])
	a.vtyp, _ = mixed.D32func(b[8:12])
	return b[BUCKET_HEADER_SIZE:], nil
}

func (a *contractBucket) Show() ([]byte, error) {
	buf := []byte{}
	data, _ := a.header.Show()
	buf = append(buf, data...)
	for _, v := range a.slots {
		buf = append(buf, mixed.E64func(v)...)
	}
	return buf, nil
}

func (a *contractBucket) Read(b []byte) ([]byte, error) {
	var err error

	a.header = &contractBucketHeader{}
	if b, err = a.header.Read(b); err == nil {
		if len(b) < int(a.header.size*8) {
			return nil, errors.New("Fate Bucket Read: Illegal Length")
		}
		a.slots = make([]uint64, int(a.header.size))
		for i := uint32(0); i < a.header.size; i++ {
			a.slots[int(i)], _ = mixed.D64func(b[:8])
			b = b[8:]
		}
		return b, nil
	} else {
		return nil, fmt.Errorf("Fate Bucket Read: %v", err)
	}
}
