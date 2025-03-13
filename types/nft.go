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
	// Type     string `json:"type,omitempty"`
	Link string `json:"link,omitempty"`
}

type MintRes struct {
	TokenID uint64
}

type ListNFTRes struct {
	NftInfos []database.NFTStore
	Length   int
}

type TweetNFTInfo struct {
	Name     string
	PostTime int64
	Tweet    string
	Images   []string
	Link     string `json:"link,omitempty"`
}

type TweetNFTInfoRes TweetNFTInfo

type UserInfoRes struct {
	Address     string
	InviteCode  string
	InvitedCode string
	Points      int64
	Referrals   int
	Space       int
}

// point types
type FinishActionReq struct {
	ActionId int
}

type PointInfoRes struct {
	Points        int64
	GodataCount   int64
	GodataSpace   int
	ChargingCount int64
	Charging      bool
}

type PointHistoryRes struct {
	History []database.ActionStore
	Length  int
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

type InviteReq struct {
	Code string
}

type RankInfo struct {
	Rank    int
	Address string
	Scores  int64
	Points  int64
}

type RankRes struct {
	RankInfo []RankInfo
	Length   int
}
