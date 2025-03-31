package dto

type ReportJsonDto struct {
	Date                   string `json:"date"`
	Language               string `json:"language"`
	Channel                string `json:"channel"`
	Generation01           int    `json:"generation01"`
	Generation02           int    `json:"generation02"`
	Generation03           int    `json:"generation03"`
	Generation04           int    `json:"generation04"`
	Generation05           int    `json:"generation05"`
	Generation06After      int    `json:"generation06After"`
	Generation02After      int    `json:"generation02After"`
	Help1                  int    `json:"help1"`
	Help2                  int    `json:"help2"`
	Help3                  int    `json:"help3"`
	Help4                  int    `json:"help4"`
	Help5                  int    `json:"help5"`
	Help6                  int    `json:"help6"`
	Help7                  int    `json:"help7"`
	Help8                  int    `json:"help8"`
	AllHelp1               int    `json:"allHelp1"`
	AllHelp2               int    `json:"allHelp2"`
	AllHelp3               int    `json:"allHelp3"`
	AllHelp4               int    `json:"allHelp4"`
	AllHelp5               int    `json:"allHelp5"`
	AllHelp6               int    `json:"allHelp6"`
	AllHelp7               int    `json:"allHelp7"`
	AllHelp8               int    `json:"allHelp8"`
	PromoteClusteringCount int    `json:"promoteClusteringCount"`
	FreeRemindCount        int    `json:"freeRemindCount"`
	PayRemindCount         int    `json:"payRemindCount"`
	SendSuccessMsgCount    int64  `json:"sendSuccessMsgCount"`
	SendFailMsgCount       int64  `json:"sendFailMsgCount"`
	SendTimeOutMsgCount    int64  `json:"sendTimeOutMsgCount"`
	NotWhiteCount          int64  `json:"notWhiteCount"`
}
