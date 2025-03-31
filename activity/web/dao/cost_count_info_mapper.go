package dao

import (
	"github.com/zhuxiujia/GoMybatis"
	"go-fission-activity/activity/model/entity"
)

var costCountInfoMapper = &CostCountInfoMapper{}

func GetCostCountInfoMapper() *CostCountInfoMapper {
	return costCountInfoMapper
}

type CostCountInfoMapper struct {
	GoMybatis.SessionSupport //session事务操作 写法1.  SpaceManageMapper.SessionSupport.NewSession()

	SelectByPrimaryKey func(id int) (entity.CostCountInfoEntity, error) `args:"id"`

	InsertSelective func(session *GoMybatis.Session, arg entity.CostCountInfoEntity) (int, error)

	UpdateByPrimaryKeySelective func(session *GoMybatis.Session, arg entity.CostCountInfoEntity) (int, error)
}
