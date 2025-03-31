package cron_task

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
	"go-fission-activity/activity/constant"
	"go-fission-activity/activity/model/dto"
	"go-fission-activity/activity/model/entity"
	"go-fission-activity/activity/task/statistics"
	"go-fission-activity/activity/third/redis_template"
	"go-fission-activity/activity/web/dao"
	"go-fission-activity/activity/web/middleware/logTracing"
	"go-fission-activity/activity/web/service"
	"go-fission-activity/config"
	"go-fission-activity/util"
	"go-fission-activity/util/config/encoder/json"
	"go-fission-activity/util/txUtil"
	"gopkg.in/gomail.v2"
	"io"
	"strconv"
	"time"
)

func reportTask(methodName string, utc int) {
	ginCtx := gin.Context{}
	ctx := &ginCtx
	// defer 异常处理
	defer func() {
		if e := recover(); e != nil {
			logTracing.LogErrorPrintf(ctx, errors.New(fmt.Sprintf("方法[%s]，发生panic异常", methodName)), logTracing.ErrorLogFmt, e)
			return
		}
	}()

	//nowCustomTime := util.GetNowCustomTime()
	//hour := nowCustomTime.Time.Hour()
	//if hour != 0 && !config.ApplicationConfig.IsDebug {
	//	logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],任务执行时间为：%v，当前时间：%v,不是任务执行时间跳过。", methodName, hour, nowCustomTime.Time))
	//	return
	//}

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
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],调用redis nx失败，本实例不执行任务，err:%v", methodName, err))
		return
	}
	if !getLock {
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

	startReportCustomTime, endReportCustomTime := util.GetReportCountTime()
	// 获取月日，格式：x月x日
	monthDay := fmt.Sprintf("%d月%d日", startReportCustomTime.Month(), startReportCustomTime.Day())

	mapper := dao.GetReportMsgInfoMapper()

	count, err := mapper.SelectCountByReportTypeAndDay(config.ApplicationConfig.Activity.Id, constant.ReportTypeExcel, monthDay)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，查询当天记录失败,monthDay：%v，err:%v", methodName, monthDay, err))
		return
	}
	if count > 0 {
		logTracing.LogPrintfP("已有报告数据，不处理任务，monthDay：%v", monthDay)
		return
	}

	// 初始化数据
	buildDataMap := getInitData(ctx, monthDay)

	query := dto.GenerationUserQueryDto{
		ActivityId:            config.ApplicationConfig.Activity.Id,
		StartReportCustomTime: startReportCustomTime.Unix(),
		EndReportCustomTime:   endReportCustomTime.Unix(),
	}

	timeRange := getTimeRange(utc)

	// 查询各代种子总数 & 查询助力情况
	err = selectReportData(ctx, methodName, query, timeRange, buildDataMap, false)
	if err != nil {
		return
	}
	err = selectReportData(ctx, methodName, query, timeRange, buildDataMap, true)
	if err != nil {
		return
	}

	//var sendSuccessKeyList []string
	//var sendFailKeyList []string
	//var sendTimeOutKeyList []string
	//var notWhiteList []string
	//var reportJsonDtoList []*dto.ReportJsonDto
	// 增加发送成功数、发送失败数、发送超时数、非白拦截
	for _, channel := range config.ApplicationConfig.Activity.ChannelList {
		for _, language := range config.ApplicationConfig.Activity.LanguageList {
			if util.ArrayStringContains(config.ApplicationConfig.Activity.ChannelList, channel) && util.ArrayStringContains(config.ApplicationConfig.Activity.LanguageList, language) {
				reportJsonDto := buildDataMap[channel][language]

				//reportJsonDtoList = append(reportJsonDtoList, reportJsonDto)

				key := constant.GetSendSuccessMsgCountKey(config.ApplicationConfig.Activity.Id, monthDay, channel, language)

				//sendSuccessKeyList = append(sendSuccessKeyList, key)

				sendSuccessMsgCount, err := service.GetIncrKeyCount(methodName, key, template)
				if err != nil {
					return
				}
				reportJsonDto.SendSuccessMsgCount += sendSuccessMsgCount

				key = constant.GetSendFailMsgCountKey(config.ApplicationConfig.Activity.Id, monthDay, channel, language)
				//sendFailKeyList = append(sendFailKeyList, key)

				sendFailMsgCount, err := service.GetIncrKeyCount(methodName, key, template)
				if err != nil {
					return
				}
				reportJsonDto.SendFailMsgCount += sendFailMsgCount

				key = constant.GetSendTimeOutMsgCountKey(config.ApplicationConfig.Activity.Id, monthDay, channel, language)
				//sendTimeOutKeyList = append(sendTimeOutKeyList, key)

				sendTimeOutMsgCount, err := service.GetIncrKeyCount(methodName, key, template)
				if err != nil {
					return
				}
				reportJsonDto.SendTimeOutMsgCount += sendTimeOutMsgCount

				key = constant.GetNotWhiteCountKey(config.ApplicationConfig.Activity.Id, monthDay, channel, language)

				//notWhiteList = append(notWhiteList, key)

				notWhiteCount, err := service.GetIncrKeyCount(methodName, key, template)
				if err != nil {
					return
				}
				reportJsonDto.NotWhiteCount += notWhiteCount
			}
		}
	}

	//sendSuccessCountList, err := template.MGet(context.Background(), sendSuccessKeyList...)
	//if err != nil {
	//	logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，批量查询sendSuccessKey统计报错,err：%v", methodName, err))
	//	return
	//}
	//
	//sendFailCountList, err := template.MGet(context.Background(), sendFailKeyList...)
	//if err != nil {
	//	logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，批量查询sendFailKey统计报错,err：%v", methodName, err))
	//	return
	//}
	//
	//sendTimeOutCountList, err := template.MGet(context.Background(), sendTimeOutKeyList...)
	//if err != nil {
	//	logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，批量查询sendTimeOut统计报错,err：%v", methodName, err))
	//	return
	//}
	//
	//notWhiteCountList, err := template.MGet(context.Background(), notWhiteList...)
	//if err != nil {
	//	logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，批量查询notWhiteKey统计报错,err：%v", methodName, err))
	//	return
	//}
	//
	//for index, reportJsonDto := range reportJsonDtoList {
	//	sendSuccessCount := sendSuccessCountList[index]
	//	sendFailCount := sendFailCountList[index]
	//	sendTimeOutCount := sendTimeOutCountList[index]
	//	notWhiteCount := notWhiteCountList[index]
	//	if sendSuccessCount != nil {
	//		// 断言为string类型
	//		str, ok := sendSuccessCount.(string)
	//		if !ok {
	//			fmt.Println("类型断言失败，不是string类型")
	//		} else {
	//			fmt.Println(str) // 输出: Hello, World!
	//		}
	//		sendSuccessMsgCount, err := strconv.ParseInt(str, 10, 64)
	//		if err != nil {
	//			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%v]，字符串转int失败:%v", methodName, err))
	//			return
	//		}
	//		reportJsonDto.SendSuccessMsgCount += sendSuccessMsgCount
	//	}
	//	if sendFailCount != nil {
	//		// 断言为string类型
	//		str, ok := sendFailCount.(string)
	//		if !ok {
	//			fmt.Println("类型断言失败，不是string类型")
	//		} else {
	//			fmt.Println(str) // 输出: Hello, World!
	//		}
	//		sendFailCount2, err := strconv.ParseInt(str, 10, 64)
	//		if err != nil {
	//			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%v]，字符串转int失败:%v", methodName, err))
	//			return
	//		}
	//		reportJsonDto.SendFailMsgCount += sendFailCount2
	//	}
	//	if sendTimeOutCount != nil {
	//		// 断言为string类型
	//		str, ok := sendTimeOutCount.(string)
	//		if !ok {
	//			fmt.Println("类型断言失败，不是string类型")
	//		} else {
	//			fmt.Println(str) // 输出: Hello, World!
	//		}
	//		sendTimeOutCount2, err := strconv.ParseInt(str, 10, 64)
	//		if err != nil {
	//			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%v]，字符串转int失败:%v", methodName, err))
	//			return
	//		}
	//		reportJsonDto.SendTimeOutMsgCount += sendTimeOutCount2
	//	}
	//	if notWhiteCount != nil {
	//		// 断言为string类型
	//		str, ok := notWhiteCount.(string)
	//		if !ok {
	//			fmt.Println("类型断言失败，不是string类型")
	//		} else {
	//			fmt.Println(str) // 输出: Hello, World!
	//		}
	//		notWhiteCount2, err := strconv.ParseInt(str, 10, 64)
	//		if err != nil {
	//			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%v]，字符串转int失败:%v", methodName, err))
	//			return
	//		}
	//		reportJsonDto.NotWhiteCount += notWhiteCount2
	//	}
	//}

	// 去重总和
	otherJsonDto := buildDataMap[constant.ChannelOther][constant.LanguageOther]
	sendSuccessMsgCount, err := service.GetAllDaysIncrKeyCount(methodName, constant.SendSuccessMsgCountKey, monthDay, template)
	if err != nil {
		return
	}
	otherJsonDto.SendSuccessMsgCount += sendSuccessMsgCount

	sendFailMsgCount, err := service.GetAllDaysIncrKeyCount(methodName, constant.SendFailMsgCountKey, monthDay, template)
	if err != nil {
		return
	}
	otherJsonDto.SendFailMsgCount += sendFailMsgCount

	sendTimeOutMsgCount, err := service.GetAllDaysIncrKeyCount(methodName, constant.SendTimeOutMsgCountKey, monthDay, template)
	if err != nil {
		return
	}
	otherJsonDto.SendTimeOutMsgCount += sendTimeOutMsgCount

	notWhiteCount, err := service.GetAllDaysIncrKeyCount(methodName, constant.NotWhiteCountKey, monthDay, template)
	if err != nil {
		return
	}
	otherJsonDto.NotWhiteCount += notWhiteCount

	// 保存今天的数据
	todayList := getSendData(ctx, buildDataMap)

	todayListBytes, err := json.NewEncoder().Encode(todayList)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，转换sendList为json失败,err:%v", methodName, err))
		return
	}

	// 保存数据
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

	nowTime := util.GetNowCustomTime()
	id := util.GetSnowFlakeIdStr(context.Background())
	reportMsgInfoEntity := entity.ReportMsgInfoEntity{
		Id:         id,
		Date:       monthDay,
		ReportType: constant.ReportTypeExcel,
		MsgStatus:  constant.NXMsgStatusOwnerSent,
		Msg:        string(todayListBytes),
		CreatedAt:  nowTime,
		UpdatedAt:  nowTime,
	}

	_, err = mapper.InsertSelective(&session, reportMsgInfoEntity)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],新增报告表失败，活动id:%v,err:%v", methodName, config.ApplicationConfig.Activity.Id, err))
		return
	}

	// 组装数据
	reportDataList, err := mapper.SelectListByReportType(config.ApplicationConfig.Activity.Id, constant.ReportTypeExcel)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，查询以往发送数据失败,err:%v", methodName, err))
		return
	}

	var sendList []*dto.ReportJsonDto
	for _, reportData := range reportDataList {
		msg := reportData.Msg
		var sendDataList []*dto.ReportJsonDto
		err = json.NewEncoder().Decode([]byte(msg), &sendDataList)
		if err != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，转换历史发送数据为实体失败,err:%v", methodName, err))
			return
		}
		sendList = append(sendList, sendDataList...)
	}
	sendList = append(sendList, todayList...)

	//sendList = append(sendList, otherJsonDto)

	// 组装excel
	fileBytes, err := generateExcelFile(ctx, sendList)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],生成excel失败，活动id:%v，err:%v", methodName, config.ApplicationConfig.Activity.Id, err))
		return
	}

	sendEmail(ctx, fileBytes, strconv.Itoa(utc))
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],邮件发送失败，活动id:%v，err:%v", methodName, config.ApplicationConfig.Activity.Id, err))
		return
	}

	if !isExist {
		session.Commit()
	}
	logTracing.LogPrintf(ctx, logTracing.TaskHandleLogFmt, fmt.Sprintf("方法[%s],任务执行完毕", methodName))
}

func selectReportData(ctx context.Context, methodName string, singleQuery dto.GenerationUserQueryDto, singleTimeRange dto.StatisticsTimeRange, buildDataMap map[string]map[string]*dto.ReportJsonDto, isAll bool) error {

	//allQuery := dto.GenerationUserQueryDto{
	//	ActivityId:            config.ApplicationConfig.Activity.Id,
	//	StartReportCustomTime: 0,
	//	EndReportCustomTime:   singleQuery.EndReportCustomTime,
	//}

	allTimeRange := dto.StatisticsTimeRange{
		StartTimestamp: 0,
		EndTimestamp:   2000000000,
	}

	//var query dto.GenerationUserQueryDto
	var timeRange dto.StatisticsTimeRange
	if isAll {
		//query = allQuery
		timeRange = allTimeRange
	} else {
		//query = singleQuery
		timeRange = singleTimeRange
	}

	// 查询各代种子总数
	//userAttendInfoMapper := dao.GetUserAttendInfoMapperV2()
	var generationUserDtoList []dto.GenerationUserDto
	var err error
	generationUserDtoList, err = statistics.GenerationInfo(ctx, timeRange)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],计算用户迭代总数，失败，活动id:%v，err:%v", methodName, config.ApplicationConfig.Activity.Id, err))
		return errors.New(fmt.Sprintf("方法[%s],计算用户迭代总数，失败，活动id:%v", methodName, config.ApplicationConfig.Activity.Id))
	}
	// 渠道-语言-数据
	for _, generationUserDto := range generationUserDtoList {
		var reportJsonDtoList []*dto.ReportJsonDto
		if isAll {
			reportJsonDtoList = append(reportJsonDtoList, buildDataMap[constant.ChannelOther][constant.LanguageOther])
		} else {
			if util.ArrayStringContains(config.ApplicationConfig.Activity.ChannelList, generationUserDto.Channel) && util.ArrayStringContains(config.ApplicationConfig.Activity.LanguageList, generationUserDto.Language) {
				reportJsonDtoList = append(reportJsonDtoList, buildDataMap[generationUserDto.Channel][generationUserDto.Language])
			} else {
				continue
			}
		}

		for _, reportJsonDto := range reportJsonDtoList {
			switch generationUserDto.Generation {
			case constant.Generation01:
				reportJsonDto.Generation01 = reportJsonDto.Generation01 + generationUserDto.Count
			case constant.Generation02:
				reportJsonDto.Generation02 = reportJsonDto.Generation02 + generationUserDto.Count
				reportJsonDto.Generation02After = reportJsonDto.Generation02After + generationUserDto.Count
			case constant.Generation03:
				reportJsonDto.Generation03 = reportJsonDto.Generation03 + generationUserDto.Count
				reportJsonDto.Generation02After = reportJsonDto.Generation02After + generationUserDto.Count
			case constant.Generation04:
				reportJsonDto.Generation04 = reportJsonDto.Generation04 + generationUserDto.Count
				reportJsonDto.Generation02After = reportJsonDto.Generation02After + generationUserDto.Count
			case constant.Generation05:
				reportJsonDto.Generation05 = reportJsonDto.Generation05 + generationUserDto.Count
				reportJsonDto.Generation02After = reportJsonDto.Generation02After + generationUserDto.Count
			default:
				reportJsonDto.Generation06After = reportJsonDto.Generation06After + generationUserDto.Count
				reportJsonDto.Generation02After = reportJsonDto.Generation02After + generationUserDto.Count
			}
		}
	}

	//todo 替代了原有的sql方法
	//helpInfoMapper := dao.GetHelpInfoMapperV2()
	//helpCountDtoList, err := helpInfoMapper.CountUserByHelpCount(singleQuery)
	helpCountDtoList, err := statistics.HelpInfo(ctx, singleTimeRange)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],计算用户助力总数，失败，活动id:%v，err:%v", methodName, config.ApplicationConfig.Activity.Id, err))
		return errors.New(fmt.Sprintf("方法[%s],计算用户助力总数，失败，活动id:%v", methodName, config.ApplicationConfig.Activity.Id))
	}

	for _, helpCountDto := range helpCountDtoList {
		var reportJsonDtoList []*dto.ReportJsonDto
		if isAll {
			reportJsonDtoList = append(reportJsonDtoList, buildDataMap[constant.ChannelOther][constant.LanguageOther])
		} else {
			if util.ArrayStringContains(config.ApplicationConfig.Activity.ChannelList, helpCountDto.Channel) && util.ArrayStringContains(config.ApplicationConfig.Activity.LanguageList, helpCountDto.Language) {
				reportJsonDtoList = append(reportJsonDtoList, buildDataMap[helpCountDto.Channel][helpCountDto.Language])
			} else {
				continue
			}
		}
		for _, reportJsonDto := range reportJsonDtoList {
			switch helpCountDto.HelpNum {
			case 1:
				reportJsonDto.Help1 = reportJsonDto.Help1 + helpCountDto.HelpNumCount
			case 2:
				reportJsonDto.Help2 = reportJsonDto.Help2 + helpCountDto.HelpNumCount
			case 3:
				reportJsonDto.Help3 = reportJsonDto.Help3 + helpCountDto.HelpNumCount
			case 4:
				reportJsonDto.Help4 = reportJsonDto.Help4 + helpCountDto.HelpNumCount
			case 5:
				reportJsonDto.Help5 = reportJsonDto.Help5 + helpCountDto.HelpNumCount
			case 6:
				reportJsonDto.Help6 = reportJsonDto.Help6 + helpCountDto.HelpNumCount
			case 7:
				reportJsonDto.Help7 = reportJsonDto.Help7 + helpCountDto.HelpNumCount
			case 8:
				reportJsonDto.Help8 = reportJsonDto.Help8 + helpCountDto.HelpNumCount
			}
		}
	}

	//helpCountDtoList, err = helpInfoMapper.CountUserByHelpCount(allQuery)
	helpCountDtoList, err = statistics.HelpInfo(ctx, allTimeRange)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],计算用户助力总数，失败，活动id:%v，err:%v", methodName, config.ApplicationConfig.Activity.Id, err))
		return errors.New(fmt.Sprintf("方法[%s],计算用户助力总数，失败，活动id:%v", methodName, config.ApplicationConfig.Activity.Id))
	}

	for _, helpCountDto := range helpCountDtoList {
		var reportJsonDtoList []*dto.ReportJsonDto
		if isAll {
			reportJsonDtoList = append(reportJsonDtoList, buildDataMap[constant.ChannelOther][constant.LanguageOther])
		} else {
			if util.ArrayStringContains(config.ApplicationConfig.Activity.ChannelList, helpCountDto.Channel) && util.ArrayStringContains(config.ApplicationConfig.Activity.LanguageList, helpCountDto.Language) {
				reportJsonDtoList = append(reportJsonDtoList, buildDataMap[helpCountDto.Channel][helpCountDto.Language])
			} else {
				continue
			}
		}
		for _, reportJsonDto := range reportJsonDtoList {
			switch helpCountDto.HelpNum {
			case 1:
				reportJsonDto.AllHelp1 = reportJsonDto.AllHelp1 + helpCountDto.HelpNumCount
			case 2:
				reportJsonDto.AllHelp2 = reportJsonDto.AllHelp2 + helpCountDto.HelpNumCount
			case 3:
				reportJsonDto.AllHelp3 = reportJsonDto.AllHelp3 + helpCountDto.HelpNumCount
			case 4:
				reportJsonDto.AllHelp4 = reportJsonDto.AllHelp4 + helpCountDto.HelpNumCount
			case 5:
				reportJsonDto.AllHelp5 = reportJsonDto.AllHelp5 + helpCountDto.HelpNumCount
			case 6:
				reportJsonDto.AllHelp6 = reportJsonDto.AllHelp6 + helpCountDto.HelpNumCount
			case 7:
				reportJsonDto.AllHelp7 = reportJsonDto.AllHelp7 + helpCountDto.HelpNumCount
			case 8:
				reportJsonDto.AllHelp8 = reportJsonDto.AllHelp8 + helpCountDto.HelpNumCount
			}
		}
	}

	// 改为催促成团下发数、免费续时下发数、付费续时下发数
	reFreeCountMap, _ := statistics.MsgInfo(ctx, timeRange)
	//msgInfoMapper := dao.GetMsgInfoMapperV2()
	//query.MsgType = constant.PromoteClusteringMsg
	//reFreeFreeCountDtoList, err := msgInfoMapper.CountReFreeMsgByPrice(query)
	//if err != nil {
	//	logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],计算免费提醒消息，失败，活动id:%v，err:%v", methodName, config.ApplicationConfig.Activity.Id, err))
	//	return errors.New(fmt.Sprintf("方法[%s],计算免费提醒消息，失败，活动id:%v，err:%v", methodName, config.ApplicationConfig.Activity.Id, err))
	//}
	promoteCountDtoList := reFreeCountMap[constant.PromoteClusteringMsg]
	for _, reFreeFreeCountDto := range promoteCountDtoList {
		if isAll {
			otherReportJsonDto := buildDataMap[constant.ChannelOther][constant.LanguageOther]
			otherReportJsonDto.PromoteClusteringCount = otherReportJsonDto.PromoteClusteringCount + reFreeFreeCountDto.Count
		} else {
			if util.ArrayStringContains(config.ApplicationConfig.Activity.ChannelList, reFreeFreeCountDto.Channel) && util.ArrayStringContains(config.ApplicationConfig.Activity.LanguageList, reFreeFreeCountDto.Language) {
				reportJsonDto := buildDataMap[reFreeFreeCountDto.Channel][reFreeFreeCountDto.Language]
				if reportJsonDto != nil {
					reportJsonDto.PromoteClusteringCount = reFreeFreeCountDto.Count
				}
			}
		}
	}

	//query.MsgType = constant.RenewFreeMsg
	//reFreeFreeCountDtoList, err = msgInfoMapper.CountReFreeMsgByPrice(query)
	//if err != nil {
	//	logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],计算免费提醒消息，失败，活动id:%v，err:%v", methodName, config.ApplicationConfig.Activity.Id, err))
	//	return errors.New(fmt.Sprintf("方法[%s],计算免费提醒消息，失败，活动id:%v，err:%v", methodName, config.ApplicationConfig.Activity.Id, err))
	//}
	reFreeFreeCountDtoList := reFreeCountMap[constant.RenewFreeMsg]
	for _, reFreeFreeCountDto := range reFreeFreeCountDtoList {
		if isAll {
			otherReportJsonDto := buildDataMap[constant.ChannelOther][constant.LanguageOther]
			otherReportJsonDto.FreeRemindCount = otherReportJsonDto.FreeRemindCount + reFreeFreeCountDto.Count
		} else {
			if util.ArrayStringContains(config.ApplicationConfig.Activity.ChannelList, reFreeFreeCountDto.Channel) && util.ArrayStringContains(config.ApplicationConfig.Activity.LanguageList, reFreeFreeCountDto.Language) {
				reportJsonDto := buildDataMap[reFreeFreeCountDto.Channel][reFreeFreeCountDto.Language]
				if reportJsonDto != nil {
					reportJsonDto.FreeRemindCount = reFreeFreeCountDto.Count
				}
			}
		}
	}

	//query.MsgType = constant.PayRenewFreeMsg
	//reFreeNotFreeCountDtoList, err := msgInfoMapper.CountReFreeMsgByPrice(query)
	//if err != nil {
	//	logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],计算付费提醒消息，失败，活动id:%v，err:%v", methodName, config.ApplicationConfig.Activity.Id, err))
	//	return errors.New(fmt.Sprintf("方法[%s],计算付费提醒消息，失败，活动id:%v，err:%v", methodName, config.ApplicationConfig.Activity.Id, err))
	//}
	reFreeNotFreeCountDtoList := reFreeCountMap[constant.PayRenewFreeMsg]
	for _, reFreeNotFreeCountDto := range reFreeNotFreeCountDtoList {
		if isAll {
			otherReportJsonDto := buildDataMap[constant.ChannelOther][constant.LanguageOther]
			otherReportJsonDto.PayRemindCount = otherReportJsonDto.PayRemindCount + reFreeNotFreeCountDto.Count
		} else {
			if util.ArrayStringContains(config.ApplicationConfig.Activity.ChannelList, reFreeNotFreeCountDto.Channel) && util.ArrayStringContains(config.ApplicationConfig.Activity.LanguageList, reFreeNotFreeCountDto.Language) {
				reportJsonDto := buildDataMap[reFreeNotFreeCountDto.Channel][reFreeNotFreeCountDto.Language]
				if reportJsonDto != nil {
					reportJsonDto.PayRemindCount = reFreeNotFreeCountDto.Count
				}
			}
		}
	}
	return nil
}

func getTimeRange(utc int) dto.StatisticsTimeRange {
	currentTimestamp := time.Now().Unix()
	diff := (currentTimestamp + int64(3600*utc)) % 86400
	endTimestamp := currentTimestamp - diff
	startTimestamp := endTimestamp - 86400
	return dto.StatisticsTimeRange{
		StartTimestamp: startTimestamp,
		EndTimestamp:   endTimestamp,
	}
}

// 创建Excel文件并返回文件流
func generateExcelFile(ctx *gin.Context, reportJsonDtoList []*dto.ReportJsonDto) ([]byte, error) {
	// 创建一个新的 Excel 文件
	f := excelize.NewFile()
	methodName := "generateExcelFile"
	// 在第一个工作表中设置表头（两行）
	sheetName := "数据日报"
	err := f.SetSheetName("Sheet1", sheetName)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],创建sheet失败，err:%v", methodName, err))
		return nil, err
	}

	// 设置 A 到 Z 列的宽度
	for col := 'A'; col <= 'Z'; col++ {
		err = f.SetColWidth(sheetName, string(col), string(col), 15) // 设置每列宽度为 20
		if err != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],设置sheet样式失败，err:%v", methodName, err))
			return nil, err
		}
	}

	err = mergeCell(ctx, f, sheetName, "A1", "C1", "基础类目")
	if err != nil {
		return nil, err
	}

	err = mergeCell(ctx, f, sheetName, "D1", "J1", "导入和裂变情况")
	if err != nil {
		return nil, err
	}

	err = mergeCell(ctx, f, sheetName, "K1", "R1", "助力滞留情况【A单日】")
	if err != nil {
		return nil, err
	}

	err = mergeCell(ctx, f, sheetName, "S1", "Z1", "助力滞留情况【B累加】")
	if err != nil {
		return nil, err
	}

	f.SetColWidth(sheetName, "AA1", "AA1", 15)
	err = f.SetCellValue(sheetName, "AA1", "催促成团下发数")
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],设置单元格AA1值失败，err:%v", methodName, err))
		return nil, err
	}

	f.SetColWidth(sheetName, "AB1", "AB1", 15)
	err = f.SetCellValue(sheetName, "AB1", "免费续时下发数")
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],设置单元格AB1值失败，err:%v", methodName, err))
		return nil, err
	}

	f.SetColWidth(sheetName, "AC1", "AC1", 15)
	err = f.SetCellValue(sheetName, "AC1", "付费续时下发数")
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],设置单元格AC1值失败，err:%v", methodName, err))
		return nil, err
	}

	f.SetColWidth(sheetName, "AD1", "AD1", 15)
	err = f.SetCellValue(sheetName, "AD1", "发送成功")
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],设置单元格AD1值失败，err:%v", methodName, err))
		return nil, err
	}

	f.SetColWidth(sheetName, "AE1", "AE1", 15)
	err = f.SetCellValue(sheetName, "AE1", "发送失败")
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],设置单元格AE1值失败，err:%v", methodName, err))
		return nil, err
	}

	f.SetColWidth(sheetName, "AF1", "AF1", 15)
	err = f.SetCellValue(sheetName, "AF1", "发送超时")
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],设置单元格AF1值失败，err:%v", methodName, err))
		return nil, err
	}

	f.SetColWidth(sheetName, "AG1", "AG1", 15)
	err = f.SetCellValue(sheetName, "AG1", "非白拦截")
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],设置单元格AG1值失败，err:%v", methodName, err))
		return nil, err
	}

	// 设置第二行的标题
	headers := []string{"日期", "语言", "渠道", "初代种子引入", "二代人数", "三代人数", "四代人数", "五代人数", "6+代人数",
		"总裂变人数（2代及之后累计）", "拉1人人数", "拉2人人数", "拉3人人数", "拉4人人数", "拉5人人数",
		"拉6人人数", "拉7人人数", "拉8人人数", "拉1人人数", "拉2人人数", "拉3人人数", "拉4人人数", "拉5人人数",
		"拉6人人数", "拉7人人数", "拉8人人数"}
	var charA = 'A'
	for i, header := range headers {
		cell := fmt.Sprintf("%s2", string(rune(int(charA)+i)))
		err = f.SetCellValue(sheetName, cell, header)
		if err != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],设置第二行的标题失败，err:%v", methodName, err))
			return nil, err
		}
	}

	// 填充数据
	for i, reportJsonDto := range reportJsonDtoList {
		row := i + 3 // 从第 3 行开始填充数据
		index := 0
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", getCloKey(index), row), reportJsonDto.Date)
		index = index + 1
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", getCloKey(index), row), reportJsonDto.Language)
		index = index + 1
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", getCloKey(index), row), reportJsonDto.Channel)
		index = index + 1
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", getCloKey(index), row), reportJsonDto.Generation01)
		index = index + 1
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", getCloKey(index), row), reportJsonDto.Generation02)
		index = index + 1
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", getCloKey(index), row), reportJsonDto.Generation03)
		index = index + 1
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", getCloKey(index), row), reportJsonDto.Generation04)
		index = index + 1
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", getCloKey(index), row), reportJsonDto.Generation05)
		index = index + 1
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", getCloKey(index), row), reportJsonDto.Generation06After)
		index = index + 1
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", getCloKey(index), row), reportJsonDto.Generation02After)
		index = index + 1
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", getCloKey(index), row), reportJsonDto.Help1)
		index = index + 1
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", getCloKey(index), row), reportJsonDto.Help2)
		index = index + 1
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", getCloKey(index), row), reportJsonDto.Help3)
		index = index + 1
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", getCloKey(index), row), reportJsonDto.Help4)
		index = index + 1
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", getCloKey(index), row), reportJsonDto.Help5)
		index = index + 1
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", getCloKey(index), row), reportJsonDto.Help6)
		index = index + 1
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", getCloKey(index), row), reportJsonDto.Help7)
		index = index + 1
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", getCloKey(index), row), reportJsonDto.Help8)
		index = index + 1
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", getCloKey(index), row), reportJsonDto.AllHelp1)
		index = index + 1
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", getCloKey(index), row), reportJsonDto.AllHelp2)
		index = index + 1
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", getCloKey(index), row), reportJsonDto.AllHelp3)
		index = index + 1
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", getCloKey(index), row), reportJsonDto.AllHelp4)
		index = index + 1
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", getCloKey(index), row), reportJsonDto.AllHelp5)
		index = index + 1
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", getCloKey(index), row), reportJsonDto.AllHelp6)
		index = index + 1
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", getCloKey(index), row), reportJsonDto.AllHelp7)
		index = index + 1
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", getCloKey(index), row), reportJsonDto.AllHelp8)
		index = index + 1
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", getCloKey(index), row), reportJsonDto.PromoteClusteringCount)
		index = index + 1
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", getCloKey(index), row), reportJsonDto.FreeRemindCount)
		index = index + 1
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", getCloKey(index), row), reportJsonDto.PayRemindCount)
		index = index + 1
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", getCloKey(index), row), reportJsonDto.SendSuccessMsgCount)
		index = index + 1
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", getCloKey(index), row), reportJsonDto.SendFailMsgCount)
		index = index + 1
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", getCloKey(index), row), reportJsonDto.SendTimeOutMsgCount)
		index = index + 1
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", getCloKey(index), row), reportJsonDto.NotWhiteCount)
	}

	// 将Excel文件保存到内存中的字节切片
	var buf bytes.Buffer
	err = f.Write(&buf)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],xlsx写字节数组失败，err:%v", methodName, err))
		return nil, err
	}

	// 返回字节切片
	return buf.Bytes(), nil
}

func getCloKey(index int) string {
	if index < 26 {
		return string(rune(int('A') + index))
	} else {
		index = index - 26
		return "A" + string(rune(int('A')+index))
	}
}

func mergeCell(ctx *gin.Context, file *excelize.File, sheetName, startIndex, endIndex, value string) error {
	methodName := "mergeCell"

	err := file.MergeCell(sheetName, startIndex, endIndex)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],合并单元格%v 失败，err:%v", methodName, startIndex, err))
		return err
	}

	// 设置合并单元格的值
	err = file.SetCellValue(sheetName, startIndex, value)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],设置合并单元格%v 的值失败，err:%v", methodName, startIndex, err))
		return err
	}
	return nil
}

func getInitData(ctx *gin.Context, monthDay string) map[string]map[string]*dto.ReportJsonDto {
	buildDataMap := make(map[string]map[string]*dto.ReportJsonDto)
	for _, channel := range config.ApplicationConfig.Activity.ChannelList {
		for _, language := range config.ApplicationConfig.Activity.LanguageList {
			reportJsonDto := &dto.ReportJsonDto{
				Date:              monthDay,
				Language:          config.ApplicationConfig.Activity.LanguageNameMap[language],
				Channel:           config.ApplicationConfig.Activity.ChannelNameMap[channel],
				Generation01:      0,
				Generation02:      0,
				Generation03:      0,
				Generation04:      0,
				Generation05:      0,
				Generation06After: 0,
				Generation02After: 0,
				Help1:             0,
				Help2:             0,
				Help3:             0,
				Help4:             0,
				Help5:             0,
				Help6:             0,
				Help7:             0,
				Help8:             0,
				FreeRemindCount:   0,
				PayRemindCount:    0,
			}
			languageMap := buildDataMap[channel]
			if languageMap == nil {
				languageMap = make(map[string]*dto.ReportJsonDto)
				buildDataMap[channel] = languageMap
			}
			languageMap[language] = reportJsonDto
		}
	}

	reportJsonDto := &dto.ReportJsonDto{
		Date:              monthDay,
		Language:          config.ApplicationConfig.Activity.LanguageNameMap[constant.LanguageOther],
		Channel:           config.ApplicationConfig.Activity.ChannelNameMap[constant.ChannelOther],
		Generation01:      0,
		Generation02:      0,
		Generation03:      0,
		Generation04:      0,
		Generation05:      0,
		Generation06After: 0,
		Generation02After: 0,
		Help1:             0,
		Help2:             0,
		Help3:             0,
		Help4:             0,
		Help5:             0,
		Help6:             0,
		Help7:             0,
		Help8:             0,
		FreeRemindCount:   0,
		PayRemindCount:    0,
	}
	languageMap := buildDataMap[constant.ChannelOther]
	if languageMap == nil {
		languageMap = make(map[string]*dto.ReportJsonDto)
		buildDataMap[constant.ChannelOther] = languageMap
	}
	languageMap[constant.LanguageOther] = reportJsonDto

	return buildDataMap
}

func getOtherInitData(ctx *gin.Context, monthDay string) *dto.ReportJsonDto {
	return &dto.ReportJsonDto{
		Date:              monthDay,
		Language:          config.ApplicationConfig.Activity.LanguageNameMap[constant.LanguageOther],
		Channel:           config.ApplicationConfig.Activity.ChannelNameMap[constant.ChannelOther],
		Generation01:      0,
		Generation02:      0,
		Generation03:      0,
		Generation04:      0,
		Generation05:      0,
		Generation06After: 0,
		Generation02After: 0,
		Help1:             0,
		Help2:             0,
		Help3:             0,
		Help4:             0,
		Help5:             0,
		Help6:             0,
		Help7:             0,
		Help8:             0,
		FreeRemindCount:   0,
		PayRemindCount:    0,
	}
}

func getSendData(ctx *gin.Context, reportMap map[string]map[string]*dto.ReportJsonDto) []*dto.ReportJsonDto {
	var buildDataList []*dto.ReportJsonDto
	for _, channel := range config.ApplicationConfig.Activity.ChannelList {
		for _, language := range config.ApplicationConfig.Activity.LanguageList {
			jsonDto := reportMap[channel][language]
			buildDataList = append(buildDataList, jsonDto)
		}
	}
	jsonDto := reportMap[constant.ChannelOther][constant.LanguageOther]
	buildDataList = append(buildDataList, jsonDto)

	return buildDataList
}

func sendEmail(ctx *gin.Context, fileData []byte, utc string) error {
	methodName := "sendEmail"
	// 创建一个新的邮件对象
	mailer := gomail.NewMessage()

	emailConfig := config.ApplicationConfig.EmailConfig

	// 设置发件人和收件人
	mailer.SetHeader("From", emailConfig.FromAddress)    // 发件人地址
	mailer.SetHeader("To", emailConfig.ToAddressList...) // 收件人地址
	mailer.SetHeader("Subject", "fission活动每日统计数据")       // 邮件主题
	mailer.SetBody("text/plain", "统计数据详见附件")             // 邮件内容

	// 将字节数组作为附件附加到邮件中
	// 使用 AttachReader 方法将字节数组包装为一个 io.Reader 来作为附件
	mailer.Attach("wa日报 UTC"+utc+".xlsx", gomail.SetCopyFunc(func(w io.Writer) error {
		_, err := w.Write(fileData)
		return err
	}))

	// 设置 SMTP 服务器的配置信息
	dialer := gomail.NewDialer(
		emailConfig.ServerHost, //邮箱的 SMTP 服务器地址
		emailConfig.ServerPort, //邮箱的 SMTP 端口
		emailConfig.ApiUser,    // user
		emailConfig.ApiKey,     // 密码（或应用专用密码）
	)
	// 设置 SSL 加密
	//dialer.SSL = true

	// 发送邮件
	if err := dialer.DialAndSend(mailer); err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],发送邮件失败，活动id:%v，err:%v", methodName, config.ApplicationConfig.Activity.Id, err))
		return err
	}
	logTracing.LogPrintf(ctx, logTracing.TaskHandleLogFmt, fmt.Sprintf("方法[%s],邮件发送成功，活动id:%v", methodName, config.ApplicationConfig.Activity.Id))
	return nil
}
