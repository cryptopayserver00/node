package adapter

import "context"

// Adapter 统一扫块入口：加新链 = 实现本接口并注册，而不是改一长串 if。
type Adapter interface {
	ChainID() uint
	Name() string
	Run(ctx context.Context)
}

// Runner 按配置启动一组 adapter。
func Runner(ctx context.Context, adapters []Adapter) {
	for _, a := range adapters {
		ad := a
		go ad.Run(ctx)
	}
}
