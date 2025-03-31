package entity

type FreeCdkInfoEntity struct {
	Id        int64  `json:"id" gm:"id"`
	WaId      string `json:"wa_id" gm:"wa_id"`
	SendState int    `json:"send_state" gm:"send_state"`
}
