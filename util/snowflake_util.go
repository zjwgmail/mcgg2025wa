package util

import (
	"context"
	"fmt"
	"github.com/bwmarrin/snowflake"
	"go-fission-activity/activity/constant"
	"go-fission-activity/activity/third/redis_template"
	"go-fission-activity/activity/web/middleware/logTracing"
	"sync"
)

var snowFlakeOnce sync.Once
var snowFlakeNode *snowflake.Node

// NewRedisTemplate 初始化redis连接
func getSnowFlakeNode(ctx context.Context) *snowflake.Node {
	if snowFlakeNode != nil {
		return snowFlakeNode
	}
	snowFlakeOnce.Do(func() {
		template := redis_template.NewRedisTemplate()

		serviceId, err := template.IncrBy(context.Background(), constant.ServiceIdKey, 1).Result()
		if err != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("getSnowFlakeNode,获取服务id失败"))
			return
		}
		// 初始化雪花 ID 生成器
		snowFlakeNode, _ = snowflake.NewNode(serviceId) //节点 ID，可以根据需要设置

		if serviceId > 888 {
			del := template.Del(context.Background(), constant.ServiceIdKey)
			if !del {
				logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("getSnowFlakeNode,删除服务id key失败"))
				return
			}
		}
	})
	return snowFlakeNode
}

func GetSnowFlakeId(ctx context.Context) int64 {
	node := getSnowFlakeNode(ctx)
	// 生成雪花 ID
	id := node.Generate()
	return id.Int64()
}

func GetSnowFlakeIdStr(ctx context.Context) string {
	node := getSnowFlakeNode(ctx)
	// 生成雪花 ID
	id := node.Generate()
	return id.String()
}
