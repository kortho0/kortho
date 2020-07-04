package sum

import (
	"hash"
)

func Sum(h hash.Hash64, data []byte) uint64 {
	h.Reset()
	h.Write(data)
	return h.Sum64()
}
