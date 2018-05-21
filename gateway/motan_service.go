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
	"github.com/Loopring/relay-lib/log"
	"github.com/Loopring/relay-lib/motan"
	"math/big"
)

type MotanService struct {
	accountManager accountmanager.AccountManager
	orderViewer    ordermanager.OrderViewer
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

func (s *MotanService) GetMinerOrders(req *motan.MinerOrdersReq) *motan.MinerOrdersRes {
	res := &motan.MinerOrdersRes{}
	res.List = s.orderViewer.MinerOrders(req.Delegate, req.TokenS, req.TokenB, req.Length, req.ReservedTime, req.StartBlockNumber, req.EndBlockNumber, req.FilterOrderHashLists...)

	log.Debugf("motan service, GetMinerOrders list length:%d", len(res.List))

	return res
}

func StartMotanService(options motan.MotanServerOptions, accountManager accountmanager.AccountManager, orderViewer ordermanager.OrderViewer) {
	service := &MotanService{}
	service.accountManager = accountManager
	service.orderViewer = orderViewer
	options.ServerInstance = service
	go motan.RunServer(options)
}
