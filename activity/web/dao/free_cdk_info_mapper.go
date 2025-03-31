package dao

import (
	"github.com/zhuxiujia/GoMybatis"
	"go-fission-activity/activity/model/entity"
)

var freeSdkInfoMapper = &FreeCdkInfoMapper{}

func GetFreeSdkInfoMapper() *FreeCdkInfoMapper {
	return freeSdkInfoMapper
}

type FreeCdkInfoMapper struct {
	GoMybatis.SessionSupport //session事务操作 写法1.  SpaceManageMapper.SessionSupport.NewSession()

	InsertIgnore func(waId string, createAt int64, sendAt int64) (int, error) `args:"waId,create_at,send_at"`

	UpdateStateByWaId func(waId string, sendState int) (int, error) `args:"waId,sendState"`

	SelectWaIdsByStateLtTimestamp func(timestamp int64, sendState int, minId int64, limit int) ([]entity.FreeCdkInfoEntity, error) `args:"timestamp,sendState,minId,limit"`
}
