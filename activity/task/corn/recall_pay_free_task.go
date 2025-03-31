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
	"go-fission-activity/util"
	"go-fission-activity/util/goroutine_pool"
	"go-fission-activity/util/txUtil"
	"time"
)

var recallPayFreeGoroutinePool = goroutine_pool.NewGoroutinePool(3)

func recallPayFreeMsgTask(methodName string, timeConfig config.TimerConfig) {
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
	if (hour < 11 || hour > 20) && !config.ApplicationConfig.IsDebug {
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

	getLock, err := template.SetNX(context.Background(), taskLockKey, "1", time.Hour*2).Result()
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],调用redis nx失败，本实例不执行任务，err:%v", methodName, err))
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

	logTracing.LogPrintf(ctx, logTracing.TaskHandleLogFmt, fmt.Sprintf("方法[%s],开始查询活动信息", methodName))

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

	logTracing.LogPrintf(ctx, logTracing.TaskHandleLogFmt, fmt.Sprintf("方法[%s],开始查询付费消息总人数", methodName))

	//msgInfoMapper := dao.GetMsgInfoMapperV2()
	//// 第二天中午11点，统计对距离上次互动超过36小时的用户，集中推送续订活动消息。
	//userAttendInfoMapper := dao.GetUserAttendInfoMapperV2()
	//currentTimestamp := time.Now().Unix()
	//diffHourTimestamp := currentTimestamp - int64(initConfig.GetReFreeNextHour()*3600)
	//renewFreeUser1stId, err := userAttendInfoMapper.Select1stIdPayRenewFree(constant.RenewFreeUnSend, diffHourTimestamp)
	//if err != nil {
	//	logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s [付费消息,未续免费用户]],统计总数失败，err:%v", methodName, err))
	//	return
	//}
	//logTracing.LogPrintf(ctx, logTracing.TaskHandleLogFmt, fmt.Sprintf("方法[%s],结束查询付费消息总人数，count:%v", methodName, renewFreeUser1stId))
	//
	//if renewFreeUser1stId > 0 {
	//	lastId := 0
	//	for {
	//		//防止发送后费用增加
	//		priceSum, err := msgInfoMapper.SumSendPriceMsg()
	//		if err != nil {
	//			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],统计消息花费失败，活动id:%v,err:%v", methodName, config.ApplicationConfig.Activity.Id, err))
	//			return
	//		}
	//
	//		if priceSum >= activityInfo.CostMax {
	//			logTracing.LogPrintf(ctx, logTracing.TaskHandleLogFmt, fmt.Sprintf("方法[%s],费用已用尽，不执行付费续订消息，priceCount:%v", methodName, priceSum))
	//			return
	//		}
	//
	//		logTracing.LogPrintf(ctx, logTracing.TaskHandleLogFmt, fmt.Sprintf("方法[%s],开始查询付费消息数据，lastId:%v", methodName, lastId))
	//
	//		userList, err := userAttendInfoMapper.SelectPayRenewFree(lastId, 200, constant.RenewFreeUnSend, diffHourTimestamp)
	//		if err != nil {
	//			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s [付费消息,未续免费用户]],查询失败，err:%v", methodName, err))
	//			break
	//		}
	//
	//		logTracing.LogPrintf(ctx, logTracing.TaskHandleLogFmt, fmt.Sprintf("方法[%s],结束查询付费消息数据，lastId:%v，本次循环：%v", methodName, lastId, len(userList)))
	//
	//		if len(userList) <= 0 {
	//			break
	//		}
	//		lastId = userList[len(userList)-1].Id
	//
	//		for _, user := range userList {
	//			recallPayFreeGoroutinePool.Execute(func(param interface{}) {
	//				u, ok := param.(entity.UserAttendInfoEntityV2) // 断言u是User类型
	//				if !ok {
	//					logTracing.LogPrintf(ctx, logTracing.TaskHandleLogFmt, fmt.Sprintf("方法[%s],断言发生错误，waId:%v", methodName, u.WaId))
	//				}
	//				logTracing.LogPrintf(ctx, logTracing.TaskHandleLogFmt, fmt.Sprintf("方法[%s],recallPayFreeGoroutinePool协程池执行任务开始，waId:%v", methodName, u.WaId))
	//				handlerPayRenewFreeUserInfo(ctx, methodName, u)
	//				logTracing.LogPrintf(ctx, logTracing.TaskHandleLogFmt, fmt.Sprintf("方法[%s],recallPayFreeGoroutinePool协程池执行任务结束，waId:%v", methodName, u.WaId))
	//			}, user)
	//		}
	//		recallPayFreeGoroutinePool.Wait()
	//		if config.ApplicationConfig.IsDebug {
	//			time.Sleep(1 * time.Minute)
	//		}
	//	}
	//}

}

func handlerPayRenewFreeUserInfo(ctx *gin.Context, methodName string, user entity.UserAttendInfoEntityV2) {
	methodName = methodName + " [付费消息,未续免费用户]"

	//waId := user.WaId
	// redis锁
	//template := redis_template.NewRedisTemplate()
	//res, err := template.SetNX(context.Background(), constant.GetUserLockKey(config.ApplicationConfig.Activity.Id, waId), "1", constant.LockTimeOut).Result()
	//if err != nil {
	//	logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，获取分布式锁报错,活动id:%v,waId:%v,err：%v", methodName, config.ApplicationConfig.Activity.Id, waId, err))
	//	return
	//}
	//if !res {
	//	logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，获取分布式锁失败,waId:%v", methodName, waId))
	//	return
	//}
	//defer func() {
	//	del := template.Del(context.Background(), constant.GetUserLockKey(config.ApplicationConfig.Activity.Id, waId))
	//	if !del {
	//		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，删除分布式锁失败,waId:%v", methodName, waId))
	//	}
	//}()
	logTracing.LogPrintf(ctx, logTracing.TaskHandleLogFmt, fmt.Sprintf("方法[%s],开始执行，waId:%v", methodName, user.WaId))

	ctx = &gin.Context{}
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

	// 更新是否发送续免费消息
	updateUser := entity.UserAttendInfoEntityV2{
		Id:                    user.Id,
		IsSendPayRenewFreeMsg: constant.RenewFreeSend,
	}
	userAttendInfoMapper := dao.GetUserAttendInfoMapperV2()
	_, err = userAttendInfoMapper.UpdateByPrimaryKeySelective(&session, updateUser)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],更新用户参与表 修改发送续免费消息状态为已发送，失败，活动id:%v，err:%v", methodName, config.ApplicationConfig.Activity.Id, err))
		return
	}

	// 发送付费-续订消息
	msgInfoEntity := &entity.MsgInfoEntityV2{
		Id:      util.GetSnowFlakeIdStr(ctx),
		Type:    "send",
		WaId:    user.WaId,
		MsgType: constant.PayRenewFreeMsg,
	}
	sendNxListParamsDtoList, err := service.PayRenewFreeMsg(ctx, msgInfoEntity, user.Language, constant.BizTypeTemplate)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，发送付费-续订消息失败,waId:%v，err:%v", methodName, user.WaId, err))
		return
	}

	if !isExist {
		session.Commit()
	}

	_, nxErr := service.SendMsgList2NX(ctx, sendNxListParamsDtoList)
	if nxErr != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，发付费-续订消息到牛信云失败,err：%v", methodName, nxErr))
		return
	}

	logTracing.LogPrintf(ctx, logTracing.TaskHandleLogFmt, fmt.Sprintf("方法[%s],执行完成，waId:%v", methodName, user.WaId))

}
