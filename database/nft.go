package database

import (
	"time"

	"golang.org/x/xerrors"
)

type NFTStore struct {
	TokenId uint64 `gorm:"index;column:tokenid"`
	Address string `gorm:"index"`
	Cid     string `gorm:"primarykey"`
	Type    string
	Time    time.Time
}

func (nft *NFTStore) CreateNFTInfo() error {
	return GlobalDataBase.Create(nft).Error
}

func GetNFTInfo(tokenId uint64) (NFTStore, error) {
	var result NFTStore
	err := GlobalDataBase.Model(&NFTStore{}).Where("tokenid = ?", tokenId).Find(&result).Error

	return result, err
}

func GetNFTInfoByCID(cid string) (NFTStore, error) {
	var result NFTStore
	err := GlobalDataBase.Model(&NFTStore{}).Where("cid = ?", cid).Find(&result).Error

	return result, err
}

func ListNFT(page, size int, address, order string) ([]NFTStore, int64, error) {
	var nfts []NFTStore
	var orderRules string
	var length int64
	switch order {
	case "date_asc":
		orderRules = "time"
	case "date_desc":
		orderRules = "time desc"
	default:
		return nil, 0, xerrors.Errorf("not spport order rules: %s", order)
	}

	err := GlobalDataBase.Model(&NFTStore{}).Where("address = ?", address).Order(orderRules).Offset((page - 1) * size).Limit(size).Find(&nfts).Error
	if err != nil {
		return nil, 0, err
	}

	err = GlobalDataBase.Model(&NFTStore{}).Where("address = ?", address).Count(&length).Error
	if err != nil {
		return nil, 0, err
	}

	return nfts, length, nil
}

func ListNFTByType(page, size int, address, order string, ntype string) ([]NFTStore, int64, error) {
	var nfts []NFTStore
	var orderRules string
	var length int64
	switch order {
	case "date_asc":
		orderRules = "time"
	case "date_desc":
		orderRules = "time desc"
	default:
		return nil, 0, xerrors.Errorf("not spport order rules: %s", order)
	}

	err := GlobalDataBase.Model(&NFTStore{}).Where("address = ? AND type = ?", address, ntype).Order(orderRules).Offset((page - 1) * size).Limit(size).Find(&nfts).Error
	if err != nil {
		return nil, 0, err
	}

	err = GlobalDataBase.Model(&NFTStore{}).Where("address = ? AND type = ?", address, ntype).Count(&length).Error
	if err != nil {
		return nil, 0, err
	}

	return nfts, length, nil
}
