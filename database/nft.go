package database

import (
	"time"

	"golang.org/x/xerrors"
)

type NFTStore struct {
	TokenId    uint64 `gorm:"primarykey;column:tokenid"`
	Address    string `gorm:"index"`
	Cid        string
	Type       string
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

func ListNFT(page, size int, address, order string) ([]NFTStore, error) {
	var nfts []NFTStore
	var orderRules string
	switch order {
	case "time_asc":
		orderRules = "create"
	case "time_desc":
		orderRules = "create desc"
	default:
		return nil, xerrors.Errorf("not spport order rules: %s", order)
	}

	err := GlobalDataBase.Model(&NFTStore{}).Where("address = ?", address).Order(orderRules).Offset((page - 1) * size).Limit(size).Find(&nfts).Error
	if err != nil {
		return nil, err
	}

	return nfts, nil
}

func ListNFTByType(page, size int, address, order string, ntype string) ([]NFTStore, error) {
	var nfts []NFTStore
	var orderRules string
	switch order {
	case "time_asc":
		orderRules = "create"
	case "time_desc":
		orderRules = "create desc"
	default:
		return nil, xerrors.Errorf("not spport order rules: %s", order)
	}

	err := GlobalDataBase.Model(&NFTStore{}).Where("address = ? AND type = ?", address, ntype).Order(orderRules).Offset((page - 1) * size).Limit(size).Find(&nfts).Error
	if err != nil {
		return nil, err
	}

	return nfts, nil
}
