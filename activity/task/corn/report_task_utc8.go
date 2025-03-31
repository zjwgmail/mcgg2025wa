package cron_task

import (
	"go-fission-activity/config"
)

func reportTaskUTC8(methodName string, timeConfig config.TimerConfig) {
	reportTask(methodName, 8)
}
