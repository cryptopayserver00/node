package redisq

import (
	"context"
	"errors"
	"fmt"
	"node/global"

	"github.com/redis/go-redis/v9"
)

func setKey(listKey string) string {
	return listKey + ":set"
}

// EnqueueUnique 仅在 value 未在去重集合中时入队。返回 true 表示新入队。
func EnqueueUnique(ctx context.Context, listKey, value string) (bool, error) {
	if global.NODE_REDIS == nil {
		return false, fmt.Errorf("redis not initialized")
	}
	n, err := global.NODE_REDIS.SAdd(ctx, setKey(listKey), value).Result()
	if err != nil {
		return false, err
	}
	if n == 0 {
		return false, nil
	}
	if err := global.NODE_REDIS.RPush(ctx, listKey, value).Err(); err != nil {
		_, _ = global.NODE_REDIS.SRem(ctx, setKey(listKey), value).Result()
		return false, err
	}
	return true, nil
}

// Enqueue 无条件入队（用于失败块高度等可接受重复的场景也可配合 EnqueueUnique）。
func Enqueue(ctx context.Context, listKey string, value any) error {
	if global.NODE_REDIS == nil {
		return fmt.Errorf("redis not initialized")
	}
	return global.NODE_REDIS.RPush(ctx, listKey, value).Err()
}

// Pop 原子弹出队首；队列空时返回 redis.Nil。
func Pop(ctx context.Context, listKey string) (string, error) {
	if global.NODE_REDIS == nil {
		return "", fmt.Errorf("redis not initialized")
	}
	return global.NODE_REDIS.LPop(ctx, listKey).Result()
}

// Ack 处理成功或放弃后，从去重集合移除。
func Ack(ctx context.Context, listKey, value string) error {
	if global.NODE_REDIS == nil {
		return fmt.Errorf("redis not initialized")
	}
	return global.NODE_REDIS.SRem(ctx, setKey(listKey), value).Err()
}

// Requeue 处理失败后重新入队（去重集合中已有该值）。
func Requeue(ctx context.Context, listKey, value string) error {
	if global.NODE_REDIS == nil {
		return fmt.Errorf("redis not initialized")
	}
	return global.NODE_REDIS.RPush(ctx, listKey, value).Err()
}

// IsEmpty 判断是否为空队列错误。
func IsEmpty(err error) bool {
	return errors.Is(err, redis.Nil)
}
