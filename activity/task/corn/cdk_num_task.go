package cron_task

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"go-fission-activity/activity/constant"
	"go-fission-activity/activity/model/dto"
	"go-fission-activity/activity/model/entity"
	"go-fission-activity/activity/model/response"
	"go-fission-activity/activity/task/statistics"
	"go-fission-activity/activity/third/http_client"
	"go-fission-activity/activity/third/redis_template"
	"go-fission-activity/activity/web/dao"
	"go-fission-activity/activity/web/middleware/logTracing"
	"go-fission-activity/activity/web/service"
	"go-fission-activity/config"
	"go-fission-activity/config/initConfig"
	"go-fission-activity/util"
	"go-fission-activity/util/config/encoder/json"
	"go-fission-activity/util/txUtil"
	"math"
	"strings"
	"time"
)

func cdkNumTask(methodName string, timeConfig config.TimerConfig) {
	ginCtx := gin.Context{}
	ctx := &ginCtx
	// defer 异常处理
	defer func() {
		if e := recover(); e != nil {
			logTracing.LogErrorPrintf(ctx, errors.New(fmt.Sprintf("方法[%s]，发生panic异常", methodName)), logTracing.ErrorLogFmt, e)
			return
		}
	}()

	logTracing.LogPrintf(ctx, logTracing.TaskHandleLogFmt, fmt.Sprintf("方法[%s],发送飞书任务开始执行", methodName))

	// 查询活动信息
	activityInfoMapper := dao.GetActivityInfoMapper()
	activityInfo, err := activityInfoMapper.SelectByPrimaryKey(config.ApplicationConfig.Activity.Id)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],查询活动信息失败，活动id:%v,err:%v", methodName, config.ApplicationConfig.Activity.Id, err))
		return
	}
	//if activityInfo.ActivityStatus == constant.ATStatusUnStart || activityInfo.ActivityStatus == constant.ATStatusEnd {
	//	logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],活动不在运行期，活动id:%v", methodName, config.ApplicationConfig.Activity.Id))
	//	return
	//}

	template := redis_template.NewRedisTemplate()
	taskLockKey := constant.GetTaskLockKey(config.ApplicationConfig.Activity.Id, methodName)

	getLock, err := template.SetNX(context.Background(), taskLockKey, "1", lockTimeout).Result()
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],调用redis nx失败，本实例不执行任务", methodName))
		return
	}
	if !getLock {
		//todo 不确定
		//template.Del(context.Background(), taskLockKey)
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],获取分布式锁失败，本实例不执行任务", methodName))
		return
	}
	defer func() {
		del := template.Del(context.Background(), taskLockKey)
		if !del {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，删除分布式锁失败", methodName))
		}
	}()

	now := util.GetNowCustomTime()
	// 获取月日，格式：x月x日
	monthDay := fmt.Sprintf("%d月%d日", now.Month(), now.Day())
	// 获取小时，格式：xx:00
	hour := fmt.Sprintf("%02d:%02d", now.Hour(), now.Minute())

	//startReportCustomTime := util.CustomTime{
	//	Time:      now.Time,
	//	IsNotZero: false,
	//}
	//
	//endReportCustomTime := util.CustomTime{
	//	Time:      now.Time,
	//	IsNotZero: false,
	//}
	// 查询各代种子总数
	//userAttendInfoMapper := dao.GetUserAttendInfoMapperV2()
	//query := dto.GenerationUserQueryDto{
	//	ActivityId:            config.ApplicationConfig.Activity.Id,
	//	StartReportCustomTime: startReportCustomTime.Unix(),
	//	EndReportCustomTime:   endReportCustomTime.Unix(),
	//}
	timeRange := dto.StatisticsTimeRange{
		StartTimestamp: 0,
		EndTimestamp:   0,
	}
	generationUserDtoList, err := statistics.GenerationInfoWithAttend(ctx, timeRange)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],计算用户迭代总数，失败，活动id:%v,err:%v", methodName, config.ApplicationConfig.Activity.Id, err))
		return
	}
	generation01Count := 0
	generation01CompleteCount := 0
	generationOtherCount := 0
	generationOtherCompleteCount := 0
	for _, generationUserDto := range generationUserDtoList {
		if constant.Generation01 == generationUserDto.Generation {
			generation01Count += generationUserDto.Count
			if constant.AttendStatusAttend != generationUserDto.AttendStatus {
				generation01CompleteCount += generationUserDto.Count
			}
		} else {
			generationOtherCount += generationUserDto.Count
			if constant.AttendStatusAttend != generationUserDto.AttendStatus {
				generationOtherCompleteCount += generationUserDto.Count
			}
		}
	}

	// 飞书
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("%s %s更新：\n", monthDay, hour))

	aa := generation01Count + generationOtherCount*3
	builder.WriteString(fmt.Sprintf("初代种子引入：%v\n初代种子完成预约：%v\n总裂变人数：%v\n裂变完成预约人数：%v\n预计覆盖人数：%v\n",
		util.AddThousandSeparators(generation01Count), util.AddThousandSeparators(generation01CompleteCount),
		util.AddThousandSeparators(generationOtherCount), util.AddThousandSeparators(generationOtherCompleteCount), util.AddThousandSeparators(aa)))

	cdkTypeList := constant.GetAllCdkType()
	for _, cdkType := range cdkTypeList {
		cdkKey := constant.GetCdkKey(config.ApplicationConfig.Activity.Id, cdkType)
		cdkNotUsedLen, err := template.LLen(context.Background(), cdkKey).Result()
		if err != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],获取%v存量的长度失败，key:%v,err:%v", methodName, cdkType, cdkKey, err))
			return
		}
		cdkInfoKey := constant.GetCdkInfoKey(config.ApplicationConfig.Activity.Id, cdkType)
		exists, err := template.Exists(context.Background(), cdkInfoKey)
		if err != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],查看%v是否存在失败，key:%v,err:%v", methodName, cdkType, cdkInfoKey, err))
			return
		}

		cdkCount := int64(0)
		percent := float64(0)
		if exists != 0 {
			cdkInfoStr, err := template.Get(context.Background(), cdkInfoKey)
			if err != nil {
				logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],获取%v总长度失败，key:%v,err:%v", methodName, cdkType, cdkInfoKey, err))
				return
			}

			cdkInfo := &response.CdkInfo{}
			err = json.NewEncoder().Decode([]byte(cdkInfoStr), cdkInfo)
			if err != nil {
				logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，cdkInfo转实体报错,cdkInfoStr:%v,err：%v", methodName, cdkInfoStr, err))
				return
			}
			cdkCount = cdkInfo.CdkCount
			percent = math.Round(float64(cdkCount-cdkNotUsedLen)/float64(cdkCount)*10000) / 100
		}

		switch cdkType {
		case constant.FreeCdk:
			builder.WriteString(fmt.Sprintf("免费奖励发放量：%v/%v；%v%%\n", util.AddThousandSeparators64(cdkCount-cdkNotUsedLen), util.AddThousandSeparators64(cdkCount), percent))
		case constant.ThreeCdk:
			cdkCount = cdkCount - 799999
			cdkNotUsedLen = cdkNotUsedLen - 799999
			builder.WriteString(fmt.Sprintf("第一档奖励发放量：%v/%v；%v%%\n", util.AddThousandSeparators64(cdkCount-cdkNotUsedLen), util.AddThousandSeparators64(cdkCount), percent))
		case constant.FiveCdk:
			builder.WriteString(fmt.Sprintf("第二档奖励发放量：%v/%v；%v%%\n", util.AddThousandSeparators64(cdkCount-cdkNotUsedLen), util.AddThousandSeparators64(cdkCount), percent))
		case constant.EightCdk:
			builder.WriteString(fmt.Sprintf("第三档奖励发放量：%v/%v；%v%%\n", util.AddThousandSeparators64(cdkCount-cdkNotUsedLen), util.AddThousandSeparators64(cdkCount), percent))
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

	msgInfoMapper := dao.GetMsgInfoMapperV2()
	priceCount, err := msgInfoMapper.SumPriceSendUnCountMsg(&session, 0)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],统计消息花费失败，活动id:%v,err:%v", methodName, config.ApplicationConfig.Activity.Id, err))
		return
	}
	percent := math.Round(priceCount/activityInfo.CostMax*10000) / 100
	//builder.WriteString(fmt.Sprintf("提醒费用（USD）：%v/%v；%v%%\n", util.AddThousandSeparators64(int64(priceCount)), activityInfo.CostMax, percent))

	builder.WriteString(fmt.Sprintf("缓冲期阈值：%v%%\n", initConfig.GetCdkLimit()))

	//发送失败，发送超时，非白拦截
	notWhiteCount, err := service.GetAllDaysIncrKeyCount(methodName, constant.NotWhiteCountKey, monthDay, template)
	if err != nil {
		return
	}

	sendSuccessMsgCount, err := service.GetAllDaysIncrKeyCount(methodName, constant.SendSuccessMsgCountKey, monthDay, template)
	if err != nil {
		return
	}

	sendFailMsgCount, err := service.GetAllDaysIncrKeyCount(methodName, constant.SendFailMsgCountKey, monthDay, template)
	if err != nil {
		return
	}

	sendTimeOutMsgCount, err := service.GetAllDaysIncrKeyCount(methodName, constant.SendTimeOutMsgCountKey, monthDay, template)
	if err != nil {
		return
	}

	msgCount := sendSuccessMsgCount + sendFailMsgCount
	percent = 0
	if msgCount > 0 {
		percent = math.Round(float64(sendFailMsgCount)/float64(msgCount)*10000) / 100
	}
	builder.WriteString(fmt.Sprintf("发送失败：%v条;%v%%\n", util.AddThousandSeparators64(sendFailMsgCount), percent))

	percent = 0
	if msgCount > 0 {
		percent = math.Round(float64(sendTimeOutMsgCount)/float64(msgCount)*10000) / 100
	}
	builder.WriteString(fmt.Sprintf("发送超时：%v条;%v%%\n", util.AddThousandSeparators64(sendTimeOutMsgCount), percent))

	//查询已经参与活动的人
	attendInfoMapper := dao.GetUserAttendInfoMapperV2()
	attendActivityCount, err := attendInfoMapper.CountUser()
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],查询参与活动总数失败，活动id:%v，err：%v", methodName, config.ApplicationConfig.Activity.Id, err))
		return
	}
	percent = 0

	userCount := int64(attendActivityCount) + notWhiteCount
	if msgCount > 0 {
		percent = math.Round(float64(notWhiteCount)/float64(userCount)*10000) / 100
	}
	builder.WriteString(fmt.Sprintf("非白拦截：%v条;%v%%", util.AddThousandSeparators64(notWhiteCount), percent))

	message := builder.String()

	nowTime := util.GetNowCustomTime()
	id := util.GetSnowFlakeIdStr(ctx)
	reportMsgInfoEntity := entity.ReportMsgInfoEntity{
		Id:         id,
		Date:       monthDay,
		Hour:       hour,
		ReportType: constant.ReportTypeFeiShu,
		MsgStatus:  constant.NXMsgStatusOwnerSent,
		Msg:        message,
		CreatedAt:  nowTime,
		UpdatedAt:  nowTime,
	}
	mapper := dao.GetReportMsgInfoMapper()

	_, err = mapper.InsertSelective(&session, reportMsgInfoEntity)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],新增报告表失败，活动id:%v,err:%v", methodName, config.ApplicationConfig.Activity.Id, err))
		return
	}

	content := map[string]any{
		"text": message,
	}
	body := map[string]any{
		"msg_type": "text",
		"content":  content,
	}

	msgStatus := constant.NXMsgStatusOwnerSent

	var res string
	for i := 1; i < 4; i++ {
		res, err = http_client.DoPostSSL(config.ApplicationConfig.Feishu.WebHook, body, nil, 2*1000, 2*1000)
		if err != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],调用飞书接口报错，message:%v,err:%v", methodName, message, err))
			time.Sleep(1 * time.Second)
			continue
		}
		break
	}
	if err != nil {
		msgStatus = constant.NXMsgStatusFailed
		res = fmt.Sprintf("发送飞书失败，err：%v", err)
	}

	updateEntity := entity.ReportMsgInfoEntity{
		Id:        id,
		MsgStatus: msgStatus,
		Res:       res,
	}
	_, err = mapper.UpdateByPrimaryKeySelective(&session, updateEntity)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],更新报告表失败，活动id:%v,err:%v", methodName, config.ApplicationConfig.Activity.Id, err))
		return
	}
	if !isExist {
		session.Commit()
	}
	logTracing.LogPrintf(ctx, logTracing.TaskHandleLogFmt, fmt.Sprintf("方法[%s],发送飞书消息执行完成，res:%v，message:%v", methodName, res, message))
}
