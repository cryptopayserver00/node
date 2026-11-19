package constant

import (
	"sync"
	"time"
)

const (
	rpcFailCooldown = 30 * time.Second
	rpcMaxFails     = 3
)

type rpcEndpoint struct {
	url       string
	fails     int
	cooldownUntil time.Time
}

type rpcPool struct {
	mu   sync.Mutex
	list []*rpcEndpoint
	next int
}

var (
	rpcPoolsMu sync.Mutex
	rpcPools   = map[uint]*rpcPool{}
)

func getRPCPool(chainId uint) *rpcPool {
	rpcs := GetAllRPCUrlByChainId(chainId)
	rpcPoolsMu.Lock()
	defer rpcPoolsMu.Unlock()

	pool, ok := rpcPools[chainId]
	if !ok {
		pool = &rpcPool{}
		rpcPools[chainId] = pool
	}

	// 配置变更时重建
	if len(pool.list) != len(rpcs) || !pool.sameURLs(rpcs) {
		list := make([]*rpcEndpoint, 0, len(rpcs))
		for _, u := range rpcs {
			list = append(list, &rpcEndpoint{url: u})
		}
		pool.list = list
		pool.next = 0
	}
	return pool
}

func (p *rpcPool) sameURLs(rpcs []string) bool {
	if len(p.list) != len(rpcs) {
		return false
	}
	for i := range rpcs {
		if p.list[i].url != rpcs[i] {
			return false
		}
	}
	return true
}

func (p *rpcPool) pick() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.list) == 0 {
		return ""
	}

	now := time.Now()
	n := len(p.list)
	for i := 0; i < n; i++ {
		idx := (p.next + i) % n
		ep := p.list[idx]
		if now.Before(ep.cooldownUntil) {
			continue
		}
		p.next = (idx + 1) % n
		return ep.url
	}
	// 全部冷却：仍返回下一个，避免完全卡死
	ep := p.list[p.next%n]
	p.next = (p.next + 1) % n
	return ep.url
}

func (p *rpcPool) report(url string, success bool) {
	if url == "" {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, ep := range p.list {
		if ep.url != url {
			continue
		}
		if success {
			ep.fails = 0
			ep.cooldownUntil = time.Time{}
			return
		}
		ep.fails++
		if ep.fails >= rpcMaxFails {
			ep.cooldownUntil = time.Now().Add(rpcFailCooldown)
			ep.fails = 0
		}
		return
	}
}

// ReportRPCSuccess 标记某 RPC 正常。
func ReportRPCSuccess(chainId uint, url string) {
	getRPCPool(chainId).report(url, true)
}

// ReportRPCFailure 标记某 RPC 失败，连续失败后进入冷却。
func ReportRPCFailure(chainId uint, url string) {
	getRPCPool(chainId).report(url, false)
}
