package request

type HelpReq struct {
	Param string `json:"param"`
}

type HelpParam struct {
	WaId      string `json:"wa_id"`
	RallyCode string `json:"rally_code"`
	IsHelp    bool   `json:"is_help"`
}

type HelpTextCountReq struct {
	HelpTextId string `json:"helpTextId"`
}
