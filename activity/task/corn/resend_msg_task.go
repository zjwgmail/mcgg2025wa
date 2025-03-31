package cron_task

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"go-fission-activity/activity/constant"
	"go-fission-activity/activity/model/dto"
	"go-fission-activity/activity/third/redis_template"
	"go-fission-activity/activity/web/dao"
	"go-fission-activity/activity/web/middleware/logTracing"
	"go-fission-activity/activity/web/service"
	"go-fission-activity/config"
	"go-fission-activity/util"
	"go-fission-activity/util/goroutine_pool"
)

var resendGoroutinePool = goroutine_pool.NewGoroutinePool(4)

func resendMsgTask(methodName string, timeConfig config.TimerConfig) {
	ginCtx := gin.Context{}
	ctx := &ginCtx
	// defer 异常处理
	defer func() {
		if e := recover(); e != nil {
			logTracing.LogErrorPrintf(ctx, errors.New(fmt.Sprintf("方法[%s]，发生panic异常", methodName)), logTracing.ErrorLogFmt, e)
			return
		}
	}()

	nowCustomTime := util.GetNowCustomTime()

	hour := nowCustomTime.Time.Hour()
	//if hour != 11 && !config.ApplicationConfig.IsDebug {
	//	logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],任务执行时间为：%v，当前时间：%v,不是任务执行时间跳过。", methodName, hour, nowCustomTime.Time))
	//	return
	//}
	if !(hour >= 10 && hour <= 21) {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],任务执行时间为：%v，当前时间：%v,不是任务执行时间跳过。", methodName, hour, nowCustomTime.Time))
		return
	}

	isDisturbTime := nowCustomTime.IsNotDisturbTime()
	if isDisturbTime {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],免打扰时间，不执行任务", methodName))
		return
	}

	template := redis_template.NewRedisTemplate()
	taskLockKey := constant.GetTaskLockKey(config.ApplicationConfig.Activity.Id, methodName)

	getLock, err := template.SetNX(context.Background(), taskLockKey, "1", lockTimeout).Result()
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],调用redis nx报错，本实例不执行任务，err:%v", methodName, err))
		return
	}
	if !getLock {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],获取分布式锁失败，本实例不执行任务", methodName))
		return
	}
	defer func() {
		del := template.Del(context.Background(), taskLockKey)
		if !del {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，删除分布式锁失败", methodName))
		}
	}()

	// 查询活动信息
	activityInfoMapper := dao.GetActivityInfoMapper()
	activityInfo, err := activityInfoMapper.SelectByPrimaryKey(config.ApplicationConfig.Activity.Id)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],查询活动信息失败，活动id:%v，err:%v", methodName, config.ApplicationConfig.Activity.Id, err))
		return
	}
	if activityInfo.ActivityStatus == constant.ATStatusUnStart || activityInfo.ActivityStatus == constant.ATStatusEnd {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],活动不在运行期，活动id:%v", methodName, config.ApplicationConfig.Activity.Id))
		return
	}

	//isUltraLimit, err := service.CostIsUltraLimit(ctx)
	//if err != nil {
	//	logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，查询费用是否超额,err：%v", methodName, err))
	//	return
	//}

	// 查询所有需要发送的消息，按照发送到的waId排序
	msgMapper := dao.GetMsgInfoMapperV2()
	// 分页参数
	//page := 1      // 当前页码
	//pageSize := 10 // 每页消息数量
	limit := 10
	minId := "0"

	// 循环直到所有消息都被处理
	for {
		//offset := (page - 1) * pageSize
		// 调用支持分页的方法获取未发送消息的waId列表
		msgOfWaIdList, err := msgMapper.SelectWaIdListOfUnSendMsg(
			minId,
			limit,
		)
		if err != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],查询未发送消息的waId列表失败，err:%v", methodName, err))
			break // 如果有错误，跳出循环
		}

		if len(msgOfWaIdList) == 0 {
			break // 如果没有更多的消息，跳出循环
		}

		// 处理当前页的消息
		for _, msgOfWaId := range msgOfWaIdList {
			minId = msgOfWaId.WaId
			if msgOfWaId.WaId == "" {
				continue
			}
			ctx2 := &gin.Context{}
			// 查询是否是免费期
			isFree, err := service.CheckIsFreeByWaId(ctx2, msgOfWaId.WaId)
			if err != nil {
				logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，根据waId查询是否是免费期报错,活动id:%v,waId:%v,err：%v", methodName, config.ApplicationConfig.Activity.Id, msgOfWaId.WaId, err))
				continue
			}
			ctx3 := &gin.Context{}
			resendGoroutinePool.Execute(func(param interface{}) {
				u, ok := param.(dto.MsgOfWaIdDto) // 断言u是User类型
				if !ok {
					logTracing.LogPrintf(ctx, logTracing.TaskHandleLogFmt, fmt.Sprintf("方法[%s],断言发生错误，waId:%v", methodName, u.WaId))
				}
				logTracing.LogPrintf(ctx, logTracing.TaskHandleLogFmt, fmt.Sprintf("方法[%s],resendGoroutinePool协程池执行任务开始，waId:%v", methodName, u.WaId))
				service.ReSendMsgByWaId(ctx3, u.WaId, isFree)
				logTracing.LogPrintf(ctx, logTracing.TaskHandleLogFmt, fmt.Sprintf("方法[%s],resendGoroutinePool协程池执行任务结束，waId:%v", methodName, u.WaId))
			}, msgOfWaId)
		}

		// 等待当前页的消息处理完成
		resendGoroutinePool.Wait()

		// 准备下一页
		//page++
	}
}
