package database

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"strings"
	"time"

	"github.com/memoio/xspace-server/config"
	"golang.org/x/xerrors"
)

type ActionStore struct {
	Id      int
	Name    string
	Address string
	Time    time.Time
}

func (action *ActionStore) CreateActionInfo() error {
	return GlobalDataBase.Create(action).Error
}

func ListActionHistory(address string, page, size int, order string) ([]ActionStore, error) {
	var actions []ActionStore
	var orderRules string
	switch order {
	case "time_asc":
		order = "time"
	case "time_desc":
		order = "time desc"
	default:
		return nil, xerrors.Errorf("not spport order rules: %d", order)
	}

	err := GlobalDataBase.Model(&ActionStore{}).Where("address = ?", address).Order(orderRules).Find(&actions).Error
	if err != nil {
		return nil, err
	}

	return actions, nil
}

func ListActionHistoryByID(address string, page, size int, order string, id int) ([]ActionStore, error) {
	var actions []ActionStore
	var orderRules string
	switch order {
	case "date_asc":
		order = "time"
	case "date_desc":
		order = "time desc"
	default:
		return nil, xerrors.Errorf("not spport order rules: %d", order)
	}

	if id == -1 {
		err := GlobalDataBase.Model(&ActionStore{}).Where("address = ?", address).Order(orderRules).Find(&actions).Error
		if err != nil {
			return nil, err
		}
	} else {
		err := GlobalDataBase.Model(&ActionStore{}).Where("address = ? and id = ?", address, id).Order(orderRules).Find(&actions).Error
		if err != nil {
			return nil, err
		}
	}

	return actions, nil
}

type UserStore struct {
	Address     string `gorm:"primarykey"`
	Points      int64
	InviteCode  string `gorm:"uniqueIndex"`
	InvitedCode string
	Referrals   int
	Space       int
	UpdateTime  time.Time `gorm:"conlum:updatetime"`
}

func (user *UserStore) CreateUserInfo() error {
	return GlobalDataBase.Create(user).Error
}

func (user *UserStore) UpdateUserInfo() error {
	return GlobalDataBase.Save(user).Error
}

func GetUserInfo(address string) (UserStore, error) {
	var user UserStore = UserStore{
		Address:     address,
		Points:      0,
		InviteCode:  "",
		InvitedCode: "",
		Referrals:   0,
		Space:       config.DefaultSpace,
		UpdateTime:  time.Now(),
	}
	err := GlobalDataBase.Model(&UserStore{}).Where("address = ?", address).Find(&user).Error
	if err != nil {
		if !strings.Contains(err.Error(), "record not found") {
			return user, err
		}

		user.InviteCode = createCode()
		return user, user.CreateUserInfo()
	}

	return user, err
}

func createCode() string {
	var length int64
	GlobalDataBase.Model(&UserStore{}).Count(&length)

	var userId = 123456789 + int32(length)
	var buffer bytes.Buffer
	binary.Write(&buffer, binary.BigEndian, userId)

	return base64.RawStdEncoding.EncodeToString(buffer.Bytes())
}
