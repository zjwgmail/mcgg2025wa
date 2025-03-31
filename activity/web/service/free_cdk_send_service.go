package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"go-fission-activity/activity/constant"
	"go-fission-activity/activity/model/dto"
	"go-fission-activity/activity/model/entity"
	"go-fission-activity/activity/model/request"
	"go-fission-activity/activity/web/dao"
	"go-fission-activity/activity/web/middleware/logTracing"
	"go-fission-activity/config"
	"go-fission-activity/util"
	"go-fission-activity/util/config/encoder/rsa"
	"go-fission-activity/util/goroutine_pool"
	"go-fission-activity/util/strUtil"
	"go-fission-activity/util/txUtil"
	"strconv"
	"sync"
)

type FreeCdkSendService struct {
	freeCdkInfoMapper    *dao.FreeCdkInfoMapper
	msgInfoMapper        *dao.MsgInfoMapperV2
	userAttendInfoMapper *dao.UserAttendInfoMapperV2
}

var freeCdkSendServiceOnce sync.Once
var globalFreeCdkSendService FreeCdkSendService

func GetFreeCdkSendService() *FreeCdkSendService {
	freeCdkSendServiceOnce.Do(func() {
		globalFreeCdkSendService = FreeCdkSendService{
			freeCdkInfoMapper:    dao.GetFreeSdkInfoMapper(),
			msgInfoMapper:        dao.GetMsgInfoMapperV2(),
			userAttendInfoMapper: dao.GetUserAttendInfoMapperV2(),
		}
		logTracing.LogPrintfP("第一次使用，globalMsgInfoService")
	})
	return &globalFreeCdkSendService
}

var freeCdkGoroutinePool = goroutine_pool.NewGoroutinePool(4)

func (u FreeCdkSendService) FreeCdkSave(ctx *gin.Context, user entity.UserAttendInfoEntityV2) ([]*dto.SendNxListParamsDto, error) {
	delay := config.ApplicationConfig.Activity.FreeCdkSendDelayHour

	insertNum, err := u.saveFreeCdkSendInfo(ctx, user.WaId)
	if err != nil {
		return nil, err
	}

	if insertNum > 0 && delay == 0 {
		build, err := freeCdkBuild(ctx, user)
		if err != nil {
			return nil, err
		}
		u.updateFreeCdkSendInfo(ctx, user.WaId, constant.CdkMsgSend)
		return build, nil

	}
	return make([]*dto.SendNxListParamsDto, 0), nil
}

func (u FreeCdkSendService) saveFreeCdkSendInfo(ctx *gin.Context, waId string) (int, error) {
	methodName := "saveFreeCdkSendInfo"
	now := util.GetNowCustomTime()
	sendTime := util.GetSendClusteringTime(config.ApplicationConfig.Activity.FreeCdkSendDelayHour, now)

	insertNum, err := u.freeCdkInfoMapper.InsertIgnore(waId, now.Unix(), sendTime.Unix())
	if nil != err {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，插入免费CD信息表失败,waId:%v,err：%v", methodName, waId, err))
		return 0, errors.New("database is error")
	}

	return insertNum, nil
}

func (u FreeCdkSendService) updateFreeCdkSendInfo(ctx *gin.Context, waId string, sendState int) error {
	methodName := "updateFreeCdkSendInfo"

	_, err := u.freeCdkInfoMapper.UpdateStateByWaId(waId, sendState)
	if nil != err {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，更新免费CD信息表失败,waId:%v,err：%v", methodName, waId, err))
		return errors.New("database is error")
	}

	return nil
}

func (u FreeCdkSendService) FreeCdkSend(ctx *gin.Context, methodName string) {
	now := util.GetNowCustomTime()
	timeSecond := now.Unix()

	minId := int64(0)
	limit := 100
	for {
		tmpEntityList, err := u.freeCdkInfoMapper.SelectWaIdsByStateLtTimestamp(timeSecond, constant.CdkMsgUnSend, minId, limit)
		if nil != err {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，查询免费CD信息表失败,timeSecond:%v，minId：%v ,err：%v", methodName, timeSecond, minId, err))
			return
		}
		if len(tmpEntityList) == 0 {
			logTracing.LogInfo(ctx, fmt.Sprintf("方法[%s]，查询免费CD信息表数据为0,结束执行,timeSecond:%v,minId:%v,err：%v", methodName, timeSecond, minId, err))
			break
		}
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，查询免费CD信息表数据条数：%v ,timeSecond:%v,minId:%v,err：%v", methodName, len(tmpEntityList), timeSecond, minId, err))
		for _, freeCdkInfoEntity := range tmpEntityList {
			tmpId := freeCdkInfoEntity.Id
			if minId < tmpId {
				minId = tmpId
			}
			waId := freeCdkInfoEntity.WaId
			freeCdkGoroutinePool.Execute(func(param interface{}) {
				ctx = &gin.Context{}
				entity, ok := param.(entity.FreeCdkInfoEntity) // 断言u是User类型
				if !ok {
					logTracing.LogPrintf(ctx, logTracing.TaskHandleLogFmt, fmt.Sprintf("方法[%s],断言发生错误，waId:%v", methodName, waId))
				}
				logTracing.LogPrintf(ctx, logTracing.TaskHandleLogFmt, fmt.Sprintf("方法[%s],freeCdkGoroutinePool协程池执行任务开始，waId:%v", methodName, entity.WaId))
				userAttendMapper := dao.GetUserAttendInfoMapperV2()
				user, err := userAttendMapper.SelectByWaId(entity.WaId)
				if err != nil {
					logTracing.LogPrintf(ctx, logTracing.TaskHandleLogFmt, fmt.Sprintf("方法[%s],查询userAttendInfo报错，waId:%v", methodName, entity.WaId))
					return
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
				build, err := freeCdkBuild(ctx, user)
				if err != nil {
					logTracing.LogPrintf(ctx, logTracing.TaskHandleLogFmt, fmt.Sprintf("方法[%s],构建免费cdk消息报错，waId:%v", methodName, entity.WaId))
					return
				}
				u.updateFreeCdkSendInfo(ctx, user.WaId, constant.CdkMsgSend)
				if !isExist {
					session.Commit()
				}

				_, nxErr := SendMsgList2NX(ctx, build)
				if nxErr != nil {
					logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，发消息到牛信云失败,waId:%v,reqParam:%v,err：%v", methodName, entity.WaId, build, nxErr))
					return
				}
				logTracing.LogPrintf(ctx, logTracing.TaskHandleLogFmt, fmt.Sprintf("方法[%s],freeCdkGoroutinePool协程池执行任务结束，waId:%v", methodName, entity.WaId))
			}, freeCdkInfoEntity)
		}
		freeCdkGoroutinePool.Wait()
	}
}

func freeCdkBuild(ctx *gin.Context, user entity.UserAttendInfoEntityV2) ([]*dto.SendNxListParamsDto, error) {
	// 改成freeCdk
	cdk, cdkIsExist, err := GetCdkByCdkType(context.Background(), constant.FreeCdk)
	if err != nil {
		return nil, err
	}
	if !cdkIsExist {
		return make([]*dto.SendNxListParamsDto, 0), nil
	}
	// 构建消息 免费红包
	msgInfoEntity := &entity.MsgInfoEntityV2{
		Id:         util.GetSnowFlakeIdStr(ctx),
		Type:       "send",
		WaId:       user.WaId,
		SourceWaId: user.WaId,
		MsgType:    constant.FreeCdkMsg,
	}
	sendNxListParamsDto, err := FreeCdkMsg2NX(ctx, msgInfoEntity, user, cdk, make([]*dto.HelpCacheDto, 0), constant.BizTypeInteractive)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，构建freeCdk失败,waId:%v", constant.MethodHelp, user.WaId))
		return nil, err
	}
	return sendNxListParamsDto, nil

	//msgInfoEntity := &entity.MsgInfoEntityV2{
	//	Id:      util.GetSnowFlakeIdStr(ctx),
	//	Type:    "send",
	//	WaId:    waId,
	//	MsgType: constant.FreeCdkMsg,
	//	SourceWaId:,
	//}
	//session, isExist, err := txUtil.GetTransaction(ctx)
	//if nil != err {
	//	logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，创建事务失败,err：%v", "freeCdkBuild", err))
	//	return errors.New("database is error")
	//}
	//if !isExist {
	//	defer func() {
	//		session.Rollback()
	//		session.Close()
	//	}()
	//}
	//_, err = u.msgInfoMapper.InsertSelective(&session, *msgInfoEntity)
	//if err != nil {
	//	return err
	//}
}

func HelpFreeCdkMsg2NX(ctx *gin.Context, msgInfoEntity *entity.MsgInfoEntityV2, user entity.UserAttendInfoEntityV2, cdk string, helpNameList []*dto.HelpCacheDto, sendNxMsgType int) ([]*dto.SendNxListParamsDto, error) {
	methodName := "HelpThreeOverMsg2NX"

	sendMsgInfo, err := getMsgInfo(ctx, msgInfoEntity, constant.HelpThreeOverMsg, user.Language)
	if err != nil {
		return nil, err
	}

	// ImageLink要修改，根据rallyCodeBeHelpCount调用合成图片上传s3接口,helpNameList 的昵称
	var nicknameList []string
	for _, helpNameEntity := range helpNameList {
		if helpNameEntity.UserNickname != "" {
			nicknameList = append(nicknameList, helpNameEntity.UserNickname)
		}
	}
	if len(nicknameList) > 0 {
		synthesisParam := &request.SynthesisParam{
			NicknameList:    nicknameList,
			CurrentProgress: int64(len(helpNameList)),
			LangNum:         user.Language,
			BizType:         constant.BizTypeInteractive,
		}
		imageUrl, err := GetImageService().GetInteractiveImageUrl(ctx, synthesisParam, msgInfoEntity.WaId)
		if err != nil {
			return nil, err
		}
		sendMsgInfo.Interactive.ImageLink = imageUrl
	} else {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，助力人昵称不存在,SourceWaId:%v,toWaId: %v,language:%v,err:%v", methodName, msgInfoEntity.SourceWaId, msgInfoEntity.WaId, user.Language, err))
		return nil, errors.New("助力人昵称不存在")
	}

	queryUser := entity.UserAttendInfoEntityV2{
		WaId:         user.WaId,
		Language:     user.Language,
		AttendStatus: user.AttendStatus,
		IsThreeStage: constant.IsStage,
		IsFiveStage:  constant.IsNotStage,
	}
	rewardStageDto, err := GetStageInfoByAttendStatus(ctx, methodName, queryUser, helpNameList)
	if err != nil {
		return nil, err
	}

	// 要将传给前端的信息拼接好发给前端，要加密成param
	cdkEncrypt, err := rsa.Encrypt(cdk, config.ApplicationConfig.Rsa.PublicKey)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，加密cdk报错,SourceWaId:%v,toWaId: %v,language:%v,err:%v", methodName, msgInfoEntity.SourceWaId, msgInfoEntity.WaId, user.Language, err))
		return nil, err
	}
	awardLink := rewardStageDto.CurrentAwardLink
	awardLink = strUtil.ReplacePlaceholders(awardLink, user.RallyCode, cdkEncrypt, user.Language, user.Channel)
	awardShortLink, err := globalShortUrlService.GetShortUrlByUrl(ctx, awardLink, msgInfoEntity.WaId)
	if err != nil {
		return nil, err
	}
	sendMsgInfo.Interactive.BodyText = strUtil.ReplacePlaceholders(sendMsgInfo.Interactive.BodyText, helpNameList[len(helpNameList)-1].UserNickname, strconv.Itoa(rewardStageDto.CurrentStageMax), awardShortLink, strconv.Itoa(rewardStageDto.NextStageMax-len(helpNameList)))

	helpText, err := GetHelpTextWeight(ctx)
	if err != nil {
		return nil, err
	}
	sendMsgInfo.Interactive.Action.Url = helpText.BodyText[config.ApplicationConfig.Activity.Scheme][user.Language]
	// url中的链接要调用接口活动，并且要用到rallyCode
	sendMsgInfo.Interactive.Action.ShortLink = strUtil.ReplacePlaceholders(sendMsgInfo.Interactive.Action.ShortLink, user.RallyCode, user.UserNickname, helpText.Id, user.Language, user.Channel)
	shortLink, err := globalShortUrlService.GetShortUrlByUrl(ctx, sendMsgInfo.Interactive.Action.ShortLink, msgInfoEntity.WaId)
	if err != nil {
		return nil, err
	}
	sendMsgInfo.Interactive.Action.Url = strUtil.ReplacePlaceholders(sendMsgInfo.Interactive.Action.Url, shortLink)
	sendMsgInfo.Interactive.Action.Url = config.ApplicationConfig.Activity.WaRedirectListPrefix + util.QueryEscape(sendMsgInfo.Interactive.Action.Url)

	if sendMsgInfo.Template != nil {
		//  模板消息未定
	}

	msgInfoEntityList := []*entity.MsgInfoEntityV2{msgInfoEntity}
	sendMsgInfoList := []*config.MsgInfo{sendMsgInfo}

	var sendNxListParamsDto []*dto.SendNxListParamsDto
	if constant.BizTypeInteractive == sendNxMsgType {
		sendNxListParamsDto, err = BuildInteractionMessage2NX(ctx, msgInfoEntityList, sendMsgInfoList)
		if err != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，发送互动信息错误,SourceWaId:%v,toWaId: %v,language:%v,err:%v", methodName, msgInfoEntity.SourceWaId, msgInfoEntity.WaId, user.Language, err))
			return nil, err
		}
	} else {
		sendNxListParamsDto, err = BuildTemplateMessage2NX(ctx, msgInfoEntityList, sendMsgInfoList)
		if err != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，发送模板信息错误,SourceWaId:%v,toWaId: %v,language:%v,err:%v", methodName, msgInfoEntity.SourceWaId, msgInfoEntity.WaId, user.Language, err))
			return nil, err
		}
	}

	return sendNxListParamsDto, nil
}
