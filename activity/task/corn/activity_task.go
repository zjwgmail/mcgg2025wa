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
	"go-fission-activity/activity/web/service"
	"go-fission-activity/config"
	"go-fission-activity/config/initConfig"
	"go-fission-activity/util"
	"go-fission-activity/util/txUtil"
)

func activityTask(methodName string, timeConfig config.TimerConfig) {
	ginCtx := gin.Context{}
	ctx := &ginCtx
	// defer 异常处理
	defer func() {
		if e := recover(); e != nil {
			logTracing.LogErrorPrintf(ctx, errors.New(fmt.Sprintf("方法[%s]，发生panic异常", methodName)), logTracing.ErrorLogFmt, e)
			return
		}
	}()

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
	// 查询活动信息
	activityInfoMapper := dao.GetActivityInfoMapper()
	activityInfo, err := activityInfoMapper.SelectByPrimaryKey(config.ApplicationConfig.Activity.Id)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],查询活动信息失败，活动id:%v", methodName, config.ApplicationConfig.Activity.Id))
		return
	}

	switch activityInfo.ActivityStatus {
	case constant.ATStatusUnStart:
		// 未开始
		startAt := activityInfo.StartAt
		nowTime := util.GetNowCustomTime()
		if nowTime.Time.After(startAt.Time) || nowTime.Time.Equal(startAt.Time) {
			// 到达开始时间
			update := entity.ActivityInfoEntity{
				Id:             config.ApplicationConfig.Activity.Id,
				ActivityStatus: constant.ATStatusStarted,
			}
			_, err := activityInfoMapper.UpdateByPrimaryKeySelective(&session, update)
			if err != nil {
				logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],更新活动信息失败，活动id:%v", methodName, config.ApplicationConfig.Activity.Id))
				return
			}
			logTracing.LogPrintf(ctx, logTracing.TaskHandleLogFmt, fmt.Sprintf("方法[%s],活动已启动;活动id:%v", methodName, config.ApplicationConfig.Activity.Id))
		} else {
			logTracing.LogPrintf(ctx, logTracing.TaskHandleLogFmt, fmt.Sprintf("方法[%s],活动未到启动时间;活动id:%v", methodName, config.ApplicationConfig.Activity.Id))
		}
	case constant.ATStatusStarted:
		// 判断最低的cdk库存是否小于90%，如果是则进入缓冲期
		cdkIsOver, err := service.IsUnderPercentCdkLen(context.Background(), initConfig.GetCdkLimit())
		if err != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],计算cdk是否超限报错，活动id:%v，err:%v", methodName, config.ApplicationConfig.Activity.Id, err))
			return
		}

		endAt := activityInfo.EndAt
		nowTime := util.GetNowCustomTime()
		if !config.ApplicationConfig.IsDebug && (nowTime.Time.After(endAt.Time) || nowTime.Time.Equal(endAt.Time) || cdkIsOver) {
			reallyEndTime := util.GetTimeOfAfterDays(activityInfo.EndBufferDay, nowTime)
			// 到达结束时间
			update := entity.ActivityInfoEntity{
				Id:             config.ApplicationConfig.Activity.Id,
				ActivityStatus: constant.ATStatusBuffer,
				EndBufferAt:    nowTime,
				ReallyEndAt:    reallyEndTime,
			}
			_, err := activityInfoMapper.UpdateByPrimaryKeySelective(&session, update)
			if err != nil {
				logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],更新活动信息失败，活动id:%v", methodName, config.ApplicationConfig.Activity.Id))
				return
			}
			logTracing.LogPrintf(ctx, logTracing.TaskHandleLogFmt, fmt.Sprintf("方法[%s],活动已结束;活动id:%v", methodName, config.ApplicationConfig.Activity.Id))
		} else {
			logTracing.LogPrintf(ctx, logTracing.TaskHandleLogFmt, fmt.Sprintf("方法[%s],活动未到结束时间;活动id:%v", methodName, config.ApplicationConfig.Activity.Id))
		}
	case constant.ATStatusBuffer:
		// 结束
		reallyEndAt := activityInfo.ReallyEndAt
		nowTime := util.GetNowCustomTime()
		if nowTime.Time.After(reallyEndAt.Time) || nowTime.Time.Equal(reallyEndAt.Time) {
			// 到达开始时间
			update := entity.ActivityInfoEntity{
				Id:             config.ApplicationConfig.Activity.Id,
				ActivityStatus: constant.ATStatusEnd,
			}
			_, err := activityInfoMapper.UpdateByPrimaryKeySelective(&session, update)
			if err != nil {
				logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],更新活动信息失败，活动id:%v", methodName, config.ApplicationConfig.Activity.Id))
				return
			}
			logTracing.LogPrintf(ctx, logTracing.TaskHandleLogFmt, fmt.Sprintf("方法[%s],活动结束;活动id:%v", methodName, config.ApplicationConfig.Activity.Id))
		} else {
			logTracing.LogPrintf(ctx, logTracing.TaskHandleLogFmt, fmt.Sprintf("方法[%s],活动未到结束时间;活动id:%v", methodName, config.ApplicationConfig.Activity.Id))
		}
	case constant.ATStatusEnd:
		logTracing.LogPrintf(ctx, logTracing.TaskHandleLogFmt, fmt.Sprintf("方法[%s],活动已结束;活动id:%v", methodName, config.ApplicationConfig.Activity.Id))
	}

	if !isExist {
		session.Commit()
	}
}
