package core

import (
	"context"
	"errors"
	"github.com/getsentry/sentry-go"
	"github.com/natefinch/lumberjack"
	"go-fission-activity/activity/constant"
	"go-fission-activity/activity/web/middleware/logTracing"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

func OpenZap() error {
	hook := lumberjack.Logger{
		Filename: getLogPath(), // ⽇志⽂件路径
		MaxSize:  100,          // 100M
		//MaxBackups: 3,       // 最多保留3个备份
		MaxAge:   30,    //days
		Compress: false, // 是否压缩 disabled by default
	}

	// 设置打印时间格式
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeDuration = zapcore.StringDurationEncoder

	var zapLogger *zap.Logger

	core := zapcore.NewTee(
		zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), zapcore.AddSync(os.Stdout), zap.NewAtomicLevelAt(zapcore.InfoLevel)),
		zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), zapcore.AddSync(&hook), zap.NewAtomicLevelAt(zapcore.InfoLevel)),
	)
	zapLogger = zap.New(core)
	logTracing.LogInfo(context.Background(), "zap启动成功！")
	//}
	zap.ReplaceGlobals(zapLogger)
	// 将 Go 自带的日志输出重定向到 Zap Logger
	zap.RedirectStdLog(zapLogger)
	return nil
}

// 实现Sentry的io.Writer接口
type sentryWriter struct{}

func (sentryWriter) Write(p []byte) (n int, err error) {
	msg := string(p)
	sentry.CaptureException(errors.New(msg))
	return len(p), nil
}
func getLogPath() string {
	if constant.DefaultActives != constant.LookupEnv(constant.ProfileActives, constant.DefaultActives) {
		return "/apps/prod-mcgg2025wa-server/logs/prod-mcgg2025wa-server.log"
	} else {
		return "D:\\apps\\logs\\prod-mcgg2025wa-server.log"
	}
}
