package entity

import "go-fission-activity/util"

type RsvMsgInfoEntity struct {
	Id             string          `json:"id" gm:"id"`
	Type           string          `json:"type" gm:"type"`
	Msg            string          `json:"msg" gm:"msg"`
	MsgStatus      string          `json:"msg_status" gm:"msg_status"`
	WaId           string          `json:"wa_id" gm:"wa_id"`
	MsgType        string          `json:"msg_type" gm:"msg_type"`
	Currency       string          `json:"currency" gm:"currency"`
	Price          float64         `json:"price" gm:"price"`
	ForeignPrice   float64         `json:"foreign_price" gm:"foreign_price"`
	WaMessageId    string          `json:"wa_message_id" gm:"wa_message_id"`
	CreatedAt      util.CustomTime `json:"created_at" gm:"created_at"`
	UpdatedAt      util.CustomTime `json:"updated_at" gm:"updated_at"`
	IsCount        int8            `json:"is_count" gm:"is_count"`
	SourceWaId     string          `json:"source_wa_id" gm:"source_wa_id"`
	ReceiveMsg     string          `json:"receive_msg" gm:"receive_msg"`
	TraceId        string          `json:"trace_id" gm:"trace_id"`
	SendRes        string          `json:"send_res" gm:"send_res"`
	BuildMsgParams string          `json:"build_msg_params" gm:"build_msg_params"`
}
