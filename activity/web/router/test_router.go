package router

import (
	"github.com/gin-gonic/gin"
	"go-fission-activity/activity/web/controller"
)

func init() {
	RouterNoCheckRole = append(RouterNoCheckRole, registerTestController)
}

func registerTestController(group *gin.RouterGroup) {
	c := controller.TestController{}
	r := group.Group("test")
	r.GET("freeSdkInfo", c.FreeSdkInfo)
	r.GET("fsvMsgInfo", c.RsvMsgInfo)
	r.GET("reportMsgInfo", c.ReportMsgInfo)
	r.GET("msgInfo", c.MsgInfo)
	r.GET("helpInfo", c.HelpInfo)
	r.GET("userAttendInfo", c.UserAttendInfo)
	r.GET("ddl", c.DDL)
}
