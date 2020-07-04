package api

import (
	"kortho/blockchain"
	"kortho/config"
	"kortho/p2p/node"
	"kortho/txpool"

	"github.com/buaazp/fasthttprouter"
)

func Start(cfg *config.APIConfigInfo, bc *blockchain.Blockchain, tp *txpool.TxPool, n node.Node) {
	greeter := newGreeter(cfg.RPCConfig, bc, tp, n)
	go greeter.RunRPC()

	blockChian = bc
	server := &Server{cfg.WEBConfig.Address, fasthttprouter.Router{}}
	server.Run()
}
