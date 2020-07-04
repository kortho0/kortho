package txpool

import (
	"kortho/transaction"
	"kortho/types"
)

type TxHeap []*transaction.Transaction

func (h TxHeap) Len() int { return len(h) }

func (h TxHeap) Less(i, j int) bool {
	if h[i].From == h[j].From {
		if h[i].GetNonce() == h[j].GetNonce() {
			return h[i].GetTime() < h[j].GetTime()
		}
		return h[i].GetNonce() < h[j].GetNonce()
	}

	return h[i].GetTime() < h[j].GetTime()
}

func (h TxHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *TxHeap) Push(x interface{}) {
	*h = append(*h, x.(*transaction.Transaction))
}

func (h *TxHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func (h *TxHeap) Get(i int) interface{} {
	old := *h
	n := len(old)
	x := old[n-i]
	return x
}

func (h *TxHeap) check(fromAddr types.Address, nonce uint64) bool {
	var count = 0
	for _, tx := range *h {
		if fromAddr == tx.From {
			if nonce == tx.Nonce {
				return false
			}
			count++
		}
	}
	if count >= 50 {
		return false
	}
	return true
}
