package redis_template

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"go-fission-activity/activity/constant"
	"go-fission-activity/activity/web/middleware/logTracing"
	"go-fission-activity/config"
	"time"
)

type redisSingleTemplate struct {
	client *redis.Client
}

// 初始化redis连接池
func (receiver *redisSingleTemplate) redisClientInit() {
	receiver.client = redis.NewClient(&redis.Options{
		Addr:         config.ApplicationConfig.Redis.Address,
		Password:     config.ApplicationConfig.Redis.Password,
		DB:           config.ApplicationConfig.Redis.Database,
		PoolSize:     config.ApplicationConfig.Redis.MaxIdle,
		MinIdleConns: config.ApplicationConfig.Redis.MaxActive,
		IdleTimeout:  time.Second * config.ApplicationConfig.Redis.IdleTimeout,
		Username:     config.ApplicationConfig.Redis.Username,
	})
}

func (receiver *redisSingleTemplate) Set(ctx context.Context, key string, value string) bool {
	return receiver.SetTimeout(ctx, key, value, 0)
}

func (receiver *redisSingleTemplate) SetTimeout(ctx context.Context, key string, value string, timeout time.Duration) bool {
	if constant.Empty == key {
		logTracing.LogPrintfP("SetTimeout key is not null")
		return false
	}
	if constant.Empty == value {
		logTracing.LogPrintfP("SetTimeout value is not null")
		return false
	}
	_, err := receiver.client.Set(ctx, key, value, timeout).Result()
	if err != nil {
		logTracing.LogPrintfP("SetTimeout redis set %s=%v fail!, err=%v", key, value, err)
		return false
	}
	return true
}

func (receiver *redisSingleTemplate) Get(ctx context.Context, key string) (string, error) {
	if "" == key {
		logTracing.LogPrintfP("key is not null")
		return "", errors.New("key is not null")
	}
	str, err := receiver.client.Get(ctx, key).Result()
	if err != nil {
		logTracing.LogPrintfP("redis get %s fail!, err=%v", key, err)
		return "", errors.New(fmt.Sprintf("redis get %s fail!, err=%v", key, err))
	}
	return str, nil
}

func (receiver *redisSingleTemplate) MGet(ctx context.Context, key ...string) ([]interface{}, error) {
	str, err := receiver.client.MGet(ctx, key...).Result()
	return str, err
}

func (receiver *redisSingleTemplate) Del(ctx context.Context, key string) bool {
	if "" == key {
		logTracing.LogPrintfP("key is not null")
		return false
	}
	_, err := receiver.client.Del(ctx, key).Result()
	if err != nil {
		logTracing.LogPrintfP("redis del %s fail!, err=%v", key, err)
		return false
	}
	return true
}

func (receiver *redisSingleTemplate) LPush(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	return receiver.client.LPush(ctx, key, values)
}

func (receiver *redisSingleTemplate) LRange(ctx context.Context, key string, start, stop int64) *redis.StringSliceCmd {
	return receiver.client.LRange(ctx, key, start, stop)
}

func (receiver *redisSingleTemplate) LPushX(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	return receiver.client.LPushX(ctx, key, values)
}

func (receiver *redisSingleTemplate) IncrBy(ctx context.Context, key string, values int64) *redis.IntCmd {
	return receiver.client.IncrBy(ctx, key, values)
}

func (receiver *redisSingleTemplate) RPush(ctx context.Context, key string, values string) *redis.IntCmd {
	return receiver.client.RPush(ctx, key, values)
}

func (receiver *redisSingleTemplate) LPop(ctx context.Context, key string) *redis.StringCmd {
	return receiver.client.LPop(ctx, key)
}

func (receiver *redisSingleTemplate) RPop(ctx context.Context, key string) *redis.StringCmd {
	return receiver.client.RPop(ctx, key)
}

func (receiver *redisSingleTemplate) BRPop(ctx context.Context, key string, timeout time.Duration) ([]string, error) {
	return receiver.client.BRPop(ctx, timeout, key).Result()
}

func (receiver *redisSingleTemplate) LLen(ctx context.Context, key string) *redis.IntCmd {
	return receiver.client.LLen(ctx, key)
}

func (receiver *redisSingleTemplate) LRem(ctx context.Context, key string, count int64, value interface{}) *redis.IntCmd {
	return receiver.client.LRem(ctx, key, count, value)
}

func (receiver *redisSingleTemplate) Exists(ctx context.Context, key string) (int64, error) {
	if "" == key {
		logTracing.LogPrintfP("key is not null")
		return 0, errors.New("key is not null")
	}
	resultInt, err := receiver.client.Exists(ctx, key).Result()
	if err != nil {
		logTracing.LogPrintfP("redis get %s fail!, err=%v", key, err)
		return 0, errors.New(fmt.Sprintf("redis get %s fail!, err=%v", key, err))
	}
	return resultInt, nil
}

func (receiver *redisSingleTemplate) Ping(ctx context.Context) string {
	statusCmd := receiver.client.Ping(ctx)
	logTracing.LogPrintfP("redis single ping status=%v", statusCmd)
	if statusCmd != nil {
		return statusCmd.Val()
	}
	return ""
}

func (receiver *redisSingleTemplate) XAdd(ctx context.Context, args *redis.XAddArgs) (string, error) {
	return receiver.client.XAdd(ctx, args).Result()
}

func (receiver *redisSingleTemplate) XReadGroup(ctx context.Context, args *redis.XReadGroupArgs) ([]redis.XStream, error) {
	return receiver.client.XReadGroup(ctx, args).Result()
}

func (receiver *redisSingleTemplate) XGroupCreateMkStream(ctx context.Context, stream, group, start string) error {
	return receiver.client.XGroupCreateMkStream(ctx, stream, group, start).Err()
}

func (receiver *redisSingleTemplate) XAck(ctx context.Context, stream, group string, ids ...string) *redis.IntCmd {
	return receiver.client.XAck(ctx, stream, group, ids...)
}

func (receiver *redisSingleTemplate) XGroupDelConsumer(ctx context.Context, stream, group, consumer string) *redis.IntCmd {
	return receiver.client.XGroupDelConsumer(ctx, stream, group, consumer)
}

func (receiver *redisSingleTemplate) HSet(ctx context.Context, key, filed string, value interface{}) *redis.IntCmd {
	return receiver.client.HSet(ctx, key, filed, value)
}

func (receiver *redisSingleTemplate) HDel(ctx context.Context, key string, fields ...string) *redis.IntCmd {
	return receiver.client.HDel(ctx, key, fields...)
}

func (receiver *redisSingleTemplate) HGetAll(ctx context.Context, key string) *redis.StringStringMapCmd {
	return receiver.client.HGetAll(ctx, key)
}

func (receiver *redisSingleTemplate) HGet(ctx context.Context, key, field string) *redis.StringCmd {
	return receiver.client.HGet(ctx, key, field)
}

func (receiver *redisSingleTemplate) HMSet(ctx context.Context, key string, fields ...interface{}) *redis.BoolCmd {
	return receiver.client.HMSet(ctx, key, fields)
}

func (receiver *redisSingleTemplate) HMGet(ctx context.Context, key string, fields ...string) *redis.SliceCmd {
	return receiver.client.HMGet(ctx, key, fields...)
}

func (receiver *redisSingleTemplate) SAdd(ctx context.Context, key string, fields ...interface{}) *redis.IntCmd {
	return receiver.client.SAdd(ctx, key, fields...)
}

func (receiver *redisSingleTemplate) SRem(ctx context.Context, key string, fields ...interface{}) *redis.IntCmd {
	return receiver.client.SRem(ctx, key, fields...)
}

func (receiver *redisSingleTemplate) SCard(ctx context.Context, key string) *redis.IntCmd {
	return receiver.client.SCard(ctx, key)
}

func (receiver *redisSingleTemplate) SetNX(ctx context.Context, key string, value string, timeout time.Duration) *redis.BoolCmd {
	return receiver.client.SetNX(ctx, key, value, timeout)
}

func (receiver *redisSingleTemplate) LIndex(ctx context.Context, key string, index int64) *redis.StringCmd {
	return receiver.client.LIndex(ctx, key, index)
}

func (receiver *redisSingleTemplate) ZAdd(ctx context.Context, key string, z *redis.Z) *redis.IntCmd {
	return receiver.client.ZAdd(ctx, key, z)
}

func (receiver *redisSingleTemplate) ZRevRangeByScore(ctx context.Context, key string, opt *redis.ZRangeBy) *redis.StringSliceCmd {
	return receiver.client.ZRevRangeByScore(ctx, key, opt)
}

func (receiver *redisSingleTemplate) ZRem(ctx context.Context, key string, members ...string) *redis.IntCmd {
	return receiver.client.ZRem(ctx, key, members)
}

func (receiver *redisSingleTemplate) Incr(ctx context.Context, key string) *redis.IntCmd {
	return receiver.client.Incr(ctx, key)
}

func (receiver *redisSingleTemplate) Keys(ctx context.Context, pattern string) ([]string, error) {
	// 使用SCAN命令迭代Redis数据库中的key
	var cursor uint64
	var keys []string
	for {
		var k []string
		var err error
		k, cursor, err = receiver.client.Scan(ctx, cursor, pattern, 0).Result()
		if err != nil {
			return nil, err
		}
		keys = append(keys, k...)
		if cursor == 0 {
			break
		}
	}
	return keys, nil
}
