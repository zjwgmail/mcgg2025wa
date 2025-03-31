package service

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go-fission-activity/activity/model/entity"
	"go-fission-activity/activity/web/dao"
	"go-fission-activity/activity/web/middleware/logTracing"
	"go-fission-activity/config"
	"go-fission-activity/util/config/encoder/json"
	"go-fission-activity/util/txUtil"
)

func ReSendMsgByWaId(ctx *gin.Context, waId string, isFree bool) {
	methodName := "ReSendMsgByWaId"

	if !isFree {
		// 非免费，跳过
		logTracing.LogPrintf(ctx, logTracing.NormalLogFmt, fmt.Sprintf("方法[%s]，用户不是免费区间，不发送此用户消息,waId:%v", methodName, waId))
		return
	}
	logTracing.LogPrintf(ctx, logTracing.TaskHandleLogFmt, fmt.Sprintf("方法[%s],开始执行，waId:%v", methodName, waId))

	msgMapper := dao.GetMsgInfoMapperV2()
	msgList, err := msgMapper.SelectMsgListOfUnSendMsg(waId)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],查询根据waId查询未发送的消息失败，err:%v", methodName, err))
		return
	}
	if len(msgList) > 0 {
		for _, msg := range msgList {
			logTracing.LogPrintf(ctx, logTracing.TaskHandleLogFmt, fmt.Sprintf("方法[%s],执行重发消息开始，waId:%v，msgId：%v", methodName, msg.WaId, msg.Id))
			reSendMsg(ctx, methodName, isFree, &msg)
			logTracing.LogPrintf(ctx, logTracing.TaskHandleLogFmt, fmt.Sprintf("方法[%s],执行重发消息结束，waId:%v，msgId：%v", methodName, msg.WaId, msg.Id))
		}
	}

	logTracing.LogPrintf(ctx, logTracing.TaskHandleLogFmt, fmt.Sprintf("方法[%s],执行完成，waId:%v", methodName, waId))

}

func reSendMsg(ctx *gin.Context, methodName string, isFree bool, msgInfoEntity *entity.MsgInfoEntityV2) {
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

	buildMsgParams := &config.MsgInfo{}
	err = json.NewEncoder().Decode([]byte(msgInfoEntity.BuildMsgParams), buildMsgParams)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，消息msg转实体报错,msg:%v,err：%v", methodName, msgInfoEntity.BuildMsgParams, err))
		return
	}

	if isFree {
		// 免费期，互动
		sendNxListParamsDto, nxErr := BuildInteractionMessage2NX(ctx, []*entity.MsgInfoEntityV2{msgInfoEntity}, []*config.MsgInfo{buildMsgParams})
		if nxErr != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，构建重发消息失败,err：%v", methodName, nxErr))
			return
		}
		if !isExist {
			session.Commit()
		}

		_, nxErr = SendMsgList2NX(ctx, sendNxListParamsDto)
		if nxErr != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，重发消息到牛信云失败,err：%v", methodName, nxErr))
			return
		}
	} else {

		//if constant.StartGroupMsg == msgInfoEntity.MsgType || constant.HelpTaskSingleSuccessMsg == msgInfoEntity.MsgType {
		//	// 替换图片   剩余模板消息可能也要替换图片
		//	nicknameList := buildMsgParams.Params.NicknameList
		//	if len(nicknameList) > 0 {
		//		synthesisParam := &request.SynthesisParam{
		//			BizType:         constant.BizTypeInteractive,
		//			LangNum:         buildMsgParams.Params.Language,
		//			NicknameList:    nicknameList,
		//			CurrentProgress: int64(len(nicknameList)),
		//		}
		//		imageId, err := service.GetImageService().GetTemplateImageId(ctx, synthesisParam)
		//		if err != nil {
		//			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，生成模板图片失败,waId:%v,err：%v", methodName, msgInfoEntity.WaId, err))
		//			return
		//		}
		//		buildMsgParams.Template.Components[0].Parameters[0].Image.Id = imageId
		//	}
		//}
		//// 模板
		//sendNxListParamsDto, nxErr := service.BuildTemplateMessage2NX(ctx, []*entity.MsgInfoEntityV2{msgInfoEntity}, []*config.MsgInfo{buildMsgParams})
		//if nxErr != nil {
		//	logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，重发消息失败,err：%v", methodName, nxErr))
		//	return
		//}
		//if !isExist {
		//	session.Commit()
		//}
		//
		//_, nxErr = service.SendMsgList2NX(ctx, sendNxListParamsDto)
		//if nxErr != nil {
		//	logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，重发消息到牛信云失败,err：%v", methodName, nxErr))
		//	return
		//}
	}

}
