package dao

import (
	"github.com/zhuxiujia/GoMybatis"
	"go-fission-activity/activity/model/dto"
	"go-fission-activity/activity/model/entity"
)

var msgInfoMapperV2 = &MsgInfoMapperV2{}

func GetMsgInfoMapperV2() *MsgInfoMapperV2 {
	return msgInfoMapperV2
}

type MsgInfoMapperV2 struct {
	GoMybatis.SessionSupport //session事务操作 写法1.  SpaceManageMapper.SessionSupport.NewSession()

	//SelectByPrimaryKey func(id string) (entity.MsgInfoEntityV2, error) `args:"id"`

	SelectByPrimaryKey2 func(session *GoMybatis.Session, id string) (entity.MsgInfoEntityV2, error) `args:"session,id"`

	InsertSelective func(session *GoMybatis.Session, arg entity.MsgInfoEntityV2) (int, error)

	UpdateByPrimaryKeySelective func(session *GoMybatis.Session, arg entity.MsgInfoEntityV2) (int, error)

	//SumSendPriceMsg func() (float64, error)

	SumPriceSendUnCountMsg func(session *GoMybatis.Session, unCounted int) (float64, error) `args:"session,un_counted"`

	UpdateCountOfSendUnCount func(session *GoMybatis.Session, unCounted int, counted int) (int, error) `args:"session,un_counted,counted"`

	//SelectWaIdListOfUnSendMsg func(msgStatus string) ([]dto.MsgOfWaIdDto, error) `args:"msg_status"`

	//SelectWaIdListOfUnSendMsgWithPagination func(msgStatus string, offset int, limit int) ([]dto.MsgOfWaIdDto, error) `args:"msg_status,offset,limit"`

	SelectWaIdListOfUnSendMsg func(minId string, limit int) ([]dto.MsgOfWaIdDto, error) `args:"minId,limit"`

	SelectMsgListOfUnSendMsg func(waId string) ([]entity.MsgInfoEntityV2, error) `args:"wa_id"`

	CountCdkMsgByWaId func(waId string, msgType string) (float64, error) `args:"wa_id,msg_type"`

	SelectByWaMessageId func(waMessageId string) (entity.MsgInfoEntityV2, error) `args:"wa_message_id"`

	//CountReFreeMsgByPrice func(params dto.GenerationUserQueryDto) ([]dto.ReFreeCountDto, error) `args:"params"`

	SelectListByMsgType func(startTimestamp int64, endTimestamp int64, minId string, limit int) ([]entity.MsgInfoEntityV2, error) `args:"startTimestamp,endTimestamp,minId,limit"`
}
