package nx

import "go-fission-activity/config"

// NxReq 牛信请求
type NxReq struct {
	Params        *NxReqParam
	CommonHeaders map[string]string
}

type NxReqParam struct {
	Appkey           string `json:"appkey"`
	BusinessPhone    string `json:"business_phone"`
	MessagingProduct string `json:"messaging_product"`
	RecipientType    string `json:"recipient_type"`
	To               string `json:"to"`
	CusMessageId     string `json:"cus_message_id"`
	Type             string `json:"type"`
	// 互动消息
	Interactive *Interactive `json:"interactive,omitempty"`
	// 模板消息
	Template *config.Template `json:"template,omitempty"`
}

type Interactive struct {
	Type   string                  `json:"type,omitempty"`
	Header *NxReqInteractiveHeader `json:"header,omitempty"`
	Body   *NxReqInteractiveBody   `json:"body,omitempty"`
	Footer *NxReqInteractiveFooter `json:"footer,omitempty"`
	Action *NxReqInteractiveAction `json:"action,omitempty"`
}

type NxReqInteractiveHeader struct {
	Type  string                 `json:"type,omitempty"`
	Image *NxReqInteractiveImage `json:"image,omitempty"`
}

type NxReqInteractiveImage struct {
	Link string `json:"link,omitempty"`
}

type NxReqInteractiveBody struct {
	Text string `json:"text,omitempty"`
}

type NxReqInteractiveFooter struct {
	Text string `json:"text,omitempty"`
}

type NxReqInteractiveAction struct {
	// cta_url
	Name       string                `json:"name,omitempty"`
	Parameters *NxReqActionParameter `json:"parameters,omitempty"`
	// button
	Buttons []*config.Button `json:"buttons,omitempty"`
}

type NxReqActionParameter struct {
	DisplayText string `json:"display_text,omitempty"`
	Url         string `json:"url,omitempty"`
}
