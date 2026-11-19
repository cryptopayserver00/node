package middleware

import (
	"context"
	"fmt"
	"net/http"
	"node/global"
	"node/model/common"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// IPRateLimit 按 system.iplimit-count / iplimit-time 限制单 IP 请求数。
// 优先用 Redis；未启用 Redis 时退回进程内计数。
func IPRateLimit() gin.HandlerFunc {
	mem := newMemIPLimiter()

	return func(c *gin.Context) {
		limit := global.NODE_CONFIG.System.LimitCountIP
		windowSec := global.NODE_CONFIG.System.LimitTimeIP
		if limit <= 0 || windowSec <= 0 {
			c.Next()
			return
		}

		ip := c.ClientIP()
		allowed, err := allowIP(c.Request.Context(), mem, ip, limit, windowSec)
		if err != nil {
			global.NODE_LOG.Error("ip rate limit error: " + err.Error())
			c.Next()
			return
		}
		if !allowed {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":    common.Error_Too_Many_Requests,
				"message": "too many requests",
			})
			return
		}
		c.Next()
	}
}

func allowIP(ctx context.Context, mem *memIPLimiter, ip string, limit, windowSec int) (bool, error) {
	if global.NODE_REDIS != nil {
		key := fmt.Sprintf("node:iplimit:%s", ip)
		n, err := global.NODE_REDIS.Incr(ctx, key).Result()
		if err != nil {
			return false, err
		}
		if n == 1 {
			if err := global.NODE_REDIS.Expire(ctx, key, time.Duration(windowSec)*time.Second).Err(); err != nil {
				return false, err
			}
		}
		return n <= int64(limit), nil
	}
	return mem.allow(ip, limit, windowSec), nil
}

type memVisitor struct {
	count int
	reset time.Time
}

type memIPLimiter struct {
	mu       sync.Mutex
	visitors map[string]*memVisitor
}

func newMemIPLimiter() *memIPLimiter {
	l := &memIPLimiter{visitors: make(map[string]*memVisitor)}
	go l.cleanup()
	return l
}

func (l *memIPLimiter) allow(ip string, limit, windowSec int) bool {
	now := time.Now()
	window := time.Duration(windowSec) * time.Second

	l.mu.Lock()
	defer l.mu.Unlock()

	v, ok := l.visitors[ip]
	if !ok || now.After(v.reset) {
		l.visitors[ip] = &memVisitor{count: 1, reset: now.Add(window)}
		return true
	}
	v.count++
	return v.count <= limit
}

func (l *memIPLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		l.mu.Lock()
		for ip, v := range l.visitors {
			if now.After(v.reset) {
				delete(l.visitors, ip)
			}
		}
		l.mu.Unlock()
	}
}
