package response

type ShortLinkResponse struct {
	//状态码
	Code int `json:"code"`
	// 响应信息
	Data *ShortLinkData `json:"data"`
}

type ShortLinkData struct {
	ShortUrl string `json:"short_url"`
}
