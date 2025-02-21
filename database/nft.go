package database

import (
	"time"
)

type NFTStore struct {
	TokenId    uint64 `gorm:"primarykey;column:tokenid"`
	Address    string `gorm:"index"`
	Cid        string
	Type       int
	CreateTime time.Time `gorm:"column:create"`
}

func (nft *NFTStore) CreateNFTInfo() error {
	return GlobalDataBase.Create(nft).Error
}

func GetNFTInfo(tokenId uint64) (NFTStore, error) {
	var result NFTStore
	err := GlobalDataBase.Model(&NFTStore{}).Where("tokenid = ?", tokenId).Find(&result).Error

	return result, err
}
