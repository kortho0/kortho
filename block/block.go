package block

import (
	"bytes"
	"encoding/json"
	"fmt"

	"kortho/contract/mixed"
	"kortho/transaction"
	"kortho/types"

	"golang.org/x/crypto/sha3"
)

type MinerTx struct {
	Amount     uint64 `json:"amount"`
	RecAddress []byte `json:"recaddress"`
}

type Block struct {
	Height       uint64                     `json:"height"`    //当前块号
	PrevHash     []byte                     `json:"prevHash"`  //上一块的hash json:"prevBlockHash --> json:"prevHash
	Hash         []byte                     `json:"hash"`      //当前块hash
	Transactions []*transaction.Transaction `json:"txs"`       //交易数据
	Root         []byte                     `json:"root"`      //默克根
	Version      uint64                     `json:"version"`   //版本号
	Timestamp    int64                      `json:"timestamp"` //时间戳
	Miner        types.Address              `json:"miner"`
	Results      map[string]uint64          `json:"res"`
}

func newBlock(height uint64, prevHash []byte, transactions []*transaction.Transaction) *Block {
	block := &Block{
		Height:       height,
		PrevHash:     prevHash,
		Transactions: transactions,
	}
	return block
}

func (b *Block) Serialize() []byte {
	block_byte, err := json.Marshal(b)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return block_byte
}

func Deserialize(data []byte) (*Block, error) {
	var block Block
	if err := json.Unmarshal(data, &block); err != nil {
		return nil, err
	}
	return &block, nil
}

func (b *Block) SetHash() {
	heightBytes := mixed.E64func(b.Height)
	txsBytes, _ := json.Marshal(b.Transactions)
	timeBytes := mixed.E64func(uint64(b.Timestamp))
	blockBytes := bytes.Join([][]byte{heightBytes, b.PrevHash, txsBytes, timeBytes}, []byte{})
	hash := sha3.Sum256(blockBytes)
	b.Hash = hash[:]
}
