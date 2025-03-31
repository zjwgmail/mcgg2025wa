package datasource

import (
	"go-fission-activity/activity/web/dao"
	"go-fission-activity/config"
)

func MybatisMapperInit() {
	InitMybatisXMLConfig(mybatisMapperPathFill("UserAttendInfoMapperV2.xml"), dao.GetUserAttendInfoMapperV2())
	InitMybatisXMLConfig(mybatisMapperPathFill("ActivityInfoMapper.xml"), dao.GetActivityInfoMapper())
	InitMybatisXMLConfig(mybatisMapperPathFill("CostCountInfoMapper.xml"), dao.GetCostCountInfoMapper())
	InitMybatisXMLConfig(mybatisMapperPathFill("HelpInfoMapperV2.xml"), dao.GetHelpInfoMapperV2())
	InitMybatisXMLConfig(mybatisMapperPathFill("MsgInfoMapperV2.xml"), dao.GetMsgInfoMapperV2())
	InitMybatisXMLConfig(mybatisMapperPathFill("RsvMsgInfoMapper.xml"), dao.GetRsvMsgInfoMapper())
	InitMybatisXMLConfig(mybatisMapperPathFill("RsvOtherMsgInfo1Mapper.xml"), dao.GetRsvOtherMsgInfo1Mapper())
	InitMybatisXMLConfig(mybatisMapperPathFill("ReportMsgInfoMapper.xml"), dao.GetReportMsgInfoMapper())
	InitMybatisXMLConfig(mybatisMapperPathFill("FreeCdkInfoMapper.xml"), dao.GetFreeSdkInfoMapper())
	InitMybatisXMLConfig(mybatisMapperPathFill("DDLMapper.xml"), dao.GetDDLMapper())
}

func mybatisMapperPathFill(xml string) string {
	return config.ApplicationConfig.Datasource.XmlPrefix + xml
}
