package txpool

import (
	"container/heap"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"kortho/block"
	"kortho/blockchain"
	"kortho/logger"
	"kortho/transaction"
	"kortho/types"
	"kortho/util/merkle"
	"sync"
	"time"

	"go.uber.org/zap"
)

const ReadyTotalQuantity = 500

var QTJPubKey []byte

type TxPool struct {
	Mutx sync.RWMutex
	List *TxHeap
	Idhc map[string]CheckBlock
}
type stateInfo struct {
	nonce   uint64
	balance uint64
}

type CheckBlock struct {
	Nodeid string
	Height uint64
	Hash   []byte
	Code   bool
}

func (pool *TxPool) SetCheckData(data []byte) error {
	pool.Mutx.Lock()
	defer pool.Mutx.Unlock()
	var cb CheckBlock
	err := json.Unmarshal(data, &cb)
	if err == nil {
		key := cb.Nodeid + string(cb.Hash)
		pool.Idhc[key] = cb
		fmt.Println("SetCheckData length = : ", len(pool.Idhc))
		return nil
	}
	return err
}

func New(address string) (*TxPool, error) {
	pool := &TxPool{List: new(TxHeap)}
	heap.Init(pool.List)

	//TODO:判断Address是否符合条件
	addr, err := types.StringToAddress(address)
	if err != nil {
		logger.Info("failed to verify address", zap.String("address", address))
		return nil, err
	}
	QTJPubKey = addr.ToPublicKey()
	return pool, nil
}

func (pool *TxPool) Add(tx *transaction.Transaction, bc blockchain.Blockchains) error {
	pool.Mutx.Lock()
	defer pool.Mutx.Unlock()

	if pool.List.Len() > 1500 {
		return errtxoutrange
	}

	if !verify(*tx, bc) {
		return errtx
	}

	if !pool.List.check(tx.From, tx.Nonce) {
		return errtomuch
	}

	heap.Push(pool.List, tx)
	return nil

}

func (pool *TxPool) Pending(Bc blockchain.Blockchains) (readyTxs []*transaction.Transaction) {
	pool.Mutx.Lock()
	defer pool.Mutx.Unlock()
	var noReadyTxs []*transaction.Transaction
	var AddrStateMap = make(map[string]stateInfo)

	for pool.List.Len() != 0 && len(readyTxs) < ReadyTotalQuantity {
		var balance, nonce uint64
		tx := heap.Pop(pool.List).(*transaction.Transaction)

		if state, ok := AddrStateMap[tx.From.String()]; !ok {
			balance, _ = Bc.GetBalance(tx.From.Bytes())
			nonce, _ = Bc.GetNonce(tx.From.Bytes())
		} else {
			nonce, balance = state.balance, state.nonce
		}

		if balance >= tx.Amount && nonce == tx.Nonce {
			if tx.IsTokenTransaction() {
				balance -= tx.Amount + tx.Fee
			} else {
				balance -= tx.Amount
			}
			nonce = tx.Nonce + 1
			AddrStateMap[tx.From.String()] = stateInfo{nonce, balance}
			readyTxs = append(readyTxs, tx)
		} else if tx.Nonce > nonce {
			noReadyTxs = append(noReadyTxs, tx)
		} else {
			logger.Info("nonce or amount error", zap.Uint64("current nonce", nonce), zap.Uint64("tx nonce", nonce),
				zap.Uint64("balance", balance), zap.Uint64("amount", tx.Amount))
		}
	}

	//TODO:要避免无法上链的tx越积越多,可以设置nonce的差距，比如大于5直接舍弃
	for _, tx := range noReadyTxs {
		pool.List.Push(tx)
	}

	return
}

func verify(tx transaction.Transaction, Bc blockchain.Blockchains) bool {
	if !tx.IsCoinBaseTransaction() {
		//1、检查地址
		if !tx.From.Verify() {
			logger.Info("faile to verify address", zap.String("from", tx.From.String()))
			return false
		}

		//2、验证签名
		if !tx.Verify() {
			logger.Info("failed to verify transaction",
				zap.String("from", tx.From.String()), zap.String("to", tx.To.String()), zap.Uint64("amount", tx.Amount))
			return false
		}

		//3、验证余额
		balance, _ := Bc.GetBalance(tx.From.Bytes())
		if tx.Amount < 500000 || tx.Amount > balance {
			logger.Info("failed to verify amount", zap.String("from", tx.From.String()),
				zap.String("to", tx.To.String()), zap.Uint64("amount", tx.Amount), zap.Uint64("unlockbalance", balance))
			return false
		}

		if tx.IsTokenTransaction() && (tx.Fee < 500000 || tx.Fee+tx.Amount > balance) {
			logger.Info("failed to verify fee", zap.String("from", tx.From.String()),
				zap.String("to", tx.To.String()), zap.Uint64("amount", tx.Amount),
				zap.Uint64("fee", tx.Fee), zap.Uint64("unlockbalance", balance))
			return false
		}

		nonce, _ := Bc.GetNonce(tx.From.Bytes())
		if tx.Nonce < nonce {
			logger.Info("failed to verify nonce", zap.String("from", tx.From.String()),
				zap.Uint64("transaction nonce", tx.Nonce), zap.Uint64("nonce", nonce))
			return false
		}
	}

	if !tx.To.Verify() {
		logger.Info("failed to verify address", zap.String("to", tx.To.String()))
		return false
	}

	return true
}

//检查区块的默克尔根
func VerifyBlcok(b block.Block, Bc blockchain.Blockchains) bool {

	trans := make([][]byte, 0, len(b.Transactions))
	for _, tx := range b.Transactions {
		trans = append(trans, tx.Serialize())
	}

	if trans != nil {
		tree := merkle.New(sha256.New(), trans)
		if ok := tree.VerifyNode(b.Root); ok {
			logger.Error("Faile to verify node")
			return false
		}

		for _, tx := range b.Transactions {
			if !verify(*tx, Bc) {
				logger.Error("Failed to verify transaction")
				return false
			}
		}
	}
	return true
}

func (pool *TxPool) Filter(block block.Block) {
	txs := []*transaction.Transaction(*pool.List)
	txsLenght := len(txs)
	now := time.Now().UTC().Unix()
	for j := 0; j < txsLenght; j++ {
		for _, btx := range block.Transactions {
			//去除与块中重叠或存在超过10s的交易
			if txs[j].EqualNonce(btx) || now-txs[j].Time > 10 {
				txs = append(txs[:j], txs[j+1:]...)
				txsLenght--
				break
			}
		}
	}

	*pool.List = TxHeap(txs[:txsLenght])
}
