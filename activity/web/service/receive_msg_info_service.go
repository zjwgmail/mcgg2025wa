package service

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"go-fission-activity/activity/constant"
	"go-fission-activity/activity/model/entity"
	"go-fission-activity/activity/web/dao"
	"go-fission-activity/activity/web/middleware/logTracing"
	"go-fission-activity/util/txUtil"
	"sync"
)

type ReceiveMsgInfoService struct {
	rsvMsgInfoMapper *dao.RsvMsgInfoMapper
}

var receiveMsgInfoServiceOnce sync.Once
var globalReceiveMsgInfoService ReceiveMsgInfoService

func GetReceiveMsgInfoService() *ReceiveMsgInfoService {
	receiveMsgInfoServiceOnce.Do(func() {
		globalReceiveMsgInfoService = ReceiveMsgInfoService{
			rsvMsgInfoMapper: dao.GetRsvMsgInfoMapper(),
		}
		logTracing.LogPrintfP("第一次使用，globalReceiveMsgInfoService")
	})
	return &globalReceiveMsgInfoService
}

func (u ReceiveMsgInfoService) InsertMsgInfo(ctx *gin.Context, msgInfoEntity *entity.RsvMsgInfoEntity) error {
	methodName := "ReceiveInsertMsgInfo"
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
		dbEntity, err := u.rsvMsgInfoMapper.SelectByPrimaryKey(msgInfoEntity.Id)
		if nil != err {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，查询消息失败,err：%v", methodName, err))
			return errors.New("database is error")
		}
		if dbEntity.Id != "" {
			_, err = u.rsvMsgInfoMapper.UpdateByPrimaryKeySelective(&session, *msgInfoEntity)
			if nil != err {
				logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，更新消息表失败,SourceWaId:%v,toWaId: %v,err：%v", constant.MethodInsertMsgInfo, msgInfoEntity.SourceWaId, msgInfoEntity.WaId, err))
				return errors.New("database is error")
			}
		} else {
			_, err = u.rsvMsgInfoMapper.InsertSelective(&session, *msgInfoEntity)
			if nil != err {
				logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，新增消息表失败,SourceWaId:%v,toWaId: %v,err：%v", constant.MethodInsertMsgInfo, msgInfoEntity.SourceWaId, msgInfoEntity.WaId, err))
				return errors.New("database is error")
			}
		}
	} else {
		_, err = u.rsvMsgInfoMapper.InsertSelective(&session, *msgInfoEntity)
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
