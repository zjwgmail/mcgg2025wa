package service

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"go-fission-activity/activity/constant"
	"go-fission-activity/activity/model/dto"
	"go-fission-activity/activity/model/entity"
	"go-fission-activity/activity/model/nx"
	"go-fission-activity/activity/model/request"
	"go-fission-activity/activity/model/response"
	"go-fission-activity/activity/third/http_client"
	"go-fission-activity/activity/web/dao"
	"go-fission-activity/activity/web/middleware/logTracing"
	"go-fission-activity/config"
	"go-fission-activity/util"
	"go-fission-activity/util/config/encoder/json"
	"go-fission-activity/util/config/encoder/rsa"
	"go-fission-activity/util/strUtil"
	"go-fission-activity/util/txUtil"
	"io"
	"math/rand"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"time"
)

type WaMsgService struct {
}

var globalWaMsgService WaMsgService
var globalImageService ImageService
var globalShortUrlService ShortUrlService

func GetWaMsgService() *WaMsgService {
	return &globalWaMsgService
}

func GetImageService() *ImageService {
	return &globalImageService
}

func GetShortUrlService() *ShortUrlService {
	return &globalShortUrlService
}

// ActivityTask2NX 参与活动消息
func ActivityTask2NX(ctx *gin.Context, msgInfoEntity *entity.MsgInfoEntityV2, language string, channel string, param *request.HelpParam) ([]*dto.SendNxListParamsDto, error) {
	methodName := "activityTask2NX"

	sendMsgInfo, err := getMsgInfo(ctx, msgInfoEntity, msgInfoEntity.MsgType, language)
	if err != nil {
		return nil, err
	}
	paramBytes, err := json.NewEncoder().Encode(param)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，转换json失败,param:%v,err:%v", methodName, param, err))
		return nil, err
	}
	paramStr := string(paramBytes)
	// 要将传给前端的信息拼接好发给前端，要加密成param
	paramStrEncrypt, err := rsa.Encrypt(paramStr, config.ApplicationConfig.Rsa.PublicKey)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，加密params报错,SourceWaId:%v,toWaId: %v,language:%v,err:%v", methodName, msgInfoEntity.SourceWaId, msgInfoEntity.WaId, language, err))
		return nil, err
	}
	paramStrEscape := util.QueryEscape(paramStrEncrypt)

	sendMsgInfo.Interactive.Action.Url = strUtil.ReplacePlaceholders(sendMsgInfo.Interactive.Action.Url, paramStrEscape, language, channel)
	sendJson, err := BuildInteractionMessage2NX(ctx, []*entity.MsgInfoEntityV2{msgInfoEntity}, []*config.MsgInfo{sendMsgInfo})
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，发送信息错误,SourceWaId:%v,toWaId: %v,language:%v,err:%v", methodName, msgInfoEntity.SourceWaId, msgInfoEntity.WaId, language, err))
		return nil, err
	}
	return sendJson, nil
}

// CannotAttendActivity2NX 不能参与活动消息
func CannotAttendActivity2NX(ctx *gin.Context, msgInfoEntity *entity.MsgInfoEntityV2, channel, language string) ([]*dto.SendNxListParamsDto, error) {
	methodName := "CannotAttendActivity2NX"
	// 获取月日，格式：x月x日
	time := util.GetNowCustomTime()
	monthDay := fmt.Sprintf("%d月%d日", time.Month(), time.Day())

	phoneSetKey := constant.GetNotWhiteSetKey(config.ApplicationConfig.Activity.Id)
	addCount, err := SAddKey(methodName, phoneSetKey, msgInfoEntity.WaId)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("方法[%s]，增加%v，报错,err：%v", methodName, phoneSetKey, err))
	}
	if addCount > 0 {
		// 给非白拦截redis增加次数
		notWhite := constant.GetNotWhiteCountKey(config.ApplicationConfig.Activity.Id, monthDay, channel, language)
		_, err := AddIncrKey(methodName, notWhite)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("方法[%s]，增加%v，报错,err：%v", methodName, notWhite, err))
		}
	}

	sendMsgInfo, err := getMsgInfo(ctx, msgInfoEntity, msgInfoEntity.MsgType, language)
	if err != nil {
		return nil, err
	}

	sendJson, err := BuildInteractionMessage2NX(ctx, []*entity.MsgInfoEntityV2{msgInfoEntity}, []*config.MsgInfo{sendMsgInfo})
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，发送信息错误,SourceWaId:%v,toWaId: %v,language:%v,err:%v", methodName, msgInfoEntity.SourceWaId, msgInfoEntity.WaId, language, err))
		return nil, err
	}
	return sendJson, nil
}

// RepeatHelpMsg2NX 重复助力消息
func RepeatHelpMsg2NX(ctx *gin.Context, msgInfoEntity *entity.MsgInfoEntityV2, nickname, language string) ([]*dto.SendNxListParamsDto, error) {
	methodName := "repeatHelpMsg2NX"
	sendMsgInfo, err := getMsgInfo(ctx, msgInfoEntity, msgInfoEntity.MsgType, language)
	if err != nil {
		return nil, err
	}
	originText := sendMsgInfo.Interactive.BodyText
	sendMsgInfo.Interactive.BodyText = strUtil.ReplacePlaceholders(originText, nickname)

	sendJson, err := BuildInteractionMessage2NX(ctx, []*entity.MsgInfoEntityV2{msgInfoEntity}, []*config.MsgInfo{sendMsgInfo})
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，发送信息错误,SourceWaId:%v,toWaId: %v,language:%v,err:%v", methodName, msgInfoEntity.SourceWaId, msgInfoEntity.WaId, language, err))
		return nil, err
	}
	return sendJson, nil
}

// StartGroupMsg2NX 开团消息
func StartGroupMsg2NX(ctx *gin.Context, msgInfoEntity *entity.MsgInfoEntityV2, user entity.UserAttendInfoEntityV2, helpNameList []entity.UserAttendInfoEntityV2, isHelp bool) ([]*dto.SendNxListParamsDto, error) {
	methodName := "StartGroupMsg2NX"
	msgType := msgInfoEntity.MsgType
	if isHelp {
		msgType = constant.HelpStartGroupMsg
	} else {
		msgType = constant.StartGroupMsg
	}

	sendMsgInfo, err := getMsgInfo(ctx, msgInfoEntity, msgType, user.Language)
	if err != nil {
		return nil, err
	}

	// ImageLink要修改，根据rallyCodeBeHelpCount调用合成图片上传s3接口,helpNameList 的昵称
	var nicknameList []string
	for _, helpNameEntity := range helpNameList {
		if helpNameEntity.Id > 0 && helpNameEntity.UserNickname != "" {
			nicknameList = append(nicknameList, helpNameEntity.UserNickname)
		}
	}
	if len(nicknameList) > 0 {
		synthesisParam := &request.SynthesisParam{
			BizType:         constant.BizTypeInteractive,
			LangNum:         user.Language,
			NicknameList:    nicknameList,
			CurrentProgress: int64(len(helpNameList)),
		}

		imageUrl, err := GetImageService().GetInteractiveImageUrl(ctx, synthesisParam, msgInfoEntity.WaId)
		if err != nil {
			return nil, err
		}
		sendMsgInfo.Interactive.ImageLink = imageUrl
	}
	//
	//sendMsgInfo.Interactive.BodyText = strUtil.ReplacePlaceholders(sendMsgInfo.Interactive.BodyText, strconv.Itoa(activityEntity.HelpMax-len(nicknameList)))

	helpText, err := GetHelpTextWeight(ctx)
	if err != nil {
		return nil, err
	}
	sendMsgInfo.Interactive.Action.Url = helpText.BodyText[config.ApplicationConfig.Activity.Scheme][user.Language]

	shortLink := user.ShortLink
	if "" == user.ShortLink {
		// url中的链接要调用接口活动，并且要用到rallyCode
		sendMsgInfo.Interactive.Action.ShortLink = strUtil.ReplacePlaceholders(sendMsgInfo.Interactive.Action.ShortLink, user.RallyCode, user.UserNickname, helpText.Id, user.Language, user.Channel)
		shortLink, err = globalShortUrlService.GetShortUrlByUrl(ctx, sendMsgInfo.Interactive.Action.ShortLink, msgInfoEntity.WaId)
		if err != nil {
			return nil, err
		}
		// 更新user表
		session, isExist, err := txUtil.GetTransaction(ctx)
		if nil != err {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，创建事务失败,err：%v", methodName, err))
			return nil, errors.New("database is error")
		}
		if !isExist {
			defer func() {
				session.Rollback()
				session.Close()
			}()
		}

		userAttendInfoEntity := entity.UserAttendInfoEntityV2{
			Id:        user.Id,
			ShortLink: shortLink,
		}
		_, err = dao.GetUserAttendInfoMapperV2().UpdateByPrimaryKeySelective(&session, userAttendInfoEntity)
		if err != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，更新短链接失败,WaId:%v", methodName, msgInfoEntity.WaId))
			return nil, err
		}

		if !isExist {
			session.Commit()
		}
	}

	sendMsgInfo.Interactive.Action.Url = strUtil.ReplacePlaceholders(sendMsgInfo.Interactive.Action.Url, shortLink)
	sendMsgInfo.Interactive.Action.Url = strUtil.ReplacePlaceholders(config.ApplicationConfig.Activity.WaRedirectListPrefix, user.Language, user.Channel, user.Generation) + util.QueryEscape(sendMsgInfo.Interactive.Action.Url)

	if sendMsgInfo.Template != nil {
		sendMsgInfo.Params = &config.Params{
			NicknameList: nicknameList,
			Language:     user.Language,
		}
		sendMsgInfo.Template.Components[1].Parameters[0].Text = strUtil.ReplacePlaceholders(sendMsgInfo.Template.Components[1].Parameters[0].Text, shortLink)
		sendMsgInfo.Template.Components[1].Parameters[0].Text = util.QueryEscape(sendMsgInfo.Template.Components[1].Parameters[0].Text)
	}

	sendJson, err := BuildInteractionMessage2NX(ctx, []*entity.MsgInfoEntityV2{msgInfoEntity}, []*config.MsgInfo{sendMsgInfo})
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，发送信息错误,SourceWaId:%v,toWaId: %v,language:%v,err:%v", methodName, msgInfoEntity.SourceWaId, msgInfoEntity.WaId, user.Language, err))
		return nil, err
	}
	return sendJson, nil
}

// FounderCanNotStartGroupMsg 主态不能开团消息
func FounderCanNotStartGroupMsg(ctx *gin.Context, msgInfoEntity *entity.MsgInfoEntityV2, language string) ([]*dto.SendNxListParamsDto, error) {
	methodName := "FounderCanNotStartGroupMsg"

	sendMsgInfo, err := getMsgInfo(ctx, msgInfoEntity, msgInfoEntity.MsgType, language)
	if err != nil {
		return nil, err
	}

	sendJson, err := BuildInteractionMessage2NX(ctx, []*entity.MsgInfoEntityV2{msgInfoEntity}, []*config.MsgInfo{sendMsgInfo})
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，发送信息错误,SourceWaId:%v,toWaId: %v,language:%v,err:%v", methodName, msgInfoEntity.SourceWaId, msgInfoEntity.WaId, language, err))
		return nil, err
	}
	return sendJson, nil
}

// CanNotStartGroupMsg 不能开团消息
func CanNotStartGroupMsg(ctx *gin.Context, msgInfoEntity *entity.MsgInfoEntityV2, language string) ([]*dto.SendNxListParamsDto, error) {
	methodName := "CanNotStartGroupMsg"

	sendMsgInfo, err := getMsgInfo(ctx, msgInfoEntity, msgInfoEntity.MsgType, language)
	if err != nil {
		return nil, err
	}
	//
	//// ImageLink要修改，根据rallyCodeBeHelpCount调用合成图片上传s3接口,helpNameList 的昵称
	//var nicknameList []string
	//for _, helpNameEntity := range helpNameList {
	//	if helpNameEntity.Id > 0 && helpNameEntity.UserNickname != "" {
	//		nicknameList = append(nicknameList, helpNameEntity.UserNickname)
	//	}
	//}
	//if len(nicknameList) > 0 {
	//	synthesisParam := &request.SynthesisParam{
	//		NicknameList:    nicknameList,
	//		CurrentProgress: int64(len(helpNameList)),
	//	}
	//	imageUrl, err := GetInteractiveImageUrl(ctx, synthesisParam)
	//	if err != nil {
	//		return "", err
	//	}
	//	sendMsgInfo.Interactive.ImageLink = imageUrl
	//}
	//
	//sendMsgInfo.Interactive.BodyText = fmt.Sprintf(sendMsgInfo.Interactive.BodyText, activityEntity.HelpMax-len(nicknameList))
	//
	//// url中的链接要调用接口活动，并且要用到rallyCode
	//sendMsgInfo.Interactive.Action.ShortLink = fmt.Sprintf(sendMsgInfo.Interactive.Action.ShortLink, rallyCode)
	//shortLink, err := GetShortUrl(ctx, sendMsgInfo.Interactive.Action.ShortLink)
	//if err != nil {
	//	return "", err
	//}
	//sendMsgInfo.Interactive.Action.Url = fmt.Sprintf(sendMsgInfo.Interactive.Action.Url, shortLink)
	//sendMsgInfo.Interactive.Action.Url = util.QueryEscape(sendMsgInfo.Interactive.Action.Url)
	if sendMsgInfo.Template != nil {
		//  模板消息未定
	}

	sendJson, err := BuildInteractionMessage2NX(ctx, []*entity.MsgInfoEntityV2{msgInfoEntity}, []*config.MsgInfo{sendMsgInfo})
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，发送信息错误,SourceWaId:%v,toWaId: %v,language:%v,err:%v", methodName, msgInfoEntity.SourceWaId, msgInfoEntity.WaId, language, err))
		return nil, err
	}
	return sendJson, nil
}

// HelpTaskSingleSuccessMsg2NX 单人助力成功信息
func HelpTaskSingleSuccessMsg2NX(ctx *gin.Context, msgInfoEntity *entity.MsgInfoEntityV2, user entity.UserAttendInfoEntityV2, helpNameList []*dto.HelpCacheDto, redPacketCode string) ([]*dto.SendNxListParamsDto, error) {
	methodName := "HelpTaskSingleSuccessMsg2NX"

	sendMsgInfo, err := getMsgInfo(ctx, msgInfoEntity, constant.HelpTaskSingleSuccessMsg, user.Language)
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
		IsThreeStage: user.IsThreeStage,
		IsFiveStage:  user.IsFiveStage,
	}
	rewardStageDto, err := GetStageInfoByAttendStatus(ctx, methodName, queryUser, helpNameList)
	if err != nil {
		return nil, err
	}
	// 消息修改 恭喜，你的好友{{1}}接受了你的邀请。再邀请{{2}}位好友助力，就能获得{{3}}奖励
	sendMsgInfo.Interactive.BodyText = strUtil.ReplacePlaceholders(sendMsgInfo.Interactive.BodyText, helpNameList[len(helpNameList)-1].UserNickname, strconv.Itoa(rewardStageDto.NextStageMax-len(helpNameList)), rewardStageDto.NextStageName)

	helpText, err := GetHelpTextWeight(ctx)
	if err != nil {
		return nil, err
	}
	sendMsgInfo.Interactive.Action.Url = helpText.BodyText[config.ApplicationConfig.Activity.Scheme][user.Language]

	shortLink := user.ShortLink
	if "" == user.ShortLink {
		// url中的链接要调用接口活动，并且要用到rallyCode
		sendMsgInfo.Interactive.Action.ShortLink = strUtil.ReplacePlaceholders(sendMsgInfo.Interactive.Action.ShortLink, user.RallyCode, user.UserNickname, helpText.Id, user.Language, user.Channel)
		shortLink, err = globalShortUrlService.GetShortUrlByUrl(ctx, sendMsgInfo.Interactive.Action.ShortLink, msgInfoEntity.WaId)
		if err != nil {
			return nil, err
		}
	}

	sendMsgInfo.Interactive.Action.Url = strUtil.ReplacePlaceholders(sendMsgInfo.Interactive.Action.Url, shortLink)
	sendMsgInfo.Interactive.Action.Url = strUtil.ReplacePlaceholders(config.ApplicationConfig.Activity.WaRedirectListPrefix, user.Language, user.Channel, user.Generation) + util.QueryEscape(sendMsgInfo.Interactive.Action.Url)

	if sendMsgInfo.Template != nil {
		sendMsgInfo.Params = &config.Params{
			NicknameList: nicknameList,
			Language:     user.Language,
		}
		sendMsgInfo.Template.Components[1].Parameters[0].Text = helpNameList[len(helpNameList)-1].UserNickname
		sendMsgInfo.Template.Components[1].Parameters[1].Text = strconv.Itoa(rewardStageDto.NextStageMax - len(helpNameList))
		sendMsgInfo.Template.Components[1].Parameters[2].Text = rewardStageDto.NextStageName
		sendMsgInfo.Template.Components[2].Parameters[0].Text = strUtil.ReplacePlaceholders(sendMsgInfo.Template.Components[1].Parameters[0].Text, shortLink)
		sendMsgInfo.Template.Components[2].Parameters[0].Text = util.QueryEscape(sendMsgInfo.Template.Components[1].Parameters[0].Text)
	}

	msgInfoEntityList := []*entity.MsgInfoEntityV2{msgInfoEntity}
	sendMsgInfoList := []*config.MsgInfo{sendMsgInfo}

	//if redPacketCode != "" {
	//	msgInfo, sendMsg, err := RedPacketSendMsg(ctx, msgInfoEntity, language, redPacketCode)
	//	if err != nil {
	//		return "", err
	//	}
	//	msgInfoEntityList = append(msgInfoEntityList, msgInfo)
	//	sendMsgInfoList = append(sendMsgInfoList, sendMsg)
	//}

	sendJson, err := BuildInteractionMessage2NX(ctx, msgInfoEntityList, sendMsgInfoList)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，发送信息错误,SourceWaId:%v,toWaId: %v,language:%v,err:%v", methodName, msgInfoEntity.SourceWaId, msgInfoEntity.WaId, user.Language, err))
		return nil, err
	}
	return sendJson, nil
}

// HelpThreeOverMsg2NX 3人助力完成信息
func HelpThreeOverMsg2NX(ctx *gin.Context, msgInfoEntity *entity.MsgInfoEntityV2, user entity.UserAttendInfoEntityV2, cdk string, helpNameList []*dto.HelpCacheDto, sendNxMsgType int) ([]*dto.SendNxListParamsDto, error) {
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
	shortLink := user.ShortLink
	if "" == user.ShortLink {
		// url中的链接要调用接口活动，并且要用到rallyCode
		sendMsgInfo.Interactive.Action.ShortLink = strUtil.ReplacePlaceholders(sendMsgInfo.Interactive.Action.ShortLink, user.RallyCode, user.UserNickname, helpText.Id, user.Language, user.Channel)
		shortLink, err = globalShortUrlService.GetShortUrlByUrl(ctx, sendMsgInfo.Interactive.Action.ShortLink, msgInfoEntity.WaId)
		if err != nil {
			return nil, err
		}
	}

	sendMsgInfo.Interactive.Action.Url = strUtil.ReplacePlaceholders(sendMsgInfo.Interactive.Action.Url, shortLink)
	sendMsgInfo.Interactive.Action.Url = strUtil.ReplacePlaceholders(config.ApplicationConfig.Activity.WaRedirectListPrefix, user.Language, user.Channel, user.Generation) + util.QueryEscape(sendMsgInfo.Interactive.Action.Url)

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

// HelpFiveOverMsg2NX 5人助力完成信息
func HelpFiveOverMsg2NX(ctx *gin.Context, msgInfoEntity *entity.MsgInfoEntityV2, user entity.UserAttendInfoEntityV2, cdk string, helpNameList []*dto.HelpCacheDto, sendNxMsgType int) ([]*dto.SendNxListParamsDto, error) {
	methodName := "HelpFiveOverMsg2NX"

	sendMsgInfo, err := getMsgInfo(ctx, msgInfoEntity, constant.HelpFiveOverMsg, user.Language)
	if err != nil {
		return nil, err
	}

	queryUser := entity.UserAttendInfoEntityV2{
		WaId:         user.WaId,
		Language:     user.Language,
		AttendStatus: user.AttendStatus,
		IsThreeStage: constant.IsStage,
		IsFiveStage:  constant.IsStage,
	}
	rewardStageDto, err := GetStageInfoByAttendStatus(ctx, methodName, queryUser, helpNameList)
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

	// 要将传给前端的信息拼接好发给前端，要加密成param
	cdkEncrypt, err := rsa.Encrypt(cdk, config.ApplicationConfig.Rsa.PublicKey)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，加密cdk报错,SourceWaId:%v,toWaId: %v,language:%v,err:%v", methodName, msgInfoEntity.SourceWaId, msgInfoEntity.WaId, user.Language, err))
		return nil, err
	}
	// 奖励短链接
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
	shortLink := user.ShortLink
	if "" == user.ShortLink {
		// url中的链接要调用接口活动，并且要用到rallyCode
		sendMsgInfo.Interactive.Action.ShortLink = strUtil.ReplacePlaceholders(sendMsgInfo.Interactive.Action.ShortLink, user.RallyCode, user.UserNickname, helpText.Id, user.Language, user.Channel)
		shortLink, err = globalShortUrlService.GetShortUrlByUrl(ctx, sendMsgInfo.Interactive.Action.ShortLink, msgInfoEntity.WaId)
		if err != nil {
			return nil, err
		}
	}

	sendMsgInfo.Interactive.Action.Url = strUtil.ReplacePlaceholders(sendMsgInfo.Interactive.Action.Url, shortLink)
	sendMsgInfo.Interactive.Action.Url = strUtil.ReplacePlaceholders(config.ApplicationConfig.Activity.WaRedirectListPrefix, user.Language, user.Channel, user.Generation) + util.QueryEscape(sendMsgInfo.Interactive.Action.Url)

	if sendMsgInfo.Template != nil {
		//  模板消息未定
	}

	msgInfoEntityList := []*entity.MsgInfoEntityV2{msgInfoEntity}
	sendMsgInfoList := []*config.MsgInfo{sendMsgInfo}
	//if redPacketCode != "" {
	//	msgInfo, sendMsg, err := RedPacketSendMsg(ctx, msgInfoEntity, language, redPacketCode)
	//	if err != nil {
	//		return "", err
	//	}
	//	msgInfoEntityList = append(msgInfoEntityList, msgInfo)
	//	sendMsgInfoList = append(sendMsgInfoList, sendMsg)
	//}

	var sendJson []*dto.SendNxListParamsDto
	if constant.BizTypeInteractive == sendNxMsgType {
		sendJson, err = BuildInteractionMessage2NX(ctx, msgInfoEntityList, sendMsgInfoList)
		if err != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，发送互动信息错误,SourceWaId:%v,toWaId: %v,language:%v,err:%v", methodName, msgInfoEntity.SourceWaId, msgInfoEntity.WaId, user.Language, err))
			return nil, err
		}
	} else {
		sendJson, err = BuildTemplateMessage2NX(ctx, msgInfoEntityList, sendMsgInfoList)
		if err != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，发送模板信息错误,SourceWaId:%v,toWaId: %v,language:%v,err:%v", methodName, msgInfoEntity.SourceWaId, msgInfoEntity.WaId, user.Language, err))
			return nil, err
		}
	}
	return sendJson, nil
}

// HelpEightOverMsg2NX 8人助力完成信息
func HelpEightOverMsg2NX(ctx *gin.Context, msgInfoEntity *entity.MsgInfoEntityV2, user entity.UserAttendInfoEntityV2, cdk string, helpNameList []*dto.HelpCacheDto, sendNxMsgType int) ([]*dto.SendNxListParamsDto, error) {
	methodName := "HelpEightOverMsg2NX"

	sendMsgInfo, err := getMsgInfo(ctx, msgInfoEntity, constant.HelpEightOverMsg, user.Language)
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

	// url中的链接要调用接口活动，可能会用到rallyCode，cdk； 还需要url加密吗
	// rallyCodeEscape := util.QueryEscape(rallyCode)
	// sendMsgInfo.Interactive.Action.Url = strUtil.ReplacePlaceholders(sendMsgInfo.Interactive.Action.Url, rallyCodeEscape)
	queryUser := entity.UserAttendInfoEntityV2{
		WaId:         user.WaId,
		Language:     user.Language,
		AttendStatus: constant.AttendStatusEightOver,
		IsThreeStage: constant.IsStage,
		IsFiveStage:  constant.IsStage,
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
	// 奖励短链接
	awardLink := rewardStageDto.CurrentAwardLink
	awardLink = strUtil.ReplacePlaceholders(awardLink, user.RallyCode, cdkEncrypt, user.Language, user.Channel)
	awardShortLink, err := globalShortUrlService.GetShortUrlByUrl(ctx, awardLink, msgInfoEntity.WaId)
	if err != nil {
		return nil, err
	}
	sendMsgInfo.Interactive.BodyText = strUtil.ReplacePlaceholders(sendMsgInfo.Interactive.BodyText, helpNameList[len(helpNameList)-1].UserNickname, strconv.Itoa(rewardStageDto.CurrentStageMax), awardShortLink)

	// 八人的行动点，改为邀请好友
	helpText, err := GetHelpTextWeight(ctx)
	if err != nil {
		return nil, err
	}
	sendMsgInfo.Interactive.Action.Url = helpText.BodyText[config.ApplicationConfig.Activity.Scheme][user.Language]
	// url中的链接要调用接口活动，并且要用到rallyCode
	shortLink := user.ShortLink
	if "" == user.ShortLink {
		// url中的链接要调用接口活动，并且要用到rallyCode
		sendMsgInfo.Interactive.Action.ShortLink = strUtil.ReplacePlaceholders(sendMsgInfo.Interactive.Action.ShortLink, user.RallyCode, user.UserNickname, helpText.Id, user.Language, user.Channel)
		shortLink, err = globalShortUrlService.GetShortUrlByUrl(ctx, sendMsgInfo.Interactive.Action.ShortLink, msgInfoEntity.WaId)
		if err != nil {
			return nil, err
		}
	}
	sendMsgInfo.Interactive.Action.Url = strUtil.ReplacePlaceholders(sendMsgInfo.Interactive.Action.Url, shortLink)
	sendMsgInfo.Interactive.Action.Url = strUtil.ReplacePlaceholders(config.ApplicationConfig.Activity.WaRedirectListPrefix, user.Language, user.Channel, user.Generation) + util.QueryEscape(sendMsgInfo.Interactive.Action.Url)

	if sendMsgInfo.Template != nil {
		//  模板消息未定
	}

	msgInfoEntityList := []*entity.MsgInfoEntityV2{msgInfoEntity}
	sendMsgInfoList := []*config.MsgInfo{sendMsgInfo}
	//if redPacketCode != "" {
	//	msgInfo, sendMsg, err := RedPacketSendMsg(ctx, msgInfoEntity, language, redPacketCode)
	//	if err != nil {
	//		return "", err
	//	}
	//	msgInfoEntityList = append(msgInfoEntityList, msgInfo)
	//	sendMsgInfoList = append(sendMsgInfoList, sendMsg)
	//}

	var sendJson []*dto.SendNxListParamsDto
	if constant.BizTypeInteractive == sendNxMsgType {
		sendJson, err = BuildInteractionMessage2NX(ctx, msgInfoEntityList, sendMsgInfoList)
		if err != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，发送互动信息错误,SourceWaId:%v,toWaId: %v,language:%v,err:%v", methodName, msgInfoEntity.SourceWaId, msgInfoEntity.WaId, user.Language, err))
			return nil, err
		}
	} else {
		sendJson, err = BuildTemplateMessage2NX(ctx, msgInfoEntityList, sendMsgInfoList)
		if err != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，发送模板信息错误,SourceWaId:%v,toWaId: %v,language:%v,err:%v", methodName, msgInfoEntity.SourceWaId, msgInfoEntity.WaId, user.Language, err))
			return nil, err
		}
	}
	return sendJson, nil
}

// FreeCdkMsg2NX 免费cdk红包
func FreeCdkMsg2NX(ctx *gin.Context, msgInfoEntity *entity.MsgInfoEntityV2, user entity.UserAttendInfoEntityV2, cdk string, helpNameList []*dto.HelpCacheDto, sendNxMsgType int) ([]*dto.SendNxListParamsDto, error) {
	methodName := "FreeCdkMsg2NX"

	sendMsgInfo, err := getMsgInfo(ctx, msgInfoEntity, constant.FreeCdkMsg, user.Language)
	if err != nil {
		return nil, err
	}

	// 要将传给前端的信息拼接好发给前端，要加密成param
	cdkEncrypt, err := rsa.Encrypt(cdk, config.ApplicationConfig.Rsa.PublicKey)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，加密cdk报错,SourceWaId:%v,toWaId: %v,language:%v,err:%v", methodName, msgInfoEntity.SourceWaId, msgInfoEntity.WaId, user.Language, err))
		return nil, err
	}
	// url中的链接要调用接口活动，并且要用到rallyCode
	sendMsgInfo.Interactive.Action.ShortLink = strUtil.ReplacePlaceholders(sendMsgInfo.Interactive.Action.ShortLink, cdkEncrypt, user.Language, user.Channel)
	//shortLink, err := globalShortUrlService.GetShortUrlByUrl(ctx, sendMsgInfo.Interactive.Action.ShortLink, msgInfoEntity.WaId)
	//if err != nil {
	//	return nil, err
	//}
	sendMsgInfo.Interactive.Action.Url = sendMsgInfo.Interactive.Action.ShortLink
	// sendMsgInfo.Interactive.Action.Url = config.ApplicationConfig.Activity.WaRedirectListPrefix + util.QueryEscape(sendMsgInfo.Interactive.Action.Url)

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

// RedPacketReadyMsg2NX 红包预发
//func RedPacketReadyMsg2NX(ctx *gin.Context, msgInfoEntity *entity.MsgInfoEntityV2, language string, userAttend entity.UserAttendInfoEntityV2) (string, error) {
//	methodName := "RedPacketReadyMsg2NX"
//
//	sendMsgInfo, err := getMsgInfo(ctx, msgInfoEntity, constant.RedPacketReadyMsg, language)
//	if err != nil {
//		return "", err
//	}
//	// url中的链接要调用接口活动，并且要用到userAttend的信息
//	//sendMsgInfo.Interactive.Action.Url = fmt.Sprintf(sendMsgInfo.Interactive.Action.Url, rallyCode)
//
//	sendMsgInfo.Interactive.Action.Url = util.QueryEscape(sendMsgInfo.Interactive.Action.Url)
//
//	sendJson, err := BuildInteractionMessage2NX(ctx, []*entity.MsgInfoEntityV2{msgInfoEntity}, []*config.MsgInfo{sendMsgInfo})
//	if err != nil {
//		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，发送信息错误,SourceWaId:%v,toWaId: %v,language:%v,err:%v", methodName, msgInfoEntity.SourceWaId, msgInfoEntity.WaId, language, err))
//		return "", err
//	}
//	return sendJson, nil
//}

//
//// RedPacketSendMsg2NX 红包发送
//func RedPacketSendMsg2NX(ctx *gin.Context, msgInfoEntity *entity.MsgInfoEntityV2, language string, userAttend entity.UserAttendInfoEntityV2, redPacketCode string) (string, error) {
//	methodName := "RedPacketSendMsg2NX"
//
//	sendMsgInfo, err := getMsgInfo(ctx, msgInfoEntity, constant.RedPacketSendMsg, language)
//	if err != nil {
//		return "", err
//	}
//	// url中的链接要调用接口活动，并且要用到userAttend的信息
//	//sendMsgInfo.Interactive.Action.Url = fmt.Sprintf(sendMsgInfo.Interactive.Action.Url, rallyCode)
//
//	sendMsgInfo.Interactive.Action.Url = util.QueryEscape(sendMsgInfo.Interactive.Action.Url)
//
//	sendJson, err := BuildInteractionMessage2NX(ctx, []*entity.MsgInfoEntityV2{msgInfoEntity}, []*config.MsgInfo{sendMsgInfo})
//	if err != nil {
//		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，发送信息错误,SourceWaId:%v,toWaId: %v,language:%v,err:%v", methodName, msgInfoEntity.SourceWaId, msgInfoEntity.WaId, language, err))
//		return "", err
//	}
//	return sendJson, nil
//}
//
//// RedPacketSendMsg 红包发放消息
//func RedPacketSendMsg(ctx *gin.Context, msgInfoEntity *entity.MsgInfoEntityV2, language string, redPacketCode string) (*entity.MsgInfoEntityV2, *config.MsgInfo, error) {
//	// methodName := "RedPacketSendMsg"
//
//	redPacketMsgInfoEntity := &entity.MsgInfoEntityV2{
//		Id:         util.GetSnowFlakeIdStr(ctx),
//		Type:       "send",
//		WaId:       msgInfoEntity.WaId,
//		SourceWaId: msgInfoEntity.SourceWaId,
//		ActivityId: config.ApplicationConfig.Activity.Id,
//		MsgType:    constant.RedPacketSendMsg,
//	}
//
//	sendMsgInfo, err := getMsgInfo(ctx, msgInfoEntity, constant.RedPacketSendMsg, language)
//	if err != nil {
//		return nil, nil, err
//	}
//
//	// url中的链接要调用接口活动，可能会用到redPacketCode； 还需要url加密吗
//	//sendMsgInfo.Interactive.Action.Url = fmt.Sprintf(sendMsgInfo.Interactive.Action.Url, rallyCode)
//
//	// sendMsgInfo.Interactive.Action.Url = util.QueryEscape(sendMsgInfo.Interactive.Action.Url)
//
//	return redPacketMsgInfoEntity, sendMsgInfo, nil
//}

// RenewFreeMsg 续免费信息
func RenewFreeMsg(ctx *gin.Context, msgInfoEntity *entity.MsgInfoEntityV2, language string, sendNxMsgType int) ([]*dto.SendNxListParamsDto, error) {
	methodName := "RenewFreeMsg"

	sendMsgInfo, err := getMsgInfo(ctx, msgInfoEntity, constant.RenewFreeMsg, language)
	if err != nil {
		return nil, err
	}

	msgInfoEntityList := []*entity.MsgInfoEntityV2{msgInfoEntity}
	sendMsgInfoList := []*config.MsgInfo{sendMsgInfo}

	var sendJson []*dto.SendNxListParamsDto

	if constant.BizTypeInteractive == sendNxMsgType {
		sendJson, err = BuildInteractionMessage2NX(ctx, msgInfoEntityList, sendMsgInfoList)
		if err != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，发送互动信息错误,SourceWaId:%v,toWaId: %v,language:%v,err:%v", methodName, msgInfoEntity.SourceWaId, msgInfoEntity.WaId, language, err))
			return nil, err
		}
	} else {
		sendJson, err = BuildTemplateMessage2NX(ctx, msgInfoEntityList, sendMsgInfoList)
		if err != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，发送模板信息错误,SourceWaId:%v,toWaId: %v,language:%v,err:%v", methodName, msgInfoEntity.SourceWaId, msgInfoEntity.WaId, language, err))
			return nil, err
		}
	}
	return sendJson, nil
}

// PayRenewFreeMsg 付费-续免费信息
func PayRenewFreeMsg(ctx *gin.Context, msgInfoEntity *entity.MsgInfoEntityV2, language string, sendNxMsgType int) ([]*dto.SendNxListParamsDto, error) {
	methodName := "PayRenewFreeMsg"

	sendMsgInfo, err := getMsgInfo(ctx, msgInfoEntity, constant.PayRenewFreeMsg, language)
	if err != nil {
		return nil, err
	}

	msgInfoEntityList := []*entity.MsgInfoEntityV2{msgInfoEntity}
	sendMsgInfoList := []*config.MsgInfo{sendMsgInfo}

	var sendJson []*dto.SendNxListParamsDto

	if constant.BizTypeInteractive == sendNxMsgType {
		sendJson, err = BuildInteractionMessage2NX(ctx, msgInfoEntityList, sendMsgInfoList)
		if err != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，发送互动信息错误,SourceWaId:%v,toWaId: %v,language:%v,err:%v", methodName, msgInfoEntity.SourceWaId, msgInfoEntity.WaId, language, err))
			return nil, err
		}
	} else {
		sendJson, err = BuildTemplateMessage2NX(ctx, msgInfoEntityList, sendMsgInfoList)
		if err != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，发送模板信息错误,SourceWaId:%v,toWaId: %v,language:%v,err:%v", methodName, msgInfoEntity.SourceWaId, msgInfoEntity.WaId, language, err))
			return nil, err
		}
	}
	return sendJson, nil
}

// PromoteClusteringMsg2NX 催促成团
func PromoteClusteringMsg2NX(ctx *gin.Context, msgInfoEntity *entity.MsgInfoEntityV2, user entity.UserAttendInfoEntityV2, sendNxMsgType int, helpNameList []entity.UserAttendInfoEntityV2) ([]*dto.SendNxListParamsDto, error) {
	methodName := "PromoteClusteringMsg2NX"

	sendMsgInfo, err := getMsgInfo(ctx, msgInfoEntity, constant.PromoteClusteringMsg, user.Language)
	if err != nil {
		return nil, err
	}
	//rewardStageDto, err := GetStageInfoByAttendStatus(ctx, methodName, user)
	//if err != nil {
	//	return nil, err
	//}

	// ImageLink要修改，根据rallyCodeBeHelpCount调用合成图片上传s3接口,helpNameList 的昵称
	var nicknameList []string
	for _, helpNameEntity := range helpNameList {
		if helpNameEntity.Id > 0 && helpNameEntity.UserNickname != "" {
			nicknameList = append(nicknameList, helpNameEntity.UserNickname)
		}
	}
	// 没有助力人就保持不变，用活动图
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
	}

	//sendMsgInfo.Interactive.BodyText = strUtil.ReplacePlaceholders(sendMsgInfo.Interactive.BodyText, rewardStageDto.NextStageName)

	helpText, err := GetHelpTextWeight(ctx)
	if err != nil {
		return nil, err
	}
	sendMsgInfo.Interactive.Action.Url = helpText.BodyText[config.ApplicationConfig.Activity.Scheme][user.Language]
	// url中的链接要调用接口活动，并且要用到rallyCode
	shortLink := user.ShortLink
	if "" == user.ShortLink {
		// url中的链接要调用接口活动，并且要用到rallyCode
		sendMsgInfo.Interactive.Action.ShortLink = strUtil.ReplacePlaceholders(sendMsgInfo.Interactive.Action.ShortLink, user.RallyCode, user.UserNickname, helpText.Id, user.Language, user.Channel)
		shortLink, err = globalShortUrlService.GetShortUrlByUrl(ctx, sendMsgInfo.Interactive.Action.ShortLink, msgInfoEntity.WaId)
		if err != nil {
			return nil, err
		}
	}
	sendMsgInfo.Interactive.Action.Url = strUtil.ReplacePlaceholders(sendMsgInfo.Interactive.Action.Url, shortLink)
	sendMsgInfo.Interactive.Action.Url = strUtil.ReplacePlaceholders(config.ApplicationConfig.Activity.WaRedirectListPrefix, user.Language, user.Channel, user.Generation) + util.QueryEscape(sendMsgInfo.Interactive.Action.Url)

	msgInfoEntityList := []*entity.MsgInfoEntityV2{msgInfoEntity}
	sendMsgInfoList := []*config.MsgInfo{sendMsgInfo}

	var sendJson []*dto.SendNxListParamsDto

	if constant.BizTypeInteractive == sendNxMsgType {
		sendJson, err = BuildInteractionMessage2NX(ctx, msgInfoEntityList, sendMsgInfoList)
		if err != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，发送互动信息错误,SourceWaId:%v,toWaId: %v,language:%v,err:%v", methodName, msgInfoEntity.SourceWaId, msgInfoEntity.WaId, user.Language, err))
			return nil, err
		}
	} else {
		sendJson, err = BuildTemplateMessage2NX(ctx, msgInfoEntityList, sendMsgInfoList)
		if err != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，发送模板信息错误,SourceWaId:%v,toWaId: %v,language:%v,err:%v", methodName, msgInfoEntity.SourceWaId, msgInfoEntity.WaId, user.Language, err))
			return nil, err
		}
	}
	return sendJson, nil
}

// EndCanNotStartGroupMsg 结束期-不能开团消息
func EndCanNotStartGroupMsg(ctx *gin.Context, msgInfoEntity *entity.MsgInfoEntityV2, language string) ([]*dto.SendNxListParamsDto, error) {
	methodName := "EndCanNotStartGroupMsg"

	sendMsgInfo, err := getMsgInfo(ctx, msgInfoEntity, msgInfoEntity.MsgType, language)
	if err != nil {
		return nil, err
	}

	sendJson, err := BuildInteractionMessage2NX(ctx, []*entity.MsgInfoEntityV2{msgInfoEntity}, []*config.MsgInfo{sendMsgInfo})
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，发送信息错误,SourceWaId:%v,toWaId: %v,language:%v,err:%v", methodName, msgInfoEntity.SourceWaId, msgInfoEntity.WaId, language, err))
		return nil, err
	}
	return sendJson, nil
}

// EndCanNotHelpMsg 结束期-不能助力消息
func EndCanNotHelpMsg(ctx *gin.Context, msgInfoEntity *entity.MsgInfoEntityV2, language string) ([]*dto.SendNxListParamsDto, error) {
	methodName := "EndCanNotHelpMsg"

	sendMsgInfo, err := getMsgInfo(ctx, msgInfoEntity, msgInfoEntity.MsgType, language)
	if err != nil {
		return nil, err
	}

	sendJson, err := BuildInteractionMessage2NX(ctx, []*entity.MsgInfoEntityV2{msgInfoEntity}, []*config.MsgInfo{sendMsgInfo})
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，发送信息错误,SourceWaId:%v,toWaId: %v,language:%v,err:%v", methodName, msgInfoEntity.SourceWaId, msgInfoEntity.WaId, language, err))
		return nil, err
	}
	return sendJson, nil
}

// RenewFreeReplyMsg 续订回复信息
func RenewFreeReplyMsg(ctx *gin.Context, msgInfoEntity *entity.MsgInfoEntityV2, user entity.UserAttendInfoEntityV2, sendNxMsgType int) ([]*dto.SendNxListParamsDto, error) {
	methodName := "RenewFreeReplyMsg"

	sendMsgInfo, err := getMsgInfo(ctx, msgInfoEntity, msgInfoEntity.MsgType, user.Language)
	if err != nil {
		return nil, err
	}

	helpText, err := GetHelpTextWeight(ctx)
	if err != nil {
		return nil, err
	}
	sendMsgInfo.Interactive.Action.Url = helpText.BodyText[config.ApplicationConfig.Activity.Scheme][user.Language]
	// url中的链接要调用接口活动，并且要用到rallyCode
	shortLink := user.ShortLink
	if "" == user.ShortLink {
		// url中的链接要调用接口活动，并且要用到rallyCode
		sendMsgInfo.Interactive.Action.ShortLink = strUtil.ReplacePlaceholders(sendMsgInfo.Interactive.Action.ShortLink, user.RallyCode, user.UserNickname, helpText.Id, user.Language, user.Channel)
		shortLink, err = globalShortUrlService.GetShortUrlByUrl(ctx, sendMsgInfo.Interactive.Action.ShortLink, msgInfoEntity.WaId)
		if err != nil {
			return nil, err
		}
	}
	sendMsgInfo.Interactive.Action.Url = strUtil.ReplacePlaceholders(sendMsgInfo.Interactive.Action.Url, shortLink)
	sendMsgInfo.Interactive.Action.Url = strUtil.ReplacePlaceholders(config.ApplicationConfig.Activity.WaRedirectListPrefix, user.Language, user.Channel, user.Generation) + util.QueryEscape(sendMsgInfo.Interactive.Action.Url)

	msgInfoEntityList := []*entity.MsgInfoEntityV2{msgInfoEntity}
	sendMsgInfoList := []*config.MsgInfo{sendMsgInfo}

	var sendJson []*dto.SendNxListParamsDto
	if constant.BizTypeInteractive == sendNxMsgType {
		sendJson, err = BuildInteractionMessage2NX(ctx, msgInfoEntityList, sendMsgInfoList)
		if err != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，发送互动信息错误,SourceWaId:%v,toWaId: %v,language:%v,err:%v", methodName, msgInfoEntity.SourceWaId, msgInfoEntity.WaId, user.Language, err))
			return nil, err
		}
	} else {
		sendJson, err = BuildTemplateMessage2NX(ctx, msgInfoEntityList, sendMsgInfoList)
		if err != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，发送模板信息错误,SourceWaId:%v,toWaId: %v,language:%v,err:%v", methodName, msgInfoEntity.SourceWaId, msgInfoEntity.WaId, user.Language, err))
			return nil, err
		}
	}
	return sendJson, nil
}

func getMsgInfo(ctx *gin.Context, msgInfoEntity *entity.MsgInfoEntityV2, msgType string, language string) (*config.MsgInfo, error) {
	methodName := "getMsgInfo"
	allMsgMap := config.MsgConfig.MsgMap

	//if activityAllMsgInfo, exists := allMsgMap[config.ApplicationConfig.Activity.Scheme]; exists {
	if singleMsgInfo, exists := allMsgMap[msgType]; exists {
		if msgInfo, exists := singleMsgInfo[language][config.ApplicationConfig.Activity.Scheme]; exists {
			sendMsgInfo := &config.MsgInfo{}
			err := util.CopyFieldsByJson(msgInfo, sendMsgInfo)
			if err != nil {
				logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，拷贝实体错误,SourceWaId:%v,toWaId: %v,language:%v，err:%v", methodName, msgInfoEntity.SourceWaId, msgInfoEntity.WaId, language, err))
				return nil, errors.New(fmt.Sprintf("方法[%s]，拷贝实体错误,SourceWaId:%v,toWaId: %v,language:%v,err:%v", methodName, msgInfoEntity.SourceWaId, msgInfoEntity.WaId, language, err))
			}
			return sendMsgInfo, nil
		} else {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，没有支持的语言配置,SourceWaId:%v,toWaId: %v,language:%v", methodName, msgInfoEntity.SourceWaId, msgInfoEntity.WaId, language))
			return nil, errors.New(fmt.Sprintf("方法[%s]，没有支持的语言配置,SourceWaId:%v,toWaId: %v,language:%v", methodName, msgInfoEntity.SourceWaId, msgInfoEntity.WaId, language))
		}
	} else {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，没有支持的语言类型,SourceWaId:%v,toWaId: %v,msgType:%v", methodName, msgInfoEntity.SourceWaId, msgInfoEntity.WaId, msgType))
		return nil, errors.New(fmt.Sprintf("方法[%s]，没有支持的语言类型,SourceWaId:%v,toWaId: %v,msgType:%v", methodName, msgInfoEntity.SourceWaId, msgInfoEntity.WaId, msgType))
	}
	//} else {
	//	logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，没有支持的活动id,SourceWaId:%v,toWaId: %v,msgType:%v", methodName, msgInfoEntity.SourceWaId, msgInfoEntity.WaId, msgType))
	//	return nil, errors.New(fmt.Sprintf("方法[%s]，没有支持的活动id,SourceWaId:%v,toWaId: %v,msgType:%v", methodName, msgInfoEntity.SourceWaId, msgInfoEntity.WaId, msgType))
	//}
}

func BuildInteractionMessage2NX(ctx *gin.Context, msgInfoEntityList []*entity.MsgInfoEntityV2, buildMsgParamsList []*config.MsgInfo) ([]*dto.SendNxListParamsDto, error) {
	methodName := "BuildInteractionMessage2NX"
	msgInfoService := GetMsgInfoService()
	var sendMsgList []nx.NxReq
	for index, msgInfoEntity := range msgInfoEntityList {
		buildMsgParams := buildMsgParamsList[index]

		var interactive *nx.Interactive

		if buildMsgParams.Interactive.Type == "cta_url" {
			interactive = &nx.Interactive{
				Type: buildMsgParams.Interactive.Type,
				Body: &nx.NxReqInteractiveBody{
					Text: buildMsgParams.Interactive.BodyText,
				},
				//Footer: &nx.NxReqInteractiveFooter{
				//	Text: buildMsgParams.Interactive.FooterText,
				//},
				//Action: &nx.NxReqInteractiveAction{
				//	Name: "cta_url",
				//	Parameters: &nx.NxReqActionParameter{
				//		DisplayText: buildMsgParams.Interactive.Action.DisplayText,
				//		Url:         buildMsgParams.Interactive.Action.Url,
				//	},
				//},
			}
			if buildMsgParams.Interactive.ImageLink != "" {
				interactive.Header = &nx.NxReqInteractiveHeader{
					Type: "image",
					Image: &nx.NxReqInteractiveImage{
						Link: buildMsgParams.Interactive.ImageLink,
					},
				}
			}

			if buildMsgParams.Interactive.Action != nil {
				interactive.Action = &nx.NxReqInteractiveAction{
					Name: "cta_url",
					Parameters: &nx.NxReqActionParameter{
						DisplayText: buildMsgParams.Interactive.Action.DisplayText,
						Url:         buildMsgParams.Interactive.Action.Url,
					},
				}
			}
		} else if buildMsgParams.Interactive.Type == "button" {
			interactive = &nx.Interactive{
				Type: buildMsgParams.Interactive.Type,
				Body: &nx.NxReqInteractiveBody{
					Text: buildMsgParams.Interactive.BodyText,
				},
				//Footer: &nx.NxReqInteractiveFooter{
				//	Text: buildMsgParams.Interactive.FooterText,
				//},
				//Action: &nx.NxReqInteractiveAction{
				//	Buttons: buildMsgParams.Interactive.Action.Buttons,
				//},
			}
			if buildMsgParams.Interactive.ImageLink != "" {
				interactive.Header = &nx.NxReqInteractiveHeader{
					Type: "image",
					Image: &nx.NxReqInteractiveImage{
						Link: buildMsgParams.Interactive.ImageLink,
					},
				}
			}
			if buildMsgParams.Interactive.Action != nil {
				interactive.Action = &nx.NxReqInteractiveAction{
					Buttons: buildMsgParams.Interactive.Action.Buttons,
				}
			}
		} else {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，不支持的互动消息类型,buildMsgParams.Interactive.Type:%v", methodName, buildMsgParams.Interactive.Type))
			return nil, errors.New("unsupported message type")
		}

		params := &nx.NxReqParam{
			Appkey:           config.ApplicationConfig.Nx.AppKey,
			BusinessPhone:    config.ApplicationConfig.Nx.BusinessPhone,
			MessagingProduct: "whatsapp",
			RecipientType:    "individual",
			To:               msgInfoEntity.WaId,
			// CusMessageId:     msgInfoEntity.Id,
			// "dr_webhook":        config.ApplicationConfig.Nx.CallBackUrl,
			Type:        "interactive",
			Interactive: interactive,
		}

		paramsBytes, err := json.NewEncoder().Encode(params)
		if err != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，params转换json失败,params:%v,err:%v", methodName, params, err))
			return nil, err
		}
		paramsStr := string(paramsBytes)
		commonHeaders := getRequestHeader("mt", paramsStr, false)

		nxReq := nx.NxReq{
			Params:        params,
			CommonHeaders: commonHeaders,
		}
		sendMsgList = append(sendMsgList, nxReq)

		// 存储请求
		sendJsonBytes, err := json.NewEncoder().Encode(nxReq)
		if err != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，nxReq转换json失败,params:%v,err:%v", methodName, params, err))
			return nil, err
		}
		sendJson := string(sendJsonBytes)

		buildMsgParamsBytes, err := json.NewEncoder().Encode(buildMsgParams)
		if err != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，buildMsgParams转换json失败,params:%v,err:%v", methodName, params, err))
			return nil, err
		}
		buildMsgParamsJson := string(buildMsgParamsBytes)

		msgInfoEntity.Msg = sendJson
		msgInfoEntity.BuildMsgParams = buildMsgParamsJson
		msgInfoEntity.MsgStatus = constant.NXMsgStatusOwnerUnSent
		// 新增
		err = msgInfoService.InsertMsgInfo(ctx, msgInfoEntity)
		if err != nil {
			return nil, err
		}
	}

	var res []*dto.SendNxListParamsDto
	for index, sendMsg := range sendMsgList {
		msgInfoEntity := msgInfoEntityList[index]

		dto := &dto.SendNxListParamsDto{
			SendMsg:       sendMsg,
			MsgInfoEntity: msgInfoEntity,
		}
		res = append(res, dto)
	}
	return res, nil
}

func BuildTemplateMessage2NX(ctx *gin.Context, msgInfoEntityList []*entity.MsgInfoEntityV2, buildMsgParamsList []*config.MsgInfo) ([]*dto.SendNxListParamsDto, error) {
	methodName := "BuildTemplateMessage2NX"
	msgInfoService := GetMsgInfoService()
	var sendMsgList []nx.NxReq
	for index, msgInfoEntity := range msgInfoEntityList {
		buildMsgParams := buildMsgParamsList[index]

		if buildMsgParams.Template == nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，模板参数信息为空,buildMsgParams.Template:%v", methodName, buildMsgParams.Template))
			return nil, errors.New("模板参数信息为空")
		}

		buildMsgParams.Template.Language.Policy = "deterministic"

		params := &nx.NxReqParam{
			Appkey:           config.ApplicationConfig.Nx.AppKey,
			BusinessPhone:    config.ApplicationConfig.Nx.BusinessPhone,
			MessagingProduct: "whatsapp",
			RecipientType:    "individual",
			To:               msgInfoEntity.WaId,
			//CusMessageId:     msgInfoEntity.Id,
			// "dr_webhook":        config.ApplicationConfig.Nx.CallBackUrl,
			Type:     "template",
			Template: buildMsgParams.Template,
		}

		paramsBytes, err := json.NewEncoder().Encode(params)
		if err != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，params转换json失败,params:%v,err:%v", methodName, params, err))
			return nil, err
		}
		paramsStr := string(paramsBytes)
		commonHeaders := getRequestHeader("mt", paramsStr, false)

		nxReq := nx.NxReq{
			Params:        params,
			CommonHeaders: commonHeaders,
		}
		sendMsgList = append(sendMsgList, nxReq)

		// 存储请求
		sendJsonBytes, err := json.NewEncoder().Encode(nxReq)
		if err != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，nxReq转换json失败,params:%v,err:%v", methodName, params, err))
			return nil, err
		}
		sendJson := string(sendJsonBytes)

		//buildMsgParamsBytes, err := json.NewEncoder().Encode(buildMsgParams)
		//if err != nil {
		//	logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，buildMsgParams转换json失败,params:%v,err:%v", methodName, params, err))
		//	return "", err
		//}
		//buildMsgParamsJson := string(buildMsgParamsBytes)

		msgInfoEntity.Msg = sendJson
		// msgInfoEntity.BuildMsgParams = buildMsgParamsJson
		msgInfoEntity.MsgStatus = constant.NXMsgStatusOwnerUnSent
		// 新增
		err = msgInfoService.InsertMsgInfo(ctx, msgInfoEntity)
		if err != nil {
			return nil, err
		}
	}

	var res []*dto.SendNxListParamsDto
	for index, sendMsg := range sendMsgList {
		msgInfoEntity := msgInfoEntityList[index]

		dto := &dto.SendNxListParamsDto{
			SendMsg:       sendMsg,
			MsgInfoEntity: msgInfoEntity,
		}
		res = append(res, dto)
	}
	return res, nil
}

func SendMsgList2NX(ctx *gin.Context, sendNxListParamsDtoList []*dto.SendNxListParamsDto) (string, error) {
	if sendNxListParamsDtoList != nil && len(sendNxListParamsDtoList) > 0 {
		for _, sendNxListParamsDto := range sendNxListParamsDtoList {
			msgInfoEntity := sendNxListParamsDto.MsgInfoEntity
			sendMsg := sendNxListParamsDto.SendMsg
			// 发送模板消息
			_, nxErr := SendNx(ctx, msgInfoEntity, sendMsg)
			if nxErr != nil {
				return "", nxErr
			}
		}
	}

	return "", nil
}

func SendMsgList2NXHelpTimeOut(ctx *gin.Context, sendNxListParamsDtoList []*dto.SendNxListParamsDto) (string, error) {
	if sendNxListParamsDtoList != nil && len(sendNxListParamsDtoList) > 0 {
		for _, sendNxListParamsDto := range sendNxListParamsDtoList {
			msgInfoEntity := sendNxListParamsDto.MsgInfoEntity
			sendMsg := sendNxListParamsDto.SendMsg
			// 发送模板消息
			_, nxErr := SendNx(ctx, msgInfoEntity, sendMsg)
			if nxErr != nil {
				return "", nxErr
			}
		}
	}

	return "", nil
}

func getRequestHeader(action string, paramsStr string, formData bool) map[string]string {
	commonHeaders := map[string]string{
		"accessKey": config.ApplicationConfig.Nx.Ak,
		"ts":        strconv.FormatInt(time.Now().UnixMilli(), 10),
		"bizType":   "2",
		"action":    action,
	}
	var sign string
	if formData {
		sign = util.CallSignFormData(commonHeaders, config.ApplicationConfig.Nx.Sk)
	} else {
		sign = util.CallSign(commonHeaders, paramsStr, config.ApplicationConfig.Nx.Sk)
	}
	commonHeaders["sign"] = sign
	return commonHeaders
}

func SendNx(ctx *gin.Context, msgInfoEntity *entity.MsgInfoEntityV2, sendMsg nx.NxReq) (*response.NXResponse, error) {
	methodName := "sendNx"

	resNx := &response.NXResponse{}
	var sendErr error
	for i := 1; i < 4; i++ {
		logTracing.LogPrintf(ctx, logTracing.WebHandleLogFmt, fmt.Sprintf("方法[%s]，第[%v]次，开始调用牛信云接口,请求：SourceWaId:%v, toWaId: %v", methodName, i, msgInfoEntity.SourceWaId, msgInfoEntity.WaId))
		res, nxErr := http_client.DoPostSSL("https://api2.nxcloud.com/api/wa/mt", sendMsg.Params, sendMsg.CommonHeaders, 10*1000*time.Second, 10*1000*time.Second)

		if nxErr != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，第[%v]次，调用牛信云接口发生错误,SourceWaId:%v,toWaId: %v,paramsStr:%v,commonHeaders:%v,err:%v", methodName, i, msgInfoEntity.SourceWaId, msgInfoEntity.WaId, sendMsg.Params, sendMsg.CommonHeaders, nxErr))
			sendErr = errors.New(fmt.Sprintf("方法[%s]，第[%v]次，调用牛信云接口发生错误,SourceWaId:%v,toWaId: %v,paramsStr:%v,commonHeaders:%v,err:%v", methodName, i, msgInfoEntity.SourceWaId, msgInfoEntity.WaId, sendMsg.Params, sendMsg.CommonHeaders, nxErr))
			continue
		}
		logTracing.LogPrintf(ctx, logTracing.WebHandleLogFmt, fmt.Sprintf("方法[%s]，第[%v]次，结束调用牛信云接口,请求：SourceWaId:%v, toWaId: %v, 返回: %v", methodName, i, msgInfoEntity.SourceWaId, msgInfoEntity.WaId, res))

		nxErr = json.NewEncoder().Decode([]byte(res), resNx)
		if nxErr != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，第[%v]次，牛信云接口返回转实体报错,res:%v,err：%v", methodName, i, res, nxErr))
			sendErr = errors.New(fmt.Sprintf("方法[%s]，第[%v]次，牛信云接口返回转实体报错,res:%v,err：%v", methodName, i, res, nxErr))
			continue
		}
		if 0 != resNx.Code {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，第[%v]次，调用牛信云接口失败,SourceWaId:%v,toWaId: %v,paramsStr:%v,commonHeaders:%v,res:%v", methodName, i, msgInfoEntity.SourceWaId, msgInfoEntity.WaId, sendMsg.Params, sendMsg.CommonHeaders, resNx))
			sendErr = errors.New(fmt.Sprintf("方法[%s]，第[%v]次，调用牛信云接口失败,SourceWaId:%v,toWaId: %v,paramsStr:%v,commonHeaders:%v,res:%v", methodName, i, msgInfoEntity.SourceWaId, msgInfoEntity.WaId, sendMsg.Params, sendMsg.CommonHeaders, resNx))
			continue
		}
		sendErr = nil
		break
	}

	if sendErr != nil {
		return nil, sendErr
	}
	logTracing.LogPrintf(ctx, logTracing.WebHandleLogFmt, fmt.Sprintf("方法[%s]，调用牛信云接口成功,请求：SourceWaId:%v, toWaId: %v, 返回: %v", methodName, msgInfoEntity.SourceWaId, msgInfoEntity.WaId, resNx))

	// 处理消息表
	traceId := resNx.TraceId
	nxSendRes := &response.NXSendRes{
		NXResponse: resNx,
	}
	nxSendResBytes, err := json.NewEncoder().Encode(nxSendRes)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，发送消息成功，nxSendResList转换json失败,nxSendRes:%v,err:%v", methodName, nxSendRes, err))
		return resNx, err
	}
	nxSendResJson := string(nxSendResBytes)
	// 更新
	updateMsg := &entity.MsgInfoEntityV2{
		Id:          msgInfoEntity.Id,
		SendRes:     nxSendResJson,
		TraceId:     traceId,
		MsgStatus:   constant.NXMsgStatusOwnerSent,
		WaMessageId: resNx.Data.Messages[0].Id,
	}
	msgInfoService := GetMsgInfoService()
	ginCtx := gin.Context{}
	err = msgInfoService.InsertMsgInfo(&ginCtx, updateMsg)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，发送消息成功，nxSendResList转换json失败,nxSendRes:%v,err:%v", methodName, nxSendRes, err))
		return resNx, err
	}
	logTracing.LogPrintf(ctx, logTracing.WebHandleLogFmt, fmt.Sprintf("方法[%s]，更新信息表成功,请求：msgInfoEntity.Id:%v", methodName, msgInfoEntity.Id))
	return resNx, nil
}

// CheckCanSendMsg2NX 检查是否可以发消息给牛信云
func CheckCanSendMsg2NX(ctx *gin.Context, waId string) (bool, bool, error) {
	methodName := "CheckCanSendMsg2NX"
	isFree, err := CheckIsFreeByWaId(ctx, waId)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，根据waId查询是否是免费期报错,活动id:%v,waId:%v,err：%v", methodName, config.ApplicationConfig.Activity.Id, waId, err))
		return false, false, errors.New("database is error")
	}
	if !isFree {
		// 不在免费时间内，查询是否超额
		isUltraLimit, err := CostIsUltraLimit(ctx)
		if err != nil {
			return false, isFree, err
		}
		if isUltraLimit {
			return false, isFree, nil
		}
	}
	return true, isFree, nil
}

// CheckIsFreeByWaId 检查是否在免费期
func CheckIsFreeByWaId(ctx *gin.Context, waId string) (bool, error) {
	methodName := "CheckIsFreeByWaId"
	userAttendInfoMapper := dao.GetUserAttendInfoMapperV2()
	session, isExist, err := txUtil.GetTransaction(ctx)
	if nil != err {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，创建事务失败,err：%v", methodName, err))
		return false, errors.New("database is error")
	}
	if !isExist {
		defer func() {
			session.Rollback()
			session.Close()
		}()
	}

	userAttendInfo, err := userAttendInfoMapper.SelectByWaIdBySession(&session, waId)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，根据waId查询userAttendInfo报错,活动id:%v,waId:%v,err：%v", methodName, config.ApplicationConfig.Activity.Id, waId, err))
		return false, errors.New("database is error")
	}
	if userAttendInfo.Id <= 0 {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，根据waId查询userAttendInfo不存在,活动id:%v,waId:%v,err：%v", methodName, config.ApplicationConfig.Activity.Id, waId, err))
		return false, errors.New("database is error")
	}
	now := util.GetNowCustomTime()
	if userAttendInfo.NewestFreeEndAt.Before(now.Time) || userAttendInfo.NewestFreeEndAt.Equal(now.Time) {
		return false, nil
	}
	if !isExist {
		session.Commit()
	}
	return true, nil
}

// 静态map定义在包级别
var templateImageIdMap = map[string]map[string]map[string]any{
	"zh_CN": {
		"openGroup": { // 开团封面id todo 需要替换
			"templateName":  "mcgg_fission_open_group_zhcn",
			"headerImageId": "2784879291678844",
		},
		"progress1": { // 进度1
			"templateName":  "mcgg_fission_progress1_zhcn",
			"headerImageId": "",
		},
	},
	// 可以添加更多的语言和场景
}

func getTemplateNXImageId(ctx *gin.Context, lang string, scene string) (string, error) {

	sceneMap, err := templateImageIdMap[lang]
	if !err {
		return "", errors.New("language not supported")
	}
	sceneConfig, err := sceneMap[scene]
	if !err {
		return "", errors.New("scene not supported")
	}
	headerImageId, err := sceneConfig["headerImageId"].(string)
	if !err {
		return "", errors.New("header image ID not found")
	}
	return headerImageId, nil
}

func (s WaMsgService) UploadTemplateImage2NX(ctx *gin.Context, path string) (string, error) {
	// 打开文件
	file, err := os.Open(path)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("打开文件失败,path:%v,err:%v", path, err))
		return "", err
	}
	defer file.Close()

	// 创建一个buffer
	var buffer bytes.Buffer

	// 创建multipart writer
	writer := multipart.NewWriter(&buffer)

	// 添加公共参数
	writer.WriteField("appkey", config.ApplicationConfig.Nx.AppKey)
	writer.WriteField("business_phone", config.ApplicationConfig.Nx.BusinessPhone)
	writer.WriteField("messaging_product", "whatsapp")
	writer.WriteField("type", "image/png")

	// 添加文件
	part, err := writer.CreateFormFile("file", "image.png")
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("创建form文件失败,err:%v", err))
		return "", err
	}
	_, err = io.Copy(part, file)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("复制文件失败,err:%v", err))
		return "", err
	}

	// 关闭writer
	err = writer.Close()
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("关闭writer失败,err:%v", err))
		return "", err
	}

	// 创建请求
	req, err := http.NewRequest("POST", "https://api2.nxcloud.com/api/wa/uploadTemplateFile", &buffer)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("创建请求失败,err:%v", err))
		return "", err
	}

	// 设置请求头
	req.Header.Set("Content-Type", writer.FormDataContentType())
	headerMap := getRequestHeader("uploadTemplateFile", "", true)
	for s2 := range headerMap {
		req.Header.Set(s2, headerMap[s2])
	}

	// 发起请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	resNx := &response.NXResponse{}
	err = json.NewEncoder().Decode(respBytes, resNx)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("转换json失败,err:%v", err))
		return "", err
	}
	if 0 != resNx.Code {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("调用牛信云接口失败,res:%v", resNx))
		return "", errors.New(fmt.Sprintf("调用牛信云接口失败,res:%v", resNx))
	}
	return resNx.Data.Id, nil
}

func (s WaMsgService) UploadMedia2NX(ctx *gin.Context, path string) (string, error) {
	// 打开文件
	file, err := os.Open(path)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("打开文件失败,path:%v,err:%v", path, err))
		return "", err
	}
	defer file.Close()

	// 创建一个buffer
	var buffer bytes.Buffer

	// 创建multipart writer
	writer := multipart.NewWriter(&buffer)

	// 添加公共参数
	writer.WriteField("appkey", config.ApplicationConfig.Nx.AppKey)
	writer.WriteField("business_phone", config.ApplicationConfig.Nx.BusinessPhone)
	writer.WriteField("messaging_product", "whatsapp")
	writer.WriteField("type", "image/png")

	// 添加文件
	part, err := writer.CreateFormFile("file", "image.png")
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("创建form文件失败,err:%v", err))
		return "", err
	}
	_, err = io.Copy(part, file)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("复制文件失败,err:%v", err))
		return "", err
	}

	// 关闭writer
	err = writer.Close()
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("关闭writer失败,err:%v", err))
		return "", err
	}

	// 创建请求
	req, err := http.NewRequest("POST", "https://api2.nxcloud.com/api/wa/uploadMedia", &buffer)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("创建请求失败,err:%v", err))
		return "", err
	}

	// 设置请求头
	req.Header.Set("Content-Type", writer.FormDataContentType())
	headerMap := getRequestHeader("uploadTemplateFile", "", true)
	for s2 := range headerMap {
		req.Header.Set(s2, headerMap[s2])
	}

	// 发起请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	resNx := &response.NXResponse{}
	err = json.NewEncoder().Decode(respBytes, resNx)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("转换json失败,err:%v", err))
		return "", err
	}
	if 0 != resNx.Code {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("调用牛信云接口失败,res:%v", resNx))
		return "", errors.New(fmt.Sprintf("调用牛信云接口失败,res:%v", resNx))
	}
	return resNx.Data.Id, nil
}

func GetStageInfoByAttendStatus(ctx *gin.Context, methodName string, user entity.UserAttendInfoEntityV2, helpNameList []*dto.HelpCacheDto) (*dto.RewardStageDto, error) {
	activity := config.ApplicationConfig.Activity
	currentStageMax := activity.Stage1Award.HelpNum
	currentStageName := activity.Stage1Award.AwardName
	currentAwardLink := activity.Stage1Award.AwardLink

	nextStageMax := activity.Stage1Award.HelpNum
	nextStageName := activity.Stage1Award.AwardName
	nextAwardLink := activity.Stage1Award.AwardLink
	helpCount := len(helpNameList)
	if helpCount < activity.Stage1Award.HelpNum {
		nextStageMax = activity.Stage1Award.HelpNum
		nextStageName = activity.Stage1Award.AwardName
		nextAwardLink = activity.Stage1Award.AwardLink

	} else if helpCount >= activity.Stage1Award.HelpNum && helpCount < activity.Stage2Award.HelpNum {
		currentStageMax = activity.Stage1Award.HelpNum
		currentStageName = activity.Stage1Award.AwardName
		currentAwardLink = activity.Stage1Award.AwardLink
		nextStageMax = activity.Stage2Award.HelpNum
		nextStageName = activity.Stage2Award.AwardName
		nextAwardLink = activity.Stage2Award.AwardLink

	} else if helpCount >= activity.Stage2Award.HelpNum && helpCount < activity.Stage3Award.HelpNum {
		currentStageMax = activity.Stage2Award.HelpNum
		currentStageName = activity.Stage2Award.AwardName
		currentAwardLink = activity.Stage2Award.AwardLink
		nextStageMax = activity.Stage3Award.HelpNum
		nextStageName = activity.Stage3Award.AwardName
		nextAwardLink = activity.Stage3Award.AwardLink

	} else if helpCount >= activity.Stage3Award.HelpNum {
		currentStageMax = activity.Stage3Award.HelpNum
		currentStageName = activity.Stage3Award.AwardName
		currentAwardLink = activity.Stage3Award.AwardLink
		nextStageMax = activity.Stage3Award.HelpNum
		nextStageName = activity.Stage3Award.AwardName
		nextAwardLink = activity.Stage3Award.AwardLink
	} else {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，不支持的状态,WaId: %v", methodName, user.WaId))
		return nil, errors.New("不支持的状态")
	}
	return &dto.RewardStageDto{
		CurrentStageMax:  currentStageMax,
		CurrentStageName: currentStageName[user.Language],
		CurrentAwardLink: currentAwardLink[user.Language],
		NextStageMax:     nextStageMax,
		NextStageName:    nextStageName[user.Language],
		NextAwardLink:    nextAwardLink[user.Language],
	}, nil
}

// 电话号码列表
var phoneNumbers = []string{
	"8618758081695",
	"60177761865",
	"601126703621",
	"60109489084",
	"60129708408",
	"60149619180",
	"6589410609",
	"8618321868434",
	"60126714138",
	"85257481920",
	"85296745569",
	"85266831314",
}

func GetRanDomMessage(ctx *gin.Context, interactive *nx.Interactive, batch int, j int) {
	// 初始化随机数种子
	rand.Seed(time.Now().UnixNano())

	toWaId := phoneNumbers[j]
	logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("nickename index %v toWaId %v", j, toWaId))

	params := &nx.NxReqParam{
		Appkey:           config.ApplicationConfig.Nx.AppKey,
		BusinessPhone:    config.ApplicationConfig.Nx.BusinessPhone,
		MessagingProduct: "whatsapp",
		RecipientType:    "individual",
		To:               toWaId,
		Type:             "interactive",
		Interactive:      interactive,
	}

	paramsBytes, err := json.NewEncoder().Encode(params)
	if err != nil {
		//logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，params转换json失败,params:%v,err:%v", GetRanDomMessage, params, err))
		return
	}
	var sendMsgList []nx.NxReq

	paramsStr := string(paramsBytes)
	commonHeaders := getRequestHeader("mt", paramsStr, false)

	nxReq := nx.NxReq{
		Params:        params,
		CommonHeaders: commonHeaders,
	}
	resNx := &response.NXResponse{}

	methodName := "TMP"
	sendMsgList = append(sendMsgList, nxReq)
	for i, sendMsg := range sendMsgList {
		res, nxErr := http_client.DoPostSSL("https://api2.nxcloud.com/api/wa/mt", sendMsg.Params, sendMsg.CommonHeaders, 10*1000*time.Second, 10*1000*time.Second)
		if nxErr != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，第[%v]次，调用牛信云接口发生错误, %v,paramsStr:%v,commonHeaders:%v,err:%v", methodName, i, sendMsg.Params, sendMsg.CommonHeaders, nxErr))
			continue
		}
		//logTracing.LogPrintf(ctx, logTracing.WebHandleLogFmt, fmt.Sprintf("方法[%s]，第[%v]次，结束调用牛信云接口,请求：SourceWaId:%v, toWaId: %v, 返回: %v", methodName, i, res))

		nxErr = json.NewEncoder().Decode([]byte(res), resNx)
		if nxErr != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，第[%v]次，牛信云接口返回转实体报错,res:%v,err：%v", methodName, i, res, nxErr))
			continue
		}
		if 0 != resNx.Code {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，第[%v]次，调用牛信云接口失败, %v,paramsStr:%v,commonHeaders:%v,res:%v", methodName, i, sendMsg.Params, sendMsg.CommonHeaders, resNx))
			continue
		}
		//fmt.Println("随机推送成功 # # #", toWaId, i, sendMsg)
		json_str, _ := json.NewEncoder().Encode(sendMsg)

		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("send success waid %v  sendMsg %v ", toWaId, string(json_str)))

	}

}
