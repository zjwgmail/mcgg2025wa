package dao

import (
	"github.com/zhuxiujia/GoMybatis"
	"go-fission-activity/activity/model/entity"
)

var rsvMsgInfoMapper = &RsvMsgInfoMapper{}

func GetRsvMsgInfoMapper() *RsvMsgInfoMapper {
	return rsvMsgInfoMapper
}

type RsvMsgInfoMapper struct {
	GoMybatis.SessionSupport //session事务操作 写法1.  SpaceManageMapper.SessionSupport.NewSession()

	SelectByPrimaryKey func(id string) (entity.RsvMsgInfoEntity, error) `args:"id"`

	InsertSelective func(session *GoMybatis.Session, arg entity.RsvMsgInfoEntity) (int, error)

	//InsertSelective2 func(arg entity.RsvMsgInfoEntity) (int, error)

	UpdateByPrimaryKeySelective func(session *GoMybatis.Session, arg entity.RsvMsgInfoEntity) (int, error)
}
