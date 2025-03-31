package response

type CdkInfo struct {
	CdkCount        int64   `json:"cdkCount"`
	NextSendPercent float64 `json:"nextSendPercent"`
}
