package request

type MsgStatusWebHookReq struct {
	Messaging_product string    `json:"messaging_product"`
	Statuses          []*Status `json:"statuses"` // 消息类型，固定值”whatsapp“
	Metadata          *Metadata `json:"metadata"`
	App_id            string    `json:"app_id"`
	Business_phone    string    `json:"business_phone"` // 商户电话
	Merchant_phone    string    `json:"merchant_phone"`
	Channel           string    `json:"channel"`
}
