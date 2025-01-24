package router

import "time"

type SpaceListInfo struct {
	Cid  string    `json:"cid"`
	Name string    `json:"name"`
	Time time.Time `json:"time"`
}

type EarningInfo struct {
	Sum       int        `json:"sum"`
	Activitys []Activity `json:"activitys"`
}

type Activity struct {
	Name     string    `json:"name"`
	Point    int       `json:"point"`
	CreateAt time.Time `json:"create_at"`
}

type ActivityInfo struct {
	ID      int    `json:"id"`
	Address string `json:"address"`
	Points  int    `json:"points"`
	Sum     int    `json:"sum"`
}

type UserInfo struct {
	Invitees int    `json:"invitees"`
	Code     string `json:"code"`
	Reward   int    `json:"reward"`
}
