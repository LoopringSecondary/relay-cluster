/*

  Copyright 2017 Loopring Project Ltd (Loopring Foundation).

  Licensed under the Apache License, Version 2.0 (the "License");
  you may not use this file except in compliance with the License.
  You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

  Unless required by applicable law or agreed to in writing, software
  distributed under the License is distributed on an "AS IS" BASIS,
  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
  See the License for the specific language governing permissions and
  limitations under the License.

*/

package dao

import (
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/jinzhu/gorm"
	"time"
)

type CityPartner struct {
	ID            int    `gorm:"column:id;primary_key;" json:"-"`
	WalletAddress string `gorm:"column:wallet_address;type:varchar(42)" json:"walletAddress"`
	CityPartner   string `gorm:"column:city_partner;type:varchar(50)" json:"cityPartner"`
	CreateTime    int64  `gorm:"column:create_time;type:bigint" json:"-"`
}

type CustumerInvitationInfo struct {
	ID           int    `gorm:"column:id;primary_key;" json:"-"`
	Device       string `gorm:"column:device;type:varchar(100)" json:"device"`
	ActivateCode string `gorm:"column:activate_code;type:varchar(50)" json:"activateCode"`
	CityPartner  string `gorm:"column:city_partner;type:varchar(50)" json:"cityPartner"`
	Activate     int    `gorm:"column:activate;type:int" json:"activate"`
	CreateTime   int64  `gorm:"column:create_time;type:bigint" json:"-"`
}

type CityPartnerReceivedDetail struct {
	ID            int    `gorm:"column:id;primary_key;" json:"-"`
	WalletAddress string `gorm:"column:wallet_address;type:varchar(50)" json:"walletAddress"`
	TokenSymbol   string `gorm:"column:token_symbol;type:varchar(50)" json:"tokenSymbol"`
	TokenAddress  string `gorm:"column:token_address;type:varchar(50)" json:"tokenAddress"`
	Amount        string `gorm:"column:amount;type:varchar(50)" json:"amount"`
	Ringhash      string `gorm:"column:ringhash;type:varchar(100)" json:"ringhash"`
	Orderhash     string `gorm:"column:orderhash;type:varchar(100)" json:"orderhash"`
	CreateTime    int64  `gorm:"column:create_time;type:bigint" json:"-"`
}

type CityPartnerReceived struct {
	ID            int    `gorm:"column:id;primary_key;" json:"-"`
	WalletAddress string `gorm:"column:wallet_address;type:varchar(50)" json:"walletAddress"`
	TokenSymbol   string `gorm:"column:token_symbol;type:varchar(50)" json:"tokenSymbol"`
	TokenAddress  string `gorm:"column:token_address;type:varchar(50)" json:"tokenAddress"`
	Amount        string `gorm:"column:amount;type:varchar(50)" json:"amount"`
	HumanAmount   string `gorm:"column:human_amount;type:varchar(50)" json:"humanAmount"`
	CreateTime    int64  `gorm:"column:create_time;type:bigint" json:"-"`
}

func (s *RdsService) SaveCityPartner(cp *CityPartner) (bool, error) {
	var count int
	err := s.Db.Model(&CityPartner{}).Where("city_partner=?", cp.CityPartner).Count(&count).Error
	if nil != err {
		return false, err
	} else {
		if count <= 0 {
			cp.CreateTime = time.Now().Unix()
			err := s.Add(cp)
			if nil != err {
				return false, err
			}
			return true, nil
		} else {
			return false, errors.New("duplicated invitation_code")
		}
	}
}

func (s *RdsService) FindCityPartnerByWalletAddress(address common.Address) (*CityPartner, error) {
	cp := &CityPartner{}
	err := s.Db.Model(&CityPartner{}).
		Where("wallet_address=?", address.Hex()).
		First(cp).Error
	return cp, err
}

func (s *RdsService) FindCityPartnerByCityPartner(cityPartner string) (*CityPartner, error) {
	cp := &CityPartner{}
	err := s.Db.Model(&CityPartner{}).
		Where("city_partner=?", cityPartner).Order("id desc").
		First(cp).Error
	return cp, err
}

type count struct {
	Count int
}

func (s *RdsService) GetCityPartnerCustomerCount(cityPartner string) (int, error) {
	c := count{}
	err := s.Db.Model(&CustumerInvitationInfo{}).
		Where("city_partner=?", cityPartner).
		Where("activate>=?", 1).Select("sum(activate) as count").
		Scan(&c).Error
	return c.Count, err
}

func (s *RdsService) FindReceivedByWalletAndToken(walletAddress, tokenAddress common.Address) (*CityPartnerReceived, error) {
	received := &CityPartnerReceived{}
	err := s.Db.Model(&CityPartnerReceived{}).
		Where("wallet_address=?", walletAddress.Hex()).
		Where("token_address=?", tokenAddress.Hex()).First(received).Error
	return received, err
}

func (s *RdsService) GetAllReceivedByWallet(walletAddress string) ([]*CityPartnerReceived, error) {
	var receiveds []*CityPartnerReceived
	err := s.Db.Model(&CityPartnerReceived{}).
		Where("wallet_address=?", walletAddress).
		Find(&receiveds).Error
	return receiveds, err
}

func (s *RdsService) SaveCustomerInvitationInfo(info *CustumerInvitationInfo) error {
	var count int
	err := s.Db.Model(&CustumerInvitationInfo{}).
		Where("activate_code=?", info.ActivateCode).
		Where("city_partner=?", info.CityPartner).
		Count(&count).Error
	if nil != err {
		return err
	} else {
		if count <= 0 {
			info.CreateTime = time.Now().Unix()
			return s.Add(info)
		} else {
			return nil
		}
	}
	//info.CreateTime = time.Now().Unix()
	//return s.Add(info)
}

func (s *RdsService) FindCustomerInvitationInfo(ip string) (*CustumerInvitationInfo, error) {
	resInfo := &CustumerInvitationInfo{}
	err := s.Db.
		Where("activate_code=?", ip).Order("id desc").
		//Where("activate=?", 0).
		First(resInfo).Error
	if nil != err {
		return nil, err
	} else {
		return resInfo, nil
	}
}

func (s *RdsService) AddCustomerInvitationActivate(info *CustumerInvitationInfo) error {
	return s.Db.Model(&CustumerInvitationInfo{}).
		Where("id=?", info.ID).
		Update("activate", gorm.Expr("activate + ?", 1)).Error
}

func (s *RdsService) UpdateCityPartnerReceived(received *CityPartnerReceived) error {
	return s.Db.Model(&CityPartnerReceived{}).
		Where("wallet_address=?", received.WalletAddress).
		Where("token_address=?", received.TokenAddress).
		Updates(map[string]interface{}{"amount": received.Amount, "human_amount": received.HumanAmount}).Error
}
