package node

import (
	"encoding/gob"
	"kortho/blockchain"
	"kortho/config"
	"kortho/p2p"
	"kortho/transaction"
	"kortho/txpool"
	"kortho/types"
)

type Node interface {
	Run()
	Stop()
	Join([]string) error
	//Broadcast(*transaction.Transaction)
	Broadcast(v interface{})
}

type node struct {
	p    p2p.P2P
	pool *txpool.TxPool
	bc   *blockchain.Blockchain
}

func init() {
	gob.Register(types.Address{})
	gob.Register(transaction.Transaction{})
}

func New(cfg *config.P2PConfigInfo, pool *txpool.TxPool, bc *blockchain.Blockchain) (*node, error) {
	n := &node{pool: pool, bc: bc}
	p, err := p2p.New(p2p.Config{cfg.BindPort, cfg.NodeName, cfg.BindAddr, cfg.AdvertiseAddr}, n, recv)
	if err != nil {
		return nil, err
	}
	n.p = p
	return n, nil
}

// func (n *node) Broadcast(tx *transaction.Transaction) {
// 	data, _ := p2p.Encode(*tx)
// 	n.p.Broadcast(data)
// }
func (n *node) Broadcast(v interface{}) {
	data, _ := p2p.Encode(v)
	n.p.Broadcast(data)
}

func (n *node) Run() {
	n.p.Run()
}

func (n *node) Stop() {
	n.p.Stop()
}

func (n *node) Join(ns []string) error {
	return n.p.Join(ns)
}

// func recv(u interface{}, data []byte) {
// 	var tx transaction.Transaction

// 	n := u.(*node)
// 	if err := p2p.Decode(data, &tx); err == nil {
// 		n.pool.Add(&tx, n.bc)
// 	}
// }

func recv(u interface{}, data []byte) {
	var tx transaction.Transaction
	n := u.(*node)
	var dt []byte
	if err := p2p.Decode(data, &dt); err == nil {
		if dt[0] == 'c' {
			n.pool.SetCheckData(dt[1:])
			return
		}
	}

	if err := p2p.Decode(data, &tx); err == nil {
		n.pool.Add(&tx, n.bc)
	}
}
