package cron_task

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"go-fission-activity/activity/constant"
	"go-fission-activity/activity/model/entity"
	"go-fission-activity/activity/third/redis_template"
	"go-fission-activity/activity/web/dao"
	"go-fission-activity/activity/web/middleware/logTracing"
	"go-fission-activity/config"
	"go-fission-activity/util/txUtil"
)

func costTask(methodName string, timeConfig config.TimerConfig) {
	ginCtx := gin.Context{}
	ctx := &ginCtx
	// defer 异常处理
	defer func() {
		if e := recover(); e != nil {
			logTracing.LogErrorPrintf(ctx, errors.New(fmt.Sprintf("方法[%s]，发生panic异常", methodName)), logTracing.ErrorLogFmt, e)
			return
		}
	}()
	logTracing.LogPrintf(ctx, logTracing.TaskHandleLogFmt, fmt.Sprintf("方法[%s],发送costTask任务开始执行", methodName))

	// 查询活动信息
	activityInfoMapper := dao.GetActivityInfoMapper()
	activityInfo, err := activityInfoMapper.SelectByPrimaryKey(config.ApplicationConfig.Activity.Id)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],查询活动信息失败，活动id:%v", methodName, config.ApplicationConfig.Activity.Id))
		return
	}
	if activityInfo.ActivityStatus == constant.ATStatusUnStart || activityInfo.ActivityStatus == constant.ATStatusEnd {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],活动不在运行期，活动id:%v", methodName, config.ApplicationConfig.Activity.Id))
		return
	}

	template := redis_template.NewRedisTemplate()
	taskLockKey := constant.GetTaskLockKey(config.ApplicationConfig.Activity.Id, methodName)

	getLock, err := template.SetNX(context.Background(), taskLockKey, "1", lockTimeout).Result()
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],调用redis nx失败，本实例不执行任务", methodName))
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

	session, isExist, err := txUtil.GetTransaction(ctx)
	if nil != err {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，创建事务失败,err：%v", methodName, err))
		return
	}
	if !isExist {
		defer func() {
			session.Rollback()
			session.Close()
		}()
	}
	// 查询信息
	msgInfoMapper := dao.GetMsgInfoMapperV2()
	priceCount, err := msgInfoMapper.SumPriceSendUnCountMsg(&session, constant.UnCounted)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],统计消息花费失败，活动id:%v", methodName, config.ApplicationConfig.Activity.Id))
		return
	}

	costCountInfoMapper := dao.GetCostCountInfoMapper()
	costCountInfo, err := costCountInfoMapper.SelectByPrimaryKey(config.ApplicationConfig.Activity.Id)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],查询花费统计表失败，活动id:%v", methodName, config.ApplicationConfig.Activity.Id))
		return
	}

	allCount := priceCount
	if costCountInfo.Id <= 0 {
		// 新增
		costCountInfo = entity.CostCountInfoEntity{
			Id:        config.ApplicationConfig.Activity.Id,
			CostCount: priceCount,
		}
		_, err := costCountInfoMapper.InsertSelective(&session, costCountInfo)
		if err != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],新增花费统计表失败，活动id:%v", methodName, config.ApplicationConfig.Activity.Id))
			return
		}
	} else {
		// 更新
		allCount = priceCount + costCountInfo.CostCount
		update := entity.CostCountInfoEntity{
			Id:        config.ApplicationConfig.Activity.Id,
			CostCount: allCount,
		}
		_, err := costCountInfoMapper.UpdateByPrimaryKeySelective(&session, update)
		if err != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],新增花费统计表失败，活动id:%v", methodName, config.ApplicationConfig.Activity.Id))
			return
		}
	}

	// 更新消息表
	_, err = msgInfoMapper.UpdateCountOfSendUnCount(&session, constant.UnCounted, constant.Counted)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],更新消息表失败，活动id:%v", methodName, config.ApplicationConfig.Activity.Id))
		return
	}

	if !isExist {
		session.Commit()
	}
	logTracing.LogPrintf(ctx, logTracing.TaskHandleLogFmt, fmt.Sprintf("方法[%s],发送costTask任务执行完成,allCount:%v", methodName, allCount))

}
