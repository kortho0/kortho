package api

import (
	"context"
	"kortho/logger"
	"strings"
	"sync"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
)

var ipLimiter = NewIPRateLimiter(100, 1000)

var ipWhiteList map[string]struct{}
var loadWhiteListOnce sync.Once
var whiteListIPs = []string{
	"127.0.0.1",
}

func loadWhiteList() map[string]struct{} {
	loadWhiteListOnce.Do(func() {
		ipWhiteList = make(map[string]struct{})
		for _, v := range whiteListIPs {
			ipWhiteList[v] = struct{}{}
		}
	})
	return ipWhiteList
}

// UnaryServerInterceptor 将超过限制的ip拦截
//var interceptor grpc.UnaryServerInterceptor
func ipInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return nil, grpc.Errorf(codes.Unavailable, "context information error")
	}

	ipWhiteList := loadWhiteList()
	strList := strings.Split(p.Addr.String(), ":")
	ip := strList[0]
	if _, ok := ipWhiteList[ip]; !ok {
		limiter := ipLimiter.GetLimiter(ip)
		if !limiter.Allow() {
			logger.Debug("ip limited", zap.String("ip", ip))
			return nil, grpc.Errorf(codes.Unavailable, "request too frequently")
		}
	}

	return handler(ctx, req)
}
