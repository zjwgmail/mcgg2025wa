package dto

type ShortDto struct {
	LongUrl         string `json:"long_url"`
	ActivityId      string `json:"activity_id"`
	ProjectId       string `json:"project_id"`
	Sign            string `json:"sign"`
	ShortLinkGenUrl string `json:"short_link_gen_url"`
	ShortLinkPrefix string `json:"short_link_prefix"`
	SignKey         string `json:"sign_key"`
}
