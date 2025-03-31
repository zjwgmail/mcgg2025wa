package dto

import (
	"go-fission-activity/activity/model/entity"
	"go-fission-activity/activity/model/nx"
)

type SendNxListParamsDto struct {
	SendMsg       nx.NxReq
	MsgInfoEntity *entity.MsgInfoEntityV2
}
