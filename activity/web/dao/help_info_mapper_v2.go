package dao

import (
	"github.com/zhuxiujia/GoMybatis"
	"go-fission-activity/activity/model/dto"
	"go-fission-activity/activity/model/entity"
)

var helpInfoMapperV2 = &HelpInfoMapperV2{}

func GetHelpInfoMapperV2() *HelpInfoMapperV2 {
	return helpInfoMapperV2
}

type HelpInfoMapperV2 struct {
	GoMybatis.SessionSupport //session事务操作 写法1.  SpaceManageMapper.SessionSupport.NewSession()

	//SelectByPrimaryKey func(id int) (entity.HelpInfoEntityV2, error) `args:"id"`

	InsertSelective func(session *GoMybatis.Session, arg entity.HelpInfoEntityV2) (int, error)

	//UpdateByPrimaryKeySelective func(session *GoMybatis.Session, arg entity.HelpInfoEntityV2) (int, error)

	//CountByRallyCode func(rallyCode string) (int, error) `args:"rally_code"`

	SelectByWaId func(waId string) (entity.HelpInfoEntityV2, error) `args:"wa_id"`

	SelectListByRallyCode func(session *GoMybatis.Session, rallyCode string) ([]entity.HelpInfoEntityV2, error) `args:"session,rally_code"`

	SelectDistinctCodeByTimestamp func(startTimestamp int64, endTimestamp int64, minId string, limit int) ([]dto.HelpCacheDto, error) `args:"startTimestamp,endTimestamp,minId,limit"`

	CountByCodesTimestamp func(codes []string, startTimestamp int64, endTimestamp int64) ([]dto.CreatorHelpCountDTO, error) `args:"codes,startTimestamp,endTimestamp"`

	//CountUserByHelpCount func(params dto.GenerationUserQueryDto) ([]dto.HelpCountDto, error) `args:"params"`
}
