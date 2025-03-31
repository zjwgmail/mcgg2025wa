package response

type NXResponse struct {

	//状态码
	Code int `json:"code"`

	//返回消息
	Message string `json:"message"`

	TraceId string `json:"traceId"`

	// 响应信息
	Data *NXData `json:"data"`
}

type NXData struct {
	MessagingProduct string `json:"messaging_product"`

	Messages []*NXMessage `json:"messages"`

	//图片上传返回的id
	Id string `json:"id"`
}

type NXMessage struct {
	Id string `json:"id"`
}

type NXSendRes struct {
	NXResponse *NXResponse `json:"NXResponse"`
}
