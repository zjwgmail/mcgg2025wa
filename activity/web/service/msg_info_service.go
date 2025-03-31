package service

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/zhuxiujia/GoMybatis/tx"
	"go-fission-activity/activity/constant"
	"go-fission-activity/activity/model/entity"
	"go-fission-activity/activity/model/request"
	"go-fission-activity/activity/web/dao"
	"go-fission-activity/activity/web/middleware/logTracing"
	"go-fission-activity/config"
	"go-fission-activity/util"
	"go-fission-activity/util/config/encoder/json"
	"go-fission-activity/util/txUtil"
	"sync"
)

type MsgInfoService struct {
	msgInfoMapper *dao.MsgInfoMapperV2
}

var msgInfoServiceOnce sync.Once
var globalMsgInfoService MsgInfoService

func GetMsgInfoService() *MsgInfoService {
	msgInfoServiceOnce.Do(func() {
		globalMsgInfoService = MsgInfoService{
			msgInfoMapper: dao.GetMsgInfoMapperV2(),
		}
		logTracing.LogPrintfP("第一次使用，globalMsgInfoService")
	})
	return &globalMsgInfoService
}

func (u MsgInfoService) InsertMsgInfo(ctx *gin.Context, msgInfoEntity *entity.MsgInfoEntityV2) error {
	methodName := "InsertMsgInfo"
	session, isExist, err := txUtil.GetTransaction(ctx)
	if nil != err {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，创建事务失败,err：%v", methodName, err))
		return errors.New("database is error")
	}
	if !isExist {
		defer func() {
			session.Rollback()
			session.Close()
		}()
	}
	if msgInfoEntity.Id != "" {
		dbEntity, err := u.msgInfoMapper.SelectByPrimaryKey2(&session, msgInfoEntity.Id)
		if nil != err {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，查询消息失败,err：%v", methodName, err))
			return errors.New("database is error")
		}
		if dbEntity.Id != "" {
			_, err = u.msgInfoMapper.UpdateByPrimaryKeySelective(&session, *msgInfoEntity)
			if nil != err {
				logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，更新消息表失败,SourceWaId:%v,toWaId: %v,err：%v", constant.MethodInsertMsgInfo, msgInfoEntity.SourceWaId, msgInfoEntity.WaId, err))
				return errors.New("database is error")
			}
		} else {
			_, err = u.msgInfoMapper.InsertSelective(&session, *msgInfoEntity)
			if nil != err {
				logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，新增消息表失败,SourceWaId:%v,toWaId: %v,err：%v", constant.MethodInsertMsgInfo, msgInfoEntity.SourceWaId, msgInfoEntity.WaId, err))
				return errors.New("database is error")
			}
		}
	} else {
		_, err = u.msgInfoMapper.InsertSelective(&session, *msgInfoEntity)
		if nil != err {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，新增消息表失败,SourceWaId:%v,toWaId: %v,err：%v", constant.MethodInsertMsgInfo, msgInfoEntity.SourceWaId, msgInfoEntity.WaId, err))
			return errors.New("database is error")
		}
	}

	if !isExist {
		session.Commit()
	}
	return nil
}

func (u MsgInfoService) UpdateMsgInfo(ctx *gin.Context, msgInfoEntity *entity.MsgInfoEntityV2) error {

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

	_, err = u.msgInfoMapper.UpdateByPrimaryKeySelective(&session, *msgInfoEntity)
	if nil != err {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，更新消息表失败,SourceWaId:%v,toWaId: %v,err：%v", constant.MethodUpdateMsgInfo, msgInfoEntity.SourceWaId, msgInfoEntity.WaId, err))
		return errors.New("database is error")
	}

	if !isExist {
		session.Commit()
	}
	return nil
}

func (u MsgInfoService) MsgStatusWebHook(ctx *gin.Context, req *request.MsgStatusWebHookReq) (bool, error) {
	encoder := json.NewEncoder()
	reqAnyEncode, err := encoder.Encode(req)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，将请求解析为二进制数组报错,err：%v", constant.MethodMsgStatusWebHook, err))
		return true, errors.New("webhook message is null")
	}

	if req.Business_phone == "" {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，Business_phone为空,req：%v", constant.MethodMsgStatusWebHook, req))
		return true, errors.New("webhook business_phone is null")
	}
	if config.ApplicationConfig.Nx.BusinessPhone != req.Business_phone {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，Business_phone与配置不匹配,req：%v", constant.MethodMsgStatusWebHook, req))
		return true, errors.New("webhook business_phone is not match")
	}
	if len(req.Statuses) <= 0 {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，Statuses为空,req：%v", constant.MethodMsgStatusWebHook, req))
		return true, errors.New("webhook Statuses is empty")
	}

	for _, statusInfo := range req.Statuses {

		if (constant.NXMsgStatusFailed != statusInfo.Status && constant.NXMsgStatusSent != statusInfo.Status) && (statusInfo.Costs == nil || len(statusInfo.Costs) <= 0) {
			logTracing.LogPrintf(ctx, logTracing.WebHandleLogFmt, fmt.Sprintf("方法[%s]，不是失败和发送状态,并且没有费用内容 不处理", constant.MethodMsgStatusWebHook))
			return true, nil
		}
		id := statusInfo.Id
		extraMap := map[string]string{
			"webhookMessage": string(reqAnyEncode),
		}
		bytes, err := json.NewEncoder().Encode(extraMap)
		if err != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，转换extra为json失败,id:%v", constant.MethodMsgStatusWebHook, id))
			return true, nil
		}

		// 获取月日，格式：x月x日
		time := util.GetNowCustomTime()
		monthDay := fmt.Sprintf("%d月%d日", time.Month(), time.Day())

		msgInfoEntity, err := u.msgInfoMapper.SelectByWaMessageId(id)
		if nil != err {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，查询消息表失败,id:%v,err：%v", constant.MethodMsgStatusWebHook, id, err))
			return true, errors.New("database is error")
		}
		if &msgInfoEntity == nil || msgInfoEntity.Id == "" {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，此消息记录不存在,id:%v,err：%v", constant.MethodMsgStatusWebHook, id, err))
			return true, errors.New("database is error")
		}

		userAttendInfoMapper := dao.GetUserAttendInfoMapperV2()
		userAttendInfo, err := userAttendInfoMapper.SelectByWaId(msgInfoEntity.WaId)
		if nil != err {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，查询用户参与表失败,id:%v,err：%v", constant.MethodMsgStatusWebHook, id, err))
			return true, errors.New("database is error")
		}
		if userAttendInfo.Id <= 0 {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，此用户参与记录不存在,id:%v,err：%v", constant.MethodMsgStatusWebHook, id, err))
			return true, errors.New("database is error")
		}

		session, err := u.msgInfoMapper.SessionSupport.NewSession()
		if nil != err {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，创建事务失败,err：%v", constant.MethodMsgStatusWebHook, err))
			return true, errors.New("database is error")
		}
		p := tx.NewPropagation("")
		session.Begin(&p)
		defer func() {
			session.Rollback()
			session.Close()
		}()
		if constant.NXMsgStatusFailed == statusInfo.Status {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，牛信云发送消息失败,id:%v,err：%v", constant.MethodMsgStatusWebHook, id, err))

			// 统计
			key := constant.GetSendFailMsgCountKey(config.ApplicationConfig.Activity.Id, monthDay, userAttendInfo.Channel, userAttendInfo.Language)
			_, err = AddIncrKey(constant.MethodMsgStatusWebHook, key)
			if err != nil {
				return true, errors.New(fmt.Sprintf("方法[%s]，查询%v 增加%v，报错,err：%v", constant.MethodMsgStatusWebHook, key, err))
			}

			if len(statusInfo.Errors) > 0 && 10002 == statusInfo.Errors[0].Code {
				// 统计
				key = constant.GetSendTimeOutMsgCountKey(config.ApplicationConfig.Activity.Id, monthDay, userAttendInfo.Channel, userAttendInfo.Language)
				_, err = AddIncrKey(constant.MethodMsgStatusWebHook, key)
				if err != nil {
					return true, errors.New(fmt.Sprintf("方法[%s]，查询%v 增加%v，报错,err：%v", constant.MethodMsgStatusWebHook, key, err))
				}
			}

			updateMsgInfo := entity.MsgInfoEntityV2{
				Id:         msgInfoEntity.Id,
				MsgStatus:  constant.NXMsgStatusFailed,
				ReceiveMsg: string(bytes),
			}

			if statusInfo.Costs != nil && len(statusInfo.Costs) > 0 {
				for _, cost := range statusInfo.Costs {
					updateMsgInfo.Currency = cost.Currency
					updateMsgInfo.Price = updateMsgInfo.Price + cost.Price
					updateMsgInfo.ForeignPrice = updateMsgInfo.ForeignPrice + cost.ForeignPrice
				}
			}

			_, err := u.msgInfoMapper.UpdateByPrimaryKeySelective(&session, updateMsgInfo)
			if nil != err {
				logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，更新消息表失败,SourceWaId:%v,toWaId: %v,err：%v", constant.MethodMsgStatusWebHook, msgInfoEntity.SourceWaId, msgInfoEntity.WaId, err))
				return true, errors.New("database is error")
			}
		} else if constant.NXMsgStatusSent == statusInfo.Status {
			logTracing.LogPrintf(ctx, logTracing.WebHandleLogFmt, fmt.Sprintf("方法[%s]，牛信云发送消息成功,id:%v,err：%v", constant.MethodMsgStatusWebHook, id, err))

			// 统计
			key := constant.GetSendSuccessMsgCountKey(config.ApplicationConfig.Activity.Id, monthDay, userAttendInfo.Channel, userAttendInfo.Language)
			_, err = AddIncrKey(constant.MethodMsgStatusWebHook, key)
			if err != nil {
				return true, errors.New(fmt.Sprintf("方法[%s]，查询%v 增加%v，报错,err：%v", constant.MethodMsgStatusWebHook, key, err))
			}

			var currency string
			var price float64
			var foreignPrice float64
			for _, cost := range statusInfo.Costs {
				currency = cost.Currency
				price = price + cost.Price
				foreignPrice = foreignPrice + cost.ForeignPrice
			}
			updateMsgInfo := entity.MsgInfoEntityV2{
				Id:           msgInfoEntity.Id,
				Currency:     currency,
				Price:        price,
				ForeignPrice: foreignPrice,
				MsgStatus:    constant.NXMsgStatusSent,
				ReceiveMsg:   string(bytes),
			}
			_, err := u.msgInfoMapper.UpdateByPrimaryKeySelective(&session, updateMsgInfo)
			if nil != err {
				logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，更新消息表失败,SourceWaId:%v,toWaId: %v,err：%v", constant.MethodMsgStatusWebHook, msgInfoEntity.SourceWaId, msgInfoEntity.WaId, err))
				return true, errors.New("database is error")
			}
		}
		session.Commit()
	}

	return false, nil
}
