package blockchain

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"kortho/block"
	"kortho/logger"
	"kortho/transaction"
	"kortho/types"
	"kortho/util/merkle"
	"kortho/util/mixed"
	"kortho/util/storage"
	"kortho/util/storage/db"
	"sync"
	"time"

	"go.uber.org/zap"
)

const (
	BlockchainDBName = "blockchain.db"
	ContractDBName   = "contract.db"
)

var (
	HeightKey = []byte("height")
	HashKey   = []byte("hash")
	NonceKey  = []byte("nonce")
)

var (
	AddrListPrefix = []byte("addr")
	HeightPrefix   = []byte("blockheight")
	TxListName     = []byte("txlist")
)

type Blockchains interface {
	NewBlock([]*transaction.Transaction, types.Address, types.Address, types.Address, types.Address) (*block.Block, error)
	AddBlock(*block.Block, []byte) error

	GetNonce([]byte) (uint64, error)
	GetBalance([]byte) (uint64, error)
	GetHeight() (uint64, error)
	GetHash(uint64) ([]byte, error)
	GetBlockByHash([]byte) (*block.Block, error)
	GetBlockByHeight(uint64) (*block.Block, error)

	GetTransactions(int64, int64) ([]*transaction.Transaction, error)
	GetTransactionByHash([]byte) (*transaction.Transaction, error)
	GetTransactionByAddr([]byte, int64, int64) ([]*transaction.Transaction, error)
	GetMaxBlockHeight() (uint64, error)

	CalculationResults(block *block.Block) *block.Block
	CheckResults(block *block.Block, Ds, Cm, qtj []byte) bool
	GetBlockSection(currentHeight, prevHeight uint64) ([]*block.Block, error)
}

type Blockchain struct {
	mu  sync.RWMutex
	db  storage.DB
	cdb storage.DB
}

func New() *Blockchain {
	bgs := db.New("blockchain.db")
	bgc := db.New("contract.db")
	bc := &Blockchain{db: bgs, cdb: bgc}

	return bc
}

func GetBlockchain() *Blockchain {
	return &Blockchain{db: db.New(BlockchainDBName), cdb: db.New(ContractDBName)}
}

func (bc *Blockchain) NewBlock(txs []*transaction.Transaction, minaddr, Ds, Cm, QTJ types.Address) (*block.Block, error) {
	var height, prevHeight uint64
	var prevHash []byte
	prevHeight, err := bc.GetHeight()
	if err != nil {
		prevHeight = 0
		logger.Info("Frist block")
	}
	height = prevHeight + 1

	if height > 1 {
		prevHash, err = bc.GetHash(prevHeight)
		if err != nil {
			logger.Error("Faied to get hash", zap.Error(err), zap.Uint64("height", prevHeight))
			return nil, err
		}
	} else {
		prevHash = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	}

	txBytesList := make([][]byte, 0, len(txs))
	for _, tx := range txs {
		tx.BlockNumber = height
		txBytesList = append(txBytesList, tx.Serialize())
	}
	tree := merkle.New(sha256.New(), txBytesList)
	root := tree.GetMtHash()

	block := &block.Block{
		Height:       height,
		PrevHash:     prevHash,
		Transactions: txs,
		Root:         root,
		Version:      1,
		Timestamp:    time.Now().Unix(),
		Miner:        minaddr,
	}
	block.SetHash()

	return block, nil
}

func (bc *Blockchain) AddBlock(block *block.Block, minaddr []byte) error {
	return nil
}

func (bc *Blockchain) GetNonce(address []byte) (uint64, error) {
	nonce, err := bc.getNonce(address)
	if err != nil {
		bc.setNonce(address, 1)
		return 1, nil
	}

	return nonce, nil
}

func (bc *Blockchain) getNonce(address []byte) (uint64, error) {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	nonceBytes, err := bc.db.Mget(address, NonceKey)
	if err != nil {
		return 0, err
	}

	return mixed.D64func(nonceBytes)
}

func (bc *Blockchain) setNonce(address []byte, nonce uint64) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	nonceBytes := mixed.E64func(nonce)
	return bc.db.Mset(NonceKey, address, nonceBytes)
}

func (bc *Blockchain) GetBalance(address []byte) (uint64, error) {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	balanceBytes, err := bc.db.Get(address)
	if err != nil {
		return 0, err
	}

	return mixed.D64func(balanceBytes)
}

func (bc *Blockchain) GetHeight() (height uint64, err error) {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	heightBytes, err := bc.db.Get(HeightKey)
	if err != nil {
		return
	}

	return mixed.D64func(heightBytes)
}

//根据块高h获取hash
func (bc *Blockchain) GetHash(height uint64) (hash []byte, err error) {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	return bc.db.Get(append(HeightPrefix, mixed.E64func(height)...))
}

//通过块hash获取块数据
func (bc *Blockchain) GetBlockByHash(hash []byte) (*block.Block, error) { //
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	blockData, err := bc.db.Get(hash)
	if err != nil {
		return nil, err
	}
	return block.Deserialize(blockData)
}

//通过块高获取块数据
func (bc *Blockchain) GetBlockByHeight(height uint64) (*block.Block, error) {

	if height < 1 {
		return nil, errors.New("parameter error")
	}

	bc.mu.RLock()
	defer bc.mu.RUnlock()

	// 1、先获取到hash
	hash, err := bc.db.Get(append(HeightPrefix, mixed.E64func(height)...))
	if err != nil {
		return nil, err
	}

	// 2、通过hash获取block
	blockData, err := bc.db.Get(hash)
	if err != nil {
		return nil, err
	}

	return block.Deserialize(blockData)
}

func (bc *Blockchain) GetTransactions(start, end int64) ([]*transaction.Transaction, error) {
	//获取hash的交易
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	hashList, err := bc.db.Lrange(TxListName, start, end)
	if err != nil {
		logger.Error("failed to get txlist", zap.Error(err))
		return nil, err
	}

	transactions := make([]*transaction.Transaction, 0, len(hashList))
	for _, hash := range hashList {
		txBytes, err := bc.db.Get(hash)
		if err != nil {
			return nil, err
		}

		transaction := &transaction.Transaction{}
		if err := json.Unmarshal(txBytes, transaction); err != nil {
			logger.Error("Failed to unmarshal bytes", zap.Error(err))
			return nil, err
		}

		transactions = append(transactions, transaction)
	}

	return transactions, err
}

func (bc *Blockchain) GetTransactionByHash(hash []byte) (*transaction.Transaction, error) {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	txBytes, err := bc.db.Get(hash)
	if err != nil {
		logger.Error("failed to get hash", zap.Error(err))
		return nil, err
	}

	transaction := &transaction.Transaction{}
	err = json.Unmarshal(txBytes, transaction)
	if err != nil {
		return nil, err
	}

	return transaction, nil
}

func (bc *Blockchain) GetTransactionByAddr(address []byte, start, end int64) ([]*transaction.Transaction, error) {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	txHashList, err := bc.db.Lrange(append(AddrListPrefix, address...), start, end)
	if err != nil {
		logger.Error("failed to get addrhashlist", zap.Error(err))
		return nil, err
	}

	transactions := make([]*transaction.Transaction, 0, len(txHashList))
	for _, hash := range txHashList {
		txBytes, err := bc.db.Get(hash)
		if err != nil {
			logger.Error("Failed to get transaction", zap.Error(err), zap.ByteString("hash", hash))
			return nil, err
		}
		var tx transaction.Transaction
		if err := json.Unmarshal(txBytes, &tx); err != nil {
			logger.Error("Failed to unmarshal bytes", zap.Error(err))
			return nil, err
		}
		transactions = append(transactions, &tx)
	}

	return transactions, nil
}

func (bc *Blockchain) GetContractDB() storage.DB {
	return bc.cdb
}

func (bc *Blockchain) GetMaxBlockHeight() (uint64, error) {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	heightBytes, err := bc.db.Get(HeightKey)
	if err != nil {
		return 0, err
	}

	return mixed.D64func(heightBytes)
}

func setMinerFee(tx storage.Transaction, to []byte, amount uint64) error {
	tobalance, err := tx.Get(to)
	if err != nil {
		tobalance = mixed.E64func(0)
	}

	toBalance, _ := mixed.D64func(tobalance)
	toBalanceBytes := mixed.E64func(toBalance + amount)

	return setBalance(tx, to, toBalanceBytes)
}

func setTxbyaddr(DBTransaction storage.Transaction, addr []byte, tx transaction.Transaction) error {
	listNmae := append(AddrListPrefix, addr...)
	_, err := DBTransaction.Llpush(listNmae, tx.Hash)
	return err
}

func setNonce(DBTransaction storage.Transaction, addr, nonce []byte) error {
	DBTransaction.Mdel(NonceKey, addr)
	return DBTransaction.Mset(NonceKey, addr, nonce)
}

func setTxList(DBTransaction storage.Transaction, tx *transaction.Transaction) error {
	if _, err := DBTransaction.Llpush(TxListName, tx.Hash); err != nil {
		logger.Error("Failed to push txhash", zap.Error(err))
		return err
	}

	//交易hash->交易数据
	txBytes, _ := json.Marshal(tx)
	if err := DBTransaction.Set(tx.Hash, txBytes); err != nil {
		logger.Error("Failed to set transaction", zap.Error(err))
		return err
	}

	return nil
}

func setToAccount(dbTransaction storage.Transaction, transaction *transaction.Transaction) error {
	var balance uint64
	balanceBytes, err := dbTransaction.Get(transaction.To.Bytes())
	if err != nil {
		balance = 0
	} else {
		balance, err = mixed.D64func(balanceBytes)
		if err != nil {
			return err
		}
	}

	newBalanceBytes := mixed.E64func(balance + transaction.Amount)
	if err := setBalance(dbTransaction, transaction.To.Bytes(), newBalanceBytes); err != nil {
		return err
	}

	return nil
}

func setAccount(DBTransaction storage.Transaction, tx *transaction.Transaction) error {
	from, to := tx.From.Bytes(), tx.To.Bytes()

	fromBalBytes, _ := DBTransaction.Get(from)
	fromBalance, _ := mixed.D64func(fromBalBytes)
	if tx.IsTokenTransaction() {
		fromBalance -= tx.Amount + tx.Fee
	} else {
		fromBalance -= tx.Amount
	}

	tobalance, err := DBTransaction.Get(to)
	if err != nil {
		setBalance(DBTransaction, to, mixed.E64func(0))
		tobalance = mixed.E64func(0)
	}

	toBalance, _ := mixed.D64func(tobalance)
	toBalance += tx.Amount

	Frombytes := mixed.E64func(fromBalance)
	Tobytes := mixed.E64func(toBalance)

	if err := setBalance(DBTransaction, from, Frombytes); err != nil {
		return err
	}
	if err := setBalance(DBTransaction, to, Tobytes); err != nil {
		return err
	}

	return nil
}

func setBalance(tx storage.Transaction, addr, balance []byte) error {
	tx.Del(addr)
	return tx.Set(addr, balance)
}

func (bc *Blockchain) CalculationResults(block *block.Block) *block.Block {
	block.Results = make(map[string]uint64)
	for _, tx := range block.Transactions {
		if !tx.IsCoinBaseTransaction() {
			if balance, ok := block.Results[tx.From.String()]; ok {
				balance -= tx.Amount
				if tx.IsTokenTransaction() {
					balance -= tx.Fee
				}
				block.Results[tx.From.String()] = balance
			} else {
				balanceBytes, err := bc.db.Get(tx.From.Bytes())
				if err != nil {
					//TODO:如果找不到改该地址如何处理
				}
				balance, _ := mixed.D64func(balanceBytes)
				block.Results[tx.From.String()] = balance - tx.Amount
			}
		}

		if balance, ok := block.Results[tx.To.String()]; ok {
			balance += tx.Amount
			block.Results[tx.To.String()] = balance
		} else {
			var balance uint64
			balanceBytes, err := bc.db.Get(tx.To.Bytes())
			if err != nil {
				balance = 0
			} else {
				balance, _ = mixed.D64func(balanceBytes)
			}
			block.Results[tx.To.String()] = balance + tx.Amount
		}
	}

	return block
}

func (bc *Blockchain) CheckResults(block *block.Block, Ds, Cm, qtj []byte) bool {

	//1、最后一笔交易必须是coinbase交易
	if !block.Transactions[len(block.Transactions)-1].IsCoinBaseTransaction() {
		logger.Info("The last transaction is not coin base")
		return false
	}

	//2、验证leader和follower的区块交易地址的余额
	currBlock := bc.CalculationResults(block)
	for _, tx := range block.Transactions {
		if !tx.IsCoinBaseTransaction() {
			if balance, ok := block.Results[tx.From.String()]; !ok {
				logger.Info("address is not exist", zap.String("from", tx.From.String()))
				return false
			} else if balance != currBlock.Results[tx.From.String()] {
				logger.Info("balance is not equal", zap.Uint64("curBalnce", balance), zap.Uint64("resBalance", currBlock.Results[tx.From.String()]))
				return false
			}
		}

		if balance, ok := block.Results[tx.To.String()]; !ok {
			logger.Info("address is not exist", zap.String("to", tx.To.String()))
			return false
		} else if balance != currBlock.Results[tx.To.String()] {
			logger.Info("balance is not equal", zap.Uint64("curBalnce", balance), zap.Uint64("resBalance", currBlock.Results[tx.To.String()]))
			return false
		}
	}

	return true
}

func (bc *Blockchain) GetBlockSection(currentHeight, prevHeight uint64) ([]*block.Block, error) {
	var blocks []*block.Block
	for i := prevHeight; i <= currentHeight; i++ {
		hash, err := bc.db.Mget(mixed.E64func(i), HashKey)
		if err != nil {
			logger.Error("Failed to get hash", zap.Error(err), zap.Uint64("height", i))
			return nil, err
		}
		B, err := bc.db.Get(hash)
		if err != nil {
			logger.Error("Failed to get block", zap.Error(err), zap.String("hash", string(hash)))
			return nil, err
		}

		blcok := &block.Block{}
		if err := json.Unmarshal(B, blcok); err != nil {
			logger.Error("Failed to unmarshal block", zap.Error(err), zap.String("hash", string(hash)))
			return nil, err
		}
		blocks = append(blocks, blcok)
	}

	return blocks, nil
}
