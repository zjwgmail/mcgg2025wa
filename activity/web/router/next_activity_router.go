package router

import (
	"github.com/gin-gonic/gin"
	"go-fission-activity/activity/web/controller"
	"go-fission-activity/activity/web/service"
	"go-fission-activity/config"
)

func init() {
	RouterNoCheckRole = append(RouterNoCheckRole, registerNextActivityController)
}

func registerNextActivityController(group *gin.RouterGroup) {
	c := controller.UserAttendInfoController{UserAttendInfoService: service.GetUserAttendInfoService(), MsgInfoService: service.GetMsgInfoService()}
	image := controller.ImageController{
		ImageService:    service.GetImageService(),
		WaMsgService:    service.GetWaMsgService(),
		ShortUrlService: service.GetShortUrlService()}
	r := group.Group("/events/mcgg2025wa/activity")

	if config.ApplicationConfig.IsActivityServer {
		r.POST("userAttendInfo", c.UserAttendInfo)
		r.POST("help", c.Help)
		//r.POST("msgStatusWebHook", c.MsgStatusWebHook)
		r.GET("getIp", c.GetIp)
		r.POST("import/data", c.ImportData)
		r.POST("helpText/count", c.HelpTextCount)
		r.POST("sql", c.ExecuteSql)
		r.POST("initDB2", c.InitSQLAndCache)
	} else {
		//图片相关g
		r.POST("preSign", image.PreSign)
		r.POST("generateImages", image.GenerateImages)
		r.POST("uploadImage2NX", image.UploadTemplateImage2NX)
		r.POST("shortUrl", image.GetShortUrl)
		r.POST("batchRandomPush", image.RandomMessage)
	}
	r.Static("/image", "./resources/image")
}
