package transaction

import (
	"bytes"
	"encoding/json"
	"kortho/types"
	miscellaneous "kortho/util/mixed"
	"time"

	"golang.org/x/crypto/ed25519"
	"golang.org/x/crypto/sha3"
)

const Lenthaddr = 44

type Address [Lenthaddr]byte

type Transaction struct {
	Nonce       uint64        `json:"nonce"`
	BlockNumber uint64        `json:"blocknumber"`
	Amount      uint64        `json:"amount"`
	From        types.Address `json:"from"`
	To          types.Address `json:"to"`
	Hash        []byte        `json:"hash"`
	Signature   []byte        `json:"signature"`
	Time        int64         `json:"time"`
	Fee         uint64        `json:"fee"`
	Root        []byte        `json:"root"`
	Script      string        `json:"script"`
}

type Option struct {
	Fee    uint64
	Script string
	Root   []byte
}

type ModOption func(option *Option)

func WithToken(fee uint64, script string, root []byte) ModOption {
	return func(option *Option) {
		option.Fee = fee
		option.Script = script
		option.Root = root
	}
}

func NewTransaction(nonce, amount uint64, from, to types.Address, modOptions ...ModOption) *Transaction {
	var option *Option
	for _, modOption := range modOptions {
		modOption(option)
	}

	tx := &Transaction{
		Nonce:  nonce,
		Amount: amount,
		From:   from,
		To:     to,
		Time:   time.Now().Unix(),
		Fee:    option.Fee,
		Script: option.Script,
		Root:   option.Root,
	}

	return tx
}

func NewCoinBaseTransaction(address types.Address, amount uint64) *Transaction {
	from := new(types.Address)
	transaction := Transaction{
		From:   *from,
		To:     address,
		Nonce:  0,
		Amount: amount,
		Time:   time.Now().Unix(),
	}
	transaction.HashTransaction()
	return &transaction
}

func (tx *Transaction) IsCoinBaseTransaction() bool {
	return tx.From.IsNil()
}

func (tx *Transaction) IsTokenTransaction() bool {
	if len(tx.Script) != 0 && tx.Fee != 0 {
		return true
	}
	return false
}

func (tx *Transaction) Serialize() []byte {
	txBytes, _ := json.Marshal(tx)
	return txBytes
}

func Deserialize(data []byte) (*Transaction, error) {
	var tx Transaction
	if err := json.Unmarshal(data, &tx); err != nil {
		return nil, err
	}
	return &tx, nil
}

func (tx *Transaction) GetTime() int64 {
	return tx.Time
}

func (tx *Transaction) GetNonce() int64 {
	return int64(tx.Nonce)
}

func Newtoken(nonce, amount, fee uint64, from, to types.Address, script string) *Transaction {
	tx := &Transaction{
		Nonce:  nonce,
		Amount: amount,
		From:   from,
		To:     to,
		Time:   time.Now().Unix(),
		Fee:    fee,
		Script: script,
	}
	tx.HashTransaction()

	return tx
}

func (tx *Transaction) HashTransaction() {
	fromBytes := tx.From[:]
	toBytes := tx.To[:]
	nonceBytes := miscellaneous.E64func(tx.Nonce)
	amountBytes := miscellaneous.E64func(tx.Amount)
	timeBytes := miscellaneous.E64func(uint64(tx.Time))
	txBytes := bytes.Join([][]byte{nonceBytes, amountBytes, fromBytes, toBytes, timeBytes}, []byte{})
	hash := sha3.Sum256(txBytes)
	tx.Hash = hash[:]
}

func (tx *Transaction) TrimmedCopy() *Transaction {
	txCopy := &Transaction{
		Nonce:  tx.Nonce,
		Amount: tx.Amount,
		From:   tx.From,
		To:     tx.To,
		Time:   tx.Time,
	}
	return txCopy
}

func (tx *Transaction) Sgin(privateKey []byte) {
	signatures := ed25519.Sign(ed25519.PrivateKey(privateKey), tx.Hash)
	tx.Signature = signatures
}

func (tx *Transaction) Verify() bool {
	txCopy := tx.TrimmedCopy()
	txCopy.HashTransaction()
	publicKey := tx.From.ToPublicKey()
	return ed25519.Verify(publicKey, txCopy.Hash, tx.Signature)
}

func (tx *Transaction) EqualNonce(transaction *Transaction) bool {
	if tx.From == transaction.From && tx.Nonce == transaction.Nonce {
		return true
	}
	return false
}
