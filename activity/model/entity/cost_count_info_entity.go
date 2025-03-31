package entity

import "go-fission-activity/util"

type CostCountInfoEntity struct {
	Id        int             `json:"id" gm:"id"`
	CostCount float64         `json:"cost_count" gm:"cost_count"`
	CreatedAt util.CustomTime `json:"created_at" gm:"created_at"`
	UpdatedAt util.CustomTime `json:"updated_at" gm:"updated_at"`
}
