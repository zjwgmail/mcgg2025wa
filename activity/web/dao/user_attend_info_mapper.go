package dao

//
//import (
//	"github.com/zhuxiujia/GoMybatis"
//	"go-fission-activity/activity/model/dto"
//	"go-fission-activity/activity/model/entity"
//)
//
//var userAttendInfoMapper = &UserAttendInfoMapper{}
//
//func GetUserAttendInfoMapper() *UserAttendInfoMapper {
//	return userAttendInfoMapper
//}
//
//type UserAttendInfoMapper struct {
//	GoMybatis.SessionSupport //session事务操作 写法1.  SpaceManageMapper.SessionSupport.NewSession()
//
//	SelectByPrimaryKey func(id int) (entity.UserAttendInfoEntity, error) `args:"id"`
//
//	InsertSelective func(session *GoMybatis.Session, arg entity.UserAttendInfoEntity) (int, error)
//
//	UpdateByPrimaryKeySelective func(session *GoMybatis.Session, arg entity.UserAttendInfoEntity) (int, error)
//
//	SelectByWaId func(activityId int, waId string) (entity.UserAttendInfoEntity, error) `args:"activity_id,wa_id"`
//
//	SelectByWaIdBySession func(session *GoMybatis.Session, activityId int, waId string) (entity.UserAttendInfoEntity, error) `args:"session,activity_id,wa_id"`
//
//	SelectByRallyCode func(activityId int, rallyCode string) (entity.UserAttendInfoEntity, error) `args:"activity_id,rally_code"`
//
//	CountReCallOfUnRedPacket func(activityId int, unRedPacketMinute int) (int, error) `args:"activity_id,unRedPacketMinute"`
//
//	SelectReCallOfUnRedPacket func(activityId int, pageStart, pageSize int, unRedPacketMinute int) ([]entity.UserAttendInfoEntity, error) `args:"activity_id,page_start,page_size,unRedPacketMinute"`
//
//	CountReCallOfSendRedPacket func(activityId int, sendRedPacketMinute int) (int, error) `args:"activity_id,sendRedPacketMinute"`
//
//	SelectReCallOfSendRedPacket func(activityId int, pageStart, pageSize int, sendRedPacketMinute int) ([]entity.UserAttendInfoEntity, error) `args:"activity_id,page_start,page_size,sendRedPacketMinute"`
//
//	CountReCallOfStartGroup func(activityId int, twoStartGroupMinute int) (int, error) `args:"activity_id,twoStartGroupMinute"`
//
//	SelectReCallOfStartGroup func(activityId int, pageStart, pageSize int, twoStartGroupMinute int) ([]entity.UserAttendInfoEntity, error) `args:"activity_id,page_start,page_size,twoStartGroupMinute"`
//
//	CountRenewFree func(activityId int, isSendRenewFreeMsg int) (int, error) `args:"activity_id,is_send_renew_free_msg"`
//
//	SelectRenewFree func(activityId int, lastId int, pageSize int, isSendRenewFreeMsg int) ([]entity.UserAttendInfoEntity, error) `args:"activity_id,last_id,page_size,is_send_renew_free_msg"`
//
//	CountNotSendCdkUser func(activityId int, isSendCdkMsg int) (int, error) `args:"activity_id,is_send_cdk_msg"`
//
//	SelectNotSendCdkUser func(activityId int, lastId, pageSize int, isSendCdkMsg int) ([]entity.UserAttendInfoEntity, error) `args:"activity_id,last_id,page_size,is_send_cdk_msg"`
//
//	CountReCallOfClustering func(activityId int, notAttendStatus string, clusteringUnSend int) (int, error) `args:"activity_id,not_attend_status,clusteringUnSend"`
//
//	SelectReCallOfClustering func(activityId int, lastId, pageSize int, notAttendStatus string, clusteringUnSend int) ([]entity.UserAttendInfoEntity, error) `args:"activity_id,last_id,page_size,not_attend_status,clusteringUnSend"`
//
//	CountUserByGeneration func(params dto.GenerationUserQueryDto) ([]dto.GenerationUserDto, error) `args:"params"`
//
//	CountPayRenewFree func(activityId int, isSendPayRenewFreeMsg int, diffHour int) (int, error) `args:"activity_id,is_send_pay_renew_free_msg,diff_hour"`
//
//	SelectPayRenewFree func(activityId int, lastId, pageSize int, isSendPayRenewFreeMsg int, diffHour int) ([]entity.UserAttendInfoEntity, error) `args:"activity_id,last_id,page_size,is_send_pay_renew_free_msg,diff_hour"`
//
//	CountUserByActivityId func(activityId int) (int, error) `args:"activity_id"`
//}
