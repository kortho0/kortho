package api

import (
	"context"
	"encoding/hex"
	"kortho/api/message"
	"kortho/config"
	"kortho/logger"
	"kortho/p2p/node"
	"kortho/transaction"
	"kortho/txpool"
	"kortho/types"
	"kortho/util"
	"net"
	"os"

	"kortho/blockchain"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
)

type Greeter struct {
	Bc      *blockchain.Blockchain
	tp      *txpool.TxPool
	n       node.Node
	Address string
	tls     tlsInfo
}
type tlsInfo struct {
	certFile string
	keyFile  string
}

func newGreeter(cfg *config.RPCConfigInfo, bc *blockchain.Blockchain, tp *txpool.TxPool, n node.Node) *Greeter {
	grpcServ := &Greeter{
		Bc:      bc,
		tp:      tp,
		n:       n,
		Address: cfg.Address,
		tls: tlsInfo{
			certFile: cfg.CertFile,
			keyFile:  cfg.KeyFile,
		},
	}

	return grpcServ
}

func (g *Greeter) RunRPC() {
	lis, err := net.Listen("tcp", g.Address)
	if err != nil {
		logger.Error("net.Listen", zap.Error(err))
		os.Exit(-1)
	}

	creds, err := credentials.NewServerTLSFromFile(g.tls.certFile, g.tls.keyFile)
	if err != nil {
		logger.Error("credentials.NewServerTLSFromFile",
			zap.Error(err), zap.String("cert file", g.tls.certFile), zap.String("key file", g.tls.keyFile))
		os.Exit(-1)
	}
	server := grpc.NewServer(grpc.Creds(creds), grpc.UnaryInterceptor(ipInterceptor))
	message.RegisterGreeterServer(server, g)
	server.Serve(lis)
}

func txToMsgTxAndOrder(tx *transaction.Transaction) (msgTx message.Tx) {
	msgTx.Hash = hex.EncodeToString(tx.Hash)
	msgTx.From = string(tx.From.Bytes())
	msgTx.Amount = tx.Amount
	msgTx.Nonce = tx.Nonce
	msgTx.To = string(tx.To.Bytes())
	msgTx.Signature = hex.EncodeToString(tx.Signature)
	msgTx.Time = tx.Time
	msgTx.Script = tx.Script

	return
}

func txToMsgTx(tx *transaction.Transaction) (msgTx message.Tx) {
	msgTx.Hash = hex.EncodeToString(tx.Hash)
	msgTx.From = string(tx.From.Bytes())
	msgTx.Amount = tx.Amount
	msgTx.Nonce = tx.Nonce
	msgTx.To = string(tx.To.Bytes())
	msgTx.Signature = hex.EncodeToString(tx.Signature)
	msgTx.Time = tx.Time
	msgTx.Script = tx.Script
	return msgTx
}

func (s *Greeter) GetBalance(ctx context.Context, in *message.ReqBalance) (*message.ResBalance, error) {

	balance, err := s.Bc.GetBalance([]byte(in.Address))
	if err != nil {
		logger.Error("s.Bc.GetBalance", zap.Error(err), zap.String("address", in.Address))
	}
	return &message.ResBalance{Balnce: balance}, nil
}

func (s *Greeter) GetBlockByNum(ctx context.Context, in *message.ReqBlockByNumber) (*message.RespBlock, error) {

	b, err := s.Bc.GetBlockByHeight(in.Height)
	if err != nil {
		logger.Error("s.Bc.GetBlockByHeight", zap.Error(err))
		return nil, grpc.Errorf(codes.InvalidArgument, "height %d not found", in.Height)
	}

	var respdata message.RespBlock
	for _, tx := range b.Transactions {
		tmpTx := txToMsgTx(tx)
		respdata.Txs = append(respdata.Txs, &tmpTx)
	}

	respdata.Height = b.Height
	respdata.Hash = hex.EncodeToString(b.Hash)
	respdata.PrevBlockHash = hex.EncodeToString(b.PrevHash)
	respdata.Root = hex.EncodeToString(b.Root)
	respdata.Timestamp = b.Timestamp
	respdata.Version = b.Version
	respdata.Miner = b.Miner.String()
	return &respdata, nil
}

func (s *Greeter) GetBlockByHash(ctx context.Context, in *message.ReqBlockByHash) (*message.RespBlock, error) {
	h, _ := hex.DecodeString(in.Hash)
	b, err := s.Bc.GetBlockByHash(h)
	if err != nil {
		logger.Error("s.Bc.GetBlockByHash", zap.Error(err), zap.String("hash", in.Hash))
		return nil, grpc.Errorf(codes.InvalidArgument, "hash %s not found", in.Hash)
	}

	var respdata message.RespBlock
	for _, tx := range b.Transactions {
		tmpTx := txToMsgTx(tx)
		respdata.Txs = append(respdata.Txs, &tmpTx)
	}

	respdata.Height = b.Height
	respdata.Hash = hex.EncodeToString(b.Hash)
	respdata.PrevBlockHash = hex.EncodeToString(b.PrevHash)
	respdata.Root = hex.EncodeToString(b.Root)
	respdata.Timestamp = b.Timestamp
	respdata.Version = b.Version
	respdata.Miner = b.Miner.String()
	return &respdata, nil
}

func (s *Greeter) GetTxsByAddr(ctx context.Context, in *message.ReqTx) (*message.ResposeTxs, error) {
	txs, err := s.Bc.GetTransactionByAddr([]byte(in.Address), 0, 9)
	if err != nil {
		logger.Error("s.Bc.GetTransactionByAddr", zap.Error(err))
		return nil, err
	}

	var respData message.ResposeTxs
	for _, tx := range txs {
		tmpTx := txToMsgTxAndOrder(tx)
		respData.Txs = append(respData.Txs, &tmpTx)
	}

	return &respData, nil
}

func (s *Greeter) GetTxByHash(ctx context.Context, in *message.ReqTxByHash) (*message.Tx, error) {
	hash, err := hex.DecodeString(in.Hash)
	if err != nil {

	}

	tx, err := s.Bc.GetTransactionByHash(hash)
	if err != nil {
		logger.Error("Faile to get transaction", zap.Error(err), zap.String("hash", string(hash)))
		return nil, grpc.Errorf(codes.InvalidArgument, "hash %s", in.Hash)
	}
	resp := txToMsgTxAndOrder(tx)
	return &resp, nil
}

func (s *Greeter) GetAddressNonceAt(ctx context.Context, in *message.ReqNonce) (*message.ResposeNonce, error) {
	nonce, err := s.Bc.GetNonce([]byte(in.Address))
	if err != nil {
		logger.Error("s.Bc.GetNonce", zap.Error(err), zap.String("address", in.Address))
		return nil, grpc.Errorf(codes.InvalidArgument, "address %s", in.Address)
	}
	return &message.ResposeNonce{Nonce: nonce}, nil
}

func (s *Greeter) SendTransaction(ctx context.Context, in *message.ReqTransaction) (*message.ResTransaction, error) {

	if in.From == in.To {
		logger.Info("From and To are the same", zap.String("from", in.From), zap.String("to", in.To))
		return nil, grpc.Errorf(codes.InvalidArgument, "from:%s,to:%s", in.From, in.To)
	}

	from, err := types.StringToAddress(in.From)
	if err != nil {
		logger.Error("Parameters error", zap.String("from", in.From), zap.String("to", in.To))
		return nil, grpc.Errorf(codes.InvalidArgument, "from:%s,to:%s", in.From, in.To)
	}

	to, err := types.StringToAddress(in.To)
	if err != nil {
		logger.Error("Parameters error", zap.String("from", in.From), zap.String("to", in.To))
		return nil, grpc.Errorf(codes.InvalidArgument, "from:%s,to:%s", in.From, in.To)
	}

	priv := util.Decode(in.Priv)
	if len(priv) != 64 {
		logger.Info("private key", zap.String("privateKey", in.Priv))
		return nil, grpc.Errorf(codes.InvalidArgument, "private key:%s", in.Priv)
	}

	tx := transaction.NewTransaction(in.Nonce, in.Amount, *from, *to)

	tx.Sgin(priv)

	if err := s.tp.Add(tx, s.Bc); err != nil {
		logger.Error("s.tp.Add", zap.Error(err))
		return nil, grpc.Errorf(codes.InvalidArgument, "data error")
	}

	s.n.Broadcast(tx)
	hash := hex.EncodeToString(tx.Hash)

	return &message.ResTransaction{Hash: hash}, nil
}

func (s *Greeter) CreateAddr(ctx context.Context, in *message.ReqCreateAddr) (*message.RespCreateAddr, error) {
	wallet := types.NewWallet()
	return &message.RespCreateAddr{Address: wallet.Address, Privkey: util.Encode(wallet.PrivateKey)}, nil
}

//GetMaxBlockNumber 获取最大的块号
func (s *Greeter) GetMaxBlockNumber(ctx context.Context, in *message.ReqMaxBlockNumber) (*message.RespMaxBlockNumber, error) {

	maxNumber, err := s.Bc.GetMaxBlockHeight()
	if err != nil {
		logger.Error("s.Bc.GetMaxBlockHeight", zap.Error(err))
		return nil, grpc.Errorf(codes.Internal, "service error")
	}
	return &message.RespMaxBlockNumber{MaxNumber: maxNumber}, nil
}

//GetAddrByPriv 通过私钥获取地址
func (s *Greeter) GetAddrByPriv(ctx context.Context, in *message.ReqAddrByPriv) (*message.RespAddrByPriv, error) {
	privBytes := util.Decode(in.Priv)
	if len(privBytes) != 64 {
		logger.Info("private key", zap.String("in.Priv", in.Priv))
	}
	addr := util.PubtoAddr(privBytes[32:])
	return &message.RespAddrByPriv{Addr: addr}, nil
}
