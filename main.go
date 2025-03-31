package main

import (
	"go-fission-activity/activity/cmd"
	"go.uber.org/zap"
)

func main() {

	cmd.Execute()
	stopService()
}

func stopService() {
	// 服务退出逻辑
	zap.S().Info("程序退出！")
	// 在退出之前调用 Sync() 方法
	zap.S().Sync()
}
