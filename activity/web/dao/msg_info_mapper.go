package dao

//
//import (
//	"github.com/zhuxiujia/GoMybatis"
//	"go-fission-activity/activity/model/dto"
//	"go-fission-activity/activity/model/entity"
//)
//
//var msgInfoMapper = &MsgInfoMapper{}
//
//func GetMsgInfoMapper() *MsgInfoMapper {
//	return msgInfoMapper
//}
//
//type MsgInfoMapper struct {
//	GoMybatis.SessionSupport //session事务操作 写法1.  SpaceManageMapper.SessionSupport.NewSession()
//
//	SelectByPrimaryKey func(id string) (entity.MsgInfoEntity, error) `args:"id"`
//
//	InsertSelective func(session *GoMybatis.Session, arg entity.MsgInfoEntity) (int, error)
//
//	UpdateByPrimaryKeySelective func(session *GoMybatis.Session, arg entity.MsgInfoEntity) (int, error)
//
//	SumSendPriceMsg func(activityId int) (float64, error) `args:"activity_id"`
//
//	SumPriceSendUnCountMsg func(session *GoMybatis.Session, activityId int, unCounted int) (float64, error) `args:"session,activity_id,un_counted"`
//
//	UpdateCountOfSendUnCount func(session *GoMybatis.Session, activityId int, unCounted int, counted int) (int, error) `args:"session,activity_id,un_counted,counted"`
//
//	SelectWaIdListOfUnSendMsg func(activityId int, msgStatus string) ([]dto.MsgOfWaIdDto, error) `args:"activity_id,msg_status"`
//
//	SelectWaIdListOfUnSendMsgWithPagination func(activityId int, msgStatus string, offset int, limit int) ([]dto.MsgOfWaIdDto, error) `args:"activity_id,msg_status,offset,limit"`
//
//	SelectMsgListOfUnSendMsg func(activityId int, waId string) ([]entity.MsgInfoEntityV2, error) `args:"activity_id,wa_id"`
//
//	CountCdkMsgByWaId func(activityId int, waId string, msgType string) (float64, error) `args:"activity_id,wa_id,msg_type"`
//
//	SelectByWaMessageId func(waMessageId string) (entity.MsgInfoEntityV2, error) `args:"wa_message_id"`
//
//	CountReFreeMsgByPrice func(params dto.GenerationUserQueryDto) ([]dto.ReFreeCountDto, error) `args:"params"`
//}
