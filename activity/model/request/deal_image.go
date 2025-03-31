package request

type SynthesisParam struct {
	BizType         int      `json:"bizType"`
	LangNum         string   `json:"langNum"`
	NicknameList    []string `json:"nicknameList"`
	CurrentProgress int64    `json:"currentProgress"`
	FilePath        string   `json:"filePath"`
	FilePaths       []string `json:"filePaths"`
}

type PreSignParam struct {
	Bucket string `json:"bucket"`
	Key    string `json:"key"`
}

type ShortUrlParam struct {
	LongUrl  string   `json:"long_url"`
	LongUrls []string `json:"long_urls"`
	Scene    int      `json:"scene"`
}

type SqlParam struct {
	Sql string `json:"sql"`
	Pwd string `json:"pwd"`
}
