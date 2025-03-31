package cron_task

import (
	"github.com/robfig/cron/v3"
	"go-fission-activity/activity/constant"
	"go-fission-activity/activity/web/middleware/logTracing"
	"go-fission-activity/config"
	"log"
	"time"
)

const PageSize = 30
const lockTimeout = time.Second * 120

type taskFunc func(string, config.TimerConfig)

func InitCron() {

	if config.ApplicationConfig.IsActivityServer {
		mapTimerConfig := config.ApplicationConfig.Timer
		for key, duration := range mapTimerConfig {
			switch key {
			case constant.RecallCdkTask:
				//todo 待测试并发
				go runTask(key, duration, recallCdkTask)
				continue
			case constant.RecallClusteringTask:
				//todo 待测试并发
				go runTask(key, duration, recallClusteringTask)
				continue
			case constant.RecallFreeTask:
				//todo 待测试并发
				go runTask(key, duration, recallFreeTask)
				continue
			case constant.RecallPayFreeMsgTask:
				//todo 待测试并发
				go runTask(key, duration, recallPayFreeMsgTask)
				continue
			case constant.CdkNumTask:
				go runTask(key, duration, cdkNumTask)
				continue
			case constant.CostTask:
				go runTask(key, duration, costTask)
				continue
			case constant.ActivityTask:
				go runTask(key, duration, activityTask)
				continue
			case constant.ResendMsgTask:
				//todo 待测试并发
				go runTask(key, duration, resendMsgTask)
				continue
			case constant.ReportTaskUTC0:
				go runTask(key, duration, reportTaskUTC0)
				continue
			case constant.ReportTaskUTC8:
				go runTask(key, duration, reportTaskUTC8)
				continue
			case constant.ReportTaskUTCMinus8:
				go runTask(key, duration, reportTaskUTCMinus8)
				continue
			case constant.FreeCdkSend:
				go runTask(key, duration, freeCdkSend)
			default:
				panic("不支持的定时任务类型")
			}
		}
	}

}

func runTask(methodName string, timerConfig config.TimerConfig, optFunc taskFunc) {
	logTracing.LogPrintfP("定时任务[%s]启动", methodName)
	c := cron.New()
	// 定义任务
	_, err := c.AddFunc(timerConfig.TimerCorn, func() {
		logTracing.LogPrintfP("定时任务[%s]开始执行：%v", methodName, time.Now())
		optFunc(methodName, timerConfig)
		logTracing.LogPrintfP("定时任务[%s]执行完成：%v", methodName, time.Now())
	})
	if err != nil {
		log.Fatalf("无法添加定时任务[%s]: %v", methodName, err)
	}

	// 启动定时任务
	c.Start()

	// 让程序保持运行
	select {}
}
