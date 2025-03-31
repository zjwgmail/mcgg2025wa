package dao

import (
	"github.com/zhuxiujia/GoMybatis"
	"go-fission-activity/activity/model/entity"
)

var reportMsgInfoMapper = &ReportMsgInfoMapper{}

func GetReportMsgInfoMapper() *ReportMsgInfoMapper {
	return reportMsgInfoMapper
}

type ReportMsgInfoMapper struct {
	GoMybatis.SessionSupport //session事务操作 写法1.  SpaceManageMapper.SessionSupport.NewSession()

	//SelectByPrimaryKey func(id int) (entity.ReportMsgInfoEntity, error) `args:"id"`

	InsertSelective func(session *GoMybatis.Session, arg entity.ReportMsgInfoEntity) (int, error)

	UpdateByPrimaryKeySelective func(session *GoMybatis.Session, arg entity.ReportMsgInfoEntity) (int, error)

	SelectListByReportType func(activityId int, reportType string) ([]entity.ReportMsgInfoEntity, error) `args:"activity_id,report_type"`

	//SelectDays func(id int) ([]string, error) `args:"id"`

	SelectCountByReportTypeAndDay func(activityId int, reportType string, monthDay string) (int64, error) `args:"activity_id,report_type,month_day"`
}
