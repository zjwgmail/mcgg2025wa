package entity

import "go-fission-activity/util"

type RsvOtherMsgInfo1Entity struct {
	TableName string `json:"table_name" gm:"table_name"`

	Id        int             `json:"id" gm:"id"`
	Msg       string          `json:"msg" gm:"msg"`
	WaId      string          `json:"wa_id" gm:"wa_id"`
	Timestamp int64           `json:"timestamp" gm:"timestamp"`
	CreatedAt util.CustomTime `json:"created_at" gm:"created_at"`
	UpdatedAt util.CustomTime `json:"updated_at" gm:"updated_at"`
}
