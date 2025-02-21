package types

import (
	"time"

	"github.com/memoio/xspace-server/database"
)

// NFT types
type MintTweetReq struct {
	Name     string
	PostTime int64
	Tweet    string
	Images   []string
}

type MintRes struct {
	TokenID uint64
}

type ListNFTRes struct {
	NftInfos []NFTInfo
}

type NFTInfo struct {
	TokenID    int64
	Type       int
	CreateTime time.Time
}

type TweetNFTInfo struct {
	Name     string
	PostTime int64
	Tweet    string
	Images   []string
}

type TweetNFTInfoRes TweetNFTInfo

type UserInfoRes struct {
	Address    string
	InviteCode string
	Points     int64
	Referrals  int
	Space      int
}

// point types
type PointInfoRes struct {
	Points        int64
	GodataCount   int
	GodataSpace   int
	ChargingCount int
	Charging      bool
}

type PointHistoryRes struct {
	History []database.ActionStore
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
