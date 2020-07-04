package api

import (
	"sync"
	"time"

	"github.com/bluele/gcache"
	"golang.org/x/time/rate"
)

const cacheSzie = 1000

type IPRateLimiter struct {
	cache gcache.Cache
	//ips map[string]*rate.Limiter
	mu *sync.RWMutex
	r  rate.Limit
	b  int
}

// NewIPRateLimiter 创建ip限制器
func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	return &IPRateLimiter{
		cache: gcache.New(cacheSzie).LRU().Build(),
		//ips: make(map[string]*rate.Limiter),
		mu: &sync.RWMutex{},
		r:  r,
		b:  b,
	}
}

// AddIP 添加ip
func (i *IPRateLimiter) AddIP(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	limiter := rate.NewLimiter(i.r, i.b)
	i.cache.SetWithExpire(ip, limiter, 24*time.Hour)
	//i.ips[ip] = limiter
	return limiter
}

// GetLimiter 获取限制器
func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.Lock()

	limiter, err := i.cache.Get(ip)
	if err != nil {
		i.mu.Unlock()
		return i.AddIP(ip)
	}
	// limiter, ok := i.ips[ip]
	// if !ok {
	// 	i.mu.Unlock()
	// 	return i.AddIP(ip)
	// }
	i.mu.Unlock()
	return limiter.(*rate.Limiter)
}
