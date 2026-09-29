package utils

import (
	"crypto/sha256"
	"sync"
	"time"
)

type dedupCache struct {
	mu   sync.Mutex
	ttl  time.Duration
	seen map[[32]byte]time.Time
}

func newDedup(ttl time.Duration) *dedupCache {
	d := &dedupCache{ttl: ttl, seen: make(map[[32]byte]time.Time)}
	go func() { // 定期清理过期项,防止 map 无限增长
		for range time.Tick(ttl) {
			d.mu.Lock()
			now := time.Now()
			for k, t := range d.seen {
				if now.Sub(t) > ttl {
					delete(d.seen, k)
				}
			}
			d.mu.Unlock()
		}
	}()
	return d
}

func (d *dedupCache) Allow(msg string) bool {
	key := sha256.Sum256([]byte(msg))
	d.mu.Lock()
	defer d.mu.Unlock()
	if t, ok := d.seen[key]; ok && time.Since(t) < d.ttl {
		return false
	}
	d.seen[key] = time.Now()
	return true
}
