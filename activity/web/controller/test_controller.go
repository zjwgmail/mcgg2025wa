package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go-fission-activity/activity/constant"
	"go-fission-activity/activity/model/dto"
	"go-fission-activity/activity/model/entity"
	"go-fission-activity/activity/task/statistics"
	"go-fission-activity/activity/web/dao"
	"go-fission-activity/activity/web/middleware/logTracing"
	"go-fission-activity/util/txUtil"
	"net/http"
	"strconv"
	"time"
)

type TestController struct {
}

func (c TestController) FreeSdkInfo(ctx *gin.Context) {
	freeSdkInfoMapper := dao.GetFreeSdkInfoMapper()
	_, _ = freeSdkInfoMapper.InsertIgnore("852550003108", time.Now().Unix(), time.Now().Unix())
	_, _ = freeSdkInfoMapper.UpdateStateByWaId("852550003108", 2)
	freeCdkInfoEntity, _ := freeSdkInfoMapper.SelectWaIdsByStateLtTimestamp(time.Now().Unix(), 2, 0, 100)
	ctx.JSON(http.StatusOK, freeCdkInfoEntity)
}

func (c TestController) RsvMsgInfo(ctx *gin.Context) {
	session, isExist, err := txUtil.GetTransaction(ctx)
	if nil != err {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，创建事务失败,err：%v", "FreeSdkInfo", err))
		return
	}
	if !isExist {
		defer func() {
			session.Rollback()
			session.Close()
		}()
	}
	rsvMsgInfoMapper := dao.GetRsvMsgInfoMapper()
	rsvMsgInfoEntity := entity.RsvMsgInfoEntity{
		Id:         "1967951514916532224",
		Type:       "receive",
		Msg:        "{\"contacts\":[{\"wa_id\":\"85257481920\",\"profile\":{\"name\":\"chen\"}}],\"messages\":[{\"from\":\"85257481920\",\"id\":\"wamid.HBgLODUyNTc0ODE5MjAVAgASGCA5QjQwRTE3RDU3QzhGRTMwOEJFRkVDQjc1MDQ5QTFGQwA=\",\"timestamp\":\"1734189319\",\"type\":\"interactive\",\"Interactive\":{\"type\":\"button_reply\",\"button_reply\":{\"id\":\"1\",\"title\":\"续订活动消息\"}},\"cost\":{\"currency\":\"USD\",\"price\":0,\"foreign_price\":0,\"cdr_type\":\"\",\"message_id\":\"wamid.HBgLODUyNTc0ODE5MjAVAgASGCA5QjQwRTE3RDU3QzhGRTMwOEJFRkVDQjc1MDQ5QTFGQwA=\",\"direction\":\"\"}}],\"metadata\":{\"display_phone_number\":\"639692369842\",\"phone_number_id\":\"296085720248808\"},\"business_phone\":\"639692369842\",\"messaging_product\":\"whatsapp\",\"app_id\":\"1533\",\"channel\":\"\",\"merchant_phone\":\"639692369842\"}",
		MsgStatus:  "receive",
		WaId:       "85257483108",
		MsgType:    "receiveMsg",
		IsCount:    1,
		SourceWaId: "85257483108",
	}
	rsvMsgInfoEntity.IsCount = 2
	_, _ = rsvMsgInfoMapper.InsertSelective(&session, rsvMsgInfoEntity)
	_, _ = rsvMsgInfoMapper.UpdateByPrimaryKeySelective(&session, rsvMsgInfoEntity)
	session.Commit()
	session.Close()
	rsvMsgInfoEntity, _ = rsvMsgInfoMapper.SelectByPrimaryKey("1967951514916532224")
	ctx.JSON(http.StatusOK, rsvMsgInfoEntity)
}

func (c TestController) ReportMsgInfo(ctx *gin.Context) {
	session, isExist, err := txUtil.GetTransaction(ctx)
	if nil != err {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，创建事务失败,err：%v", "FreeSdkInfo", err))
		return
	}
	if !isExist {
		defer func() {
			session.Rollback()
			session.Close()
		}()
	}
	reportMsgInfoMapper := dao.GetReportMsgInfoMapper()
	reportMsgInfoEntity := entity.ReportMsgInfoEntity{
		Id:         "1975882677471625216",
		Date:       "1月13日",
		ReportType: "excel",
		MsgStatus:  "owner_un_send",
		Msg:        "[{\"date\":\"1月4日\",\"language\":\"中文\",\"channel\":\"端内-128通路\",\"generation01\":0,\"generation02\":0,\"generation03\":0,\"generation04\":0,\"generation05\":0,\"generation06After\":0,\"generation02After\":0,\"help1\":0,\"help2\":0,\"help3\":0,\"help4\":0,\"help5\":0,\"help6\":0,\"help7\":0,\"help8\":0,\"allHelp1\":0,\"allHelp2\":0,\"allHelp3\":0,\"allHelp4\":0,\"allHelp5\":0,\"allHelp6\":0,\"allHelp7\":0,\"allHelp8\":0,\"promoteClusteringCount\":0,\"freeRemindCount\":0,\"payRemindCount\":0,\"sendSuccessMsgCount\":0,\"sendFailMsgCount\":0,\"sendTimeOutMsgCount\":0,\"notWhiteCount\":0},{\"date\":\"1月4日\",\"language\":\"英语\",\"channel\":\"端内-128通路\",\"generation01\":0,\"generation02\":0,\"generation03\":0,\"generation04\":0,\"generation05\":0,\"generation06After\":0,\"generation02After\":0,\"help1\":0,\"help2\":0,\"help3\":0,\"help4\":0,\"help5\":0,\"help6\":0,\"help7\":0,\"help8\":0,\"allHelp1\":0,\"allHelp2\":0,\"allHelp3\":0,\"allHelp4\":0,\"allHelp5\":0,\"allHelp6\":0,\"allHelp7\":0,\"allHelp8\":0,\"promoteClusteringCount\":0,\"freeRemindCount\":0,\"payRemindCount\":0,\"sendSuccessMsgCount\":0,\"sendFailMsgCount\":0,\"sendTimeOutMsgCount\":0,\"notWhiteCount\":0},{\"date\":\"1月4日\",\"language\":\"马来语\",\"channel\":\"端内-128通路\",\"generation01\":0,\"generation02\":0,\"generation03\":0,\"generation04\":0,\"generation05\":0,\"generation06After\":0,\"generation02After\":0,\"help1\":0,\"help2\":0,\"help3\":0,\"help4\":0,\"help5\":0,\"help6\":0,\"help7\":0,\"help8\":0,\"allHelp1\":0,\"allHelp2\":0,\"allHelp3\":0,\"allHelp4\":0,\"allHelp5\":0,\"allHelp6\":0,\"allHelp7\":0,\"allHelp8\":0,\"promoteClusteringCount\":0,\"freeRemindCount\":0,\"payRemindCount\":0,\"sendSuccessMsgCount\":0,\"sendFailMsgCount\":0,\"sendTimeOutMsgCount\":0,\"notWhiteCount\":0},{\"date\":\"1月4日\",\"language\":\"中文\",\"channel\":\"端内-邮件推送\",\"generation01\":0,\"generation02\":0,\"generation03\":0,\"generation04\":0,\"generation05\":0,\"generation06After\":0,\"generation02After\":0,\"help1\":0,\"help2\":0,\"help3\":0,\"help4\":0,\"help5\":0,\"help6\":0,\"help7\":0,\"help8\":0,\"allHelp1\":0,\"allHelp2\":0,\"allHelp3\":0,\"allHelp4\":0,\"allHelp5\":0,\"allHelp6\":0,\"allHelp7\":0,\"allHelp8\":0,\"promoteClusteringCount\":0,\"freeRemindCount\":0,\"payRemindCount\":0,\"sendSuccessMsgCount\":0,\"sendFailMsgCount\":0,\"sendTimeOutMsgCount\":0,\"notWhiteCount\":0},{\"date\":\"1月4日\",\"language\":\"英语\",\"channel\":\"端内-邮件推送\",\"generation01\":0,\"generation02\":0,\"generation03\":0,\"generation04\":0,\"generation05\":0,\"generation06After\":0,\"generation02After\":0,\"help1\":0,\"help2\":0,\"help3\":0,\"help4\":0,\"help5\":0,\"help6\":0,\"help7\":0,\"help8\":0,\"allHelp1\":0,\"allHelp2\":0,\"allHelp3\":0,\"allHelp4\":0,\"allHelp5\":0,\"allHelp6\":0,\"allHelp7\":0,\"allHelp8\":0,\"promoteClusteringCount\":0,\"freeRemindCount\":0,\"payRemindCount\":0,\"sendSuccessMsgCount\":0,\"sendFailMsgCount\":0,\"sendTimeOutMsgCount\":0,\"notWhiteCount\":0},{\"date\":\"1月4日\",\"language\":\"马来语\",\"channel\":\"端内-邮件推送\",\"generation01\":0,\"generation02\":0,\"generation03\":0,\"generation04\":0,\"generation05\":0,\"generation06After\":0,\"generation02After\":0,\"help1\":0,\"help2\":0,\"help3\":0,\"help4\":0,\"help5\":0,\"help6\":0,\"help7\":0,\"help8\":0,\"allHelp1\":0,\"allHelp2\":0,\"allHelp3\":0,\"allHelp4\":0,\"allHelp5\":0,\"allHelp6\":0,\"allHelp7\":0,\"allHelp8\":0,\"promoteClusteringCount\":0,\"freeRemindCount\":0,\"payRemindCount\":0,\"sendSuccessMsgCount\":0,\"sendFailMsgCount\":0,\"sendTimeOutMsgCount\":0,\"notWhiteCount\":0},{\"date\":\"1月4日\",\"language\":\"中文\",\"channel\":\"端内-任务达人\",\"generation01\":0,\"generation02\":0,\"generation03\":0,\"generation04\":0,\"generation05\":0,\"generation06After\":0,\"generation02After\":0,\"help1\":0,\"help2\":0,\"help3\":0,\"help4\":0,\"help5\":0,\"help6\":0,\"help7\":0,\"help8\":0,\"allHelp1\":0,\"allHelp2\":0,\"allHelp3\":0,\"allHelp4\":0,\"allHelp5\":0,\"allHelp6\":0,\"allHelp7\":0,\"allHelp8\":0,\"promoteClusteringCount\":0,\"freeRemindCount\":0,\"payRemindCount\":0,\"sendSuccessMsgCount\":0,\"sendFailMsgCount\":0,\"sendTimeOutMsgCount\":0,\"notWhiteCount\":0},{\"date\":\"1月4日\",\"language\":\"英语\",\"channel\":\"端内-任务达人\",\"generation01\":0,\"generation02\":0,\"generation03\":0,\"generation04\":0,\"generation05\":0,\"generation06After\":0,\"generation02After\":0,\"help1\":0,\"help2\":0,\"help3\":0,\"help4\":0,\"help5\":0,\"help6\":0,\"help7\":0,\"help8\":0,\"allHelp1\":0,\"allHelp2\":0,\"allHelp3\":0,\"allHelp4\":0,\"allHelp5\":0,\"allHelp6\":0,\"allHelp7\":0,\"allHelp8\":0,\"promoteClusteringCount\":0,\"freeRemindCount\":0,\"payRemindCount\":0,\"sendSuccessMsgCount\":0,\"sendFailMsgCount\":0,\"sendTimeOutMsgCount\":0,\"notWhiteCount\":0},{\"date\":\"1月4日\",\"language\":\"马来语\",\"channel\":\"端内-任务达人\",\"generation01\":0,\"generation02\":0,\"generation03\":0,\"generation04\":0,\"generation05\":0,\"generation06After\":0,\"generation02After\":0,\"help1\":0,\"help2\":0,\"help3\":0,\"help4\":0,\"help5\":0,\"help6\":0,\"help7\":0,\"help8\":0,\"allHelp1\":0,\"allHelp2\":0,\"allHelp3\":0,\"allHelp4\":0,\"allHelp5\":0,\"allHelp6\":0,\"allHelp7\":0,\"allHelp8\":0,\"promoteClusteringCount\":0,\"freeRemindCount\":0,\"payRemindCount\":0,\"sendSuccessMsgCount\":0,\"sendFailMsgCount\":0,\"sendTimeOutMsgCount\":0,\"notWhiteCount\":0},{\"date\":\"1月4日\",\"language\":\"中文\",\"channel\":\"端外-FB\",\"generation01\":0,\"generation02\":0,\"generation03\":0,\"generation04\":0,\"generation05\":0,\"generation06After\":0,\"generation02After\":0,\"help1\":0,\"help2\":0,\"help3\":0,\"help4\":0,\"help5\":0,\"help6\":0,\"help7\":0,\"help8\":0,\"allHelp1\":0,\"allHelp2\":0,\"allHelp3\":0,\"allHelp4\":0,\"allHelp5\":0,\"allHelp6\":0,\"allHelp7\":0,\"allHelp8\":0,\"promoteClusteringCount\":0,\"freeRemindCount\":0,\"payRemindCount\":0,\"sendSuccessMsgCount\":0,\"sendFailMsgCount\":0,\"sendTimeOutMsgCount\":0,\"notWhiteCount\":0},{\"date\":\"1月4日\",\"language\":\"英语\",\"channel\":\"端外-FB\",\"generation01\":0,\"generation02\":0,\"generation03\":0,\"generation04\":0,\"generation05\":0,\"generation06After\":0,\"generation02After\":0,\"help1\":0,\"help2\":0,\"help3\":0,\"help4\":0,\"help5\":0,\"help6\":0,\"help7\":0,\"help8\":0,\"allHelp1\":0,\"allHelp2\":0,\"allHelp3\":0,\"allHelp4\":0,\"allHelp5\":0,\"allHelp6\":0,\"allHelp7\":0,\"allHelp8\":0,\"promoteClusteringCount\":0,\"freeRemindCount\":0,\"payRemindCount\":0,\"sendSuccessMsgCount\":0,\"sendFailMsgCount\":0,\"sendTimeOutMsgCount\":0,\"notWhiteCount\":0},{\"date\":\"1月4日\",\"language\":\"马来语\",\"channel\":\"端外-FB\",\"generation01\":0,\"generation02\":0,\"generation03\":0,\"generation04\":0,\"generation05\":0,\"generation06After\":0,\"generation02After\":0,\"help1\":0,\"help2\":0,\"help3\":0,\"help4\":0,\"help5\":0,\"help6\":0,\"help7\":0,\"help8\":0,\"allHelp1\":0,\"allHelp2\":0,\"allHelp3\":0,\"allHelp4\":0,\"allHelp5\":0,\"allHelp6\":0,\"allHelp7\":0,\"allHelp8\":0,\"promoteClusteringCount\":0,\"freeRemindCount\":0,\"payRemindCount\":0,\"sendSuccessMsgCount\":0,\"sendFailMsgCount\":0,\"sendTimeOutMsgCount\":0,\"notWhiteCount\":0},{\"date\":\"1月4日\",\"language\":\"中文\",\"channel\":\"端外-INS\",\"generation01\":0,\"generation02\":0,\"generation03\":0,\"generation04\":0,\"generation05\":0,\"generation06After\":0,\"generation02After\":0,\"help1\":0,\"help2\":0,\"help3\":0,\"help4\":0,\"help5\":0,\"help6\":0,\"help7\":0,\"help8\":0,\"allHelp1\":0,\"allHelp2\":0,\"allHelp3\":0,\"allHelp4\":0,\"allHelp5\":0,\"allHelp6\":0,\"allHelp7\":0,\"allHelp8\":0,\"promoteClusteringCount\":0,\"freeRemindCount\":0,\"payRemindCount\":0,\"sendSuccessMsgCount\":0,\"sendFailMsgCount\":0,\"sendTimeOutMsgCount\":0,\"notWhiteCount\":0},{\"date\":\"1月4日\",\"language\":\"英语\",\"channel\":\"端外-INS\",\"generation01\":0,\"generation02\":0,\"generation03\":0,\"generation04\":0,\"generation05\":0,\"generation06After\":0,\"generation02After\":0,\"help1\":0,\"help2\":0,\"help3\":0,\"help4\":0,\"help5\":0,\"help6\":0,\"help7\":0,\"help8\":0,\"allHelp1\":0,\"allHelp2\":0,\"allHelp3\":0,\"allHelp4\":0,\"allHelp5\":0,\"allHelp6\":0,\"allHelp7\":0,\"allHelp8\":0,\"promoteClusteringCount\":0,\"freeRemindCount\":0,\"payRemindCount\":0,\"sendSuccessMsgCount\":0,\"sendFailMsgCount\":0,\"sendTimeOutMsgCount\":0,\"notWhiteCount\":0},{\"date\":\"1月4日\",\"language\":\"马来语\",\"channel\":\"端外-INS\",\"generation01\":0,\"generation02\":0,\"generation03\":0,\"generation04\":0,\"generation05\":0,\"generation06After\":0,\"generation02After\":0,\"help1\":0,\"help2\":0,\"help3\":0,\"help4\":0,\"help5\":0,\"help6\":0,\"help7\":0,\"help8\":0,\"allHelp1\":0,\"allHelp2\":0,\"allHelp3\":0,\"allHelp4\":0,\"allHelp5\":0,\"allHelp6\":0,\"allHelp7\":0,\"allHelp8\":0,\"promoteClusteringCount\":0,\"freeRemindCount\":0,\"payRemindCount\":0,\"sendSuccessMsgCount\":0,\"sendFailMsgCount\":0,\"sendTimeOutMsgCount\":0,\"notWhiteCount\":0},{\"date\":\"1月4日\",\"language\":\"中文\",\"channel\":\"端外-UA加热\",\"generation01\":0,\"generation02\":0,\"generation03\":0,\"generation04\":0,\"generation05\":0,\"generation06After\":0,\"generation02After\":0,\"help1\":0,\"help2\":0,\"help3\":0,\"help4\":0,\"help5\":0,\"help6\":0,\"help7\":0,\"help8\":0,\"allHelp1\":0,\"allHelp2\":0,\"allHelp3\":0,\"allHelp4\":0,\"allHelp5\":0,\"allHelp6\":0,\"allHelp7\":0,\"allHelp8\":0,\"promoteClusteringCount\":0,\"freeRemindCount\":0,\"payRemindCount\":0,\"sendSuccessMsgCount\":0,\"sendFailMsgCount\":0,\"sendTimeOutMsgCount\":0,\"notWhiteCount\":0},{\"date\":\"1月4日\",\"language\":\"英语\",\"channel\":\"端外-UA加热\",\"generation01\":0,\"generation02\":0,\"generation03\":0,\"generation04\":0,\"generation05\":0,\"generation06After\":0,\"generation02After\":0,\"help1\":0,\"help2\":0,\"help3\":0,\"help4\":0,\"help5\":0,\"help6\":0,\"help7\":0,\"help8\":0,\"allHelp1\":0,\"allHelp2\":0,\"allHelp3\":0,\"allHelp4\":0,\"allHelp5\":0,\"allHelp6\":0,\"allHelp7\":0,\"allHelp8\":0,\"promoteClusteringCount\":0,\"freeRemindCount\":0,\"payRemindCount\":0,\"sendSuccessMsgCount\":0,\"sendFailMsgCount\":0,\"sendTimeOutMsgCount\":0,\"notWhiteCount\":0},{\"date\":\"1月4日\",\"language\":\"马来语\",\"channel\":\"端外-UA加热\",\"generation01\":0,\"generation02\":0,\"generation03\":0,\"generation04\":0,\"generation05\":0,\"generation06After\":0,\"generation02After\":0,\"help1\":0,\"help2\":0,\"help3\":0,\"help4\":0,\"help5\":0,\"help6\":0,\"help7\":0,\"help8\":0,\"allHelp1\":0,\"allHelp2\":0,\"allHelp3\":0,\"allHelp4\":0,\"allHelp5\":0,\"allHelp6\":0,\"allHelp7\":0,\"allHelp8\":0,\"promoteClusteringCount\":0,\"freeRemindCount\":0,\"payRemindCount\":0,\"sendSuccessMsgCount\":0,\"sendFailMsgCount\":0,\"sendTimeOutMsgCount\":0,\"notWhiteCount\":0},{\"date\":\"1月4日\",\"language\":\"中文\",\"channel\":\"端外-备用1\",\"generation01\":0,\"generation02\":0,\"generation03\":0,\"generation04\":0,\"generation05\":0,\"generation06After\":0,\"generation02After\":0,\"help1\":0,\"help2\":0,\"help3\":0,\"help4\":0,\"help5\":0,\"help6\":0,\"help7\":0,\"help8\":0,\"allHelp1\":0,\"allHelp2\":0,\"allHelp3\":0,\"allHelp4\":0,\"allHelp5\":0,\"allHelp6\":0,\"allHelp7\":0,\"allHelp8\":0,\"promoteClusteringCount\":0,\"freeRemindCount\":0,\"payRemindCount\":0,\"sendSuccessMsgCount\":0,\"sendFailMsgCount\":0,\"sendTimeOutMsgCount\":0,\"notWhiteCount\":0},{\"date\":\"1月4日\",\"language\":\"英语\",\"channel\":\"端外-备用1\",\"generation01\":0,\"generation02\":0,\"generation03\":0,\"generation04\":0,\"generation05\":0,\"generation06After\":0,\"generation02After\":0,\"help1\":0,\"help2\":0,\"help3\":0,\"help4\":0,\"help5\":0,\"help6\":0,\"help7\":0,\"help8\":0,\"allHelp1\":0,\"allHelp2\":0,\"allHelp3\":0,\"allHelp4\":0,\"allHelp5\":0,\"allHelp6\":0,\"allHelp7\":0,\"allHelp8\":0,\"promoteClusteringCount\":0,\"freeRemindCount\":0,\"payRemindCount\":0,\"sendSuccessMsgCount\":0,\"sendFailMsgCount\":0,\"sendTimeOutMsgCount\":0,\"notWhiteCount\":0},{\"date\":\"1月4日\",\"language\":\"马来语\",\"channel\":\"端外-备用1\",\"generation01\":0,\"generation02\":0,\"generation03\":0,\"generation04\":0,\"generation05\":0,\"generation06After\":0,\"generation02After\":0,\"help1\":0,\"help2\":0,\"help3\":0,\"help4\":0,\"help5\":0,\"help6\":0,\"help7\":0,\"help8\":0,\"allHelp1\":0,\"allHelp2\":0,\"allHelp3\":0,\"allHelp4\":0,\"allHelp5\":0,\"allHelp6\":0,\"allHelp7\":0,\"allHelp8\":0,\"promoteClusteringCount\":0,\"freeRemindCount\":0,\"payRemindCount\":0,\"sendSuccessMsgCount\":0,\"sendFailMsgCount\":0,\"sendTimeOutMsgCount\":0,\"notWhiteCount\":0},{\"date\":\"1月4日\",\"language\":\"中文\",\"channel\":\"端外-备用2\",\"generation01\":0,\"generation02\":0,\"generation03\":0,\"generation04\":0,\"generation05\":0,\"generation06After\":0,\"generation02After\":0,\"help1\":0,\"help2\":0,\"help3\":0,\"help4\":0,\"help5\":0,\"help6\":0,\"help7\":0,\"help8\":0,\"allHelp1\":0,\"allHelp2\":0,\"allHelp3\":0,\"allHelp4\":0,\"allHelp5\":0,\"allHelp6\":0,\"allHelp7\":0,\"allHelp8\":0,\"promoteClusteringCount\":0,\"freeRemindCount\":0,\"payRemindCount\":0,\"sendSuccessMsgCount\":0,\"sendFailMsgCount\":0,\"sendTimeOutMsgCount\":0,\"notWhiteCount\":0},{\"date\":\"1月4日\",\"language\":\"英语\",\"channel\":\"端外-备用2\",\"generation01\":0,\"generation02\":0,\"generation03\":0,\"generation04\":0,\"generation05\":0,\"generation06After\":0,\"generation02After\":0,\"help1\":0,\"help2\":0,\"help3\":0,\"help4\":0,\"help5\":0,\"help6\":0,\"help7\":0,\"help8\":0,\"allHelp1\":0,\"allHelp2\":0,\"allHelp3\":0,\"allHelp4\":0,\"allHelp5\":0,\"allHelp6\":0,\"allHelp7\":0,\"allHelp8\":0,\"promoteClusteringCount\":0,\"freeRemindCount\":0,\"payRemindCount\":0,\"sendSuccessMsgCount\":0,\"sendFailMsgCount\":0,\"sendTimeOutMsgCount\":0,\"notWhiteCount\":0},{\"date\":\"1月4日\",\"language\":\"马来语\",\"channel\":\"端外-备用2\",\"generation01\":0,\"generation02\":0,\"generation03\":0,\"generation04\":0,\"generation05\":0,\"generation06After\":0,\"generation02After\":0,\"help1\":0,\"help2\":0,\"help3\":0,\"help4\":0,\"help5\":0,\"help6\":0,\"help7\":0,\"help8\":0,\"allHelp1\":0,\"allHelp2\":0,\"allHelp3\":0,\"allHelp4\":0,\"allHelp5\":0,\"allHelp6\":0,\"allHelp7\":0,\"allHelp8\":0,\"promoteClusteringCount\":0,\"freeRemindCount\":0,\"payRemindCount\":0,\"sendSuccessMsgCount\":0,\"sendFailMsgCount\":0,\"sendTimeOutMsgCount\":0,\"notWhiteCount\":0},{\"date\":\"1月4日\",\"language\":\"去重合并多语言\",\"channel\":\"去重合并多渠道\",\"generation01\":1,\"generation02\":1,\"generation03\":0,\"generation04\":0,\"generation05\":0,\"generation06After\":0,\"generation02After\":1,\"help1\":0,\"help2\":0,\"help3\":0,\"help4\":0,\"help5\":0,\"help6\":0,\"help7\":0,\"help8\":0,\"allHelp1\":0,\"allHelp2\":0,\"allHelp3\":0,\"allHelp4\":0,\"allHelp5\":0,\"allHelp6\":0,\"allHelp7\":0,\"allHelp8\":0,\"promoteClusteringCount\":0,\"freeRemindCount\":2,\"payRemindCount\":0,\"sendSuccessMsgCount\":0,\"sendFailMsgCount\":0,\"sendTimeOutMsgCount\":0,\"notWhiteCount\":0}]",
	}
	_, _ = reportMsgInfoMapper.InsertSelective(&session, reportMsgInfoEntity)
	reportMsgInfoEntity.MsgStatus = "owner_send"
	reportMsgInfoEntity.Res = "res"
	_, _ = reportMsgInfoMapper.UpdateByPrimaryKeySelective(&session, reportMsgInfoEntity)
	session.Commit()
	session.Close()
	reportType, _ := reportMsgInfoMapper.SelectListByReportType(1, "excel")
	day, _ := reportMsgInfoMapper.SelectCountByReportTypeAndDay(1, "excel", "1月13日")
	logTracing.LogInfo(ctx, strconv.Itoa(int(day)))
	ctx.JSON(http.StatusOK, reportType)
}

func (c TestController) MsgInfo(ctx *gin.Context) {
	session, isExist, err := txUtil.GetTransaction(ctx)
	if nil != err {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，创建事务失败,err：%v", "FreeSdkInfo", err))
		return
	}
	if !isExist {
		defer func() {
			session.Rollback()
			session.Close()
		}()
	}
	msgInfoMapper := dao.GetMsgInfoMapperV2()
	//msgInfoEntity := entity.MsgInfoEntityV2{
	//	Id:             "1978742030323044352",
	//	Type:           "send",
	//	Msg:            "{\"Params\":{\"appkey\":\"ETrw2BEq\",\"business_phone\":\"6285873165264\",\"messaging_product\":\"whatsapp\",\"recipient_type\":\"individual\",\"to\":\"852500000087\",\"cus_message_id\":\"\",\"type\":\"interactive\",\"interactive\":{\"type\":\"button\",\"header\":{\"type\":\"image\",\"image\":{\"link\":\"https://mlbbmy.outweb.mobilelegends.com/1733302147252-HJ44TLEQi6.png\"}},\"body\":{\"text\":\"【重要通知】由于您长时间没有活动进度，即将收不到[MLBB：Teguh Bersama²  ]的消息推送。\\n\\n为避免获取不到奖励，请尽快续订活动消息。\"},\"action\":{\"buttons\":[{\"type\":\"reply\",\"reply\":{\"id\":\"1\",\"title\":\"续订活动消息\"}}]}}},\"CommonHeaders\":{\"accessKey\":\"PcCx5hDi\",\"action\":\"mt\",\"bizType\":\"2\",\"sign\":\"73bc24b176b6f4522627a5779a6a31fb\",\"ts\":\"1736761980025\"}}",
	//	MsgStatus:      "owner_un_send",
	//	WaId:           "852500000087",
	//	MsgType:        "renewFreeMsg",
	//	WaMessageId:    "wamid.141a6d9710e140bf88efa024447218f7",
	//	IsCount:        1,
	//	TraceId:        "fb378c33e7a843168a576feabda21177",
	//	SendRes:        "{\"NXResponse\":{\"code\":0,\"message\":\"Success\",\"traceId\":\"fb378c33e7a843168a576feabda21173\",\"data\":{\"messaging_product\":\"whatsapp\",\"messages\":[{\"id\":\"wamid.141f6d9710e140bf88efa024447218f7\"}],\"id\":\"\"}}}",
	//	BuildMsgParams: "{\"interactive\":{\"Type\":\"button\",\"ImageLink\":\"https://mlbbmy.outweb.mobilelegends.com/1733302147252-HJ44TLEQi6.png\",\"BodyText\":\"【重要通知】由于您长时间没有活动进度，即将收不到[MLBB：Teguh Bersama²  ]的消息推送。\\n\\n为避免获取不到奖励，请尽快续订活动消息。\",\"FooterText\":\"\",\"Action\":{\"DisplayText\":\"\",\"Url\":\"\",\"ShortLink\":\"\",\"Buttons\":[{\"type\":\"reply\",\"reply\":{\"id\":\"1\",\"title\":\"续订活动消息\"}}]}}}",
	//}
	//_, _ = msgInfoMapper.InsertSelective(&session, msgInfoEntity)
	//msgInfoEntity.MsgStatus = "sent"
	//_, _ = msgInfoMapper.UpdateByPrimaryKeySelective(&session, msgInfoEntity)
	//_, _ = msgInfoMapper.UpdateCountOfSendUnCount(&session, 1, 2)
	//msgInfo, _ := msgInfoMapper.SelectByPrimaryKey2(&session, "1978742030323044352")
	//msgInfo, _ := msgInfoMapper.SelectListByMsgType(0, time.Now().Unix(), "0", 100)
	//msgInfo, _ := msgInfoMapper.SelectMsgListOfUnSendMsg("852500000087")
	//msgInfo, _ := msgInfoMapper.SelectWaIdListOfUnSendMsg("0", 100)
	//msgInfo, _ := msgInfoMapper.SelectByWaMessageId("wamid.141a6d9710e140bf88efa024447218f7")
	//msgInfo, _ := msgInfoMapper.SumPriceSendUnCountMsg(&session, 1)
	msgInfo, _ := msgInfoMapper.CountCdkMsgByWaId("852500000087", "renewFreeMsg")
	session.Commit()
	session.Close()
	ctx.JSON(http.StatusOK, msgInfo)
}

func (c TestController) HelpInfo(ctx *gin.Context) {
	session, isExist, err := txUtil.GetTransaction(ctx)
	if nil != err {
		logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，创建事务失败,err：%v", "FreeSdkInfo", err))
		return
	}
	if !isExist {
		defer func() {
			session.Rollback()
			session.Close()
		}()
	}
	//helpInfoMapper := dao.GetHelpInfoMapperV2()
	//helpInfoEntity := entity.HelpInfoEntityV2{
	//	RallyCode:  "a0201djsgg",
	//	WaId:       "852601001001",
	//	HelpStatus: "efficien",
	//}
	//_, _ = helpInfoMapper.InsertSelective(&session, helpInfoEntity)
	//msgInfo, _ := helpInfoMapper.CountByCodesTimestamp([]string{"a0201djsgg", "a0201djggu"}, 0, time.Now().Unix())
	//msgInfo, _ := helpInfoMapper.SelectByWaId("852601001001")
	//msgInfo, _ := helpInfoMapper.SelectListByRallyCode(&session, "a0201djsgg")
	//msgInfo, _ := helpInfoMapper.SelectDistinctCodeByTimestamp(0, time.Now().Unix(), "0", 100)
	//session.Commit()
	//session.Close()
	allTimeRange := dto.StatisticsTimeRange{
		StartTimestamp: 0,
		EndTimestamp:   2000000000,
	}
	_, _ = statistics.HelpInfo(ctx, allTimeRange)
	ctx.JSON(http.StatusOK, nil)
}

func (c TestController) UserAttendInfo(ctx *gin.Context) {
	//session, isExist, err := txUtil.GetTransaction(ctx)
	//if nil != err {
	//	logTracing.LogWarn(ctx, logTracing.WarnLogFmt, fmt.Sprintf("方法[%s]，创建事务失败,err：%v", "FreeSdkInfo", err))
	//	return
	//}
	//if !isExist {
	//	defer func() {
	//		session.Rollback()
	//		session.Close()
	//	}()
	//}
	//userAttendInfoMapper := dao.GetUserAttendInfoMapperV2()
	//user, _ := userAttendInfoMapper.CountUser()
	//user, _ := userAttendInfoMapper.CountReCallOfStartGroup(time.Now().Unix())
	//user, _ := userAttendInfoMapper.CountRenewFree(1, time.Now().Unix())
	//user, _ := userAttendInfoMapper.CountNotSendCdkUser(1)
	//user, _ := userAttendInfoMapper.CountReCallOfClustering(1, time.Now().Unix())
	//user, _ := userAttendInfoMapper.SelectByWaId("852574819200500500000002")
	//user, _ := userAttendInfoMapper.SelectListByWaIdsWithSession(&session, []string{"852574819200500500000002"})
	//user, _ := userAttendInfoMapper.SelectListByWaIds([]string{"852574819200500500000002"})
	//user, _ := userAttendInfoMapper.SelectByWaIdBySession(&session, "852574819200500500000002")
	//user, _ := userAttendInfoMapper.SelectByRallyCode("a0402djgwr")
	//user, _ := userAttendInfoMapper.SelectReCallOfStartGroup(0, 30, time.Now().Unix())
	//user, _ := userAttendInfoMapper.SelectRenewFree(0, 30, 1, time.Now().Unix())
	//user, _ := userAttendInfoMapper.SelectNotSendCdkUser(0, 30, 1)
	//user, _ := userAttendInfoMapper.SelectReCallOfClustering(0, 30, 1, time.Now().Unix())
	//user, _ := userAttendInfoMapper.SelectListByGeneration(0, time.Now().Unix(), 0, 100)
	//user, _ := userAttendInfoMapper.SelectListByCodes([]string{"a0402djgwr"})
	//session.Commit()
	//session.Close()
	timeRange := dto.StatisticsTimeRange{
		StartTimestamp: 0,
		EndTimestamp:   0,
	}
	generationUserDtoList, _ := statistics.GenerationInfoWithAttend(ctx, timeRange)
	generation01Count := 0
	generation01CompleteCount := 0
	generationOtherCount := 0
	generationOtherCompleteCount := 0
	for _, generationUserDto := range generationUserDtoList {
		if constant.Generation01 == generationUserDto.Generation {
			generation01Count += generationUserDto.Count
			if constant.AttendStatusAttend != generationUserDto.AttendStatus {
				generation01CompleteCount += generationUserDto.Count
			}
		} else {
			generationOtherCount += generationUserDto.Count
			if constant.AttendStatusAttend != generationUserDto.AttendStatus {
				generationOtherCompleteCount += generationUserDto.Count
			}
		}
	}
	ctx.JSON(http.StatusOK, strconv.Itoa(generation01Count)+", "+strconv.Itoa(generation01CompleteCount)+", "+strconv.Itoa(generationOtherCount)+", "+strconv.Itoa(generationOtherCompleteCount))
}

func (c TestController) DDL(ctx *gin.Context) {
	ddlMapper := dao.GetDDLMapper()
	_ = ddlMapper.DropActivityInfo()
	_ = ddlMapper.CreateActivityInfo()
	_ = ddlMapper.InitActivityInfo()
	_ = ddlMapper.DropCostCountInfo()
	_ = ddlMapper.CreateCostCountInfo()
	_ = ddlMapper.DropFreeCdkInfo()
	_ = ddlMapper.CreateFreeCdkInfo()
	_ = ddlMapper.DropHelpInfo()
	_ = ddlMapper.CreateHelpInfo()
	_ = ddlMapper.DropMsgInfo()
	_ = ddlMapper.CreateMsgInfo()
	_ = ddlMapper.DropReportMsgInfo()
	_ = ddlMapper.CreateReportMsgInfo()
	_ = ddlMapper.DropRsvMsgInfo()
	_ = ddlMapper.CreateRsvMsgInfo()
	_ = ddlMapper.DropRsvOtherMsgInfo()
	_ = ddlMapper.CreateRsvOtherMsgInfo()
	_ = ddlMapper.DropUserAttendInfo()
	_ = ddlMapper.CreateUserAttendInfo()
	ctx.JSON(http.StatusOK, nil)
}
