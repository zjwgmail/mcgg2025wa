package entity

//
//import "go-fission-activity/util"
//
//type UserAttendInfoEntity struct {
//	Id                    int             `json:"id" gm:"id"`
//	ActivityId            int             `json:"activity_id" gm:"activity_id"`
//	Channel               string          `json:"channel" gm:"channel"`
//	Language              string          `json:"language" gm:"language"`
//	Generation            string          `json:"generation" gm:"generation"`
//	IdentificationCode    string          `json:"identification_code" gm:"identification_code"`
//	WaId                  string          `json:"wa_id" gm:"wa_id"`
//	RallyCode             string          `json:"rally_code" gm:"rally_code"`
//	UserNickname          string          `json:"user_nickname" gm:"user_nickname"`
//	ThreeCdkCode          string          `json:"three_cdk_code" gm:"three_cdk_code"`
//	FiveCdkCode           string          `json:"five_cdk_code" gm:"five_cdk_code"`
//	EightCdkCode          string          `json:"eight_cdk_code" gm:"eight_cdk_code"`
//	AttendAt              util.CustomTime `json:"attend_at" gm:"attend_at"`
//	StartGroupAt          util.CustomTime `json:"start_group_at" gm:"start_group_at"`
//	NewestFreeStartAt     util.CustomTime `json:"newest_free_start_at" gm:"newest_free_start_at"`
//	NewestFreeEndAt       util.CustomTime `json:"newest_free_end_at" gm:"newest_free_end_at"`
//	SendRenewFreeAt       util.CustomTime `json:"send_renew_free_at" gm:"send_renew_free_at"`
//	IsSendRenewFreeMsg    int8            `json:"is_send_renew_free_msg" gm:"is_send_renew_free_msg"`
//	NewestHelpAt          util.CustomTime `json:"newest_help_at" gm:"newest_help_at"`
//	ThreeOverAt           util.CustomTime `json:"three_over_at" gm:"three_over_at"`
//	FiveOverAt            util.CustomTime `json:"five_over_at" gm:"five_over_at"`
//	EightOverAt           util.CustomTime `json:"eight_over_at" gm:"eight_over_at"`
//	AttendStatus          string          `json:"attend_status" gm:"attend_status"`
//	IsThreeStage          int8            `json:"is_three_stage" gm:"is_three_stage"`
//	IsFiveStage           int8            `json:"is_five_stage" gm:"is_five_stage"`
//	CreatedAt             util.CustomTime `json:"created_at" gm:"created_at"`
//	UpdatedAt             util.CustomTime `json:"updated_at" gm:"updated_at"`
//	RedPacketReadyAt      util.CustomTime `json:"red_packet_ready_at" gm:"red_packet_ready_at"`
//	RedPacketSendAt       util.CustomTime `json:"red_packet_send_at" gm:"red_packet_send_at"`
//	Extra                 string          `json:"extra" gm:"extra"`
//	RedPacketCode         string          `json:"red_packet_code" gm:"red_packet_code"`
//	RedPacketStatus       string          `json:"red_packet_status" gm:"red_packet_status"`
//	IsSendCdkMsg          int8            `json:"is_send_cdk_msg" gm:"is_send_cdk_msg"`
//	IsSendClusteringMsg   int8            `json:"is_send_clustering_msg" gm:"is_send_clustering_msg"`
//	SendClusteringAt      util.CustomTime `json:"send_clustering_at" gm:"send_clustering_at"`
//	IsSendPayRenewFreeMsg int8            `json:"is_send_pay_renew_free_msg" gm:"is_send_pay_renew_free_msg"`
//	ShortLink             string          `json:"short_link" gm:"short_link"`
//}
