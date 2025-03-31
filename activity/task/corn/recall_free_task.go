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

var recallFreeGoroutinePool = goroutine_pool.NewGoroutinePool(4)

func recallFreeTask(methodName string, timeConfig config.TimerConfig) {
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
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],调用redis nx报错，本实例不执行任务,err:%v", methodName, err))
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
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],查询活动信息失败，活动id:%v,err:%v", methodName, config.ApplicationConfig.Activity.Id, err))
		return
	}
	if activityInfo.ActivityStatus == constant.ATStatusUnStart || activityInfo.ActivityStatus == constant.ATStatusEnd {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],活动不在运行期，活动id:%v", methodName, config.ApplicationConfig.Activity.Id))
		return
	}

	userAttendInfoMapper := dao.GetUserAttendInfoMapperV2()

	// 查询将要达到续费时间的用户 使用：SendRenewFreeAt
	currentTimestamp := time.Now().Unix()
	renewFreeUserCount, err := userAttendInfoMapper.CountRenewFree(constant.RenewFreeUnSend, currentTimestamp)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s [未续免费用户]],统计总数失败,err:%v", methodName, err))
		return
	}
	if renewFreeUserCount > 0 {
		lastId := 0 // 初始的起始ID为0
		for {
			// 查询当前页数据
			userList, err := userAttendInfoMapper.SelectRenewFree(lastId, PageSize, constant.RenewFreeUnSend, currentTimestamp)
			if err != nil {
				logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s [未续免费用户]],查询失败,err:%v", methodName, err))
				break // 如果查询失败，退出循环
			}

			// 如果查询结果为空，说明分页结束
			if len(userList) == 0 {
				break
			}

			// 遍历查询结果
			for _, user := range userList {
				recallFreeGoroutinePool.Execute(func(param interface{}) {
					u, ok := param.(entity.UserAttendInfoEntityV2) // 断言u是User类型
					if !ok {
						logTracing.LogPrintf(ctx, logTracing.TaskHandleLogFmt, fmt.Sprintf("方法[%s],断言发生错误，waId:%v", methodName, u.WaId))
					}
					logTracing.LogPrintf(ctx, logTracing.TaskHandleLogFmt, fmt.Sprintf("方法[%s],recallFreeGoroutinePool协程池执行任务开始，waId:%v", methodName, u.WaId))
					handlerRenewFreeUserInfo(ctx, methodName, u)
					logTracing.LogPrintf(ctx, logTracing.TaskHandleLogFmt, fmt.Sprintf("方法[%s],recallFreeGoroutinePool协程池执行任务结束，waId:%v", methodName, u.WaId))
				}, user)
			}

			// 更新 lastId 为当前结果中的最大ID
			lastId = userList[len(userList)-1].Id

			recallFreeGoroutinePool.Wait()
		}
	}

}

func handlerRenewFreeUserInfo(ctx *gin.Context, methodName string, user entity.UserAttendInfoEntityV2) {
	methodName = methodName + " [RenewFree]"
	waId := user.WaId
	logTracing.LogPrintf(ctx, logTracing.TaskHandleLogFmt, fmt.Sprintf("方法[%s],开始执行，waId:%v", methodName, waId))

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
		Id:                 user.Id,
		IsSendRenewFreeMsg: constant.RenewFreeSend,
	}
	userAttendInfoMapper := dao.GetUserAttendInfoMapperV2()
	_, err = userAttendInfoMapper.UpdateByPrimaryKeySelective(&session, updateUser)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],更新用户参与表 修改发送续免费消息状态为已发送，失败，活动id:%v,err:%v", methodName, config.ApplicationConfig.Activity.Id, err))
		return
	}

	// 发送续免费
	msgInfoEntity := &entity.MsgInfoEntityV2{
		Id:      util.GetSnowFlakeIdStr(ctx),
		Type:    "send",
		WaId:    user.WaId,
		MsgType: constant.RenewFreeMsg,
	}
	sendNxListParamsDtoList, err := service.RenewFreeMsg(ctx, msgInfoEntity, user.Language, constant.BizTypeInteractive)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，发送续免费信息失败,waId:%v,err:%v", methodName, user.WaId, err))
		return
	}

	if !isExist {
		session.Commit()
	}

	_, nxErr := service.SendMsgList2NX(ctx, sendNxListParamsDtoList)
	if nxErr != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，发消息到牛信云失败,err：%v", methodName, nxErr))
		return
	}
	logTracing.LogPrintf(ctx, logTracing.TaskHandleLogFmt, fmt.Sprintf("方法[%s],执行完成，waId:%v", methodName, waId))

}
