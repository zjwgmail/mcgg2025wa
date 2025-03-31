package entity

import "go-fission-activity/util"

type HelpInfoEntityV2 struct {
	Id         int             `json:"id" gm:"id"`
	RallyCode  string          `json:"rally_code" gm:"rally_code"`
	WaId       string          `json:"wa_id" gm:"wa_id"`
	CreatedAt  util.CustomTime `json:"created_at" gm:"created_at"`
	UpdatedAt  util.CustomTime `json:"updated_at" gm:"updated_at"`
	HelpStatus string          `json:"help_status" gm:"help_status"`
}
