package api

import (
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"

	"kortho/logger"

	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

type Server struct {
	port string
	fasthttprouter.Router
}

func (s *Server) Run() {

	s.GET("/block", s.GetBlockHandler)
	s.GET("/balance", s.GetBalanceHandler)
	s.GET("/transaction", s.GetTransactionHandler)

	if err := fasthttp.ListenAndServe(s.port, s.Handler); err != nil {
		logger.Error("asthttp.ListenAndServe failed", zap.Error(err))
		os.Exit(-1)
	}
}

func (s *Server) GetBalanceHandler(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.Set("Access-Control-Allow-Origin", "*")
	ctx.Response.Header.Set("Content-Type", "application/json")
	var result resultInfo
	defer func() {
		jsbyte, _ := json.Marshal(result)
		ctx.Write(jsbyte)
	}()

	address := ctx.QueryArgs().Peek("address")
	if len(address) == 0 {
		result.Code = failedCode
		result.Message = ErrParameters
		ctx.Response.SetStatusCode(http.StatusBadRequest)
		return
	}

	balance, _ := blockChian.GetBalance(address)

	result.Code = successCode
	result.Message = OK
	result.Data = balance
	ctx.Response.SetStatusCode(http.StatusOK)
	return
}

func (s *Server) GetBlockHandler(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.Set("Access-Control-Allow-Origin", "*")
	ctx.Response.Header.Set("Content-Type", "application/json")
	var result resultInfo
	defer func() {
		jsbyte, _ := json.Marshal(result)
		ctx.Write(jsbyte)
	}()

	args := ctx.QueryArgs()
	number, numErr := args.GetUint("number")
	page, pageErr := args.GetUint("page")
	size, sizeErr := args.GetUint("size")
	if numErr != nil && (pageErr != nil || sizeErr != nil) {
		logger.Error("Failed to get parameters", zap.Int("number", number), zap.Errors("errors", []error{pageErr, sizeErr}))
		result.Code = failedCode
		result.Message = ErrParameters
		ctx.Response.SetStatusCode(http.StatusBadRequest)
		return
	}

	var viewBlocks []Block
	if number > 0 {
		block, err := blockChian.GetBlockByHeight(uint64(number))
		if err != nil {
			logger.Error("Failed to get blockHeight", zap.Error(err), zap.Int("number", number))
			result.Code = failedCode
			result.Message = ErrParameters
			ctx.Response.SetStatusCode(http.StatusBadRequest)
			return
		}
		viewBlocks = append(viewBlocks, changeBlock(block))
	} else if pageErr == nil && sizeErr == nil {
		maxHeight, err := blockChian.GetMaxBlockHeight()
		if err != nil {
			logger.Error("Failed to get maxHeight", zap.Error(err))
			result.Code = failedCode
			result.Message = ErrParameters
			ctx.Response.SetStatusCode(http.StatusBadRequest)
			return
		}
		var start, end uint64
		start = maxHeight - uint64((page-1)*size)
		if maxHeight < uint64(page*size) {
			end = 0
		} else {
			end = maxHeight - uint64(page*size)
		}

		if start < 0 {
			logger.Error("Parameters error", zap.Uint64("max height", maxHeight), zap.Int("page", page), zap.Int("size", size))
			result.Code = failedCode
			result.Message = ErrParameters
			ctx.Response.SetStatusCode(http.StatusBadRequest)
			return
		}

		for ; start > end; start-- {
			block, err := blockChian.GetBlockByHeight(start)
			if err != nil {
				logger.Error("Failed to get block", zap.Error(err), zap.Uint64("height", start))
				result.Code = failedCode
				result.Message = ErrParameters
				ctx.Response.SetStatusCode(http.StatusBadRequest)
				return
			}
			viewBlocks = append(viewBlocks, changeBlock(block))
		}
	} else {
		logger.Error("Parameters error", zap.Int("number", number), zap.Int("page", page), zap.Int("size", size))
		result.Code = failedCode
		result.Message = ErrParameters
		ctx.Response.SetStatusCode(http.StatusBadRequest)
		return
	}

	result.Code = successCode
	result.Message = OK
	result.Data = viewBlocks
	ctx.Response.SetStatusCode(http.StatusOK)
	return
}

func (s *Server) GetTransactionHandler(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.Set("Access-Control-Allow-Origin", "*")
	ctx.Response.Header.Set("Content-Type", "application/json")
	var result resultInfo
	defer func() {
		jsbyte, _ := json.Marshal(result)
		ctx.Write(jsbyte)
	}()

	args := ctx.QueryArgs()
	address := args.Peek("address")
	hash := args.Peek("hash")
	page, _ := args.GetUint("page")
	size, _ := args.GetUint("size")

	start := (page - 1) * size
	end := start + size - 1

	var viewTxs []Transaction
	if len(address) != 0 {

		txs, err := blockChian.GetTransactionByAddr(address, int64(start), int64(end))
		if err != nil {
			logger.Error("Failed to get transactions", zap.Error(err), zap.String("address", string(address)),
				zap.Int("start", start), zap.Int("end", end))
			result.Code = -1
			result.Message = "failed"
			return
		}
		for _, tx := range txs {
			viewTxs = append(viewTxs, changeTransaction(tx))
		}
		result.Data = viewTxs
	} else if len(hash) != 0 {
		hash, _ = hex.DecodeString(string(hash))
		tx, err := blockChian.GetTransactionByHash(hash)
		if err != nil {
			logger.Error("Failed to get transactions", zap.Error(err), zap.String("hash", string(hash)))
			result.Code = failedCode
			result.Message = ErrParameters
			return
		}
		result.Data = append(viewTxs, changeTransaction(tx))
	} else {
		txs, err := blockChian.GetTransactions(int64(start), int64(end))
		if err != nil {
			logger.Error("Failed to get transactions", zap.Error(err), zap.Int("start", start), zap.Int("end", end))
			result.Code = failedCode
			result.Message = ErrParameters
			return
		}

		for _, tx := range txs {
			viewTxs = append(viewTxs, changeTransaction(tx))
		}
		result.Data = viewTxs
	}
	result.Code = successCode
	result.Message = OK
	ctx.Response.SetStatusCode(http.StatusOK)
	return
}
