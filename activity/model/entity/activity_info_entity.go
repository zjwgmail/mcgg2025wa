package entity

import "go-fission-activity/util"

type ActivityInfoEntity struct {
	Id             int             `json:"id" gm:"id"`
	ActivityName   string          `json:"activity_name" gm:"activity_name"`
	ActivityStatus string          `json:"activity_status" gm:"activity_status"`
	CreatedAt      util.CustomTime `json:"created_at" gm:"created_at"`
	UpdatedAt      util.CustomTime `json:"updated_at" gm:"updated_at"`
	StartAt        util.CustomTime `json:"start_at" gm:"start_at"`
	EndAt          util.CustomTime `json:"end_at" gm:"end_at"`
	EndBufferDay   int             `json:"end_buffer_day" gm:"end_buffer_day"`
	EndBufferAt    util.CustomTime `json:"end_buffer_at" gm:"end_buffer_at"`
	ReallyEndAt    util.CustomTime `json:"really_end_at" gm:"really_end_at"`
	HelpMax        int             `json:"help_max" gm:"help_max"`
	CostMax        float64         `json:"cost_max" gm:"cost_max"`
}
