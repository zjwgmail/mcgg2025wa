package dao

import (
	"github.com/zhuxiujia/GoMybatis"
	"go-fission-activity/activity/model/entity"
)

var activityInfoMapper = &ActivityInfoMapper{}

func GetActivityInfoMapper() *ActivityInfoMapper {
	return activityInfoMapper
}

type ActivityInfoMapper struct {
	GoMybatis.SessionSupport //session事务操作 写法1.  SpaceManageMapper.SessionSupport.NewSession()

	SelectByPrimaryKey func(id int) (entity.ActivityInfoEntity, error) `args:"id"`

	InsertSelective func(session *GoMybatis.Session, arg entity.ActivityInfoEntity) (int, error)

	UpdateByPrimaryKeySelective func(session *GoMybatis.Session, arg entity.ActivityInfoEntity) (int, error)

	SelectStatusByPrimaryKey func(id int) (string, error) `args:"id"`
}
