package config

type MsgYml struct {
	MsgMap map[string]map[string]map[string]MsgInfo
}

type MsgInfo struct {
	Interactive *Interactive `json:"interactive,omitempty"`
	Template    *Template    `json:"template,omitempty"`
	Params      *Params      `json:"params,omitempty"`
}

type Params struct {
	NicknameList []string
	Language     string
}

type Interactive struct {
	Type       string
	ImageLink  string
	BodyText   string
	FooterText string
	Action     *Action
}

type Action struct {
	DisplayText string
	Url         string
	ShortLink   string
	Buttons     []*Button
}

type Button struct {
	Type  string `json:"type,omitempty"`
	Reply Reply  `json:"reply,omitempty"`
}

type Reply struct {
	Id    string `json:"id,omitempty"`
	Title string `json:"title,omitempty"`
}

type Template struct {
	Name       string                    `json:"name,omitempty"`
	Language   *NxReqTemplateLanguage    `json:"language,omitempty"`
	Components []*NxReqTemplateComponent `json:"components,omitempty"`
}

type NxReqTemplateLanguage struct {
	Policy string `json:"policy,omitempty"`
	Code   string `json:"code,omitempty"`
}

type NxReqTemplateComponent struct {
	Type       string                             `json:"type,omitempty"` // type为body时 参数text ，type为header时参数image，type为button时参数text
	Parameters []*NxReqTemplateComponentParameter `json:"parameters,omitempty"`
	// type为button时有
	SubType string `json:"sub_type,omitempty"`
	Index   int    `json:"index,omitempty"`
}

type NxReqTemplateComponentParameter struct {
	Type string `json:"type,omitempty"`
	// type为text时
	Text string `json:"text,omitempty"`
	// type为image时
	Image *NxReqTemplateComponentImage `json:"image,omitempty"`
}

type NxReqTemplateComponentImage struct {
	Id string `json:"id,omitempty"`
}

var MsgConfig = new(MsgYml)
