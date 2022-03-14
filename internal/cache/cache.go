package cache

import (
	"time"

	"github.com/patrickmn/go-cache"
)

var memCache *cache.Cache

func init() {
	// 创建一个默认过期时间为1分钟，清理间隔时间为2分钟的高速缓存
	memCache = cache.New(1*time.Minute, 2*time.Minute)
}

func Set(k string) {
	memCache.Set(k, k, cache.DefaultExpiration)
}

func Check(key string) bool {
	_, ok := memCache.Get(key)
	return ok
}
