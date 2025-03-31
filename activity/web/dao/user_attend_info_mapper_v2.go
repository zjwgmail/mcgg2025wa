package dao

import (
	"github.com/zhuxiujia/GoMybatis"
	"go-fission-activity/activity/model/dto"
	"go-fission-activity/activity/model/entity"
)

var userAttendInfoMapperV2 = &UserAttendInfoMapperV2{}

func GetUserAttendInfoMapperV2() *UserAttendInfoMapperV2 {
	return userAttendInfoMapperV2
}

type UserAttendInfoMapperV2 struct {
	GoMybatis.SessionSupport //session事务操作 写法1.  SpaceManageMapper.SessionSupport.NewSession()

	//SelectByPrimaryKey func(id int) (entity.UserAttendInfoEntityV2, error) `args:"id"`

	InsertSelective func(session *GoMybatis.Session, arg entity.UserAttendInfoEntityV2) (int, error)

	UpdateByPrimaryKeySelective func(session *GoMybatis.Session, arg entity.UserAttendInfoEntityV2) (int, error)

	SelectByWaId func(waId string) (entity.UserAttendInfoEntityV2, error) `args:"wa_id"`

	SelectListByWaIdsWithSession func(session *GoMybatis.Session, waIds []string) ([]entity.UserAttendInfoEntityV2, error) `args:"session,wa_ids"`

	SelectListByWaIds func(waIds []string) ([]entity.UserAttendInfoEntityV2, error) `args:"wa_ids"`

	SelectByWaIdBySession func(session *GoMybatis.Session, waId string) (entity.UserAttendInfoEntityV2, error) `args:"session,wa_id"`

	SelectByRallyCode func(rallyCode string) (entity.UserAttendInfoEntityV2, error) `args:"rally_code"`

	CountReCallOfStartGroup func(twoStartGroupTimestamp int64) (int, error) `args:"twoStartGroupTimestamp"`

	SelectReCallOfStartGroup func(pageStart, pageSize int, twoStartGroupTimestamp int64) ([]entity.UserAttendInfoEntityV2, error) `args:"page_start,page_size,twoStartGroupTimestamp"`

	CountRenewFree func(isSendRenewFreeMsg int, currentTimestamp int64) (int, error) `args:"is_send_renew_free_msg,currentTimestamp"`

	SelectRenewFree func(lastId int, pageSize int, isSendRenewFreeMsg int, currentTimestamp int64) ([]entity.UserAttendInfoEntityV2, error) `args:"last_id,page_size,is_send_renew_free_msg,currentTimestamp"`

	CountNotSendCdkUser func(isSendCdkMsg int) (int, error) `args:"is_send_cdk_msg"`

	SelectNotSendCdkUser func(lastId, pageSize int, isSendCdkMsg int) ([]entity.UserAttendInfoEntityV2, error) `args:"last_id,page_size,is_send_cdk_msg"`

	CountReCallOfClustering func(clusteringUnSend int, currentTimestamp int64) (int, error) `args:"clusteringUnSend,currentTimestamp"`

	SelectReCallOfClustering func(lastId int, pageSize int, clusteringUnSend int, currentTimestamp int64) ([]entity.UserAttendInfoEntityV2, error) `args:"last_id,page_size,clusteringUnSend,currentTimestamp"`

	SelectListByGeneration func(startTimestamp int64, endTimestamp int64, minId int, limit int) ([]dto.GenerationUserDto, error) `args:"startTimestamp,endTimestamp,minId,limit"`

	//CountUserByGeneration func(params dto.GenerationUserQueryDto) ([]dto.GenerationUserDto, error) `args:"params"`

	//Select1stIdPayRenewFree func(isSendPayRenewFreeMsg int, diffHourTimestamp int64) (int, error) `args:"is_send_pay_renew_free_msg,diff_hour_timestamp"`

	//CountPayRenewFree func(isSendPayRenewFreeMsg int, diffHourTimestamp int64) (int, error) `args:"is_send_pay_renew_free_msg,diff_hour_timestamp"`

	//SelectPayRenewFree func(lastId, pageSize int, isSendPayRenewFreeMsg int, diffHourTimestamp int64) ([]entity.UserAttendInfoEntityV2, error) `args:"last_id,page_size,is_send_pay_renew_free_msg,diff_hour_timestamp"`

	CountUser func() (int, error)

	SelectListByCodes func(codes []string) ([]entity.UserAttendInfoEntityV2, error) `args:"codes"`
}
