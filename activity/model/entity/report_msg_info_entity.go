package entity

import "go-fission-activity/util"

type ReportMsgInfoEntity struct {
	Id         string          `json:"id" gm:"id"`
	Date       string          `json:"date" gm:"date"`
	Hour       string          `json:"hour" gm:"hour"`
	ReportType string          `json:"report_type" gm:"report_type"`
	MsgStatus  string          `json:"msg_status" gm:"msg_status"`
	Msg        string          `json:"msg" gm:"msg"`
	CountMsg   string          `json:"count_msg" gm:"count_msg"`
	Res        string          `json:"res" gm:"res"`
	CreatedAt  util.CustomTime `json:"created_at" gm:"created_at"`
	UpdatedAt  util.CustomTime `json:"updated_at" gm:"updated_at"`
}
