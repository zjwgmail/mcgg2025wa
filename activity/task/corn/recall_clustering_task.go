package cron_task

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"go-fission-activity/activity/constant"
	"go-fission-activity/activity/model/dto"
	"go-fission-activity/activity/model/entity"
	"go-fission-activity/activity/model/request"
	"go-fission-activity/activity/third/redis_template"
	"go-fission-activity/activity/web/dao"
	"go-fission-activity/activity/web/middleware/logTracing"
	"go-fission-activity/activity/web/service"
	"go-fission-activity/config"
	"go-fission-activity/util"
	"go-fission-activity/util/goroutine_pool"
	"go-fission-activity/util/txUtil"
	"strconv"
	"time"
)

var recallClusteringGoroutinePool = goroutine_pool.NewGoroutinePool(3)

func recallClusteringTask(methodName string, timeConfig config.TimerConfig) {
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

	//// 查询五小时内未成团或未做开团任务的用户
	//
	//userCount, err := userAttendInfoMapper.CountReCallOfUnRedPacket(config.ApplicationConfig.Activity.Id, config.ApplicationConfig.Activity.UnRedPacketMinute)
	//if err != nil {
	//	logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s [五小时内未成团或未做开团任务的用户]],统计总数失败", methodName))
	//	return
	//}
	//
	//if userCount > 0 {
	//	pageCount := userCount/PageSize + 1
	//
	//	for pageIndex := 0; pageIndex < pageCount; pageIndex++ {
	//		pageStart := pageIndex * PageSize
	//		userList, err := userAttendInfoMapper.SelectReCallOfUnRedPacket(config.ApplicationConfig.Activity.Id, pageStart, PageSize, config.ApplicationConfig.Activity.UnRedPacketMinute)
	//		if err != nil {
	//			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s [五小时内未成团或未做开团任务的用户]],查询失败", methodName))
	//			continue
	//		}
	//
	//		for _, user := range userList {
	//			if util.CheckDiffTime(user.NewestHelpAt, nowCustomTime, config.ApplicationConfig.Activity.UnRedPacketMinute) {
	//				handlerUnRedPacketUserInfo(ctx, methodName, user)
	//			}
	//		}
	//	}
	//}
	//
	//// 查询三小时内,的用户
	//sendRedPacketUserCount, err := userAttendInfoMapper.CountReCallOfSendRedPacket(config.ApplicationConfig.Activity.Id, config.ApplicationConfig.Activity.SendRedPacketMinute)
	//if err != nil {
	//	logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s [三小时内未领取红包的用户]],统计总数失败", methodName))
	//	return
	//}
	//if sendRedPacketUserCount > 0 {
	//	pageCount := sendRedPacketUserCount/PageSize + 1
	//	for pageIndex := 0; pageIndex < pageCount; pageIndex++ {
	//		pageStart := pageIndex * PageSize
	//		userList, err := userAttendInfoMapper.SelectReCallOfSendRedPacket(config.ApplicationConfig.Activity.Id, pageStart, PageSize, config.ApplicationConfig.Activity.SendRedPacketMinute)
	//		if err != nil {
	//			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s [三小时内未领取红包的用户]],查询失败", methodName))
	//			continue
	//		}
	//		for _, user := range userList {
	//			if util.CheckDiffTime(user.RedPacketReadyAt, nowCustomTime, config.ApplicationConfig.Activity.SendRedPacketMinute) {
	//				handlerSendRedPacketUserInfo(ctx, methodName, user)
	//			}
	//
	//		}
	//	}
	//}

	currentTimestamp := time.Now().Unix()
	if config.ApplicationConfig.Activity.NeedSubscribe {
		// 查询两小时未开团的用户，重新发送开团消息
		twoStartGroupTimestamp := currentTimestamp - int64(config.ApplicationConfig.Activity.TwoStartGroupMinute*60)
		sendStartGroupUserCount, err := userAttendInfoMapper.CountReCallOfStartGroup(twoStartGroupTimestamp)
		if err != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s [两小时内未开团的用户]],统计总数失败，err:%v", methodName, err))
			return
		}
		if sendStartGroupUserCount > 0 {
			pageCount := sendStartGroupUserCount/PageSize + 1
			for pageIndex := 0; pageIndex < pageCount; pageIndex++ {
				pageStart := 0
				userList, err := userAttendInfoMapper.SelectReCallOfStartGroup(pageStart, PageSize, twoStartGroupTimestamp)
				if err != nil {
					logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s [两小时内未开团的用户]],查询失败，err:%v", methodName, err))
					continue
				}
				for _, user := range userList {
					if util.CheckDiffTime(util.NewCustomTime(time.Unix(user.AttendAt, 0)), nowCustomTime, config.ApplicationConfig.Activity.TwoStartGroupMinute) {
						recallClusteringGoroutinePool.Execute(func(param interface{}) {
							u, ok := param.(entity.UserAttendInfoEntityV2) // 断言u是User类型
							if !ok {
								logTracing.LogPrintf(ctx, logTracing.TaskHandleLogFmt, fmt.Sprintf("方法[%s],断言发生错误，waId:%v", methodName, u.WaId))

							}
							logTracing.LogPrintf(ctx, logTracing.TaskHandleLogFmt, fmt.Sprintf("方法[%s],recallClusteringGoroutinePool协程池执行任务开始，waId:%v", methodName, u.WaId))
							handlerStartGroupUserInfo(ctx, methodName, u)
							logTracing.LogPrintf(ctx, logTracing.TaskHandleLogFmt, fmt.Sprintf("方法[%s],recallClusteringGoroutinePool协程池执行任务结束，waId:%v", methodName, u.WaId))
						}, user)
					}
				}
				recallClusteringGoroutinePool.Wait()
			}
		}
	}

	// 查找5小时之后未成团的，发送催促成团
	userCount, err := userAttendInfoMapper.CountReCallOfClustering(constant.ClusteringUnSend, currentTimestamp)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s [24小时之后未成团的用户]],统计总数失败，err:%v", methodName, err))
		return
	}

	if userCount > 0 {
		lastId := 0    // 初始起始ID为0
		pageSize := 30 // 每页大小

		for {
			// 查询未发送成团消息的用户
			userList, err := userAttendInfoMapper.SelectReCallOfClustering(
				lastId,
				pageSize,
				constant.ClusteringUnSend,
				currentTimestamp,
			)
			if err != nil {
				logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s [24小时之后未成团的用户]],查询失败，err:%v", methodName, err))
				break // 查询失败，退出循环
			}

			// 如果查询结果为空，说明没有更多用户，退出循环
			if len(userList) == 0 {
				break
			}

			// 遍历当前批次用户
			for _, user := range userList {
				recallClusteringGoroutinePool.Execute(func(param interface{}) {
					u, ok := param.(entity.UserAttendInfoEntityV2) // 断言u是User类型
					if !ok {
						logTracing.LogPrintf(ctx, logTracing.TaskHandleLogFmt, fmt.Sprintf("方法[%s],断言发生错误，waId:%v", methodName, u.WaId))
					}
					logTracing.LogPrintf(ctx, logTracing.TaskHandleLogFmt, fmt.Sprintf("方法[%s],recallClusteringGoroutinePool协程池执行任务开始，waId:%v", methodName, u.WaId))
					handlerUnClusteringUserInfo(ctx, methodName, u)
					logTracing.LogPrintf(ctx, logTracing.TaskHandleLogFmt, fmt.Sprintf("方法[%s],recallClusteringGoroutinePool协程池执行任务结束，waId:%v", methodName, u.WaId))
				}, user)
			}

			// 更新 lastId 为当前结果中的最大ID
			lastId = userList[len(userList)-1].Id

			recallClusteringGoroutinePool.Wait()
		}

	}

}

//
//func handlerUnRedPacketUserInfo(ctx *gin.Context, methodName string, user entity.UserAttendInfoEntityV2) {
//	methodName = methodName + " [五小时内未成团或未做开团任务的用户]"
//	waId := user.WaId
//	// redis锁
//	template := redis_template.NewRedisTemplate()
//	res, err := template.SetNX(context.Background(), constant.GetUserLockKey(config.ApplicationConfig.Activity.Id, waId), "1", constant.LockTimeOut).Result()
//	if err != nil {
//		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，获取分布式锁报错,活动id:%v,waId:%v,err：%v", methodName, config.ApplicationConfig.Activity.Id, waId, err))
//		return
//	}
//	if !res {
//		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，获取分布式锁失败,waId:%v", methodName, waId))
//		return
//	}
//	defer func() {
//		del := template.Del(context.Background(), constant.GetUserLockKey(config.ApplicationConfig.Activity.Id, waId))
//		if !del {
//			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，删除分布式锁失败,waId:%v", methodName, waId))
//		}
//	}()
//
//	ginCtx := &gin.Context{}
//	session, isExist, err := txUtil.GetTransaction(ginCtx)
//	if nil != err {
//		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，创建事务失败,err：%v", methodName, err))
//		return
//	}
//	if !isExist {
//		defer func() {
//			session.Rollback()
//			session.Close()
//		}()
//	}
//
//	updateEntity := entity.UserAttendInfoEntityV2{
//		Id:               user.Id,
//		RedPacketStatus:  constant.RedPacketStatusReady,
//		RedPacketReadyAt: util.GetNowCustomTime(),
//	}
//	userAttendInfoMapper := dao.GetUserAttendInfoMapperV2()
//	_, err = userAttendInfoMapper.UpdateByPrimaryKeySelective(&session, updateEntity)
//	if err != nil {
//		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s], 更新用户预发红包码失败，waId:%v", methodName, user.WaId))
//		return
//	}
//
//	// 发送预发红包消息
//	msgInfoEntity := &entity.MsgInfoEntityV2{
//		Id:         util.GetSnowFlakeIdStr(ctx),
//		Type:       "send",
//		WaId:       user.WaId,
//		ActivityId: config.ApplicationConfig.Activity.Id,
//		MsgType:    constant.RedPacketReadyMsg,
//	}
//	_, err = service.RedPacketReadyMsg2NX(ctx, msgInfoEntity, user.Language, user)
//	if err != nil {
//		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，发送预发红包消息失败,waId:%v", methodName, user.WaId))
//		return
//	}
//
//	if !isExist {
//		session.Commit()
//	}
//}
//
//func handlerSendRedPacketUserInfo(ctx *gin.Context, methodName string, user entity.UserAttendInfoEntityV2) {
//	methodName = methodName + " [三小时内未领取红包的用户]"
//	waId := user.WaId
//	// redis锁
//	template := redis_template.NewRedisTemplate()
//	res, err := template.SetNX(context.Background(), constant.GetUserLockKey(config.ApplicationConfig.Activity.Id, waId), "1", constant.LockTimeOut).Result()
//	if err != nil {
//		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，获取分布式锁报错,活动id:%v,waId:%v,err：%v", methodName, config.ApplicationConfig.Activity.Id, waId, err))
//		return
//	}
//	if !res {
//		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，获取分布式锁失败,waId:%v", methodName, waId))
//		return
//	}
//	defer func() {
//		del := template.Del(context.Background(), constant.GetUserLockKey(config.ApplicationConfig.Activity.Id, waId))
//		if !del {
//			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，删除分布式锁失败,waId:%v", methodName, waId))
//		}
//	}()
//
//	ginCtx := &gin.Context{}
//	session, isExist, err := txUtil.GetTransaction(ginCtx)
//	if nil != err {
//		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，创建事务失败,err：%v", methodName, err))
//		return
//	}
//	if !isExist {
//		defer func() {
//			session.Rollback()
//			session.Close()
//		}()
//	}
//
//	// 红包发放信息
//	redPacketKey := constant.GetRedPacketKey(config.ApplicationConfig.Activity.Id)
//	redPacket, err := template.BRPop(ctx, redPacketKey, constant.RedPacketTimeOut)
//	if err != nil {
//		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，获取redPacket失败,waId:%v", methodName, user.WaId))
//		return
//	}
//	redPacketCode := redPacket[0]
//	updateEntity := entity.UserAttendInfoEntityV2{
//		Id:              user.Id,
//		RedPacketCode:   redPacketCode,
//		RedPacketStatus: constant.RedPacketStatusSend,
//		RedPacketSendAt: util.GetNowCustomTime(),
//	}
//	userAttendInfoMapper := dao.GetUserAttendInfoMapperV2()
//	_, err = userAttendInfoMapper.UpdateByPrimaryKeySelective(&session, updateEntity)
//	if nil != err {
//		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，更新红包码失败,err：%v", methodName, err))
//		return
//	}
//
//	// 发送红包消息
//	msgInfoEntity := &entity.MsgInfoEntityV2{
//		Id:         util.GetSnowFlakeIdStr(ctx),
//		Type:       "send",
//		WaId:       user.WaId,
//		ActivityId: config.ApplicationConfig.Activity.Id,
//		MsgType:    constant.RedPacketSendMsg,
//	}
//	_, err = service.RedPacketSendMsg2NX(ctx, msgInfoEntity, user.Language, user, redPacketCode)
//	if err != nil {
//		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，发送红包发放信息失败,waId:%v", methodName, user.WaId))
//		return
//	}
//
//	if !isExist {
//		session.Commit()
//	}
//}

func handlerStartGroupUserInfo(ctx *gin.Context, methodName string, user entity.UserAttendInfoEntityV2) {
	methodName = methodName + " [recallNotStartGroupUserInfo]"
	waId := user.WaId
	logTracing.LogPrintf(ctx, logTracing.TaskHandleLogFmt, fmt.Sprintf("方法[%s],开始执行，waId:%v", methodName, waId))

	//// redis锁
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

	ginCtx := &gin.Context{}
	session, isExist, err := txUtil.GetTransaction(ginCtx)
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

	userAttendInfoMapper := dao.GetUserAttendInfoMapperV2()
	userAttendInfoEntity := entity.UserAttendInfoEntityV2{
		Id:                  user.Id,
		IsSendStartGroupMsg: constant.ClusteringSend,
	}
	_, err = userAttendInfoMapper.UpdateByPrimaryKeySelective(&session, userAttendInfoEntity)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，更新用户是否发送了催促成团消息失败,waId:%v，err:%v", methodName, user.WaId, err))
		return
	}

	// 发送参与活动消息
	msgInfoEntity := &entity.MsgInfoEntityV2{
		Id:      util.GetSnowFlakeIdStr(ctx),
		Type:    "send",
		WaId:    user.WaId,
		MsgType: constant.ActivityTaskMsg,
	}
	param := &request.HelpParam{
		WaId:      waId,
		IsHelp:    false,
		RallyCode: user.RallyCode,
	}
	sendNxListParamsDto, err := service.ActivityTask2NX(ctx, msgInfoEntity, user.Language, user.Channel, param)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，发送红包发放信息失败,waId:%v，err:%v", methodName, user.WaId, err))
		return
	}

	if !isExist {
		session.Commit()
	}

	_, nxErr := service.SendMsgList2NX(ctx, sendNxListParamsDto)
	if nxErr != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，发消息到牛信云失败,err：%v", methodName, nxErr))
		return
	}
	logTracing.LogPrintf(ctx, logTracing.TaskHandleLogFmt, fmt.Sprintf("方法[%s],执行完成，waId:%v", methodName, waId))

}

func sendCdk(ctx *gin.Context, methodName string, user entity.UserAttendInfoEntityV2, cdkType string, sendNxMsgType int) error {
	msgType := ""
	num := 0
	switch cdkType {
	case constant.ThreeCdk:
		msgType = constant.HelpThreeOverMsg
		num = config.ApplicationConfig.Activity.Stage1Award.HelpNum
	case constant.FiveCdk:
		msgType = constant.HelpFiveOverMsg
		num = config.ApplicationConfig.Activity.Stage2Award.HelpNum
	case constant.EightCdk:
		msgType = constant.HelpEightOverMsg
		num = config.ApplicationConfig.Activity.Stage3Award.HelpNum
	default:
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，不支持的cdk类型,waId:%v,cdkType:%v", methodName, user.WaId, cdkType))
		return errors.New("不支持的cdk类型")
	}

	logTracing.LogPrintf(ctx, logTracing.TaskHandleLogFmt, fmt.Sprintf("方法[%s],开始执行，waId:%v,cdkType:%v", methodName, user.WaId, cdkType))

	msgInfoMapper := dao.GetMsgInfoMapperV2()
	// 查询用户是否发送过了三人cdk消息
	msgCount, err := msgInfoMapper.CountCdkMsgByWaId(user.WaId, msgType)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，查询是否发送过%v cdk失败,waId:%v，err:%v", methodName, msgType, user.WaId, err))
		return err
	}
	if msgCount < 1 {
		// 发送三人cdk
		cdk, cdkIsExist, err := service.GetCdkByCdkType(context.Background(), cdkType)
		if err != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，查询redis中%v cdk失败,waId:%v，err:%v", methodName, msgType, user.WaId, err))
			return err
		}
		if !cdkIsExist {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，redis中%v cdk为空,waId:%v", methodName, msgType, user.WaId))
			return errors.New("redis" + cdkType + "类型cdk不存在")
		}

		key := constant.GetHelpInfoCacheKey(config.ApplicationConfig.Activity.Id, user.RallyCode)
		helpNameList, err := service.QueryHelpInfoCache(methodName, key, int64(num))
		if nil != err {
			return err
		}
		if len(helpNameList) < num {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，redis中缓存助力人数少于发cdk的需要人数,%v,waId:%v", methodName, msgType, user.WaId))
			return errors.New("redis中助力人数" + strconv.Itoa(len(helpNameList)) + "小于要求人数" + string(rune(num)))
		}

		ctx = &gin.Context{}
		session, isExist, err := txUtil.GetTransaction(ctx)
		if nil != err {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，创建事务失败,err：%v", methodName, err))
			return err
		}
		if !isExist {
			defer func() {
				session.Rollback()
				session.Close()
			}()
		}

		userAttendInfoMapper := dao.GetUserAttendInfoMapperV2()
		//helpInfoMapper := dao.GetHelpInfoMapperV2()

		//helpNameList, err := helpInfoMapper.SelectHelpNameByRallyCode(&session, config.ApplicationConfig.Activity.Id, user.RallyCode)
		//if nil != err || len(helpNameList) < num {
		//	logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，根据被助力人rallyCode查询助力人昵称失败,或助力人小于%v,rallyCode：%v,err：%v", methodName, num, user.RallyCode, err))
		//	return errors.New("database is error")
		//}
		//helpNameList = helpNameList[:num]

		userAttendInfoEntity := entity.UserAttendInfoEntityV2{
			Id: user.Id,
		}
		switch cdkType {
		case constant.ThreeCdk:
			userAttendInfoEntity.ThreeCdkCode = cdk
		case constant.FiveCdk:
			userAttendInfoEntity.FiveCdkCode = cdk
		case constant.EightCdk:
			userAttendInfoEntity.EightCdkCode = cdk
		}
		_, err = userAttendInfoMapper.UpdateByPrimaryKeySelective(&session, userAttendInfoEntity)
		if err != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，更新%v cdk失败,waId:%v，err:%v", methodName, msgType, user.WaId, err))
			return err
		}

		// 发送助力完成消息
		msgInfoEntity := &entity.MsgInfoEntityV2{
			Id:         util.GetSnowFlakeIdStr(ctx),
			Type:       "send",
			WaId:       user.WaId,
			SourceWaId: helpNameList[len(helpNameList)-1].WaId,
			MsgType:    msgType,
		}

		var sendNxListParamsDtoList []*dto.SendNxListParamsDto
		// cdk消息
		switch cdkType {
		case constant.ThreeCdk:
			sendNxListParamsDto, err := service.HelpThreeOverMsg2NX(ctx, msgInfoEntity, user, cdk, helpNameList, sendNxMsgType)
			if err != nil {
				logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，发送%v人助力完成消息失败,WaId:%v，err:%v", methodName, num, user.WaId, err))
				return err
			}
			sendNxListParamsDtoList = append(sendNxListParamsDtoList, sendNxListParamsDto...)
		case constant.FiveCdk:
			sendNxListParamsDto, err := service.HelpFiveOverMsg2NX(ctx, msgInfoEntity, user, cdk, helpNameList, sendNxMsgType)
			if err != nil {
				logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，发送%v人助力完成消息失败,WaId:%v，err:%v", methodName, num, user.WaId, err))
				return err
			}
			sendNxListParamsDtoList = append(sendNxListParamsDtoList, sendNxListParamsDto...)
		case constant.EightCdk:
			sendNxListParamsDto, err := service.HelpEightOverMsg2NX(ctx, msgInfoEntity, user, cdk, helpNameList, sendNxMsgType)
			if err != nil {
				logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，发送%v人助力完成消息失败,WaId:%v，err:%v", methodName, num, user.WaId, err))
				return err
			}
			sendNxListParamsDtoList = append(sendNxListParamsDtoList, sendNxListParamsDto...)
		}

		if !isExist {
			session.Commit()
		}
		_, nxErr := service.SendMsgList2NX(ctx, sendNxListParamsDtoList)
		if nxErr != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，发消息到牛信云失败,err：%v", methodName, nxErr))
			return nil
		}
	}
	logTracing.LogPrintf(ctx, logTracing.TaskHandleLogFmt, fmt.Sprintf("方法[%s],执行完成，waId:%v,cdkType:%v", methodName, user.WaId, cdkType))
	return nil
}

func handlerUnClusteringUserInfo(ctx *gin.Context, methodName string, user entity.UserAttendInfoEntityV2) {
	methodName = methodName + " [UnClusteringUserInfo]"

	ctx = &gin.Context{}

	logTracing.LogPrintf(ctx, logTracing.TaskHandleLogFmt, fmt.Sprintf("方法[%s],开始执行，waId:%v", methodName, user.WaId))

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

	_, isFree, err := service.CheckCanSendMsg2NX(ctx, user.WaId)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s], 判断用户是否是免费期失败，waId:%v，err:%v", methodName, user.WaId, err))
		return
	}

	var sendNxListParamsDtoList []*dto.SendNxListParamsDto

	helpInfoList, err := dao.GetHelpInfoMapperV2().SelectListByRallyCode(&session, user.RallyCode)
	if nil != err {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，根据开团人rallyCode查询助力人信息失败,rallyCode：%v,err：%v", constant.MethodHelp, user.RallyCode, err))
		return
	}
	waIds := make([]string, len(helpInfoList))
	for i, helpInfo := range helpInfoList {
		waIds[i] = helpInfo.WaId
	}

	userAttendInfoList := make([]entity.UserAttendInfoEntityV2, 0)
	if len(waIds) > 0 {
		userAttendInfoList, err = dao.GetUserAttendInfoMapperV2().SelectListByWaIdsWithSession(&session, waIds)
		if nil != err {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，根据助力人waId查询助力人昵称失败,rallyCode：%v,err：%v", constant.MethodHelp, user.RallyCode, err))
			return
		}
	}

	//  催促成团需要发送上次推送进度图
	// 发送催促成团消息
	msgInfoEntity := &entity.MsgInfoEntityV2{
		Id:      util.GetSnowFlakeIdStr(ctx),
		Type:    "send",
		WaId:    user.WaId,
		MsgType: constant.PromoteClusteringMsg,
	}
	//if isFree {
	sendNxListParamsDtoList, err = service.PromoteClusteringMsg2NX(ctx, msgInfoEntity, user, constant.BizTypeInteractive, userAttendInfoList)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，构建催促成团互动消息失败,waId:%v，err:%v", methodName, user.WaId, err))
		return
	}
	//} else {
	//	sendNxListParamsDtoList, err = service.PromoteClusteringMsg2NX(ctx, msgInfoEntity, user, constant.BizTypeTemplate)
	//	if err != nil {
	//		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，构建催促成团模板消息失败,waId:%v", methodName, user.WaId))
	//		return
	//	}
	//}

	userAttendInfoMapper := dao.GetUserAttendInfoMapperV2()
	userAttendInfoEntity := entity.UserAttendInfoEntityV2{
		Id:                  user.Id,
		IsSendClusteringMsg: constant.ClusteringSend,
	}
	_, err = userAttendInfoMapper.UpdateByPrimaryKeySelective(&session, userAttendInfoEntity)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，更新用户是否发送了催促成团消息失败,waId:%v，err:%v", methodName, user.WaId, err))
		return
	}

	if !isExist {
		session.Commit()
	}

	if isFree {
		_, nxErr := service.SendMsgList2NX(ctx, sendNxListParamsDtoList)
		if nxErr != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，发消息到牛信云失败,err：%v", methodName, nxErr))
			return
		}
	} else {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，用户非免费,不发送催促开团消息，且,waId:%v", methodName, user.WaId))
		return
	}
	logTracing.LogPrintf(ctx, logTracing.TaskHandleLogFmt, fmt.Sprintf("方法[%s],执行完成，waId:%v", methodName, user.WaId))

}
