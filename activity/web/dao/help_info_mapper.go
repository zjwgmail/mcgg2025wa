package dao

//
//import (
//	"github.com/zhuxiujia/GoMybatis"
//	"go-fission-activity/activity/model/dto"
//	"go-fission-activity/activity/model/entity"
//)
//
//var helpInfoMapper = &HelpInfoMapper{}
//
//func GetHelpInfoMapper() *HelpInfoMapper {
//	return helpInfoMapper
//}
//
//type HelpInfoMapper struct {
//	GoMybatis.SessionSupport //session事务操作 写法1.  SpaceManageMapper.SessionSupport.NewSession()
//
//	SelectByPrimaryKey func(id int) (entity.HelpInfoEntity, error) `args:"id"`
//
//	InsertSelective func(session *GoMybatis.Session, arg entity.HelpInfoEntity) (int, error)
//
//	UpdateByPrimaryKeySelective func(session *GoMybatis.Session, arg entity.HelpInfoEntity) (int, error)
//
//	CountByRallyCode func(activityId int, rallyCode string) (int, error) `args:"activity_id,rally_code"`
//
//	SelectCountByWaId func(activityId int, waId string) (int, error) `args:"activity_id,wa_id"`
//
//	SelectByWaIdAndActivityId func(activityId int, waId string) (dto.HelpCacheDto, error) `args:"activity_id,wa_id"`
//
//	SelectHelpNameByRallyCode func(session *GoMybatis.Session, activityId int, rallyCode string) ([]entity.UserAttendInfoEntity, error) `args:"session,activity_id,rally_code"`
//
//	CountUserByHelpCount func(params dto.GenerationUserQueryDto) ([]dto.HelpCountDto, error) `args:"params"`
//}
