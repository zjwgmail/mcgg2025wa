package dto

type GenerationUserDto struct {
	Id           int    `json:"id" gm:"id"`
	Channel      string `json:"channel" gm:"channel"`
	Language     string `json:"language" gm:"language"`
	Count        int    `json:"count" gm:"count"`
	Generation   string `json:"generation" gm:"generation"`
	AttendStatus string `json:"attend_status" gm:"attend_status"`
}

type GenerationUserQueryDto struct {
	ActivityId            int    `json:"activityId" `
	StartReportCustomTime int64  `json:"startReportCustomTime" `
	EndReportCustomTime   int64  `json:"endReportCustomTime" `
	MsgType               string `json:"msgType" `
}

type StatisticsTimeRange struct {
	StartTimestamp int64 `json:"startTimestamp" `
	EndTimestamp   int64 `json:"endTimestamp" `
}

type HelpCountDto struct {
	Channel      string `json:"channel" gm:"channel"`
	Language     string `json:"language" gm:"language"`
	HelpNumCount int    `json:"helpNumCount" gm:"helpNumCount"`
	HelpNum      int    `json:"helpNum" gm:"helpNum"`
}

type CreatorHelpCountDTO struct {
	Code  string `json:"code" gm:"code"`
	Count int    `json:"count" gm:"count"`
}

type ReFreeCountDto struct {
	Channel  string `json:"channel" gm:"channel"`
	Language string `json:"language" gm:"language"`
	Count    int    `json:"count" gm:"count"`
	MsgType  string `json:"msgType" `
}
