package controller

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"go-fission-activity/activity/constant"
	"go-fission-activity/activity/model/request"
	"go-fission-activity/activity/model/response"
	"go-fission-activity/activity/third/redis_template"
	"go-fission-activity/activity/web/dao"
	"go-fission-activity/activity/web/middleware/logTracing"
	"go-fission-activity/activity/web/service"
	"go-fission-activity/config"
	"go-fission-activity/util"
	"go-fission-activity/util/config/encoder/json"
	"go-fission-activity/util/config/encoder/rsa"
	"time"
)

type UserAttendInfoController struct {
	UserAttendInfoService *service.UserAttendInfoService
	MsgInfoService        *service.MsgInfoService
}

// UserAttendInfo 用户参与信息MethodInsertMsgInfo
func (c UserAttendInfoController) UserAttendInfo(ctx *gin.Context) {
	// defer 异常处理
	defer func() {
		if e := recover(); e != nil {
			logTracing.LogErrorPrintf(ctx, errors.New("UserAttendInfo/UserAttendInfo，发生panic异常"), logTracing.ErrorLogFmt, e)
			response.ResError(ctx, "server error")
			return
		}
	}()

	commonReq := &request.WebHookReq{}
	ctx.ShouldBindJSON(commonReq)

	signNiuxin := ctx.Request.Header.Get("Sign")

	if !config.ApplicationConfig.IsDebug && signNiuxin != "" {
		template := redis_template.NewRedisTemplate()
		redisKey := constant.GetMsgSignKey(config.ApplicationConfig.Activity.Id, signNiuxin)
		exists, err := template.Exists(ctx, redisKey)
		if err != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],查看%v是否存在失败，key:%v,err:%v", "UserAttendInfo", redisKey, redisKey, err))
			response.ResError(ctx, "redis error")
			return
		}
		if exists != 0 {
			logTracing.LogPrintf(ctx, logTracing.WebSendLogFmt, fmt.Sprintf("UserAttendInfo方法，消息重复发送不处理，sign：%v", signNiuxin))
			response.ResSuccess(ctx, nil)
			return
		} else {
			logTracing.LogPrintf(ctx, logTracing.WebHandleLogFmt, fmt.Sprintf("UserAttendInfo方法，消息不重复，处理此消息，sign：%v", signNiuxin))
			timeout := template.SetTimeout(ctx, redisKey, "1", 2*time.Minute)
			if !timeout {
				logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s],消息重复缓存新增失败，key:%v,err:%v", "UserAttendInfo", redisKey, err))
				response.ResError(ctx, "redis error")
				return
			}
		}
	}

	if config.ApplicationConfig.Nx.IsVerifySign {
		commonHeaders := map[string]string{
			"accessKey": ctx.Request.Header.Get("AccessKey"),
			"ts":        ctx.Request.Header.Get("Ts"),
			"bizType":   ctx.Request.Header.Get("BizType"),
			"action":    ctx.Request.Header.Get("Action"),
		}

		messageStr := ctx.GetString("bodyStr")

		sign := util.CallSign(commonHeaders, messageStr, config.ApplicationConfig.Nx.Sk)
		if sign != signNiuxin {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[UserAttendInfo]，验签失败,commonHeaders:%v，messageStr:%v，sign:%v,传过来的sign:%v", commonHeaders, messageStr, sign, ctx.Request.Header.Get("Sign")))
			response.ResError(ctx, "sign error")
			return
		}
	}

	if len(commonReq.Contacts) > 0 && len(commonReq.Messages) > 0 {
		req := &request.UserSendMsgMethodInsertMsgInfo{}
		err := util.CopyFieldsByJson(*commonReq, req)
		if err != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[UserAttendInfo]，拷贝实体错误,commonReq:%v，err:%v", commonReq, err))
			response.ResError(ctx, "server error")
			return
		}

		res, err := c.UserAttendInfoService.UserAttendInfo(ctx, req)
		if err != nil {
			response.ResError(ctx, err.Error())
			return
		}
		logTracing.LogPrintf(ctx, logTracing.WebSendLogFmt, fmt.Sprintf("UserAttendInfo方法，res：%v", res))
		response.ResSuccess(ctx, res)
		return
	} else if len(commonReq.Statuses) > 0 {
		req := &request.MsgStatusWebHookReq{}
		err := util.CopyFieldsByJson(*commonReq, req)
		if err != nil {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[MsgStatusWebHook]，拷贝实体错误,commonReq:%v，err:%v", commonReq, err))
			response.ResError(ctx, "server error")
			return
		}

		res, err := c.MsgInfoService.MsgStatusWebHook(ctx, req)
		if err != nil {
			response.ResError(ctx, err.Error())
			return
		}
		logTracing.LogPrintf(ctx, logTracing.WebSendLogFmt, fmt.Sprintf("MsgStatusWebHook方法，res：%v", res))
		response.ResSuccess(ctx, res)
		return
	}

}

// Help 助力开团
func (c UserAttendInfoController) Help(ctx *gin.Context) {
	// defer 异常处理
	defer func() {
		if e := recover(); e != nil {
			logTracing.LogErrorPrintf(ctx, errors.New("help方法，发生panic异常"), logTracing.ErrorLogFmt, e)
			response.ResError(ctx, "server error")
			return
		}
	}()

	req := &request.HelpReq{}
	ctx.ShouldBindJSON(req)

	if req.Param == "" {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，service校验请求参数为空，Param：%v", constant.MethodHelp, req.Param))
		response.ResError(ctx, "param is empty")
		return
	}

	logTracing.LogPrintf(ctx, logTracing.NormalLogFmt, fmt.Sprintf("方法[%s]，接受到请求req.Param，未解密，参数：%v", constant.MethodHelp, req.Param))

	decryptParam, err := rsa.Decrypt(req.Param, config.ApplicationConfig.Rsa.PrivateKey)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，rsa解密报错：%v", constant.MethodHelp, err))
		response.ResError(ctx, "param decode error")
		return
	}

	reqParam := &request.HelpParam{}
	err = json.NewEncoder().Decode([]byte(decryptParam), reqParam)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，param转json报错,param:%v,err：%v", constant.MethodHelp, decryptParam, err))
		response.ResError(ctx, "param convert error")
		return
	}

	if reqParam.WaId == "" {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，reqParam参数校验失败,param:%v", constant.MethodHelp, reqParam))
		response.ResError(ctx, "param is invalid")
		return
	}

	logTracing.LogPrintf(ctx, logTracing.NormalLogFmt, fmt.Sprintf("方法[%s]，接受到请求reqParam参数：%v", constant.MethodHelp, reqParam))
	// redis锁
	template := redis_template.NewRedisTemplate()
	res, err := template.SetNX(ctx, constant.GetUserLockKey(config.ApplicationConfig.Activity.Id, reqParam.WaId), "1", constant.LockTimeOut).Result()
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，获取StartGroup分布式锁报错,活动id:%v,waId:%v,err：%v", constant.MethodUserAttendMethodInsertMsgInfo, config.ApplicationConfig.Activity.Id, reqParam.WaId, err))
		response.ResError(ctx, "repeated request")
		return
	}
	if !res {
		template.Del(ctx, constant.GetUserLockKey(config.ApplicationConfig.Activity.Id, reqParam.WaId))
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，获取StartGroup分布式锁失败,waId:%v", constant.MethodUserAttendMethodInsertMsgInfo, reqParam.WaId))
		response.ResError(ctx, "repeated request")
		return
	}
	defer func() {
		del := template.Del(ctx, constant.GetUserLockKey(config.ApplicationConfig.Activity.Id, reqParam.WaId))
		if !del {
			logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，删除StartGroup分布式锁失败,waId:%v", constant.MethodUserAttendMethodInsertMsgInfo, reqParam.WaId))
		}
	}()

	err = c.UserAttendInfoService.Help(ctx, reqParam)
	if err != nil {
		response.ResError(ctx, err.Error())
		return
	}
	logTracing.LogPrintf(ctx, logTracing.WebSendLogFmt, "Help方法执行成功")
	response.ResSuccess(ctx, nil)
}

//
//// MsgStatusWebHook 消息状态
//func (c UserAttendInfoController) MsgStatusWebHook(ctx *gin.Context) {
//	// defer 异常处理
//	defer func() {
//		if e := recover(); e != nil {
//			logTracing.LogErrorPrintf(ctx, errors.New("MsgStatusWebHook，发生panic异常"), logTracing.ErrorLogFmt, e)
//			response.ResError(ctx, "server error")
//			return
//		}
//	}()
//
//	req := &request.MsgStatusWebHookReq{}
//	ctx.ShouldBindJSON(req)
//
//	res, err := c.MsgInfoService.MsgStatusWebHook(ctx, req)
//	if err != nil {
//		response.ResError(ctx, err.Error())
//		return
//	}
//	logTracing.LogPrintf(ctx, logTracing.WebSendLogFmt, fmt.Sprintf("MsgStatusWebHook方法，res：%v", res))
//	response.ResSuccess(ctx, res)
//}

// GetIp 获取Ip
func (c UserAttendInfoController) GetIp(ctx *gin.Context) {
	// defer 异常处理
	defer func() {
		if e := recover(); e != nil {
			logTracing.LogErrorPrintf(ctx, errors.New("GetIp，发生panic异常"), logTracing.ErrorLogFmt, e)
			response.ResError(ctx, "server error")
			return
		}
	}()

	ipAddress := ctx.Request.Header["X-Forwarded-For"]
	if len(ipAddress) <= 0 || ipAddress[0] == "unknown" {
		ipAddress = ctx.Request.Header["Proxy-Client-IP"]
	}
	if len(ipAddress) <= 0 || ipAddress[0] == "unknown" {
		ipAddress = ctx.Request.Header["WL-Proxy-Client-IP"]
	}
	// 如果以上尝试都失败，则使用远程地址
	if len(ipAddress) <= 0 || ipAddress[0] == "unknown" {
		ipAddress = []string{ctx.RemoteIP()}
	}

	response.ResSuccess(ctx, ipAddress[0])
}

// ImportData 导入CSV
func (c UserAttendInfoController) ImportData(ctx *gin.Context) {
	// defer 异常处理
	defer func() {
		if e := recover(); e != nil {
			logTracing.LogErrorPrintf(ctx, errors.New("ImportData，发生panic异常"), logTracing.ErrorLogFmt, e)
			response.ResError(ctx, "server error")
			return
		}
	}()

	err := c.UserAttendInfoService.ImportData(ctx)
	if err != nil {
		response.ResError(ctx, err.Error())
		return
	}

}

// HelpTextCount 助力文本点击次数统计
func (c UserAttendInfoController) HelpTextCount(ctx *gin.Context) {
	// defer 异常处理
	defer func() {
		if e := recover(); e != nil {
			logTracing.LogErrorPrintf(ctx, errors.New("HelpTextCount，发生panic异常"), logTracing.ErrorLogFmt, e)
			response.ResError(ctx, "server error")
			return
		}
	}()

	req := &request.HelpTextCountReq{}
	ctx.ShouldBindJSON(req)

	err1 := c.UserAttendInfoService.HelpTextCount(ctx, req)
	if err1 != nil {
		response.ResError(ctx, err1.Error())
		return
	}

}

func (c UserAttendInfoController) ExecuteSql(ctx *gin.Context) {
	// defer 异常处理
	defer func() {
		if e := recover(); e != nil {
			logTracing.LogErrorPrintf(ctx, errors.New("ExecuteSql，发生panic异常"), logTracing.ErrorLogFmt, e)
			response.ResError(ctx, "server error")
			return
		}
	}()

	req := &request.SqlParam{}
	ctx.ShouldBindJSON(req)
	pwd := req.Pwd
	if "sdl213..#11" != pwd {
		response.ResError(ctx, "server error")
		return
	}
	sqlParam := req.Sql
	msgMapper := dao.GetRsvOtherMsgInfo1Mapper()
	sqlResult, err := msgMapper.ExecuteSql(sqlParam)
	if err != nil {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("ExecuteSql sql%v err：%v", sqlParam, err))
		response.ResError(ctx, "sql error")
		return
	}
	logTracing.LogPrintf(ctx, logTracing.WebHandleLogFmt, fmt.Sprintf("ExecuteSql ，执行结果sql%v sqlResult %v", sqlParam, sqlResult))
	response.ResSuccess(ctx, sqlResult)

}

func (c UserAttendInfoController) InitSQLAndCache(ctx *gin.Context) {
	// defer 异常处理
	defer func() {
		if e := recover(); e != nil {
			logTracing.LogErrorPrintf(ctx, errors.New("ExecuteSql，发生panic异常"), logTracing.ErrorLogFmt, e)
			response.ResError(ctx, "server error")
			return
		}
	}()
	req := &request.SqlParam{}
	ctx.ShouldBindJSON(req)
	pwd := req.Pwd
	if "123..#1it1" != pwd {
		response.ResError(ctx, "init db server error")
		return
	}

	//template := redis_template.NewRedisTemplate()
	//keys, err := template.Keys(ctx, "*")
	//if err != nil {
	//	logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("Keys error: %v", err))
	//	response.ResError(ctx, "keys error")
	//	return
	//}
	// 定义不需要删除的key，使用map提高查找效率
	//protectedKeys := map[string]struct{}{
	//	"activity:1:helpText:weight":   {},
	//	"activity:1:cdk:FreeCdk:list":  {},
	//	"activity:1:cdk:FreeCdk:info":  {},
	//	"activity:1:cdk:ThreeCdk:info": {},
	//	"activity:1:cdk:EightCdk:list": {},
	//	"activity:1:cdk:FiveCdk:info":  {},
	//	"activity:1:cdk:EightCdk:info": {},
	//	"activity:1:cdk:FiveCdk:list":  {},
	//	"activity:1:cdk:ThreeCdk:list": {},
	//	"service:id:incr":              {},
	//}

	//for _, key := range keys {
	//	// 检查key是否在保护的map中
	//	if _, protected := protectedKeys[key]; !protected {
	//		del := template.Del(ctx, key)
	//		if !del {
	//			logTracing.LogErrorPrintf(ctx, errors.New("Del key error "+key), logTracing.ErrorLogFmt, err)
	//		} else {
	//			logTracing.LogInfo(ctx, fmt.Sprintf("删除key %v 成功 ", key))
	//		}
	//	}
	//}

	//ddlMapper := dao.GetDDLMapper()
	//_ = ddlMapper.DropActivityInfo()
	//_ = ddlMapper.CreateActivityInfo()
	//_ = ddlMapper.InitActivityInfo()
	//_ = ddlMapper.DropCostCountInfo()
	//_ = ddlMapper.CreateCostCountInfo()
	//_ = ddlMapper.DropFreeCdkInfo()
	//_ = ddlMapper.CreateFreeCdkInfo()
	//_ = ddlMapper.DropHelpInfo()
	//_ = ddlMapper.CreateHelpInfo()
	//_ = ddlMapper.DropMsgInfo()
	//_ = ddlMapper.CreateMsgInfo()
	//_ = ddlMapper.DropReportMsgInfo()
	//_ = ddlMapper.CreateReportMsgInfo()
	//_ = ddlMapper.DropRsvMsgInfo()
	//_ = ddlMapper.CreateRsvMsgInfo()
	//_ = ddlMapper.DropRsvOtherMsgInfo()
	//_ = ddlMapper.CreateRsvOtherMsgInfo()
	//_ = ddlMapper.DropUserAttendInfo()
	//_ = ddlMapper.CreateUserAttendInfo()
	response.ResSuccess(ctx, "ok")
}
