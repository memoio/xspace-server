package router

import "time"

// NFT types
type MintTweetReq struct {
	Address  string
	Name     string
	PostTime int64
	Tweet    string
	Images   []string
}

type MintRes struct {
	TokenID int64
}

type ListNFTRes struct {
	NftInfos []NFTInfo
}

type NFTInfo struct {
	TokenID    int64
	Type       int
	CreateTime time.Time
}

type TweetNFTInfoRes struct {
	Name     string
	PostTime int64
	Tweet    string
	Images   []string
}

// point types
type PointInfoRes struct {
	Points        int64
	GodataCount   int
	GodataSpace   int
	ChargingCount int
	Charging      bool
}

type PointInfo struct {
	Point      int64
	Time       time.Time
	ActionName string
}

type PointHistoryRes struct {
	History []PointInfo
}

type ProjectInfo struct {
	ProjectID int
	Name      string
	Start     time.Time
	End       time.Time
}

type ListProjectsRes struct {
	Projects []ProjectInfo
}

type RankInfo struct {
	Rank    int
	Address string
	Scores  int64
	Points  int64
}

type RankRes struct {
	RnakInfo []RankInfo
}
