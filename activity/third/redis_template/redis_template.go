package redis_template

import (
	"context"
	"github.com/go-redis/redis/v8"
	"go-fission-activity/config"
	"log"
	"strings"
	"sync"
	"time"
)

var once sync.Once
var redisTemplate RedisTemplate

// NewRedisTemplate 初始化redis连接
func NewRedisTemplate() RedisTemplate {
	if redisTemplate != nil {
		return redisTemplate
	}
	once.Do(func() {
		contains := strings.Contains(config.ApplicationConfig.Redis.Address, ",")
		log.Printf("初始化redis连接%s,是否是分布式：%v", config.ApplicationConfig.Redis.Address, contains)
		if contains {
			template := redisClusterTemplate{}
			template.redisClientClusterInit()
			redisTemplate = &template
		} else {
			template := redisSingleTemplate{}
			template.redisClientInit()
			redisTemplate = &template
		}
		redisTemplate.Ping(context.Background())
	})
	return redisTemplate
}

type RedisTemplate interface {
	Set(ctx context.Context, key string, value string) bool
	SetTimeout(ctx context.Context, key string, value string, timeout time.Duration) bool

	Get(ctx context.Context, key string) (string, error)

	MGet(ctx context.Context, key ...string) ([]interface{}, error)

	Del(ctx context.Context, key string) bool

	LPush(ctx context.Context, key string, values ...interface{}) *redis.IntCmd

	LRange(ctx context.Context, key string, start, stop int64) *redis.StringSliceCmd

	LPushX(ctx context.Context, key string, values ...interface{}) *redis.IntCmd

	IncrBy(ctx context.Context, key string, values int64) *redis.IntCmd

	RPush(ctx context.Context, key string, values string) *redis.IntCmd

	LPop(ctx context.Context, key string) *redis.StringCmd

	RPop(ctx context.Context, key string) *redis.StringCmd

	BRPop(ctx context.Context, key string, timeout time.Duration) ([]string, error)

	LLen(ctx context.Context, key string) *redis.IntCmd

	LRem(ctx context.Context, key string, count int64, value interface{}) *redis.IntCmd

	// Exists 校验key是否存在，若键不存在则返回0
	Exists(ctx context.Context, key string) (int64, error)

	Ping(ctx context.Context) string

	XAdd(ctx context.Context, args *redis.XAddArgs) (string, error)

	XReadGroup(ctx context.Context, args *redis.XReadGroupArgs) ([]redis.XStream, error)

	XGroupCreateMkStream(ctx context.Context, stream, group, start string) error

	XAck(ctx context.Context, stream, group string, ids ...string) *redis.IntCmd

	XGroupDelConsumer(ctx context.Context, stream, group, consumer string) *redis.IntCmd

	HSet(ctx context.Context, key, filed string, value interface{}) *redis.IntCmd

	HDel(ctx context.Context, key string, fields ...string) *redis.IntCmd

	HGetAll(ctx context.Context, key string) *redis.StringStringMapCmd

	HGet(ctx context.Context, key, field string) *redis.StringCmd

	HMSet(ctx context.Context, key string, fields ...interface{}) *redis.BoolCmd

	HMGet(ctx context.Context, key string, fields ...string) *redis.SliceCmd

	SAdd(ctx context.Context, key string, fields ...interface{}) *redis.IntCmd

	SRem(ctx context.Context, key string, fields ...interface{}) *redis.IntCmd

	SCard(ctx context.Context, key string) *redis.IntCmd

	SetNX(ctx context.Context, key string, value string, timeout time.Duration) *redis.BoolCmd

	LIndex(ctx context.Context, key string, index int64) *redis.StringCmd

	ZAdd(ctx context.Context, key string, z *redis.Z) *redis.IntCmd

	ZRevRangeByScore(ctx context.Context, key string, opt *redis.ZRangeBy) *redis.StringSliceCmd

	ZRem(ctx context.Context, key string, members ...string) *redis.IntCmd

	Incr(ctx context.Context, key string) *redis.IntCmd

	Keys(ctx context.Context, pattern string) ([]string, error)
}
