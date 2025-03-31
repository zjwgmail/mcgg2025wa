package cron_task

import (
	"go-fission-activity/config"
)

func reportTaskUTC0(methodName string, timeConfig config.TimerConfig) {
	reportTask(methodName, 0)
}
