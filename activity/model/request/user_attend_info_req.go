package request

type UserSendMsgMethodInsertMsgInfo struct {
	Contacts          []*Contact `json:"contacts"`
	Messages          []*Message `json:"messages"`
	Metadata          *Metadata  `json:"metadata"`
	Business_phone    string     `json:"business_phone"`
	Messaging_product string     `json:"messaging_product"`
	App_id            string     `json:"app_id"`
	Channel           string     `json:"channel"`
	Merchant_phone    string     `json:"merchant_phone"`
}
