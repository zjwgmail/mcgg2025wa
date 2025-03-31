package cron_task

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"go-fission-activity/activity/web/middleware/logTracing"
	"go-fission-activity/activity/web/service"
	"go-fission-activity/config"
)

func freeCdkSend(methodName string, timeConfig config.TimerConfig) {
	ginCtx := gin.Context{}
	ctx := &ginCtx
	// defer 异常处理
	defer func() {
		if e := recover(); e != nil {
			logTracing.LogErrorPrintf(ctx, errors.New(fmt.Sprintf("方法[%s]，发生panic异常", methodName)), logTracing.ErrorLogFmt, e)
			return
		}
	}()

	freeCdkSendService := service.GetFreeCdkSendService()
	freeCdkSendService.FreeCdkSend(ctx, methodName)

}
