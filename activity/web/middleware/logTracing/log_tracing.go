package logTracing

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"runtime"
	"time"
)

const (
	WebRcLogFmt      = "[go-fission-activity]接收[web]请求：%v"
	WebSendLogFmt    = "[go-fission-activity]发送[web]结果：%v"
	WebHandleLogFmt  = "[go-fission-activity]处理[web]请求：%v"
	TaskHandleLogFmt = "[go-fission-activity]处理[task]任务：%v"
	NormalLogFmt     = "[go-fission-activity]日志：%v"
	WarnLogFmt       = "[go-fission-activity]发生告警：%v"
	ErrorLogFmt      = "[go-fission-activity]发生异常：%v"
)

func LogErrorPrintf(ctx context.Context, err error, format string, v ...any) {
	LogError(ctx, err, fmt.Sprintf(format, v...))
}

func LogError(ctx context.Context, err error, message string) {
	pc, _, line, ok := runtime.Caller(2)
	funcName := ""
	if ok {
		funcName = runtime.FuncForPC(pc).Name()
	}
	startTime, traceId := GetInfoByContext(ctx)

	zap.S().Error(fmt.Sprintf("[function]:%v,[line]:%v,[startTime]:%v,[traceId]:%v,[msg]:%v", funcName, line, startTime, traceId, fmt.Sprintf(message, err)))
}

func LogPrintf(ctx context.Context, format string, v ...any) {
	LogInfo(ctx, fmt.Sprintf(format, v...))
}

func LogPrintfP(format string, v ...any) {
	LogInfo(context.Background(), fmt.Sprintf(format, v...))
}

func LogInfo(ctx context.Context, message string) {
	pc, _, line, ok := runtime.Caller(1)
	funcName := ""
	if ok {
		funcName = runtime.FuncForPC(pc).Name()
	}
	startTime, traceId := GetInfoByContext(ctx)

	zap.S().Info(fmt.Sprintf("[function]:%v,[line]:%v,[startTime]:%v,[traceId]:%v,[msg]:%v", funcName, line, startTime, traceId, message))
}

func LogWarn(ctx context.Context, format string, v ...any) {
	pc, _, line, ok := runtime.Caller(1)
	funcName := ""
	if ok {
		funcName = runtime.FuncForPC(pc).Name()
	}
	message := fmt.Sprintf(format, v...)
	startTime, traceId := GetInfoByContext(ctx)

	zap.S().Warn(fmt.Sprintf("[function]:%v,[line]:%v,[startTime]:%v,[traceId]:%v,[msg]:%v", funcName, line, startTime, traceId, message))
}

func GetInfoByContext(ctx context.Context) (time.Time, string) {
	var startTime time.Time
	var traceId string
	var ok bool
	if ctx == nil {
		return startTime, traceId
	}
	if ginCtx, ok := ctx.(*gin.Context); ok {
		if ginCtx == nil || ginCtx.Request == nil {
			return startTime, traceId
		}
	}
	value := ctx.Value("startTime")
	if value != nil {
		// 使用类型断言将 value 转换为 time.Time 类型
		if startTime, ok = value.(time.Time); !ok {
			//log.Println("无法将 value 转换为 time.Time 类型")
		}
	}

	traceIdAny := ctx.Value("traceId")
	if traceIdAny != nil {
		// 使用类型断言将 value 转换为 time.Time 类型
		if traceId, ok = traceIdAny.(string); !ok {
			//log.Println("无法将 value 转换为 time.Time 类型")
		}
	}

	return startTime, traceId
}
