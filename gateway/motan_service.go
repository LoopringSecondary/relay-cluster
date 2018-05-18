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

package gateway

import (
	"github.com/Loopring/relay-cluster/accountmanager"
	"github.com/Loopring/relay-cluster/ordermanager"
	"github.com/Loopring/relay-lib/motan"
	"math/big"
)

type MotanService struct {
	accountManager accountmanager.AccountManager
	orderManager   ordermanager.OrderManager
}

func (s *MotanService) GetBalanceAndAllowance(req *motan.AccountBalanceAndAllowanceReq) *motan.AccountBalanceAndAllowanceRes {
	res := &motan.AccountBalanceAndAllowanceRes{}
	if balance, allowance, err := accountmanager.GetBalanceAndAllowance(req.Owner, req.Token, req.Spender); nil != err {
		res.Allowance = big.NewInt(int64(0))
		res.Balance = big.NewInt(int64(0))
		//res.Err = err.Error()
	} else {
		res.Balance = new(big.Int).Set(balance)
		res.Allowance = new(big.Int).Set(allowance)
		//res.Err = ""
	}
	return res
}

func StartMotanService(options motan.MotanServerOptions, accountManager accountmanager.AccountManager) {
	service := &MotanService{}
	service.accountManager = accountManager
	options.ServerInstance = service
	go motan.RunServer(options)
}
