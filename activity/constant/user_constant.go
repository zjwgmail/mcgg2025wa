package constant

const (
	AttendStatusAttend     = "attend"      // 参与活动
	AttendStatusStartGroup = "start_group" // 开团
	AttendStatusThreeOver  = "three_over"  // 3人助力成功
	AttendStatusFiveOver   = "five_over"   // 5人助力成功
	AttendStatusEightOver  = "eight_over"  // 8人助力成功

	RedPacketStatusUn    = "un_red_packet"    // 未发红包
	RedPacketStatusReady = "red_packet_ready" // 预发红包
	RedPacketStatusSend  = "red_packet_send"  // 红包已发送

	FirstIdentificationCode = "00000" // 初代识别码

	RenewFreeSend   = 2 // 已提醒续费
	RenewFreeUnSend = 1 // 未提醒续费

	CdkMsgSend   = 2 // 已发送
	CdkMsgUnSend = 1 // 未发送

	ClusteringSend   = 2 // 已催促成团
	ClusteringUnSend = 1 // 未催促成团

	IsStage    = 2 // 是那个阶段
	IsNotStage = 1 // 不是那个阶段

)
