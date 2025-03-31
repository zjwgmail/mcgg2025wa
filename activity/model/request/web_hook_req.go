package request

type WebHookReq struct {
	Messaging_product string    `json:"messaging_product"`
	Statuses          []*Status `json:"statuses"` // 消息类型，固定值”whatsapp“
	Metadata          *Metadata `json:"metadata"`
	App_id            string    `json:"app_id"`
	Business_phone    string    `json:"business_phone"` // 商户电话
	Merchant_phone    string    `json:"merchant_phone"`
	Channel           string    `json:"channel"`

	Contacts []*Contact `json:"contacts"`
	Messages []*Message `json:"messages"`
}

// Status 状态对象
type Status struct {
	Conversation          Conversation             `json:"conversation"`             // 联系人的 WhatsApp ID
	Errors                []*MsgStatusWebHookError `json:"errors"`                   // 错误信息
	RecipientId           string                   `json:"recipient_id"`             // 收件人WhatsApp_id
	Timestamp             string                   `json:"timestamp"`                // 回调时间戳
	Status                string                   `json:"status"`                   // 消息的状态，sent（已发送），delivered（已送达)，read（已读），failed（发送失败）
	Id                    string                   `json:"id"`                       // 消息ID（发送消息时返回的ID）
	Costs                 []*Cost                  `json:"costs"`                    // 费用信息
	MetaMessageId         string                   `json:"meta_message_id"`          // meta原始消息ID, 该字段不一定存在； 发送引用消息时可能用到该字段值，id与meta_message_id 并存时使用meta_message_id作为被引用的消息id，否则使用id
	BizOpaqueCallbackData string                   `json:"biz_opaque_callback_data"` // 发送消息时携带的追踪参数

}

// MsgStatusWebHookError 错误信息
type MsgStatusWebHookError struct {
	Code     int    `json:"code"`      // 平台错误码
	MetaCode int    `json:"meta_code"` // meta错误码
	Title    string `json:"title"`     // 错误信息
}

// Conversation 会话信息
type Conversation struct {
	Id                  string `json:"id"`                   // 会话ID
	ExpirationTimestamp string `json:"expiration_timestamp"` // 会话过期时间戳
	Origin              Origin `json:"origin"`               // 会话类型信息
}

// Origin 会话类型信息
type Origin struct {
	Type string `json:"type"` // 会话类型，marketing（营销会话），utility（通知会话），authentication（验证会话），service（服务会话），referral_conversion（免费会话）
}

// Contact 提供联系人的信息
type Contact struct {
	Wa_id   string   `json:"wa_id"`   // 联系人的 WhatsApp ID
	Profile *Profile `json:"profile"` // 配置文件对象
}
type Profile struct {
	Name string `json:"name"` // 昵称
}

// Message 入站消息
type Message struct {
	From      string `json:"from"`      // 发件人的 WhatsApp ID
	Id        string `json:"id"`        // 消息标识，此 ID 可用于将消息标记为已读
	Timestamp string `json:"timestamp"` // 消息接收时间戳
	/*
		支持接收的消息类型
		1. text 文本
		2. image 图片
		3. video 视频
		4. voice 语音
		5. audio 音频
		6. document 文件
		7. location 位置
		8. sticker 贴图表情
		9. interactive 互动消息
		10. order 下单消息
		11. referral 被广告引流来的客户消息
		12. reaction 心情消息
	*/
	Type        string       `json:"type"`
	Text        *Text        `json:"text,omitempty"`
	Button      *Button      `json:"button,omitempty"`
	Interactive *Interactive `json:"Interactive,omitempty"`
	Cost        *Cost        `json:"cost"` //费用信息
}

type Text struct {
	Body string `json:"body"`
}

type Button struct {
	Text    string `json:"text"`
	Payload string `json:"payload"`
}

type Interactive struct {
	Type        string       `json:"type"`
	ButtonReply *ButtonReply `json:"button_reply"`
}

type ButtonReply struct {
	Id    string `json:"id"`
	Title string `json:"title"`
}

// Cost 费用信息
type Cost struct {
	Currency     string  `json:"currency"`      // 币种
	Price        float64 `json:"price"`         // 客户售价（本币CNY）
	ForeignPrice float64 `json:"foreign_price"` // 客户售价（外币）
	CdrType      string  `json:"cdr_type"`      // cdr类型，4营销 5通知 6验证 7服务 8广告推广这五个类型
	MessageId    string  `json:"message_id"`    // wa消息id
	Direction    string  `json:"direction"`     // 方向，1（下行），2（上行）
}
type Metadata struct {
	Display_phone_number string `json:"display_phone_number"`
	Phone_number_id      string `json:"phone_number_id"`
}
