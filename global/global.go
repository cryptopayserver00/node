package global

import (
	"node/config"
	"sync"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	NODE_DB       *gorm.DB
	NODE_DBList   map[string]*gorm.DB
	NODE_REDIS    *redis.Client
	NODE_MEMCACHE *memcache.Client
	NODE_CONFIG   config.Server
	NODE_VP       *viper.Viper
	NODE_LOG      *zap.Logger
	// NODE_Timer time.Timer = timer

	NODE_MUTEX sync.Mutex
)
