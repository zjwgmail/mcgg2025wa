package dao

import (
	"github.com/zhuxiujia/GoMybatis"
	"go-fission-activity/activity/model/entity"
)

var rsvOtherMsgInfo1Mapper = &RsvOtherMsgInfo1Mapper{}

func GetRsvOtherMsgInfo1Mapper() *RsvOtherMsgInfo1Mapper {
	return rsvOtherMsgInfo1Mapper
}

type RsvOtherMsgInfo1Mapper struct {
	GoMybatis.SessionSupport //session事务操作 写法1.  SpaceManageMapper.SessionSupport.NewSession()

	SelectByPrimaryKey func(id int) (entity.RsvOtherMsgInfo1Entity, error) `args:"id"`

	InsertSelective func(session *GoMybatis.Session, arg entity.RsvOtherMsgInfo1Entity) (int, error)

	InsertSelective2 func(arg entity.RsvOtherMsgInfo1Entity) (int, error)

	ExecuteSql func(sql string) ([]map[string]interface{}, error) `args:"sql"`
}
