package controller

import (
	"github.com/gin-gonic/gin"
	"go-fission-activity/activity/web/middleware/logTracing"
	"net/http"
)

type HealthController struct {
}

// Ping 健康检查
func (c HealthController) Ping(ctx *gin.Context) {
	logTracing.LogPrintf(ctx, logTracing.WebRcLogFmt, "接收到健康检查请求")
	ctx.JSON(http.StatusOK, nil)
}
