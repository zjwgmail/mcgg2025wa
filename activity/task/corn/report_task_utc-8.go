package cron_task

import (
	"go-fission-activity/config"
)

func reportTaskUTCMinus8(methodName string, timeConfig config.TimerConfig) {
	reportTask(methodName, -8)
}
