package redis_template

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"go-fission-activity/activity/constant"
	"go-fission-activity/activity/web/middleware/logTracing"
	"go-fission-activity/config"
	"strings"
	"time"
)

type redisClusterTemplate struct {
	client *redis.ClusterClient
}

// 初始化redis连接池
func (receiver *redisClusterTemplate) redisClientClusterInit() {
	addrArr := strings.Split(config.ApplicationConfig.Redis.Address, ",")
	receiver.client = redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:        addrArr,
		Password:     config.ApplicationConfig.Redis.Password,
		Username:     config.ApplicationConfig.Redis.Username,
		PoolSize:     config.ApplicationConfig.Redis.MaxIdle,
		MinIdleConns: config.ApplicationConfig.Redis.MaxActive,
		IdleTimeout:  time.Second * config.ApplicationConfig.Redis.IdleTimeout,
		//ReadTimeout:  time.Minute * 1, // 设置读取操作的最大超时时间为1分钟
		//WriteTimeout: time.Minute * 1, // 设置写入操作的最大超时时间为1分钟
	})
}

func (receiver *redisClusterTemplate) Set(ctx context.Context, key string, value string) bool {
	return receiver.SetTimeout(ctx, key, value, 0)
}

func (receiver *redisClusterTemplate) SetTimeout(ctx context.Context, key string, value string, timeout time.Duration) bool {
	if constant.Empty == key {
		logTracing.LogPrintfP("key is not null")
		return false
	}
	if constant.Empty == value {
		logTracing.LogPrintfP("value is not null")
		return false
	}
	_, err := receiver.client.Set(ctx, key, value, timeout).Result()
	if err != nil {
		logTracing.LogPrintfP("redis set %s=%v fail!, err=%v", key, value, err)
		return false
	}
	return true
}

func (receiver *redisClusterTemplate) Get(ctx context.Context, key string) (string, error) {
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

func (receiver *redisClusterTemplate) MGet(ctx context.Context, key ...string) ([]interface{}, error) {
	str, err := receiver.client.MGet(ctx, key...).Result()
	return str, err
}

func (receiver *redisClusterTemplate) Del(ctx context.Context, key string) bool {
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

func (receiver *redisClusterTemplate) LPush(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	return receiver.client.LPush(ctx, key, values)
}

func (receiver *redisClusterTemplate) LRange(ctx context.Context, key string, start, stop int64) *redis.StringSliceCmd {
	return receiver.client.LRange(ctx, key, start, stop)
}

func (receiver *redisClusterTemplate) LPushX(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	return receiver.client.LPushX(ctx, key, values)
}

func (receiver *redisClusterTemplate) IncrBy(ctx context.Context, key string, values int64) *redis.IntCmd {
	return receiver.client.IncrBy(ctx, key, values)
}

func (receiver *redisClusterTemplate) RPush(ctx context.Context, key string, values string) *redis.IntCmd {
	return receiver.client.RPush(ctx, key, values)
}

func (receiver *redisClusterTemplate) LPop(ctx context.Context, key string) *redis.StringCmd {
	return receiver.client.LPop(ctx, key)
}

func (receiver *redisClusterTemplate) RPop(ctx context.Context, key string) *redis.StringCmd {
	return receiver.client.RPop(ctx, key)
}

func (receiver *redisClusterTemplate) BRPop(ctx context.Context, key string, timeout time.Duration) ([]string, error) {
	return receiver.client.BRPop(ctx, timeout, key).Result()
}

func (receiver *redisClusterTemplate) LLen(ctx context.Context, key string) *redis.IntCmd {
	return receiver.client.LLen(ctx, key)
}

func (receiver *redisClusterTemplate) LRem(ctx context.Context, key string, count int64, value interface{}) *redis.IntCmd {
	return receiver.client.LRem(ctx, key, count, value)
}

func (receiver *redisClusterTemplate) Exists(ctx context.Context, key string) (int64, error) {
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

func (receiver *redisClusterTemplate) Ping(ctx context.Context) string {
	statusCmd := receiver.client.Ping(ctx)
	logTracing.LogPrintfP("redis cluster ping status=%v", statusCmd)
	if statusCmd != nil {
		return statusCmd.Val()
	}
	return ""
}

func (receiver *redisClusterTemplate) XAdd(ctx context.Context, args *redis.XAddArgs) (string, error) {
	return receiver.client.XAdd(ctx, args).Result()
}

func (receiver *redisClusterTemplate) XReadGroup(ctx context.Context, args *redis.XReadGroupArgs) ([]redis.XStream, error) {
	return receiver.client.XReadGroup(ctx, args).Result()
}

func (receiver *redisClusterTemplate) XGroupCreateMkStream(ctx context.Context, stream, group, start string) error {
	return receiver.client.XGroupCreateMkStream(ctx, stream, group, start).Err()
}

func (receiver *redisClusterTemplate) XAck(ctx context.Context, stream, group string, ids ...string) *redis.IntCmd {
	return receiver.client.XAck(ctx, stream, group, ids...)
}

func (receiver *redisClusterTemplate) XGroupDelConsumer(ctx context.Context, stream, group, consumer string) *redis.IntCmd {
	return receiver.client.XGroupDelConsumer(ctx, stream, group, consumer)
}

func (receiver *redisClusterTemplate) HSet(ctx context.Context, key, filed string, value interface{}) *redis.IntCmd {
	return receiver.client.HSet(ctx, key, filed, value)
}

func (receiver *redisClusterTemplate) HDel(ctx context.Context, key string, fields ...string) *redis.IntCmd {
	return receiver.client.HDel(ctx, key, fields...)
}

func (receiver *redisClusterTemplate) HGetAll(ctx context.Context, key string) *redis.StringStringMapCmd {
	return receiver.client.HGetAll(ctx, key)
}

func (receiver *redisClusterTemplate) HGet(ctx context.Context, key, field string) *redis.StringCmd {
	return receiver.client.HGet(ctx, key, field)
}

func (receiver *redisClusterTemplate) HMSet(ctx context.Context, key string, fields ...interface{}) *redis.BoolCmd {
	return receiver.client.HMSet(ctx, key, fields)
}

func (receiver *redisClusterTemplate) HMGet(ctx context.Context, key string, fields ...string) *redis.SliceCmd {
	return receiver.client.HMGet(ctx, key, fields...)
}

func (receiver *redisClusterTemplate) SAdd(ctx context.Context, key string, fields ...interface{}) *redis.IntCmd {
	return receiver.client.SAdd(ctx, key, fields...)
}

func (receiver *redisClusterTemplate) SRem(ctx context.Context, key string, fields ...interface{}) *redis.IntCmd {
	return receiver.client.SRem(ctx, key, fields...)
}

func (receiver *redisClusterTemplate) SCard(ctx context.Context, key string) *redis.IntCmd {
	return receiver.client.SCard(ctx, key)
}

func (receiver *redisClusterTemplate) SetNX(ctx context.Context, key string, value string, timeout time.Duration) *redis.BoolCmd {
	return receiver.client.SetNX(ctx, key, value, timeout)
}

func (receiver *redisClusterTemplate) LIndex(ctx context.Context, key string, index int64) *redis.StringCmd {
	return receiver.client.LIndex(ctx, key, index)
}

func (receiver *redisClusterTemplate) ZAdd(ctx context.Context, key string, z *redis.Z) *redis.IntCmd {
	return receiver.client.ZAdd(ctx, key, z)
}

func (receiver *redisClusterTemplate) ZRevRangeByScore(ctx context.Context, key string, opt *redis.ZRangeBy) *redis.StringSliceCmd {
	return receiver.client.ZRevRangeByScore(ctx, key, opt)
}

func (receiver *redisClusterTemplate) ZRem(ctx context.Context, key string, members ...string) *redis.IntCmd {
	return receiver.client.ZRem(ctx, key, members)
}

func (receiver *redisClusterTemplate) Incr(ctx context.Context, key string) *redis.IntCmd {
	return receiver.client.Incr(ctx, key)
}

func (receiver *redisClusterTemplate) Keys(ctx context.Context, pattern string) ([]string, error) {
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
