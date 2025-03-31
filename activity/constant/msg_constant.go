package constant

const (
	ReceiveMsg                 = "receiveMsg"                 // 接收的消息
	ActivityTaskMsg            = "activityTaskMsg"            // 参与活动消息
	CannotAttendActivityMsg    = "cannotAttendActivityMsg"    // 不能参与活动消息
	RepeatHelpMsg              = "repeatHelpMsg"              // 重复助力消息
	StartGroupMsg              = "startGroupMsg"              // 开团消息
	HelpStartGroupMsg          = "helpStartGroupMsg"          // 受邀人开团消息
	FounderCanNotStartGroupMsg = "founderCanNotStartGroupMsg" // 主态不能开团消息
	CanNotStartGroupMsg        = "canNotStartGroupMsg"        // 不能开团消息
	HelpTaskSingleStartMsg     = "helpTaskSingleStartMsg"     // 助力人参与活动信息
	HelpTaskSingleSuccessMsg   = "helpTaskSingleSuccessMsg"   // 被人助力成功信息
	HelpThreeOverMsg           = "helpThreeOverMsg"           // 3人助力完成信息
	HelpFiveOverMsg            = "helpFiveOverMsg"            // 5人助力完成信息
	HelpEightOverMsg           = "helpEightOverMsg"           // 8人助力完成信息
	FreeCdkMsg                 = "freeCdkMsg"                 // 免费CDK信息
	RedPacketReadyMsg          = "redPacketReadyMsg"          // 红包预发信息
	RedPacketSendMsg           = "redPacketSendMsg"           // 红包发放信息
	RenewFreeMsg               = "renewFreeMsg"               // 续免费信息
	PayRenewFreeMsg            = "payRenewFreeMsg"            // 付费-续免费信息
	PromoteClusteringMsg       = "promoteClusteringMsg"       // 催促成团消息
	EndCanNotStartGroupMsg     = "endCanNotStartGroupMsg"     // 结束期-不能开团消息
	EndCanNotHelpMsg           = "endCanNotHelpMsg"           // 结束期-不能助力消息
	RenewFreeReplyMsg          = "renewFreeReplyMsg"          // 续订回复信息

	NXMsgStatusReceive     = "receive"
	NXMsgStatusOwnerUnSent = "owner_un_send"
	NXMsgStatusOwnerSent   = "owner_send"
	NXMsgStatusSent        = "sent"
	NXMsgStatusFailed      = "failed"

	WaRedirectListPrefix = "https://wa.me/?text="

	Counted   = 2 // 已统计
	UnCounted = 1 // 未统计

	BizTypeInteractive = 1 // 互动消息类型
	BizTypeTemplate    = 2 // 模板消息类型

)
