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
	ID             int            `gorm:"column:id;primary_key;" json:"-"`
	WalletAddress  common.Address `gorm:"column:wallet_address;type:varchar(42)" json:"walletAddress"`
	InvitationCode string         `gorm:"column:invitation_code;type:varchar(50)" json:"invitationCode"`
}

type CustumerInvitationInfo struct {
	ID             int    `gorm:"column:id;primary_key;" json:"-"`
	Device string `gorm:"column:device;type:varchar(100)" json:"device"`
	ActivateCode   string `gorm:"column:activate_code;type:varchar(50)" json:"activateCode"`
	InvitationCode string `gorm:"column:invitation_code;type:varchar(50)" json:"invitationCode"`
	Activate       int    `gorm:"column:activate;type:int" json:"activate"`
}

type CityPartnerReceivedDetail struct {
	ID            int    `gorm:"column:id;primary_key;" json:"-"`
	WalletAddress string `gorm:"column:wallet_address;type:varchar(50)" json:"walletAddress"`
	TokenSymbol   string `gorm:"column:token_symbol;type:varchar(50)" json:"tokenSymbol"`
	TokenAddress  string `gorm:"column:token_address;type:varchar(50)" json:"tokenAddress"`
	Amount        string `gorm:"column:amount;type:varchar(50)" json:"amount"`
	Ringhash      string `gorm:"column:ringhash;type:varchar(100)" json:"ringhash"`
	Orderhash     string `gorm:"column:orderhash;type:varchar(100)" json:"orderhash"`
}

type CityPartnerReceived struct {
	ID            int    `gorm:"column:id;primary_key;" json:"-"`
	WalletAddress string `gorm:"column:wallet_address;type:varchar(50)" json:"walletAddress"`
	TokenSymbol   string `gorm:"column:token_symbol;type:varchar(50)" json:"tokenSymbol"`
	TokenAddress  string `gorm:"column:token_address;type:varchar(50)" json:"tokenAddress"`
	Amount        string `gorm:"column:amount;type:varchar(50)" json:"amount"`
	HumanAmount   string `gorm:"column:human_amount;type:varchar(50)" json:"humanAmount"`
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
	cp := &CityPartner{}
	err := s.Db.Model(&CityPartner{}).
		Where("wallet_address=?", address).
		First(&cp).Error
	return cp, err
}

func (s *RdsService) FindCityPartnerByInvitationCode(invitationCode string) (*CityPartner, error) {
	cp := &CityPartner{}
	err := s.Db.Model(&CityPartner{}).
		Where("invitation_code=?", invitationCode).
		FirstOrInit(&cp).Error
	return cp, err
}

func (s *RdsService) GetCityPartnerCustomerCount(invitationCode string) (int, error) {
	var count int
	err := s.Db.Model(&CustumerInvitationInfo{}).Where("invitation_code=?", invitationCode).Count(&count).Error
	return count, err
}

func (s *RdsService) GetAllActivateCode(invitationCode string) ([]string, error) {
	var codes []string
	err := s.Db.Model(&CustumerInvitationInfo{}).
		Where("invitation_code=?", invitationCode).
		Where("activate=?", 0).
		Count(&codes).Error
	return codes, err
}

func (s *RdsService) FindReceivedByWalletAndToken(walletAddress, tokenAddress common.Address) (*CityPartnerReceived, error) {
	received := &CityPartnerReceived{}
	err := s.Db.Model(&CityPartnerReceived{}).
		Where("wallet_address=?", walletAddress.Hex()).
		Where("token_address=?", tokenAddress.Hex()).First(&received).Error
	return received, err
}

func (s *RdsService) GetAllReceivedByWallet(walletAddress common.Address) ([]*CityPartnerReceived, error) {
	receiveds := []*CityPartnerReceived{}
	err := s.Db.Model(&CityPartnerReceived{}).
		Where("wallet_address=?", walletAddress.Hex()).
		Find(&receiveds).Error
	return receiveds, err
}

func (s *RdsService) SaveCustumerInvitationInfo(info *CustumerInvitationInfo) error {
	var count int
	err := s.Db.Model(&CustumerInvitationInfo{}).
		Where("activate_code=?", info.ActivateCode).
		Where("activate=?", 0).
		Count(&count).Error
	if nil != err {
		return err
	} else {
		if count <= 0 {
			s.Add(info)
			return nil
		} else {
			return errors.New("duplicated device info")
		}
	}
}

func (s *RdsService) FindCustumerInvitationInfo(info CustumerInvitationInfo) (*CustumerInvitationInfo, error) {
	resInfo := &CustumerInvitationInfo{}
	err := s.Db.Model(&CustumerInvitationInfo{}).
		Where("activate_code=?", info.ActivateCode).
		Where("activate=?", 0).
		Count(resInfo).Error
	if nil != err {
		return nil, err
	} else {
		return resInfo, nil
	}
}

func (s *RdsService) AddCustumerInvitationActivate(info *CustumerInvitationInfo) error {
	return s.Db.Model(&CustumerInvitationInfo{}).
		Where("activate_code=?", info.ActivateCode).
		Where("invitation_code=?", info.InvitationCode).
		Update("activate", info.Activate).Error
}
