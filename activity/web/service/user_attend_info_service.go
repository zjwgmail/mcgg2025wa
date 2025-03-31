package service

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gocarina/gocsv"
	"go-fission-activity/activity/constant"
	"go-fission-activity/activity/model/dto"
	"go-fission-activity/activity/model/entity"
	"go-fission-activity/activity/model/request"
	"go-fission-activity/activity/model/response"
	"go-fission-activity/activity/third/redis_template"
	"go-fission-activity/activity/web/dao"
	"go-fission-activity/activity/web/middleware/logTracing"
	"go-fission-activity/config"
	"go-fission-activity/util"
	"go-fission-activity/util/config/encoder/json"
	"go-fission-activity/util/goroutine_pool"
	"go-fission-activity/util/txUtil"
	"strconv"
	"strings"
	"sync"
	"time"
)

type UserAttendInfoService struct {
	userAttendInfoMapper  *dao.UserAttendInfoMapperV2
	activityInfoMapper    *dao.ActivityInfoMapper
	helpInfoMapper        *dao.HelpInfoMapperV2
	rsvMsgInfoMapper      *dao.RsvMsgInfoMapper
	rsvOtherMsgInfoMapper *dao.RsvOtherMsgInfo1Mapper
	freeCdkSendService    *FreeCdkSendService
}

var attendInfoServiceOnce sync.Once
var globalUserAttendInfoService UserAttendInfoService
var helpGoroutinePool = goroutine_pool.NewGoroutinePool(2000)

func GetUserAttendInfoService() *UserAttendInfoService {
	attendInfoServiceOnce.Do(func() {
		globalUserAttendInfoService = UserAttendInfoService{
			userAttendInfoMapper:  dao.GetUserAttendInfoMapperV2(),
			activityInfoMapper:    dao.GetActivityInfoMapper(),
			helpInfoMapper:        dao.GetHelpInfoMapperV2(),
			rsvMsgInfoMapper:      dao.GetRsvMsgInfoMapper(),
			rsvOtherMsgInfoMapper: dao.GetRsvOtherMsgInfo1Mapper(),
			freeCdkSendService:    GetFreeCdkSendService(),
		}
		logTracing.LogPrintfP("第一次使用，globalUserAttendInfoService")
	})
	return &globalUserAttendInfoService
}

func (u UserAttendInfoService) UserAttendInfo(ctx *gin.Context, req *request.UserSendMsgMethodInsertMsgInfo) (bool, error) {
	ginCtx := ctx.Copy()
	encoder := json.NewEncoder()
	reqAnyEncode, err := encoder.Encode(req)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，将请求解析为二进制数组报错,err：%v", constant.MethodUserAttendMethodInsertMsgInfo, err))
		return constant.MethodInsertMsgInfoReturnFail, errors.New("MethodInsertMsgInfo message is null")
	}
	if len(req.Messages) <= 0 {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，Messages为空,req：%v", constant.MethodUserAttendMethodInsertMsgInfo, req))
		return constant.MethodInsertMsgInfoReturnFail, errors.New("MethodInsertMsgInfo messages is empty")
	}

	if req.Business_phone == "" {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，Business_phone为空,req：%v", constant.MethodUserAttendMethodInsertMsgInfo, req))
		return constant.MethodInsertMsgInfoReturnFail, errors.New("MethodInsertMsgInfo business_phone is null")
	}
	if config.ApplicationConfig.Nx.BusinessPhone != req.Business_phone {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，Business_phone与配置不匹配,req：%v", constant.MethodUserAttendMethodInsertMsgInfo, req))
		return constant.MethodInsertMsgInfoReturnFail, errors.New("MethodInsertMsgInfo business_phone is not match")
	}
	if len(req.Contacts) <= 0 {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，Contacts为空,req：%v", constant.MethodUserAttendMethodInsertMsgInfo, req))
		return constant.MethodInsertMsgInfoReturnFail, errors.New("MethodInsertMsgInfo contacts is empty")
	}
	waId := req.Contacts[0].Wa_id
	if waId == "" {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，wa_id为空,req：%v", constant.MethodUserAttendMethodInsertMsgInfo, req))
		return constant.MethodInsertMsgInfoReturnFail, errors.New("MethodInsertMsgInfo waId is null")
	}

	profile := req.Contacts[0].Profile
	if profile == nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，profile为空,req：%v", constant.MethodUserAttendMethodInsertMsgInfo, req))
		return constant.MethodInsertMsgInfoReturnFail, errors.New("MethodInsertMsgInfo profile is null")
	}
	userNickName := profile.Name
	if userNickName == "" {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，profile.Name为空,req：%v", constant.MethodUserAttendMethodInsertMsgInfo, req))
		return constant.MethodInsertMsgInfoReturnFail, errors.New("MethodInsertMsgInfo profile.Name is null")
	}

	var textBody string
	if req.Messages[0].Type == "text" {
		if req.Messages[0].Text == nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，messages中Text为空,req：%v", constant.MethodUserAttendMethodInsertMsgInfo, req))
			return constant.MethodInsertMsgInfoReturnFail, errors.New("MethodInsertMsgInfo messages‘s text is null")
		} else {
			textBody = req.Messages[0].Text.Body
		}
	} else if req.Messages[0].Type == "button" {
		if req.Messages[0].Button == nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，messages中Button为空,req：%v", constant.MethodUserAttendMethodInsertMsgInfo, req))
			return constant.MethodInsertMsgInfoReturnFail, errors.New("MethodInsertMsgInfo messages‘s button is null")
		} else {
			textBody = req.Messages[0].Button.Text
		}
	} else if req.Messages[0].Type == "interactive" {
		if req.Messages[0].Interactive == nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，messages中interactive为空,req：%v", constant.MethodUserAttendMethodInsertMsgInfo, req))
			return constant.MethodInsertMsgInfoReturnFail, errors.New("MethodInsertMsgInfo messages‘s interactive is null")
		} else {
			if req.Messages[0].Interactive.Type == "button_reply" {
				if req.Messages[0].Interactive.ButtonReply == nil {
					logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，messages中interactive中的button_reply为空,req：%v", constant.MethodUserAttendMethodInsertMsgInfo, req))
					return constant.MethodInsertMsgInfoReturnFail, errors.New("MethodInsertMsgInfo messages‘s interactive button_reply is null")
				} else {
					textBody = req.Messages[0].Interactive.ButtonReply.Title
				}
			}
		}
	}

	var msgRecTime util.CustomTime
	var timestampDB int64 = 0
	if req.Messages[0].Timestamp == "" {
		msgRecTime = util.GetNowCustomTime()
	} else {
		timestamp := req.Messages[0].Timestamp
		msgRecTime, err = util.GetCustomTimeByTime(timestamp)
		if err != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，转换消息中的时间戳失败,req：%v", constant.MethodUserAttendMethodInsertMsgInfo, req))
			return constant.MethodInsertMsgInfoReturnFail, errors.New("MethodInsertMsgInfo receive messages time convert error")
		}
		timestampDB, _ = strconv.ParseInt(timestamp, 10, 64)
	}

	receiveMsgInfoEntity2 := entity.RsvOtherMsgInfo1Entity{
		TableName: config.ApplicationConfig.Activity.InsertOtherRsvMsgTable,
		WaId:      waId,
		Msg:       string(reqAnyEncode),
		Timestamp: timestampDB,
		CreatedAt: msgRecTime,
		UpdatedAt: msgRecTime,
	}
	_, err = u.rsvOtherMsgInfoMapper.InsertSelective2(receiveMsgInfoEntity2)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，新增用户发送的信息,waid:%v,err：%v", constant.RsvMsgInsert2, waId, err))
	} else {
		logTracing.LogPrintf(ctx, logTracing.NormalLogFmt, fmt.Sprintf("方法[%s]，新增用户发送的信息,waid:%v, msg：%v", constant.RsvMsgInsert2, waId, string(reqAnyEncode)))
	}

	sendRallyCode := ""
	containsPrefix := false
	isHelp := false
	for _, prefix := range config.ApplicationConfig.MethodInsertMsgInfo[config.ApplicationConfig.Activity.Scheme].UserAttendPrefixList {
		if strings.Contains(textBody, prefix) {
			containsPrefix = true
			codeStr := strings.Split(textBody, prefix)
			sendRallyCode = strings.TrimSpace(codeStr[len(codeStr)-1])
			break
		}
	}
	if !containsPrefix {
		for _, prefix := range config.ApplicationConfig.MethodInsertMsgInfo[config.ApplicationConfig.Activity.Scheme].UserAttendOfHelpPrefixList {
			if strings.Contains(textBody, prefix) {
				containsPrefix = true
				isHelp = true
				codeStr := strings.Split(textBody, prefix)
				sendRallyCode = strings.TrimSpace(codeStr[len(codeStr)-1])
				break
			}
		}
	}

	dbUserAttendInfo, err := u.userAttendInfoMapper.SelectByWaId(waId)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，根据waid查询userAttendInfo报错,waid:%v,err：%v", constant.MethodUserAttendMethodInsertMsgInfo, waId, err))
		return constant.MethodInsertMsgInfoReturnFail, errors.New("database is error")
	}

	if !containsPrefix {
		// 是续免费的消息
		if dbUserAttendInfo.Id > 0 {

			session, isExist, err := txUtil.GetTransaction(ctx)
			if nil != err {
				logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，创建事务失败,err：%v", constant.MethodHelp, err))
				return true, errors.New("database is error")
			}
			if !isExist {
				defer func() {
					session.Rollback()
					session.Close()
				}()
			}

			receiveMsgInfoService := GetReceiveMsgInfoService()
			receiveMsgInfoEntity := &entity.RsvMsgInfoEntity{
				Id:         util.GetSnowFlakeIdStr(ctx),
				Type:       "receive",
				WaId:       waId,
				SourceWaId: waId,
				MsgType:    constant.ReceiveMsg,
				Msg:        string(reqAnyEncode),
				MsgStatus:  constant.NXMsgStatusReceive,
				CreatedAt:  msgRecTime,
				UpdatedAt:  msgRecTime,
			}
			// 新增接收的消息
			err = receiveMsgInfoService.InsertMsgInfo(ctx, receiveMsgInfoEntity)
			if err != nil {
				logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，新增接收信息失败,req：%v,err:%v", constant.MethodUserAttendMethodInsertMsgInfo, req, err))
				return constant.MethodInsertMsgInfoReturnFail, errors.New("MethodInsertMsgInfo receive messages insert error")
			}

			endTime := util.GetTimeOfAfterDays(1, msgRecTime)
			sendRenewFreeAt := util.GetSendRenewMsgTime(1, msgRecTime)
			upUserAttendInfo := entity.UserAttendInfoEntityV2{
				Id:                 dbUserAttendInfo.Id,
				NewestFreeStartAt:  msgRecTime.Unix(),
				NewestFreeEndAt:    endTime,
				SendRenewFreeAt:    sendRenewFreeAt.Unix(),
				IsSendRenewFreeMsg: constant.RenewFreeUnSend,
			}
			_, err = u.userAttendInfoMapper.UpdateByPrimaryKeySelective(&session, upUserAttendInfo)
			if nil != err {
				logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，修改用户免费信息失败,waId:%v,err：%v", constant.MethodUserAttendMethodInsertMsgInfo, waId, err))
				return true, errors.New("database is error")
			}
			if !isExist {
				session.Commit()
			}

			ctx = &gin.Context{}
			// 判断是否是点击续费消息的回复
			for _, prefix := range config.ApplicationConfig.MethodInsertMsgInfo[config.ApplicationConfig.Activity.Scheme].RenewFreePrefixList {
				if strings.Contains(textBody, prefix) {
					containsPrefix = true
					isHelp = true
					codeStr := strings.Split(textBody, prefix)
					sendRallyCode = strings.TrimSpace(codeStr[len(codeStr)-1])
					break
				}
			}
			if containsPrefix {
				// 是点击续费消息的回复，发送续费回复消息
				msgInfoEntity := &entity.MsgInfoEntityV2{
					Id:         util.GetSnowFlakeIdStr(ctx),
					Type:       "send",
					WaId:       waId,
					SourceWaId: waId,
					MsgType:    constant.RenewFreeReplyMsg,
				}
				sendNxListParamsDto, err := RenewFreeReplyMsg(ctx, msgInfoEntity, dbUserAttendInfo, constant.BizTypeInteractive)
				if err != nil {
					logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，发送续费回复消息息失败,waId:%v", constant.MethodHelp, waId))
					return true, err
				}
				_, nxErr := SendMsgList2NX(ctx, sendNxListParamsDto)
				if nxErr != nil {
					logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，发结束期-不能开团消息到牛信云失败,err：%v", constant.MethodUserAttendMethodInsertMsgInfo, nxErr))
					return true, nxErr
				}
				return false, nil
			}
			ctx = &gin.Context{}
			// 立刻重发
			ReSendMsgByWaId(ctx, dbUserAttendInfo.WaId, true)

		} else {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，不包含指定前缀，并且非参与活动用户,waId:%v,err：%v", constant.MethodUserAttendMethodInsertMsgInfo, waId, err))
			return true, errors.New("database is error")
		}
		return false, nil
	}

	activityStatus, err := u.activityInfoMapper.SelectStatusByPrimaryKey(config.ApplicationConfig.Activity.Id)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，根据活动id查询activityInfo报错,活动id:%v,err：%v", constant.MethodUserAttendMethodInsertMsgInfo, config.ApplicationConfig.Activity.Id, err))
		return constant.MethodInsertMsgInfoReturnFail, errors.New("database is error")
	}
	if "" == activityStatus || constant.ATStatusUnStart == activityStatus {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，活动尚未开始,活动id:%v,err：%v", constant.MethodUserAttendMethodInsertMsgInfo, config.ApplicationConfig.Activity.Id, err))
		return constant.MethodInsertMsgInfoReturnFail, errors.New("activity is end")
	}

	// redis锁
	template := redis_template.NewRedisTemplate()
	res, err := template.SetNX(ctx, constant.GetUserLockKey(config.ApplicationConfig.Activity.Id, waId), "1", constant.LockTimeOut).Result()
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，获取分布式锁报错,活动id:%v,waId:%v,err：%v", constant.MethodUserAttendMethodInsertMsgInfo, config.ApplicationConfig.Activity.Id, waId, err))
		return constant.MethodInsertMsgInfoReturnFail, errors.New("repeated request")
	}
	if !res {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，获取分布式锁失败,waId:%v", constant.MethodUserAttendMethodInsertMsgInfo, waId))
		return constant.MethodInsertMsgInfoReturnFail, errors.New("repeated request")
	}
	defer func() {
		del := template.Del(ctx, constant.GetUserLockKey(config.ApplicationConfig.Activity.Id, waId))
		if !del {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，删除分布式锁失败,waId:%v", constant.MethodUserAttendMethodInsertMsgInfo, waId))
		}
	}()

	session, isExist, err := txUtil.GetTransaction(ctx)
	if nil != err {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，创建事务失败,err：%v", constant.MethodHelp, err))
		return true, errors.New("database is error")
	}
	if !isExist {
		defer func() {
			session.Rollback()
			session.Close()
		}()
	}

	receiveMsgInfoService := GetReceiveMsgInfoService()
	receiveMsgInfoEntity := &entity.RsvMsgInfoEntity{
		Id:         util.GetSnowFlakeIdStr(ctx),
		Type:       "receive",
		WaId:       waId,
		SourceWaId: waId,
		MsgType:    constant.ReceiveMsg,
		Msg:        string(reqAnyEncode),
		MsgStatus:  constant.NXMsgStatusReceive,
		CreatedAt:  msgRecTime,
		UpdatedAt:  msgRecTime,
	}
	// 新增接收的消息
	err = receiveMsgInfoService.InsertMsgInfo(ctx, receiveMsgInfoEntity)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，新增接收信息失败,req：%v,err:%v", constant.MethodUserAttendMethodInsertMsgInfo, req, err))
		return constant.MethodInsertMsgInfoReturnFail, errors.New("MethodInsertMsgInfo receive messages insert error")
	}

	if sendRallyCode == "" || len(sendRallyCode) < 5 {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，messages中Text的body中去除前缀后的字符串不符合要求,req：%v", constant.MethodUserAttendMethodInsertMsgInfo, req))
		return constant.MethodInsertMsgInfoReturnFail, errors.New("MethodInsertMsgInfo messages‘s text’s body not contains prefix,or empty after the prefix is removed")
	}

	sendChannel := sendRallyCode[0:1]
	sendLanguage := sendRallyCode[1:3]
	sendGeneration := sendRallyCode[3:5]
	sendIdentificationCode := sendRallyCode[5:]

	if !isHelp && (sendGeneration != constant.Generation01 || sendIdentificationCode != constant.FirstIdentificationCode) {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，初代发送参与活动活动码格式不对,waid:%v,err：%v", constant.MethodUserAttendMethodInsertMsgInfo, waId, err))
		return constant.MethodInsertMsgInfoReturnFail, errors.New("RallyCode is error")
	}

	// 校验waId是否正确
	if !util.StartsWithPrefix(waId, config.ApplicationConfig.Activity.WaIdPrefixList) {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，wa_id前缀不是允许国码,req：%v", constant.MethodUserAttendMethodInsertMsgInfo, req))
		// 发送不允许参与消息
		msgInfoEntity := &entity.MsgInfoEntityV2{
			Id:         util.GetSnowFlakeIdStr(ctx),
			Type:       "send",
			WaId:       waId,
			SourceWaId: waId,
			MsgType:    constant.CannotAttendActivityMsg,
		}
		sendNxListParamsDto, err := CannotAttendActivity2NX(ctx, msgInfoEntity, sendChannel, sendLanguage)
		if nil != err {
			return true, errors.New("send nx msg is error")
		}
		if !isExist {
			session.Commit()
		}
		_, nxErr := SendMsgList2NX(ctx, sendNxListParamsDto)
		if nxErr != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，发消息到牛信云失败,err：%v", constant.MethodUserAttendMethodInsertMsgInfo, nxErr))
			return true, nxErr
		}
		return false, nil
	}

	generation := sendGeneration
	if sendIdentificationCode != constant.FirstIdentificationCode {
		// 非初代，迭代
		generation, err = util.GetNewGeneration(generation)
		if err != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，升级传播代数失败,waid:%v,err：%v", constant.MethodUserAttendMethodInsertMsgInfo, waId, err))
			return constant.MethodInsertMsgInfoReturnFail, errors.New("generation convert is error")
		}
	}

	// 需要预约，才要判断数据库是否有值，并且是预约活动才能助力
	if config.ApplicationConfig.Activity.NeedSubscribe {
		if dbUserAttendInfo.Id > 0 {
			if isHelp {
				if dbUserAttendInfo.AttendStatus != constant.AttendStatusAttend {
					// 立刻重发
					ginCtx1 := gin.Context{}
					ReSendMsgByWaId(&ginCtx1, dbUserAttendInfo.WaId, true)

					// 走助力逻辑
					param := &request.HelpParam{
						WaId:      waId,
						IsHelp:    isHelp,
						RallyCode: sendRallyCode,
					}
					// 另开一个上下文
					err = u.Help(ginCtx, param)
					if err != nil {
						return constant.MethodInsertMsgInfoReturnFail, errors.New("help is error")
					}
					if !isExist {
						session.Commit()
					}
					return false, nil
				} else {
					logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，用户状态不是attend不能重复参与,waid:%v", constant.MethodUserAttendMethodInsertMsgInfo, waId))
					return constant.MethodInsertMsgInfoReturnFail, errors.New("convert json is error")
				}
			} else {
				logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，之前参与过活动不再发送消息,waid:%v", constant.MethodUserAttendMethodInsertMsgInfo, waId))
				return constant.MethodInsertMsgInfoReturnFail, errors.New("repeat attend activity")
			}
		}
	} else {
		if isHelp {
			if dbUserAttendInfo.Id <= 0 {
				// 新增
				extraMap := map[string]string{
					"UserAttendInfoMessage": string(reqAnyEncode),
				}
				bytes, err := json.NewEncoder().Encode(extraMap)
				if err != nil {
					logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，转换extra为json失败,waid:%v", constant.MethodUserAttendMethodInsertMsgInfo, waId))
					return constant.MethodInsertMsgInfoReturnFail, errors.New("convert json is error")
				}

				endTime := util.GetTimeOfAfterDays(1, msgRecTime)
				sendRenewFreeAt := util.GetSendRenewMsgTime(1, msgRecTime)
				// 之前未参与过,落库
				dbUserAttendInfo = entity.UserAttendInfoEntityV2{
					Channel:            sendChannel,
					Language:           sendLanguage,
					Generation:         generation,
					WaId:               waId,
					Extra:              string(bytes),
					UserNickname:       userNickName,
					NewestFreeStartAt:  msgRecTime.Unix(),
					NewestFreeEndAt:    endTime,
					SendRenewFreeAt:    sendRenewFreeAt.Unix(),
					IsSendRenewFreeMsg: constant.RenewFreeUnSend,
					AttendAt:           time.Now().Unix(),
				}

				//if !config.ApplicationConfig.Activity.NeedSubscribe {
				//	// 直接开团
				//	nowCustomTime := util.GetNowCustomTime()
				//	sendClusteringAt := util.GetSendClusteringTime(5, nowCustomTime)
				//
				//	dbUserAttendInfo.AttendStatus = constant.AttendStatusStartGroup
				//	dbUserAttendInfo.StartGroupAt = nowCustomTime
				//	dbUserAttendInfo.NewestHelpAt = nowCustomTime
				//	dbUserAttendInfo.IsSendClusteringMsg = constant.ClusteringUnSend
				//	dbUserAttendInfo.SendClusteringAt = sendClusteringAt
				//}

				_, err = u.userAttendInfoMapper.InsertSelective(&session, dbUserAttendInfo)
				if nil != err {
					logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，新增用户参与信息失败,waId:%v,err：%v", constant.MethodUserAttendMethodInsertMsgInfo, waId, err))
					return true, errors.New("database is error")
				}

				insertEntity, err := u.userAttendInfoMapper.SelectByWaIdBySession(&session, waId)
				if nil != err {
					logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，查询用户参与信息失败,waId:%v,err：%v", constant.MethodUserAttendMethodInsertMsgInfo, waId, err))
					return true, errors.New("database is error")
				}

				base32 := util.ToBase32(insertEntity.Id)
				upUserAttendInfo := entity.UserAttendInfoEntityV2{
					Id:                 insertEntity.Id,
					IdentificationCode: base32,
					RallyCode:          sendChannel + sendLanguage + generation + base32,
					UpdatedAt:          util.GetNowCustomTime(),
				}

				_, err = u.userAttendInfoMapper.UpdateByPrimaryKeySelective(&session, upUserAttendInfo)
				if nil != err {
					logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，修改用户参与信息失败,waId:%v,err：%v", constant.MethodUserAttendMethodInsertMsgInfo, waId, err))
					return true, errors.New("database is error")
				}
			}
			if !isExist {
				session.Commit()
			}
			// 立刻重发
			ginCtx1 := gin.Context{}
			ReSendMsgByWaId(&ginCtx1, dbUserAttendInfo.WaId, true)

			// 走助力逻辑
			param := &request.HelpParam{
				WaId:      waId,
				IsHelp:    isHelp,
				RallyCode: sendRallyCode,
			}
			// 另开一个上下文
			err = u.Help(ginCtx, param)
			if err != nil {
				return constant.MethodInsertMsgInfoReturnFail, errors.New("help is error")
			}

			return false, nil

		}
	}

	switch activityStatus {
	case constant.ATStatusEnd:
		// 发送结束期-不能开团消息
		msgInfoEntity := &entity.MsgInfoEntityV2{
			Id:         util.GetSnowFlakeIdStr(ctx),
			Type:       "send",
			WaId:       waId,
			SourceWaId: waId,
			MsgType:    constant.EndCanNotStartGroupMsg,
		}

		msgList := make([]*dto.SendNxListParamsDto, 0)
		if isHelp {
			msgInfoEntity.MsgType = constant.EndCanNotHelpMsg
			msgList, err = EndCanNotHelpMsg(ctx, msgInfoEntity, sendLanguage)
			if err != nil {
				logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，发送结束期-不能助力消息失败,waId:%v", constant.MethodHelp, waId))
				return true, err
			}
		} else {
			msgList, err = EndCanNotStartGroupMsg(ctx, msgInfoEntity, sendLanguage)
			if err != nil {
				logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，发送结束期-不能开团消息失败,waId:%v", constant.MethodHelp, waId))
				return true, err
			}
		}

		if !isExist {
			session.Commit()
		}
		_, nxErr := SendMsgList2NX(ctx, msgList)
		if nxErr != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，发结束期-不能开团消息到牛信云失败,err：%v", constant.MethodUserAttendMethodInsertMsgInfo, nxErr))
			return true, nxErr
		}
		return false, nil
	case constant.ATStatusBuffer:
		// todo 1211 缓冲期重复开团是否发送消息,已完成
		if !config.ApplicationConfig.Activity.NeedSubscribe || (dbUserAttendInfo.Id > 0 && dbUserAttendInfo.AttendStatus == constant.AttendStatusAttend) {
			if dbUserAttendInfo.Id > 0 && dbUserAttendInfo.AttendStatus != constant.AttendStatusAttend && !isHelp {
				//表示重复开团
				msgInfoEntity := &entity.MsgInfoEntityV2{
					Id:         util.GetSnowFlakeIdStr(ctx),
					Type:       "send",
					WaId:       dbUserAttendInfo.WaId,
					SourceWaId: dbUserAttendInfo.WaId,
					MsgType:    constant.StartGroupMsg,
				}
				helpNameList := make([]entity.UserAttendInfoEntityV2, 0)
				sendNxListParamsDtoList, err := StartGroupMsg2NX(ctx, msgInfoEntity, dbUserAttendInfo, helpNameList, false)
				if err != nil {
					logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，发送开团消息失败,rallyCode:%v", constant.MethodUserAttendMethodInsertMsgInfo, dbUserAttendInfo.RallyCode))
					return true, err
				}

				if !isExist {
					session.Commit()
				}

				_, nxErr := SendMsgList2NX(ctx, sendNxListParamsDtoList)
				if nxErr != nil {
					logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，发消息到牛信云失败,err：%v", constant.MethodUserAttendMethodInsertMsgInfo, nxErr))
					return false, nxErr
				}
				return true, nil
			}
			// 发送不能开团消息
			msgInfoEntity := &entity.MsgInfoEntityV2{
				Id:         util.GetSnowFlakeIdStr(ctx),
				Type:       "send",
				WaId:       waId,
				SourceWaId: waId,
				MsgType:    constant.FounderCanNotStartGroupMsg,
			}
			sendNxListParamsDto, err := FounderCanNotStartGroupMsg(ctx, msgInfoEntity, sendLanguage)
			if err != nil {
				logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，发送主态不能开团消息失败,waId:%v", constant.MethodHelp, waId))
				return true, err
			}
			if !isExist {
				session.Commit()
			}
			_, nxErr := SendMsgList2NX(ctx, sendNxListParamsDto)
			if nxErr != nil {
				logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，发主态不能开团消息到牛信云失败,err：%v", constant.MethodUserAttendMethodInsertMsgInfo, nxErr))
				return true, nxErr
			}
			return false, nil
		}

	}

	extraMap := map[string]string{
		"UserAttendInfoMessage": string(reqAnyEncode),
	}
	bytes, err := json.NewEncoder().Encode(extraMap)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，转换extra为json失败,waid:%v", constant.MethodUserAttendMethodInsertMsgInfo, waId))
		return constant.MethodInsertMsgInfoReturnFail, errors.New("convert json is error")
	}

	endTime := util.GetTimeOfAfterDays(1, msgRecTime)
	sendRenewFreeAt := util.GetSendRenewMsgTime(1, msgRecTime)

	if dbUserAttendInfo.Id > 0 {
		// todo 1211 重复开团消息。开团消息 已完成
		msgInfoEntity := &entity.MsgInfoEntityV2{
			Id:         util.GetSnowFlakeIdStr(ctx),
			Type:       "send",
			WaId:       dbUserAttendInfo.WaId,
			SourceWaId: dbUserAttendInfo.WaId,
			MsgType:    constant.StartGroupMsg,
		}
		helpNameList := make([]entity.UserAttendInfoEntityV2, 0)
		sendNxListParamsDtoList, err := StartGroupMsg2NX(ctx, msgInfoEntity, dbUserAttendInfo, helpNameList, false)
		if err != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，发送开团消息失败,rallyCode:%v", constant.MethodUserAttendMethodInsertMsgInfo, dbUserAttendInfo.RallyCode))
			return true, err
		}

		if !isExist {
			session.Commit()
		}

		_, nxErr := SendMsgList2NX(ctx, sendNxListParamsDtoList)
		if nxErr != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，发消息到牛信云失败,err：%v", constant.MethodUserAttendMethodInsertMsgInfo, nxErr))
			return true, nxErr
		}
		return false, nil
	}

	// 之前未参与过,落库
	dbUserAttendInfo = entity.UserAttendInfoEntityV2{
		Channel:            sendChannel,
		Language:           sendLanguage,
		Generation:         generation,
		WaId:               waId,
		Extra:              string(bytes),
		UserNickname:       userNickName,
		NewestFreeStartAt:  msgRecTime.Unix(),
		NewestFreeEndAt:    endTime,
		SendRenewFreeAt:    sendRenewFreeAt.Unix(),
		IsSendRenewFreeMsg: constant.RenewFreeUnSend,
		AttendAt:           time.Now().Unix(),
	}

	if !config.ApplicationConfig.Activity.NeedSubscribe {
		// 直接开团
		nowCustomTime := util.GetNowCustomTime()
		//todo 1212测试 这里不需要一定是整点吧
		sendClusteringAt := util.GetSendClusteringTime(5, msgRecTime)

		dbUserAttendInfo.AttendStatus = constant.AttendStatusStartGroup
		dbUserAttendInfo.IsSendStartGroupMsg = constant.ClusteringSend
		dbUserAttendInfo.StartGroupAt = nowCustomTime
		dbUserAttendInfo.NewestHelpAt = nowCustomTime
		dbUserAttendInfo.IsSendClusteringMsg = constant.ClusteringUnSend
		dbUserAttendInfo.SendClusteringAt = sendClusteringAt.Unix()
	}

	_, err = u.userAttendInfoMapper.InsertSelective(&session, dbUserAttendInfo)
	if nil != err {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，新增用户参与信息失败,waId:%v,err：%v", constant.MethodUserAttendMethodInsertMsgInfo, waId, err))
		return true, errors.New("database is error")
	}

	insertEntity, err := u.userAttendInfoMapper.SelectByWaIdBySession(&session, waId)
	if nil != err {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，查询用户参与信息失败,waId:%v,err：%v", constant.MethodUserAttendMethodInsertMsgInfo, waId, err))
		return true, errors.New("database is error")
	}

	base32 := util.ToBase32(insertEntity.Id)
	upUserAttendInfo := entity.UserAttendInfoEntityV2{
		Id:                 insertEntity.Id,
		IdentificationCode: base32,
		RallyCode:          sendChannel + sendLanguage + generation + base32,
		UpdatedAt:          util.GetNowCustomTime(),
	}

	_, err = u.userAttendInfoMapper.UpdateByPrimaryKeySelective(&session, upUserAttendInfo)
	if nil != err {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，修改用户参与信息失败,waId:%v,err：%v", constant.MethodUserAttendMethodInsertMsgInfo, waId, err))
		return true, errors.New("database is error")
	}

	var sendNxListParamsDtoList []*dto.SendNxListParamsDto
	if !config.ApplicationConfig.Activity.NeedSubscribe {
		// 发送开团消息
		msgInfoEntity := &entity.MsgInfoEntityV2{
			Id:         util.GetSnowFlakeIdStr(ctx),
			Type:       "send",
			WaId:       dbUserAttendInfo.WaId,
			SourceWaId: dbUserAttendInfo.WaId,
			MsgType:    constant.StartGroupMsg,
		}
		dbUserAttendInfo.RallyCode = upUserAttendInfo.RallyCode
		//helpNameList, err := u.helpInfoMapper.SelectHelpNameByRallyCode(&session, config.ApplicationConfig.Activity.Id, dbUserAttendInfo.RallyCode)
		//if nil != err {
		//	logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，根据开团人rallyCode查询助力人昵称失败,rallyCode：%v,err：%v", constant.MethodUserAttendMethodInsertMsgInfo, dbUserAttendInfo.RallyCode, err))
		//	return true, errors.New("database is error")
		//}
		helpNameList := make([]entity.UserAttendInfoEntityV2, 0)
		sendNxListParamsDtoList, err = StartGroupMsg2NX(ctx, msgInfoEntity, dbUserAttendInfo, helpNameList, false)
		if err != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，发送开团消息失败,rallyCode:%v", constant.MethodUserAttendMethodInsertMsgInfo, dbUserAttendInfo.RallyCode))
			return true, err
		}
	} else {
		msgType := constant.HelpTaskSingleStartMsg
		if generation == constant.Generation01 {
			msgType = constant.ActivityTaskMsg
		}
		msgInfoEntity := &entity.MsgInfoEntityV2{
			Id:         util.GetSnowFlakeIdStr(ctx),
			Type:       "send",
			WaId:       waId,
			SourceWaId: waId,
			MsgType:    msgType,
		}

		if generation == constant.Generation01 {
			sendRallyCode = upUserAttendInfo.RallyCode
		}
		param := &request.HelpParam{
			WaId:      waId,
			IsHelp:    isHelp,
			RallyCode: sendRallyCode,
		}

		sendNxListParamsDtoList, err = ActivityTask2NX(ctx, msgInfoEntity, sendLanguage, sendChannel, param)
		if nil != err {
			return true, errors.New("send nx msg is error")
		}

	}
	if !isExist {
		session.Commit()
	}

	_, nxErr := SendMsgList2NX(ctx, sendNxListParamsDtoList)
	if nxErr != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，发消息到牛信云失败,err：%v", constant.MethodUserAttendMethodInsertMsgInfo, nxErr))
		return true, nxErr
	}
	return false, nil
}

func (u UserAttendInfoService) Help(ctx *gin.Context, reqParam *request.HelpParam) error {

	if reqParam.WaId == "" {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，reqParam参数校验失败,param:%v", constant.MethodHelp, reqParam))
		return errors.New("param is invalid")
	}
	//template := redis_template.NewRedisTemplate()

	beHelpedUserInfo, err := u.selectByRallyCode(ctx, reqParam.RallyCode)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，查询被助力人报错,活动id:%v,reqParam.RallyCode:%v,err：%v", constant.MethodUserAttendMethodInsertMsgInfo, config.ApplicationConfig.Activity.Id, reqParam.RallyCode, err))
		return errors.New("database is error")
	}
	if beHelpedUserInfo.Id <= 0 {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，被助力人不存在,活动id:%v,reqParam.RallyCode:%v,err：%v", constant.MethodUserAttendMethodInsertMsgInfo, config.ApplicationConfig.Activity.Id, reqParam.RallyCode, err))
		return errors.New("database is error")
	}

	if reqParam.IsHelp {
		if beHelpedUserInfo.WaId == reqParam.WaId {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，被助力人和助力人是一个,活动id:%v,waId:%v,err：%v", constant.MethodUserAttendMethodInsertMsgInfo, config.ApplicationConfig.Activity.Id, beHelpedUserInfo.WaId, err))
			return errors.New("database is error")
		}
	}

	session, isExist, err := txUtil.GetTransaction(ctx)
	if nil != err {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，创建事务失败,err：%v", constant.MethodHelp, err))
		return errors.New("database is error")
	}
	if !isExist {
		defer func() {
			session.Rollback()
			session.Close()
		}()
	}

	userAttendInfoEntity, err := u.userAttendInfoMapper.SelectByWaIdBySession(&session, reqParam.WaId)
	if nil != err {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，根据waId查询userAttendInfo失败,WaId：%v,err：%v", constant.MethodHelp, reqParam.WaId, err))
		return errors.New("database is error")
	}

	if userAttendInfoEntity.Id <= 0 {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，根据waId查询userAttendInfo失败,WaId：%v,err：%v", constant.MethodHelp, reqParam.WaId, err))
		return errors.New("database is error")
	}

	// 校验waId是否正确
	if !util.StartsWithPrefix(reqParam.WaId, config.ApplicationConfig.Activity.WaIdPrefixList) {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，wa_id前缀不是允许国码,req：%v", constant.MethodHelp, reqParam))
		// 发送不允许参与消息
		msgInfoEntity := &entity.MsgInfoEntityV2{
			Id:         util.GetSnowFlakeIdStr(ctx),
			Type:       "send",
			WaId:       reqParam.WaId,
			SourceWaId: reqParam.WaId,
			MsgType:    constant.CannotAttendActivityMsg,
		}
		sendNxListParamsDto, err := CannotAttendActivity2NX(ctx, msgInfoEntity, userAttendInfoEntity.Channel, userAttendInfoEntity.Language)
		if nil != err {
			return errors.New("send nx msg is error")
		}
		if !isExist {
			session.Commit()
		}
		_, nxErr := SendMsgList2NX(ctx, sendNxListParamsDto)
		if nxErr != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，发消息到牛信云失败,err：%v", constant.MethodUserAttendMethodInsertMsgInfo, nxErr))
			return nxErr
		}
		return nil
	}

	// 判断是否可以助力、开团
	activityEntity, err := u.activityInfoMapper.SelectByPrimaryKey(config.ApplicationConfig.Activity.Id)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，根据活动id查询activityInfo报错,活动id:%v,err：%v", constant.MethodHelp, config.ApplicationConfig.Activity.Id, err))
		return errors.New("database is error")
	}
	var sendNxListParamsDtoList []*dto.SendNxListParamsDto

	switch activityEntity.ActivityStatus {
	case constant.ATStatusEnd:
		// 发送结束期-不能助力消息
		msgInfoEntity := &entity.MsgInfoEntityV2{
			Id:         util.GetSnowFlakeIdStr(ctx),
			Type:       "send",
			WaId:       userAttendInfoEntity.WaId,
			SourceWaId: userAttendInfoEntity.WaId,
			MsgType:    constant.EndCanNotHelpMsg,
		}
		sendNxListParamsDto, err := EndCanNotHelpMsg(ctx, msgInfoEntity, userAttendInfoEntity.Language)
		if err != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，发送结束期-不能助力消息失败,rallyCode:%v", constant.MethodHelp, userAttendInfoEntity.RallyCode))
			return err
		}
		sendNxListParamsDtoList = append(sendNxListParamsDtoList, sendNxListParamsDto...)
	case constant.ATStatusUnStart:
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，活动尚未开始或已结束,活动id:%v", constant.MethodHelp, config.ApplicationConfig.Activity.Id))
		return errors.New("activity is error")
	case constant.ATStatusBuffer:
		// 缓冲期，不许开团，允许助力
		// 正常发送，存储数据库，若不在免费期间等待11点统一发。
		if userAttendInfoEntity.AttendStatus == constant.AttendStatusAttend {
			// 发送不能开团消息
			msgInfoEntity := &entity.MsgInfoEntityV2{
				Id:         util.GetSnowFlakeIdStr(ctx),
				Type:       "send",
				WaId:       userAttendInfoEntity.WaId,
				SourceWaId: userAttendInfoEntity.WaId,
				MsgType:    constant.CanNotStartGroupMsg,
			}
			sendNxListParamsDto, err := CanNotStartGroupMsg(ctx, msgInfoEntity, userAttendInfoEntity.Language)
			if err != nil {
				logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，发送不能开团消息失败,rallyCode:%v", constant.MethodHelp, userAttendInfoEntity.RallyCode))
				return err
			}
			sendNxListParamsDtoList = append(sendNxListParamsDtoList, sendNxListParamsDto...)
			//// 红包
			//sendNxListParamsDto2, err := u.freeCdkSendService.FreeCdkSave(ctx, userAttendInfoEntity)
			//if err != nil {
			//	logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，构建免费cdk红包失败,waId:%v", constant.MethodHelp, userAttendInfoEntity.WaId))
			//	return err
			//}
			//sendNxListParamsDtoList = append(sendNxListParamsDtoList, sendNxListParamsDto2...)
		}

		//else {
		//	helpNameList, err := u.helpInfoMapper.SelectHelpNameByRallyCode(&session, config.ApplicationConfig.Activity.Id, userAttendInfoEntity.RallyCode)
		//	if nil != err {
		//		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，根据开团人rallyCode查询助力人昵称失败,rallyCode：%v,err：%v", constant.MethodHelp, userAttendInfoEntity.RallyCode, err))
		//		return errors.New("database is error")
		//	}
		//	// 发送开团消息
		//	msgInfoEntity := &entity.MsgInfoEntityV2{
		//		Id:         util.GetSnowFlakeIdStr(ctx),
		//		Type:       "send",
		//		WaId:       userAttendInfoEntity.WaId,
		//		SourceWaId: userAttendInfoEntity.WaId,
		//		ActivityId: config.ApplicationConfig.Activity.Id,
		//		MsgType:    constant.StartGroupMsg,
		//	}
		//	sendNxListParamsDto, err := StartGroupMsg2NX(ctx, msgInfoEntity, userAttendInfoEntity, helpNameList, activityEntity)
		//	if err != nil {
		//		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，发送开团消息失败,rallyCode:%v", constant.MethodHelp, userAttendInfoEntity.RallyCode))
		//		return err
		//	}
		//	sendNxListParamsDtoList = append(sendNxListParamsDtoList, sendNxListParamsDto...)
		//}
	case constant.ATStatusStarted:
		if userAttendInfoEntity.AttendStatus == constant.AttendStatusAttend {
			nowCustomTime := util.GetNowCustomTime()
			sendClusteringAt := util.GetSendClusteringTime(5, nowCustomTime)
			// 运行期
			updateUserAttendInfoEntity := entity.UserAttendInfoEntityV2{
				Id:                  userAttendInfoEntity.Id,
				AttendStatus:        constant.AttendStatusStartGroup,
				StartGroupAt:        nowCustomTime,
				NewestHelpAt:        nowCustomTime,
				IsSendClusteringMsg: constant.ClusteringUnSend,
				SendClusteringAt:    sendClusteringAt.Unix(),
				IsSendStartGroupMsg: constant.ClusteringSend,
			}
			_, err = u.userAttendInfoMapper.UpdateByPrimaryKeySelective(&session, updateUserAttendInfoEntity)
			if nil != err {
				logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，根据更新开团信息失败,WaId：%v,err：%v", constant.MethodHelp, reqParam.WaId, err))
				return errors.New("database is error")
			}
		}

		helpInfoList, err := u.helpInfoMapper.SelectListByRallyCode(&session, userAttendInfoEntity.RallyCode)
		if nil != err {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，根据开团人rallyCode查询助力人信息失败,rallyCode：%v,err：%v", constant.MethodHelp, userAttendInfoEntity.RallyCode, err))
			return errors.New("database is error")
		}
		waIds := make([]string, len(helpInfoList))
		for i, helpInfo := range helpInfoList {
			waIds[i] = helpInfo.WaId
		}
		var userAttendInfoList []entity.UserAttendInfoEntityV2
		if len(waIds) > 0 {
			userAttendInfoList, err = u.userAttendInfoMapper.SelectListByWaIdsWithSession(&session, waIds)
			if nil != err {
				logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，根据助力人waId查询助力人昵称失败,rallyCode：%v,err：%v", constant.MethodHelp, userAttendInfoEntity.RallyCode, err))
				return errors.New("database is error")
			}
		}

		// 发送开团消息
		msgInfoEntity := &entity.MsgInfoEntityV2{
			Id:         util.GetSnowFlakeIdStr(ctx),
			Type:       "send",
			WaId:       userAttendInfoEntity.WaId,
			SourceWaId: userAttendInfoEntity.WaId,
			MsgType:    constant.StartGroupMsg,
		}

		logTracing.LogPrintf(ctx, logTracing.NormalLogFmt, fmt.Sprintf("开始构建开团消息,waId:%v", userAttendInfoEntity.WaId))

		sendNxListParamsDto, err := StartGroupMsg2NX(ctx, msgInfoEntity, userAttendInfoEntity, userAttendInfoList, reqParam.IsHelp)
		if err != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，发送开团消息失败,rallyCode:%v", constant.MethodHelp, userAttendInfoEntity.RallyCode))
			return err
		}
		logTracing.LogPrintf(ctx, logTracing.NormalLogFmt, fmt.Sprintf("结束构建开团消息,waId:%v", userAttendInfoEntity.WaId))

		// 发送红包消息
		//u.freeCdkSendService.FreeCdkSave()
		sendNxListParamsDtoList = append(sendNxListParamsDtoList, sendNxListParamsDto...)

		logTracing.LogPrintf(ctx, logTracing.NormalLogFmt, fmt.Sprintf("开始构建freeCdk消息,waId:%v", userAttendInfoEntity.WaId))
		// 红包
		sendNxListParamsDto2, err := u.freeCdkSendService.FreeCdkSave(ctx, userAttendInfoEntity)
		if err != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，构建免费cdk红包失败,waId:%v", constant.MethodHelp, userAttendInfoEntity.WaId))
			return err
		}
		logTracing.LogPrintf(ctx, logTracing.NormalLogFmt, fmt.Sprintf("结束构建freeCdk消息,waId:%v", userAttendInfoEntity.WaId))

		sendNxListParamsDtoList = append(sendNxListParamsDtoList, sendNxListParamsDto2...)
	default:
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，不支持的状态,活动id:%v", constant.MethodHelp, config.ApplicationConfig.Activity.Id))
		return errors.New("activity status error")
	}

	switch activityEntity.ActivityStatus {
	case constant.ATStatusBuffer:
		fallthrough
	case constant.ATStatusStarted:
		if reqParam.IsHelp {
			logTracing.LogPrintf(ctx, logTracing.NormalLogFmt, fmt.Sprintf("开始处理助力消息,waId:%v", userAttendInfoEntity.WaId))

			// 助力
			sendNxListParamsDto, err := u.helpHandler(ctx, activityEntity, userAttendInfoEntity, reqParam.RallyCode)
			if err != nil {
				return errors.New("HelpHandler is error")
			}
			logTracing.LogPrintf(ctx, logTracing.NormalLogFmt, fmt.Sprintf("结束处理助力消息,waId:%v", userAttendInfoEntity.WaId))

			sendNxListParamsDtoList = append(sendNxListParamsDtoList, sendNxListParamsDto...)
		}
	}

	if !isExist {
		session.Commit()
	}

	// 延迟
	helpGoroutinePool.Execute(func(param interface{}) {
		dto, ok := param.([]*dto.SendNxListParamsDto) // 断言u是User类型
		if !ok {
			logTracing.LogPrintf(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],断言发生错误，reqParam:%v", constant.MethodHelp, reqParam))
		}
		logTracing.LogPrintf(ctx, logTracing.WebHandleLogFmt, fmt.Sprintf("方法[%s],helpGoroutinePool协程池执行任务开始休眠八秒,线程池channel长度:%v,reqParam:%v", constant.MethodHelp, len(helpGoroutinePool.Ch), reqParam))
		time.Sleep(8 * time.Second)
		logTracing.LogPrintf(ctx, logTracing.WebHandleLogFmt, fmt.Sprintf("方法[%s],helpGoroutinePool协程池执行任务休眠八秒结束，任务正式开始，reqParam:%v", constant.MethodHelp, reqParam))
		_, nxErr := SendMsgList2NXHelpTimeOut(ctx, dto)
		if nxErr != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，发消息到牛信云失败,reqParam:%v,err：%v", constant.MethodHelp, reqParam, nxErr))
			return
		}
		logTracing.LogPrintf(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],helpGoroutinePool协程池执行任务结束，reqParam:%v", constant.MethodHelp, reqParam))
	}, sendNxListParamsDtoList)
	//_, nxErr := SendMsgList2NXHelpTimeOut(ctx, sendNxListParamsDtoList)
	//if nxErr != nil {
	//	logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，发消息到牛信云失败,err：%v", constant.MethodHelp, nxErr))
	//	return nil
	//}
	return nil
}

func (u UserAttendInfoService) helpHandler(ctx *gin.Context, activityEntity entity.ActivityInfoEntity, userAttendInfoEntity entity.UserAttendInfoEntityV2, rallyCode string) ([]*dto.SendNxListParamsDto, error) {
	waId := userAttendInfoEntity.WaId
	beHelpedUserInfo, err := u.selectByRallyCode(ctx, rallyCode)
	if err != nil {
		return nil, errors.New("database is error")
	}

	// 八人应该是不发送助力消息
	if constant.AttendStatusEightOver == beHelpedUserInfo.AttendStatus {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，db判断用户已助力满不需要再次助力,返回空", constant.MethodHelp))
		return make([]*dto.SendNxListParamsDto, 0), nil
		//logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，助力已成功不能再助力,活动id:%v,rallyCode:%v", constant.MethodHelp, config.ApplicationConfig.Activity.Id, rallyCode))
		//return nil, errors.New("help is over")
	}

	oldHelpInfoEntity, err := u.helpInfoMapper.SelectByWaId(waId)
	if nil != err {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，查询助力人助力信息,err：%v", constant.MethodHelp, err))
		return nil, errors.New("database is error")
	}
	if oldHelpInfoEntity.Id > 0 {
		ctx = &gin.Context{}
		session, isExist, err := txUtil.GetTransaction(ctx)
		if nil != err {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，创建事务失败,err：%v", constant.MethodHelp, err))
			return nil, errors.New("database is error")
		}
		if !isExist {
			defer func() {
				session.Rollback()
				session.Close()
			}()
		}
		// todo 1211 重复助力、超出助力次数,已完成
		oldUserAttendInfoEntity, err := u.userAttendInfoMapper.SelectByRallyCode(oldHelpInfoEntity.RallyCode)
		if nil != err {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，查询开团人信息,err：%v", constant.MethodHelp, err))
			return nil, errors.New("database is error")
		}
		//被助力人
		beHelpedNickname := oldUserAttendInfoEntity.UserNickname
		// 发送不允许参与消息
		msgInfoEntity := &entity.MsgInfoEntityV2{
			Id:         util.GetSnowFlakeIdStr(ctx),
			Type:       "send",
			WaId:       waId,
			SourceWaId: waId,
			MsgType:    constant.RepeatHelpMsg,
		}

		if !isExist {
			session.Commit()
		}

		sendNxListParamsDto, err := RepeatHelpMsg2NX(ctx, msgInfoEntity, beHelpedNickname, userAttendInfoEntity.Language)
		if nil != err {
			return nil, errors.New("send nx msg is error")
		}
		_, nxErr := SendMsgList2NX(ctx, sendNxListParamsDto)
		if nxErr != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，发消息到牛信云失败,err：%v", constant.MethodUserAttendMethodInsertMsgInfo, nxErr))
			return nil, errors.New("SendMsgList2NX error")
		}
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，之前助力过不能再次助力,Activity：%v，rallyCode：%v，waId:%v", constant.MethodHelp, config.ApplicationConfig.Activity.Id, rallyCode, waId))
		return nil, errors.New("database is error")
	}

	//redPacketCode := ""
	//if constant.RedPacketStatusReady == beHelpedUserInfo.RedPacketStatus {
	//	// 红包发放信息
	//	template := redis_template.NewRedisTemplate()
	//	redPacketKey := constant.GetRedPacketKey(config.ApplicationConfig.Activity.Id)
	//	redPacket, err := template.BRPop(ctx, redPacketKey, constant.RedPacketTimeOut)
	//	if err != nil {
	//		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，获取redPacket失败,waId:%v", constant.MethodHelp, beHelpedUserInfo.WaId))
	//		return errors.New("redis is error")
	//	}
	//	redPacketCode = redPacket[0]
	//	updateEntity := entity.UserAttendInfoEntityV2{
	//		Id:              beHelpedUserInfo.Id,
	//		RedPacketCode:   redPacketCode,
	//		RedPacketStatus: constant.RedPacketStatusSend,
	//		RedPacketSendAt: util.GetNowCustomTime(),
	//	}
	//	_, err = u.userAttendInfoMapper.UpdateByPrimaryKeySelective(&session, updateEntity)
	//	if nil != err {
	//		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，更新红包码失败,err：%v", constant.MethodHelp, err))
	//		return errors.New("database is error")
	//	}
	//}
	values := &dto.HelpCacheDto{
		WaId:         waId,
		UserNickname: userAttendInfoEntity.UserNickname,
		RallyCode:    rallyCode,
	}
	key := constant.GetHelpInfoCacheKey(config.ApplicationConfig.Activity.Id, rallyCode)
	newCount, helpCacheDtoList, err := AddHelpInfoCache(constant.MethodHelp, key, values)
	if nil != err {
		return nil, errors.New("redis is error")
	}
	if newCount > config.ApplicationConfig.Activity.Stage3Award.HelpNum {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，redis判断用户已助力满不需要再次助力,返回空", constant.MethodHelp))
		return make([]*dto.SendNxListParamsDto, 0), nil
	}
	defer func() {
		if e := recover(); e != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，发生painc异常，减去新增的助力缓存，key：%v", constant.MethodHelp, key))
			RemoveHelpInfoCache(constant.MethodHelp, key)
			panic(e)
		} else if err != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，发生错误，减去新增的助力缓存，key：%v", constant.MethodHelp, key))
			RemoveHelpInfoCache(constant.MethodHelp, key)
		}
	}()

	//rallyCodeBeHelpCount, err := u.helpInfoMapper.CountByRallyCode(rallyCode)
	//if nil != err {
	//	logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，查询助力总数失败,err：%v", constant.MethodHelp, err))
	//	return nil, errors.New("database is error")
	//}
	//helpNameList, err := u.helpInfoMapper.SelectHelpNameByRallyCode(&session, rallyCode)
	//if nil != err || len(helpNameList) <= 0 {
	//	logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，根据被助力人rallyCode查询助力人昵称失败,rallyCode：%v,err：%v", constant.MethodHelp, rallyCode, err))
	//	return nil, errors.New("database is error")
	//}

	session, isExist, err := txUtil.GetTransaction(ctx)
	if nil != err {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，创建事务失败,err：%v", constant.MethodHelp, err))
		return nil, errors.New("database is error")
	}
	if !isExist {
		defer func() {
			session.Rollback()
			session.Close()
		}()
	}

	logTracing.LogPrintf(ctx, logTracing.NormalLogFmt, fmt.Sprintf("开始新增助力信息到数据库,waId:%v", waId))

	helpInfoEntity := entity.HelpInfoEntityV2{
		WaId:       waId,
		RallyCode:  rallyCode,
		HelpStatus: "efficien",
	}
	_, err = u.helpInfoMapper.InsertSelective(&session, helpInfoEntity)
	if nil != err {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，新增助力失败,err：%v", constant.MethodHelp, err))
		return nil, errors.New("database is error")
	}
	logTracing.LogPrintf(ctx, logTracing.NormalLogFmt, fmt.Sprintf("新增助力信息到数据库成功，开始更新attend,waId:%v，newCount：%v", waId, newCount))

	switch newCount {
	case config.ApplicationConfig.Activity.Stage1Award.HelpNum:
		// 三人助力
		userAttendInfoEntity := entity.UserAttendInfoEntityV2{
			Id:           beHelpedUserInfo.Id,
			IsThreeStage: constant.IsStage,
			ThreeOverAt:  util.GetNowCustomTime(),
			NewestHelpAt: util.GetNowCustomTime(),
			IsSendCdkMsg: constant.CdkMsgUnSend,
		}
		_, err = u.userAttendInfoMapper.UpdateByPrimaryKeySelective(&session, userAttendInfoEntity)
		if err != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，更新被助力人助力成功失败,rallyCode:%v", constant.MethodHelp, rallyCode))
			return nil, err
		}
	case config.ApplicationConfig.Activity.Stage2Award.HelpNum:
		// 五人助力
		userAttendInfoEntity := entity.UserAttendInfoEntityV2{
			Id:           beHelpedUserInfo.Id,
			IsFiveStage:  constant.IsStage,
			FiveOverAt:   util.GetNowCustomTime(),
			NewestHelpAt: util.GetNowCustomTime(),
			IsSendCdkMsg: constant.CdkMsgUnSend,
		}
		_, err = u.userAttendInfoMapper.UpdateByPrimaryKeySelective(&session, userAttendInfoEntity)
		if err != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，更新被助力人助力成功失败,rallyCode:%v", constant.MethodHelp, rallyCode))
			return nil, err
		}
	case config.ApplicationConfig.Activity.Stage3Award.HelpNum:
		// 八人助力
		userAttendInfoEntity := entity.UserAttendInfoEntityV2{
			Id:           beHelpedUserInfo.Id,
			AttendStatus: constant.AttendStatusEightOver,
			EightOverAt:  util.GetNowCustomTime(),
			NewestHelpAt: util.GetNowCustomTime(),
			IsSendCdkMsg: constant.CdkMsgUnSend,
		}
		_, err = u.userAttendInfoMapper.UpdateByPrimaryKeySelective(&session, userAttendInfoEntity)
		if err != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，更新被助力人助力成功失败,rallyCode:%v", constant.MethodHelp, rallyCode))
			return nil, err
		}
	default:
		userAttendInfoEntity := entity.UserAttendInfoEntityV2{
			Id:           beHelpedUserInfo.Id,
			HasHelper:    2,
			NewestHelpAt: util.GetNowCustomTime(),
		}
		_, err = u.userAttendInfoMapper.UpdateByPrimaryKeySelective(&session, userAttendInfoEntity)
		if err != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，更新被助力人助力成功失败,rallyCode:%v", constant.MethodHelp, rallyCode))
			return nil, err
		}
	}

	logTracing.LogPrintf(ctx, logTracing.NormalLogFmt, fmt.Sprintf("结束新增助力信息到数据库,waId:%v", waId))

	var sendNxListParamsDtoList []*dto.SendNxListParamsDto

	handler, err := u.helpSendHandler(ctx, newCount, userAttendInfoEntity, beHelpedUserInfo, helpCacheDtoList)
	if nil != err {
		return sendNxListParamsDtoList, err
	}

	sendNxListParamsDtoList = append(sendNxListParamsDtoList, handler...)

	if !isExist {
		session.Commit()
	}
	return sendNxListParamsDtoList, nil
}

func (u UserAttendInfoService) helpSendHandler(ctx *gin.Context, newCount int, userAttendInfoEntity entity.UserAttendInfoEntityV2, beHelpedUserInfo *entity.UserAttendInfoEntityV2, helpNameList []*dto.HelpCacheDto) ([]*dto.SendNxListParamsDto, error) {
	waId := userAttendInfoEntity.WaId
	rallyCode := beHelpedUserInfo.RallyCode

	msgInfoEntity := &entity.MsgInfoEntityV2{
		Id:         util.GetSnowFlakeIdStr(ctx),
		Type:       "send",
		WaId:       beHelpedUserInfo.WaId,
		SourceWaId: waId,
	}

	session, isExist, err := txUtil.GetTransaction(ctx)
	if nil != err {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，创建事务失败,err：%v", constant.MethodHelp, err))
		return nil, errors.New("database is error")
	}
	if !isExist {
		defer func() {
			session.Rollback()
			session.Close()
		}()
	}

	var sendNxListParamsDtoList []*dto.SendNxListParamsDto
	switch newCount {
	case config.ApplicationConfig.Activity.Stage1Award.HelpNum:
		// 三人助力
		cdk, cdkIsExist, err := GetCdkByCdkType(ctx, constant.ThreeCdk)
		if err != nil {
			return nil, err
		}
		if cdkIsExist {
			// cdk没有，就不发消息，等待cdk补充。发送三人助力完成消息
			userAttendInfo := entity.UserAttendInfoEntityV2{
				Id:           beHelpedUserInfo.Id,
				ThreeCdkCode: cdk,
				IsSendCdkMsg: constant.CdkMsgSend,
			}
			_, err = u.userAttendInfoMapper.UpdateByPrimaryKeySelective(&session, userAttendInfo)
			if err != nil {
				logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，更新cdk消息未发送失败,rallyCode:%v", constant.MethodHelp, rallyCode))
				return nil, err
			}

			logTracing.LogPrintf(ctx, logTracing.NormalLogFmt, fmt.Sprintf("开始构建三人cdk,waId:%v", waId))

			msgInfoEntity.MsgType = constant.HelpThreeOverMsg
			sendNxListParamsDto, err := HelpThreeOverMsg2NX(ctx, msgInfoEntity, *beHelpedUserInfo, cdk, helpNameList, constant.BizTypeInteractive)
			if err != nil {
				logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，发送3人助力完成消息失败,rallyCode:%v", constant.MethodHelp, rallyCode))
				return nil, err
			}
			logTracing.LogPrintf(ctx, logTracing.NormalLogFmt, fmt.Sprintf("结束构建三人cdk,waId:%v", waId))

			sendNxListParamsDtoList = append(sendNxListParamsDtoList, sendNxListParamsDto...)
		}
	case config.ApplicationConfig.Activity.Stage2Award.HelpNum:
		// 五人助力
		cdk, cdkIsExist, err := GetCdkByCdkType(ctx, constant.FiveCdk)
		if err != nil {
			return nil, err
		}

		if cdkIsExist {
			// cdk没有，就不发消息，等待cdk补充。发送三人助力完成消息
			userAttendInfo := entity.UserAttendInfoEntityV2{
				Id:           beHelpedUserInfo.Id,
				FiveCdkCode:  cdk,
				IsSendCdkMsg: constant.CdkMsgSend,
			}
			_, err = u.userAttendInfoMapper.UpdateByPrimaryKeySelective(&session, userAttendInfo)
			if err != nil {
				logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，更新cdk消息未发送失败,rallyCode:%v", constant.MethodHelp, rallyCode))
				return nil, err
			}

			logTracing.LogPrintf(ctx, logTracing.NormalLogFmt, fmt.Sprintf("开始构建五人cdk,waId:%v", waId))

			msgInfoEntity.MsgType = constant.HelpFiveOverMsg
			sendNxListParamsDto, err := HelpFiveOverMsg2NX(ctx, msgInfoEntity, *beHelpedUserInfo, cdk, helpNameList, constant.BizTypeInteractive)
			if err != nil {
				logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，发送5人助力完成消息失败,rallyCode:%v", constant.MethodHelp, rallyCode))
				return nil, err
			}
			logTracing.LogPrintf(ctx, logTracing.NormalLogFmt, fmt.Sprintf("结束构建五人cdk,waId:%v", waId))

			sendNxListParamsDtoList = append(sendNxListParamsDtoList, sendNxListParamsDto...)
		}
	case config.ApplicationConfig.Activity.Stage3Award.HelpNum:
		// 8人助力
		cdk, cdkIsExist, err := GetCdkByCdkType(ctx, constant.EightCdk)
		if err != nil {
			return nil, err
		}

		if cdkIsExist {
			// cdk没有，就不发消息，等待cdk补充。发送8人助力完成消息
			userAttendInfo := entity.UserAttendInfoEntityV2{
				Id:           beHelpedUserInfo.Id,
				EightCdkCode: cdk,
				IsSendCdkMsg: constant.CdkMsgSend,
			}
			_, err = u.userAttendInfoMapper.UpdateByPrimaryKeySelective(&session, userAttendInfo)
			if err != nil {
				logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，更新cdk消息未发送失败,rallyCode:%v", constant.MethodHelp, rallyCode))
				return nil, err
			}

			logTracing.LogPrintf(ctx, logTracing.NormalLogFmt, fmt.Sprintf("开始构建八人cdk,waId:%v", waId))

			msgInfoEntity.MsgType = constant.HelpEightOverMsg
			sendNxListParamsDto, err := HelpEightOverMsg2NX(ctx, msgInfoEntity, *beHelpedUserInfo, cdk, helpNameList, constant.BizTypeInteractive)
			if err != nil {
				logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，发送8人助力完成消息失败,rallyCode:%v", constant.MethodHelp, rallyCode))
				return nil, err
			}
			logTracing.LogPrintf(ctx, logTracing.NormalLogFmt, fmt.Sprintf("结束构建八人cdk,waId:%v", waId))

			sendNxListParamsDtoList = append(sendNxListParamsDtoList, sendNxListParamsDto...)
		}
	default:
		userAttendInfoEntity := entity.UserAttendInfoEntityV2{
			Id:           beHelpedUserInfo.Id,
			NewestHelpAt: util.GetNowCustomTime(),
		}
		_, err := u.userAttendInfoMapper.UpdateByPrimaryKeySelective(&session, userAttendInfoEntity)
		if err != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，更新被助力人助力成功失败,rallyCode:%v", constant.MethodHelp, rallyCode))
			return nil, err
		}
		logTracing.LogPrintf(ctx, logTracing.NormalLogFmt, fmt.Sprintf("开始构建单人助力成功,waId:%v", waId))

		// 单人助力成功
		msgInfoEntity.MsgType = constant.HelpTaskSingleSuccessMsg
		sendNxListParamsDto, err := HelpTaskSingleSuccessMsg2NX(ctx, msgInfoEntity, *beHelpedUserInfo, helpNameList, "")
		if err != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，发送单人助力成功消息失败,rallyCode:%v", constant.MethodHelp, rallyCode))
			return nil, err
		}
		logTracing.LogPrintf(ctx, logTracing.NormalLogFmt, fmt.Sprintf("结束构建单人助力成功,waId:%v", waId))

		sendNxListParamsDtoList = append(sendNxListParamsDtoList, sendNxListParamsDto...)
	}

	if !isExist {
		session.Commit()
	}

	return sendNxListParamsDtoList, nil
}

func (u UserAttendInfoService) ImportData(ctx *gin.Context) error {
	form, err := ctx.MultipartForm()
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("ImportData方法，获取MultipartForm失败,err：%v", err))
		return errors.New("error parsing form")
	}
	// redis
	template := redis_template.NewRedisTemplate()
	// 获取文件列表
	files := form.File["file"]
	fileTypeList := form.Value["fileType"]

	for index, file := range files {
		fileType := fileTypeList[index]

		if !constant.ContainsCdkType(fileType) {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("ImportData方法，不支持的文件类型,fileType:%v,err：%v", fileType, err))
			return errors.New("error file type")
		}

		csvFile, err := file.Open()
		if err != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("ImportData方法，打开文件失败,fileType:%v,err：%v", fileType, err))
			return errors.New("error opening file")
		}
		defer csvFile.Close()

		// 将CSV文件解析到CSVRecord结构体切片中
		var records []request.CSVRecord
		if err := gocsv.Unmarshal(csvFile, &records); err != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("ImportData方法，解析csv行失败,fileType:%v,err：%v", fileType, err))
			return errors.New("error parsing CSV")
		}
		elements := make([]interface{}, 0)

		cdkKey := constant.GetCdkKey(config.ApplicationConfig.Activity.Id, fileType)

		addCdkLen := len(records)
		for index, record := range records {
			if len(elements) >= 10000 {
				_, err := template.LPush(ctx, cdkKey, elements...).Result()
				if err != nil {
					logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("ImportData方法，插入redis，部分报错，fileType:%v,endIndex:%v,cdk:%v,err：%v", fileType, index, record.Data, err))
					return errors.New(fmt.Sprintf("error add redis,end index:%v,cdk:%v", index, record.Data))
				}
				elements = make([]interface{}, 0)
			}
			elements = append(elements, record.Data)
		}
		if len(elements) > 0 {
			_, err := template.LPush(ctx, cdkKey, elements...).Result()
			if err != nil {
				logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("ImportData方法，插入redis，部分报错，fileType:%v,endIndex:%v,cdk:%v,err：%v", fileType, len(records), elements[len(elements)-1], err))
				return errors.New(fmt.Sprintf("error add redis,end index:%v,cdk:%v", len(records), elements[len(elements)-1]))
			}
			elements = make([]interface{}, 0)
		}
		//len, err := template.LLen(ctx, cdkKey).Result()
		//if err != nil {
		//	logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("ImportData方法，查询cdk list长度报错,err：%v", err))
		//	return errors.New(fmt.Sprintf("error query llen redis,key:%v", cdkKey))
		//}

		cdkCountKey := constant.GetCdkInfoKey(config.ApplicationConfig.Activity.Id, fileType)
		exists, err := template.Exists(ctx, cdkCountKey)
		if err != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("ImportData方法，判断cdkCountKey是否存在报错,fileType:%v,err：%v", fileType, err))
			return errors.New(fmt.Sprintf("error exists cdkCountKey redis,key:%v", cdkKey))
		}

		cdkInfo := &response.CdkInfo{
			CdkCount:        0,
			NextSendPercent: 90.0,
		}
		if exists != 0 {
			cdkInfoStr, err := template.Get(ctx, cdkCountKey)
			if err != nil {
				logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("ImportData方法，查询cdkContKey报错,fileType:%v,err：%v", fileType, err))
				return errors.New(fmt.Sprintf("failed set count redis,key:%v", cdkCountKey))
			}
			err = json.NewEncoder().Decode([]byte(cdkInfoStr), cdkInfo)
			if err != nil {
				logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("ImportData方法，cdkInfo转实体报错,cdkInfoStr:%v,err：%v", cdkInfoStr, err))
				return errors.New(fmt.Sprintf("cdkInfo convert json error,key:%v", cdkCountKey))
			}
		}

		cdkInfo.CdkCount = cdkInfo.CdkCount + int64(addCdkLen)

		cdkInfoBytes, err := json.NewEncoder().Encode(cdkInfo)
		if err != nil {
			return errors.New(fmt.Sprintf("mportData方法,转换cdkInfo报错，cdkInfo:%v", cdkInfo))
		}

		set := template.Set(ctx, cdkCountKey, string(cdkInfoBytes))
		if !set {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("ImportData方法，查询cdk list长度报错,err：%v", err))
			return errors.New(fmt.Sprintf("failed set count redis,key:%v", cdkCountKey))
		}
		logTracing.LogPrintf(ctx, logTracing.WebHandleLogFmt, fmt.Sprintf("ImportData方法，导入成功"))
		//template.IncrBy(ctx, constant.CdkInfoKey, int64(len(records)))
	}
	return nil
}

func (u UserAttendInfoService) selectByRallyCode(ctx *gin.Context, rallyCode string) (*entity.UserAttendInfoEntityV2, error) {
	helpUserInfo, err := u.userAttendInfoMapper.SelectByRallyCode(rallyCode)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，根据RallyCode查询userAttendInfo报错,活动id:%v,RallyCode:%v,err：%v", constant.MethodSelectByRallyCode, config.ApplicationConfig.Activity.Id, rallyCode, err))
		return nil, errors.New("database is error")
	}
	if helpUserInfo.Id <= 0 {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，根据RallyCode查询userAttendInfo不存在,活动id:%v,RallyCode:%v,err：%v", constant.MethodSelectByRallyCode, config.ApplicationConfig.Activity.Id, rallyCode, err))
		return nil, errors.New("database is error")
	}
	return &helpUserInfo, nil
}

func (u UserAttendInfoService) HelpTextCount(ctx *gin.Context, req *request.HelpTextCountReq) error {
	methodName := "HelpTextCount"
	if req.HelpTextId == "" {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，HelpTextId为空,req：%v", methodName, req))
		return errors.New("HelpTextCount HelpTextId is null")
	}
	template := redis_template.NewRedisTemplate()
	helpTextWeightKey := constant.GetHelpTextWeightKey(config.ApplicationConfig.Activity.Id)

	code, err := template.Exists(ctx, helpTextWeightKey)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，查询%v 是否存在报错,err：%v", methodName, helpTextWeightKey, err))
		return errors.New(fmt.Sprintf("方法[%s]，查询%v 是否存在报错,err：%v", methodName, helpTextWeightKey, err))
	}
	if code == 0 {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，查询%v helpText数据不存在,err：%v", methodName, helpTextWeightKey, err))
		return errors.New(fmt.Sprintf("方法[%s]，查询%v helpText数据不存在,err：%v", methodName, helpTextWeightKey, err))
	}

	helpTextClickAllCountKey := constant.GetHelpTextClickAllCountKey(config.ApplicationConfig.Activity.Id)
	newCount, err := template.Incr(ctx, helpTextClickAllCountKey).Result()
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，增加%v 报错,err：%v", methodName, helpTextClickAllCountKey, err))
		return errors.New(fmt.Sprintf("方法[%s]，查询%v 增加%v，报错,err：%v", methodName, helpTextClickAllCountKey, err))
	}

	helpTextClickCountKey := constant.GetHelpTextClickCountKey(config.ApplicationConfig.Activity.Id, req.HelpTextId)
	_, err = template.Incr(ctx, helpTextClickCountKey).Result()
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，增加%v 报错,err：%v", methodName, helpTextClickCountKey, err))
		return errors.New(fmt.Sprintf("方法[%s]，查询%v 增加%v，报错,err：%v", methodName, helpTextClickCountKey, err))
	}

	//  需要确认,重新计算权重
	if newCount > 1000 {

	}
	return nil
}
