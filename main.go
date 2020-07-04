package main

import (
	"fmt"
	"os"

	"kortho/api"
	"kortho/blockchain"
	"kortho/config"
	"kortho/logger"
	"kortho/p2p/node"
	"kortho/txpool"
	_ "net/http/pprof"

	"go.uber.org/zap"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Println("load config failed:", err)
		os.Exit(-1)
	}

	if err = logger.InitLogger(cfg.LogConfig); err != nil {
		fmt.Println("logger.InitLogger failed:", err)
		os.Exit(-1)
	}

	tp, err := txpool.New(cfg.ConsensusConfig.QTJ)
	if err != nil {
		logger.Error("Failed to new txpool", zap.Error(err))
		os.Exit(-1)
	}

	bc := blockchain.New()

	n, err := node.New(cfg.P2PConfig, tp, bc)
	if err != nil {
		logger.Error("failed to new p2p node", zap.Error(err))
	}
	go n.Run()
	for _, member := range cfg.P2PConfig.Members {
		if err := n.Join([]string{member}); err != nil {
			logger.Info("Failed to join p2p", zap.Error(err), zap.String("node id", member))
		}
	}
	//go bftNode.NewBftNode(cfg.ConsensusConfig, bc, n, tp)
	api.Start(cfg.APIConfig, bc, tp, n)
}
