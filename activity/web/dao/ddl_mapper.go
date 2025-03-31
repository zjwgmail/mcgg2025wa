package dao

import (
	"github.com/zhuxiujia/GoMybatis"
)

var ddlMapper = &DDLMapper{}

func GetDDLMapper() *DDLMapper {
	return ddlMapper
}

type DDLMapper struct {
	GoMybatis.SessionSupport //session事务操作 写法1.  SpaceManageMapper.SessionSupport.NewSession()

	DropActivityInfo func() error

	CreateActivityInfo func() error

	InitActivityInfo func() error

	DropCostCountInfo func() error

	CreateCostCountInfo func() error

	DropFreeCdkInfo func() error

	CreateFreeCdkInfo func() error

	DropHelpInfo func() error

	CreateHelpInfo func() error

	DropMsgInfo func() error

	CreateMsgInfo func() error

	DropReportMsgInfo func() error

	CreateReportMsgInfo func() error

	DropRsvMsgInfo func() error

	CreateRsvMsgInfo func() error

	DropRsvOtherMsgInfo func() error

	CreateRsvOtherMsgInfo func() error

	DropUserAttendInfo func() error

	CreateUserAttendInfo func() error
}
