package response

type AttendInfo struct {
	AttendCode string `json:"attendCode"`
	CdkCode    string `json:"cdkCode"`
	WaId       string `json:"waId"`
}
