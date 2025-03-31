package router

import (
	"github.com/gin-gonic/gin"
	"go-fission-activity/activity/web/controller"
)

func init() {
	RouterNoCheckRole = append(RouterNoCheckRole, registerHealthController)
}

func registerHealthController(group *gin.RouterGroup) {
	c := controller.HealthController{}
	r := group.Group("/events/mcgg2025wa/ping")
	r.GET("", c.Ping)

}
