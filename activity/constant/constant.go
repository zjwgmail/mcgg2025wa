package constant

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	ProfileActives = "PROFILE_ACTIVES"
	DefaultActives = LocalActives
	LocalActives   = "local"
	DevActives     = "dev"
	FatActives     = "fat"
	ProdActives    = "prod"

	AppName = "prod-mcgg2025wa-server"
	Empty   = ""
	Comma   = ","
	Hyphen  = "-"

	ERROR = "error"

	//grpc
	GrpcRunTimeOut = 3 * time.Minute

	LockTimeOut       = 60 * time.Second
	AttendLockTimeOut = 60 * time.Second
	CdkKeyTimeOut     = 60 * time.Second
	RedPacketTimeOut  = 60 * time.Second
	// redis
	UserLock = "activity:v2:%v:lock:user:%v"

	TaskLock                 = "activity:v2:%v:lock:task:%v"
	CdkKey                   = "activity:v2:%v:cdk:%v:list"
	CdkInfoKey               = "activity:v2:%v:cdk:%v:info"
	ServiceIdKey             = "service::v2:id:incr"
	HelpTextLockKey          = "activity:v2:%v:helpText:lock"
	HelpTextClickAllCountKey = "activity:v2:%v:helpText:click:all:count"
	HelpTextClickCountKey    = "activity:v2:%v:helpText:click:%v:count"
	HelpTextWeightKey        = "activity:v2:%v:helpText:weight"
	MsgSignKey               = "activity:v2:%v:msg:sign:%v"

	NotWhiteSetKey         = "activity:v2:%v:notWhite:phoneSet:"
	NotWhiteCountKey       = "activity:v2:%v:notWhite:count:%v:%v:%v"
	SendSuccessMsgCountKey = "activity:v2:%v:sendSuccessMsg:count:%v:%v:%v"
	SendFailMsgCountKey    = "activity:v2:%v:sendFailMsg:count:%v:%v:%v"
	SendTimeOutMsgCountKey = "activity:v2:%v:sendTimeOutMsg:count:%v:%v:%v"

	HelpInfoCacheKey = "activity:v2:%v:helpInfoCache:%v"

	MethodUserAttendMethodInsertMsgInfo = "UserSendMsgMethodInsertMsgInfo"
	RsvMsgInsert2                       = "RsvMsgInsert2"
	MethodUpdateMsgInfo                 = "UpdateMsgInfo"
	MethodInsertMsgInfo                 = "InsertMsgInfo"
	MethodMsgStatusWebHook              = "MsgStatusWebHook"
	MethodHelp                          = "Help"
	MethodSelectByRallyCode             = "SelectByRallyCode"

	MethodInsertMsgInfoReturnFail    = true
	MethodInsertMsgInfoReturnSuccess = false
)

var (
	AppYaml       string = "resources/config/application-" + LookupEnv(ProfileActives, DefaultActives) + ".yml"
	MsgConfigYaml string = "resources/config/msg-" + LookupEnv(ProfileActives, DefaultActives) + ".yml"
)

func LookupEnv(key, defVal string) string {
	val, ok := os.LookupEnv(key)
	if ok {
		return val
	}
	return defVal
}

func GetEnv(key string) string {
	val := os.Getenv(key)
	return val
}

func GetCdkKey(activity int, cdkType string) string {
	return fmt.Sprintf(CdkKey, strconv.Itoa(activity), cdkType)
}

func GetCdkInfoKey(activity int, cdkType string) string {
	return fmt.Sprintf(CdkInfoKey, strconv.Itoa(activity), cdkType)
}

func GetMsgSignKey(activity int, sign string) string {
	return fmt.Sprintf(MsgSignKey, strconv.Itoa(activity), sign)
}

func GetUserLockKey(activity int, waId string) string {
	return fmt.Sprintf(UserLock, strconv.Itoa(activity), waId)
}

func GetTaskLockKey(activity int, taskName string) string {
	return fmt.Sprintf(TaskLock, strconv.Itoa(activity), taskName)
}

func GetHelpTextLockKey(activity int) string {
	return fmt.Sprintf(HelpTextLockKey, strconv.Itoa(activity))
}

func GetHelpTextClickAllCountKey(activity int) string {
	return fmt.Sprintf(HelpTextClickAllCountKey, strconv.Itoa(activity))
}

func GetHelpTextClickCountKey(activity int, helpTextId string) string {
	return fmt.Sprintf(HelpTextClickCountKey, strconv.Itoa(activity), helpTextId)
}

func GetHelpTextWeightKey(activity int) string {
	return fmt.Sprintf(HelpTextWeightKey, strconv.Itoa(activity))
}

func GetNotWhiteSetKey(activity int) string {
	return fmt.Sprintf(NotWhiteSetKey, strconv.Itoa(activity))
}

func GetNotWhiteCountKey(activity int, date, channel, language string) string {
	date = ReplaceChineseMonthDay(date)
	return fmt.Sprintf(NotWhiteCountKey, strconv.Itoa(activity), date, channel, language)
}

func GetSendSuccessMsgCountKey(activity int, date, channel, language string) string {
	date = ReplaceChineseMonthDay(date)
	return fmt.Sprintf(SendSuccessMsgCountKey, strconv.Itoa(activity), date, channel, language)
}

func GetSendFailMsgCountKey(activity int, date, channel, language string) string {
	date = ReplaceChineseMonthDay(date)
	return fmt.Sprintf(SendFailMsgCountKey, strconv.Itoa(activity), date, channel, language)
}

func GetSendTimeOutMsgCountKey(activity int, date, channel, language string) string {
	date = ReplaceChineseMonthDay(date)
	return fmt.Sprintf(SendTimeOutMsgCountKey, strconv.Itoa(activity), date, channel, language)
}

func GetHelpInfoCacheKey(activity int, rallyCode string) string {
	return fmt.Sprintf(HelpInfoCacheKey, strconv.Itoa(activity), rallyCode)
}

func GetTempCsvPath() string {
	if LookupEnv(ProfileActives, DefaultActives) == DefaultActives {
		return "D:\\apps\\tempCsv\\"
	}
	return "/apps/tempCsv/"
}

func ReplaceChineseMonthDay(s string) string {
	// 将字符串中的"月"替换为"M"
	s = strings.ReplaceAll(s, "月", "M")
	// 将字符串中的"日"或"号"替换为"D"
	s = strings.ReplaceAll(s, "日", "D")
	s = strings.ReplaceAll(s, "号", "D")
	return s
}
