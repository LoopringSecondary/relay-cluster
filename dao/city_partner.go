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
)

type CityPartner struct {
	ID             int            `gorm:"column:id;primary_key;"`
	WalletAddress  common.Address `gorm:"column:wallet_address;type:varchar(42)"`
	InvitationCode string         `gorm:"column:invitation_code;type:varchar(50)"`
}

type CustumerInvitationInfo struct {
	ID             int    `gorm:"column:id;primary_key;"`
	Device         string `gorm:"column:device;type:varchar(50)"`
	InvitationCode string `gorm:"column:invitation_code;type:varchar(50)"`
	Activate       int    `gorm:"column:activate;type:int"`
}

type CityPartnerReceivedDetail struct {
	ID            int    `gorm:"column:id;primary_key;"`
	WalletAddress string `gorm:"column:wallet_address;type:varchar(50)"`
	TokenSymbol   string `gorm:"column:token_symbol;type:varchar(50)"`
	TokenAddress  string `gorm:"column:token_address;type:varchar(50)"`
	Amount        string `gorm:"column:amount;type:varchar(50)"`
	Ringhash      string `gorm:"column:ringhash;type:varchar(100)"`
	Orderhash     string `gorm:"column:orderhash;type:varchar(100)"`
}

type CityPartnerReceived struct {
	ID            int    `gorm:"column:id;primary_key;"`
	WalletAddress string `gorm:"column:wallet_address;type:varchar(50)"`
	TokenSymbol   string `gorm:"column:token_symbol;type:varchar(50)"`
	TokenAddress  string `gorm:"column:token_address;type:varchar(50)"`
	Amount        string `gorm:"column:amount;type:varchar(50)"`
	HumanAmount   string `gorm:"column:human_amount;type:varchar(50)"`
}

func (s *RdsService) SaveCityPartner(cp CityPartner) (bool, error) {
	var count int
	err := s.Db.Model(&CityPartner{}).Where("invitation_code=?", cp.InvitationCode).Count(&count).Error
	if nil != err {
		return false, err
	} else {
		if count <= 0 {
			s.Add(cp)
			return true, nil
		} else {
			return false, errors.New("duplicated invitation_code")
		}
	}
}

func (s *RdsService) FindCityPartnerByWalletAddress(address common.Address) (*CityPartner, error) {

}

func (s *RdsService) FindCityPartnerByInvitationCode(invitationCode string) (*CityPartner, error) {

}

func (s *RdsService) GetCityPartnerCustomerCount(invitationCode string) (int, error) {
	var count int
	err := s.Db.Model(&CustumerInvitationInfo{}).Where("invitation_code=?", invitationCode).Count(&count).Error
	return count, err
}

func (s *RdsService) FindReceivedByWalletAndToken(walletAddress, tokenAddress common.Address) (*CityPartnerReceived, error) {

}

func (s *RdsService) GetAllReceivedByWallet(walletAddress common.Address) ([]*CityPartnerReceived, error) {

}

func (s *RdsService) SaveCustumerInvitationInfo(info CustumerInvitationInfo) (bool, error) {
	var count int
	err := s.Db.Model(&CustumerInvitationInfo{}).Where("device=?", info.Device).Count(&count).Error
	if nil != err {
		return false, err
	} else {
		if count <= 0 {
			s.Add(info)
			return true, nil
		} else {
			return false, errors.New("duplicated device info")
		}
	}
}

func (s *RdsService) FindCustumerInvitationInfo(info CustumerInvitationInfo) (*CustumerInvitationInfo, error) {
	resInfo := &CustumerInvitationInfo{}
	err := s.Db.Model(&CustumerInvitationInfo{}).Where("device=?", info.Device).Count(resInfo).Error
	if nil != err {
		return nil, err
	} else {
		return resInfo, nil
	}
}

func (s *RdsService) AddCustumerInvitationActivate(info *CustumerInvitationInfo) error {
	return s.Db.Model(&CustumerInvitationInfo{}).Where("device=?", info.Device).Update("activate", info.Activate).Error
}
