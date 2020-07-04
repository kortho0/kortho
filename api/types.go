package api

import (
	"encoding/hex"
	"kortho/block"
	"kortho/blockchain"
	"kortho/transaction"
)

var blockChian blockchain.Blockchains

type Transaction struct {
	Nonce       uint64 `json:"nonce"`
	BlockNumber uint64 `json:"blocknumber"`
	Amount      uint64 `json:"amount"`
	From        string `json:"from"`
	To          string `json:"to"`
	Hash        string `json:"hash"`
	Signature   string `json:"signature"`
	Time        int64  `json:"time"`
	Script      string `json:"script"`
}

type Block struct {
	Version       uint64        `json:"version"`
	Height        uint64        `json:"height"`
	PrevBlockHash string        `json:"prevblockhash"`
	Hash          string        `json:"hash"`
	Root          string        `json:"root"`
	Timestamp     int64         `json:"timestamp"`
	Miner         string        `json:"miner"`
	Txs           []Transaction `json:"txs"`
}

func changeTransaction(tx *transaction.Transaction) (result Transaction) {
	result.Hash = hex.EncodeToString(tx.Hash)
	result.From = tx.From.String()
	result.Amount = tx.Amount
	result.Nonce = tx.Nonce
	result.To = tx.To.String()
	result.Signature = hex.EncodeToString(tx.Signature)
	result.Time = tx.Time
	result.BlockNumber = tx.BlockNumber

	return
}

func changeBlock(b *block.Block) (result Block) {
	result.Height = b.Height
	result.Hash = hex.EncodeToString(b.Hash)
	result.PrevBlockHash = hex.EncodeToString(b.PrevHash)
	result.Root = hex.EncodeToString(b.Root)
	result.Timestamp = b.Timestamp
	result.Version = b.Version
	result.Miner = b.Miner.String()

	for _, tx := range b.Transactions {
		result.Txs = append(result.Txs, changeTransaction(tx))
	}
	return
}
