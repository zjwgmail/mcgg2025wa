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
)

var recallCdkGoroutinePool = goroutine_pool.NewGoroutinePool(3)

func recallCdkTask(methodName string, timeConfig config.TimerConfig) {
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
	isDisturbTime := nowCustomTime.IsNotDisturbTime()
	if isDisturbTime {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],免打扰时间，不执行任务", methodName))
		return
	}

	template := redis_template.NewRedisTemplate()
	taskLockKey := constant.GetTaskLockKey(config.ApplicationConfig.Activity.Id, methodName)

	getLock, err := template.SetNX(context.Background(), taskLockKey, "1", lockTimeout).Result()
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

	userAttendInfoMapper := dao.GetUserAttendInfoMapperV2()

	// 查询未发送cdk的消息的用户，查看cdk是否充足并且发送cdk
	notSendCdkUserCount, err := userAttendInfoMapper.CountNotSendCdkUser(constant.CdkMsgUnSend)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s [未发送cdk消息用户]],统计总数失败，err:%v", methodName, err))
		return
	}
	if notSendCdkUserCount > 0 {
		//todo 1212测试暂缓
		lastId := 0 // 初始起始ID为0
		for {
			// 查询未发送CDK消息的用户
			userList, err := userAttendInfoMapper.SelectNotSendCdkUser(lastId, PageSize, constant.CdkMsgUnSend)
			if err != nil {
				logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s [未发送cdk消息用户]],查询失败，err:%v", methodName, err))
				break // 查询失败，退出循环
			}

			// 如果查询结果为空，说明没有更多用户，退出循环
			if len(userList) == 0 {
				break
			}

			// 遍历当前批次用户
			for _, user := range userList {
				recallCdkGoroutinePool.Execute(func(param interface{}) {
					u, ok := param.(entity.UserAttendInfoEntityV2) // 断言u是User类型
					if !ok {
						logTracing.LogPrintf(ctx, logTracing.TaskHandleLogFmt, fmt.Sprintf("方法[%s],断言发生错误，waId:%v", methodName, u.WaId))

					}
					logTracing.LogPrintf(ctx, logTracing.TaskHandleLogFmt, fmt.Sprintf("方法[%s],recallCdkGoroutinePool协程池执行任务开始，waId:%v", methodName, u.WaId))
					handlerNotSendCdkUserInfo(ctx, methodName, u)
					logTracing.LogPrintf(ctx, logTracing.TaskHandleLogFmt, fmt.Sprintf("方法[%s],recallCdkGoroutinePool协程池执行任务结束，waId:%v", methodName, u.WaId))
				}, user)
			}

			// 更新 lastId 为当前结果中的最大ID
			lastId = userList[len(userList)-1].Id

			recallCdkGoroutinePool.Wait()
		}
	}

}

func handlerNotSendCdkUserInfo(ctx *gin.Context, methodName string, user entity.UserAttendInfoEntityV2) {
	methodName = methodName + " [NotSendCdkUser]"
	waId := user.WaId
	logTracing.LogPrintf(ctx, logTracing.TaskHandleLogFmt, fmt.Sprintf("方法[%s],开始执行，waId:%v", methodName, waId))

	ctx = &gin.Context{}
	userAttendInfoMapper := dao.GetUserAttendInfoMapperV2()

	// 判断用户状态
	if constant.IsStage != user.IsThreeStage && constant.IsStage != user.IsFiveStage && constant.AttendStatusEightOver != user.AttendStatus {
		return
	}

	_, isFree, err := service.CheckCanSendMsg2NX(ctx, user.WaId)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，判断用户是否可以发送消息报错,waId:%v，err:%v", methodName, user.WaId, err))
		return
	}
	if !isFree {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，且用户不在免费期！,waId:%v，err:%v", methodName, user.WaId, err))
		return
	}
	sendNxMsgType := constant.BizTypeInteractive
	//if !isFree {
	//	sendNxMsgType = constant.BizTypeTemplate
	//}

	err = sendCdk(ctx, methodName, user, constant.ThreeCdk, sendNxMsgType)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，补发三人cdk成功消息报错,活动id:%v,waId:%v,err：%v", methodName, config.ApplicationConfig.Activity.Id, waId, err))
		return
	}
	if constant.IsStage == user.IsFiveStage || constant.AttendStatusEightOver == user.AttendStatus {
		err = sendCdk(ctx, methodName, user, constant.FiveCdk, sendNxMsgType)
		if err != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，补发五人cdk成功消息报错,活动id:%v,waId:%v,err：%v", methodName, config.ApplicationConfig.Activity.Id, waId, err))
			return
		}
		if constant.AttendStatusEightOver == user.AttendStatus {
			err = sendCdk(ctx, methodName, user, constant.EightCdk, sendNxMsgType)
			if err != nil {
				logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，补发八人cdk成功消息报错,活动id:%v,waId:%v,err：%v", methodName, config.ApplicationConfig.Activity.Id, waId, err))
				return
			}
		}
	}

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
	userAttendInfo := entity.UserAttendInfoEntityV2{
		Id:           user.Id,
		IsSendCdkMsg: constant.CdkMsgSend,
	}
	_, err = userAttendInfoMapper.UpdateByPrimaryKeySelective(&session, userAttendInfo)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，更新cdk消息未发送失败,waId:%v，err:%v", methodName, user.WaId, err))
		return
	}
	if !isExist {
		session.Commit()
	}
	logTracing.LogPrintf(ctx, logTracing.TaskHandleLogFmt, fmt.Sprintf("方法[%s],执行完成，waId:%v", methodName, waId))

}
