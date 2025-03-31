package dto

type HelpCacheDto struct {
	// 助力人
	WaId string `json:"wa_id"`
	// 助力人昵称
	UserNickname string `json:"user_nick_name"`
	// 助力码
	RallyCode string `json:"rally_code"`
}
