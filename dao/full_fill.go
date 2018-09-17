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
	"github.com/Loopring/relay-cluster/ringtrackermanager/types"
	types2 "github.com/Loopring/relay-lib/types"
	"strconv"
)

type FullFillEvent struct {
	Id              int     `gorm:"column:id;primary_key;" json:"id"`
	Protocol        string  `gorm:"column:contract_address;type:varchar(42)" json:"protocol"`
	DelegateAddress string  `gorm:"column:delegate_address;type:varchar(42)" json:"delegateAddress"`
	Owner           string  `gorm:"column:owner;type:varchar(42)" json:"owner"`
	RingIndex       int64   `gorm:"column:ring_index;" json:"ringIndex"`
	FillIndex       int64   `gorm:"column:fill_index;" json:"fillIndex"`
	CreateTime      int64   `gorm:"column:create_time" json:"createTime"`
	RingHash        string  `gorm:"column:ring_hash;varchar(82)" json:"ringHash"`
	TxHash          string  `gorm:"column:tx_hash;type:varchar(82)" json:"txHash"`
	OrderHash       string  `gorm:"column:order_hash;type:varchar(82)" json:"orderHash"`
	TokenS          string  `gorm:"column:token_s;type:varchar(42)" json:"tokenS"`
	SymbolS         string  `gorm:"column:symbol_s;type:varchar(42)"`
	TokenB          string  `gorm:"column:token_b;type:varchar(42)" json:"tokenB"`
	SymbolB         string  `gorm:"column:symbol_b;type:varchar(42)"`
	AmountS         string  `gorm:"column:amount_s;type:varchar(40)" json:"amountS"`
	AmountB         string  `gorm:"column:amount_b;type:varchar(40)" json:"amountB"`
	LrcFee          string  `gorm:"column:lrc_fee;type:varchar(40)" json:"LrcFee"`
	Market          string  `gorm:"column:market;type:varchar(42)" json:"market"`
	Side            string  `gorm:"column:side" json:"side"`
	Miner           string  `gorm:"column:miner;type:varchar(42)"`
	WalletAddress   string  `gorm:"column:wallet_address;type:varchar(42)"`
	LrcCal          float64 `gorm:"column:lrc_cal;type:float"`
	TokenAmountCal  float64 `gorm:"column:token_amount_cal;type:float"`
	AmountBCal      float64 `gorm:"column:amount_s_cal;type:float"`
	OrderType       string  `gorm:"column:order_type;type:varchar(50)" json:"orderType"`
	Relay           string  `gorm:"column:relay;type:varchar(100)" json:"relay"`
}

type Relay struct {
	Id      int    `gorm:"column:id;primary_key;"`
	Relay   string `gorm:"column:relay;type:varchar(100)" `
	Miner   string `gorm:"column:miner;type:varchar(42)"`
	Website string `gorm:"column:website;type:text"`
}

type Dex struct {
	Id            int    `gorm:"column:id;primary_key;"`
	Dex           string `gorm:"column:dex;type:varchar(100)" `
	WalletAddress string `gorm:"column:wallet_address;type:varchar(42)"`
	Website       string `gorm:"column:website;type:text"`
}

type FailFill struct {
	TxHash string `gorm:"column:tx_hash;type:varchar(82)"`
}

func (f *FullFillEvent) ConvertToFullFill(src *types2.OrderFilledEvent, miner string) {
	f.AmountS = src.AmountS.String()
	f.AmountB = src.AmountB.String()
	f.LrcFee = src.LrcFee.String()
	f.Protocol = src.Protocol.Hex()
	f.DelegateAddress = src.DelegateAddress.Hex()
	f.RingIndex = src.RingIndex.Int64()
	f.FillIndex = src.FillIndex.Int64()
	f.CreateTime = src.BlockTime
	f.RingHash = src.Ringhash.Hex()
	f.TxHash = src.TxHash.Hex()
	f.OrderHash = src.OrderHash.Hex()
	f.TokenS = src.TokenS.Hex()
	f.TokenB = src.TokenB.Hex()
	f.Owner = src.Owner.Hex()
	f.Market = src.Market
	f.Miner = miner
}

func (s *RdsService) GetAmount() (res types.AmountResp) {
	s.Db.Raw("select count(*) trades, " +
		"count(distinct token_b) tokens, " +
		"(select count(distinct relay) from lpr_relays) relays, " +
		"(select count(*) from lpr_dexes) dexs, " +
		"(select count(*) from lpr_ring_mined_events) rings " +
		"from lpr_full_fill_events " +
		"where order_type = 'market_order'").Scan(&res)
	return
}

func (s *RdsService) GetTrend(duration types.Duration, trendType types.TrendType, keyword string, currency types.Currency, date string, startDate int64) (res []types.TrendFill) {
	sql := "select " +
		"count(*) trade, " +
		"sum(a.lrc_cal*b.coin_amount) fee, " +
		"sum(a.token_amount_cal*b.coin_amount) volume, " +
		date + " date " +
		"from lpr_full_fill_events a left join lpr_token_price_trends b on UNIX_TIMESTAMP(date_format(FROM_UNIXTIME(a.create_time), '%Y-%m-%d')) = b.time " +
		"where a.order_type = 'market_order' " +
		"and a.side = 'buy' " +
		"and a.create_time >= " + strconv.FormatInt(startDate, 10) + " " +
		"and b.coin_name = '" + string(currency) + "' "
	switch trendType {
	case types.TOKEN:
		sql += "and (a.token_s = '" + keyword + "' or a.token_b = '" + keyword + "') "
	case types.DEX:
		sql += "and a.wallet_address = (select wallet_address from lpr_dexes where dex = '" + keyword + "') "
	case types.RING:
		sql += "and a.miner = '" + keyword + "' "
	case types.RELAY:
		sql += "and a.miner in (select miner from lpr_relays where relay = '" + keyword + "') "
	}
	s.Db.Raw(sql + "group by date order by date desc").Scan(&res)
	return
}

func (s *RdsService) GetEcoStatFill(currency types.Currency, trendType types.TrendType, indicator types.Indicator, startDate int64) (res []types.EcoStatFill) {
	sql := "select e.name, sum(e.value) value from (select @r:=@r+1 id, if(@r > 5, 'others', d.name) name, d.value value from (select "
	switch trendType {
	case types.TOKEN:
		sql += "a.symbol_b name, "
	case types.DEX:
		sql += "if(c.dex is null, 'others', c.dex) name, "
	case types.RELAY:
		sql += "if(c.relay is null, 'others', c.relay) name, "
	}
	switch indicator {
	case types.VOLUME:
		sql += "sum(a.token_amount_cal*b.coin_amount) value "
	case types.TRADE:
		sql += "count(*) value "
	case types.FEE:
		sql += "sum(a.lrc_cal*b.coin_amount) value "
	}
	switch trendType {
	case types.TOKEN:
		sql += "from lpr_full_fill_events a left join lpr_token_price_trends b on UNIX_TIMESTAMP(date_format(FROM_UNIXTIME(a.create_time), '%Y-%m-%d')) = b.time " +
			"where b.coin_name = '" + string(currency) + "' "
	case types.DEX:
		sql += "from lpr_dexes c right join lpr_full_fill_events a on a.wallet_address = c.wallet_address left join lpr_token_price_trends b on UNIX_TIMESTAMP(date_format(FROM_UNIXTIME(a.create_time), '%Y-%m-%d')) = b.time " +
			"where b.coin_name = '" + string(currency) + "' "
	case types.RELAY:
		sql += "from lpr_relays c right join lpr_full_fill_events a on a.miner = c.miner left join lpr_token_price_trends b on UNIX_TIMESTAMP(date_format(FROM_UNIXTIME(a.create_time), '%Y-%m-%d')) = b.time " +
			"where b.coin_name = '" + string(currency) + "' "
	}
	sql += "and a.order_type = 'market_order' and a.create_time >= " + strconv.FormatInt(startDate, 10) + " group by name order by value desc) d, (select @r:=0) f) e group by e.name order by case when name != 'others' then 0 else 1 end, value desc"
	s.Db.Raw(sql).Scan(&res)
	return
}

func (s *RdsService) GetAllFills(miner, txHash string) (res []*FullFillEvent) {
	s.Db.Raw("select " +
		"b.contract_address, b.delegate_address, b.owner, b.ring_index, b.fill_index, b.create_time, b.ring_hash, b.tx_hash, b.order_hash, " +
		"b.token_s, b.token_b, b.amount_s, b.amount_b, b.lrc_fee, b.market, b.side, '" + miner + "' miner, a.wallet_address, b.order_type " +
		"from lpr_orders a join lpr_fill_events b on a.order_hash = b.order_hash  " +
		"where b.tx_hash = '" + txHash + "'").Scan(&res)
	return
}

func (s *RdsService) GetAllFullFills(currency types.Currency, trendType types.TrendType, keyword, search string, pageIndex, pageSize int) (res []FullFillEvent) {
	sql := "select b.id, " +
		"(select d.relay from lpr_relays d where d.miner = b.miner) relay, b.contract_address, b.delegate_address, b.owner, b.ring_index, b.fill_index, b.create_time, b.ring_hash, b.tx_hash, b.order_hash, " +
		"b.symbol_s token_s, b.symbol_b token_b, concat('0x', conv(b.amount_s, 10, 16)) amount_s, concat('0x', conv(b.amount_b, 10, 16)) amount_b, concat('0x', conv(b.lrc_fee, 10, 16)) lrc_fee, b.market, b.side, b.miner, b.order_type, b.wallet_address, " +
		"b.lrc_cal*a.coin_amount lrc_cal, b.token_amount_cal*a.coin_amount token_amount_cal " +
		"from lpr_token_price_trends a right join lpr_full_fill_events b on UNIX_TIMESTAMP(date_format(FROM_UNIXTIME(b.create_time), '%Y-%m-%d')) = a.time " +
		"where a.coin_name = '" + string(currency) + "' " +
		"and b.order_type = 'market_order' "
	switch trendType {
	case types.TOKEN:
		sql += "and (b.token_s = '" + keyword + "' or b.token_b = '" + keyword + "') "
	case types.RELAY:
		sql += "and b.miner in (select miner from lpr_relays where relay = '" + keyword + "') "
	case types.DEX:
		sql += "and b.wallet_address = (select wallet_address from lpr_dexes where dex = '" + keyword + "') "
	case types.RING:
		sql += "and b.miner = '" + keyword + "' "
	}
	if len(search) > 0 {
		sql += "and (b.order_hash like '%" + search + "%' or b.tx_hash like '%" + search + "%' or b.ring_hash like '%" + search + "%' " +
			"or b.owner like '%" + search + "%' or b.wallet_address like '%" + search + "%' or b.miner like '%" + search + "%' or b.miner in (select miner from lpr_relays where relay like '%" + search + "%')) "
	}
	sql += "order by b.create_time desc"
	s.Db.Raw(sql).Offset((pageIndex - 1) * pageSize).Limit(pageSize).Scan(&res)
	return
}

func (s *RdsService) GetFullFill(id int64) (res FullFillEvent) {
	s.Db.Model(&FullFillEvent{}).Where("id = ?", id).First(&res)
	return
}

func (s *RdsService) CountFullFills(trendType types.TrendType, keyword, search string) (res int) {
	sql := "order_type = 'market_order' "
	switch trendType {
	case types.TOKEN:
		sql += "and (token_s = '" + keyword + "' or token_b = '" + keyword + "') "
	case types.RELAY:
		sql += "and miner in (select miner from lpr_relays where relay = '" + keyword + "') "
	case types.DEX:
		sql += "and wallet_address = (select wallet_address from lpr_dexes where dex = '" + keyword + "') "
	case types.RING:
		sql += "and miner = '" + keyword + "' "
	}
	if len(search) > 0 {
		sql += "and (order_hash like '%" + search + "%' or tx_hash like '%" + search + "%' or ring_hash like '%" + search + "%' " +
			"or owner like '%" + search + "%' or wallet_address like '%" + search + "%' or miner like '%" + search + "%' or miner in (select miner from lpr_relays where relay like '%" + search + "%')) "
	}
	s.Db.Model(&FullFillEvent{}).Where(sql).Count(&res)
	return
}

func (s *RdsService) GetTokenSymbols() (res []string) {
	s.Db.Model(&FullFillEvent{}).Where("order_type = ?", "market_order").Pluck("DISTINCT symbol_b", &res)
	return
}

func (s *RdsService) GetFillsByToken(currency types.Currency, indicator types.Indicator, pageIndex, pageSize int) (res []types.TokenFill) {
	sql := "select a.token_b token, a.symbol_b symbol, count(*) trade, " +
		"sum(a.token_amount_cal*b.coin_amount) volume, " +
		"sum(a.amount_b_cal) token_volume, " +
		"sum(a.lrc_cal*b.coin_amount) fee " +
		"from lpr_full_fill_events a left join lpr_token_price_trends b on UNIX_TIMESTAMP(date_format(FROM_UNIXTIME(a.create_time), '%Y-%m-%d')) = b.time " +
		"where a.order_type = 'market_order' and b.coin_name = '" + string(currency) + "' " +
		"group by token, symbol order by " + string(indicator) + " desc"
	s.Db.Raw(sql).Offset((pageIndex - 1) * pageSize).Limit(pageSize).Scan(&res)
	return
}

func (s *RdsService) GetFillsByRelay(currency types.Currency, indicator types.Indicator, pageIndex, pageSize int) (res []types.RelayFill) {
	sql := "select c.relay, c.website, count(*) trade, " +
		"sum(a.token_amount_cal*b.coin_amount) volume, " +
		"sum(a.lrc_cal*b.coin_amount) fee " +
		"from lpr_relays c left join lpr_full_fill_events a on a.miner = c.miner left join lpr_token_price_trends b on UNIX_TIMESTAMP(date_format(FROM_UNIXTIME(a.create_time), '%Y-%m-%d')) = b.time " +
		"where a.order_type = 'market_order' and b.coin_name = '" + string(currency) + "' " +
		"group by c.relay, c.website order by " + types.IndicatorToStr(indicator) + " desc"
	s.Db.Raw(sql).Offset((pageIndex - 1) * pageSize).Limit(pageSize).Scan(&res)
	return
}

func (s *RdsService) CountRelays() (res int) {
	s.Db.Model(&Relay{}).Select("count(DISTINCT relay)").Count(&res)
	return
}

func (s *RdsService) GetFillsByDex(currency types.Currency, indicator types.Indicator, pageIndex, pageSize int) (res []types.DexFill) {
	sql := "select c.dex, count(*) trade, " +
		"sum(a.token_amount_cal*b.coin_amount) volume, " +
		"sum(a.lrc_cal*b.coin_amount) fee " +
		"from lpr_dexes c left join lpr_full_fill_events a on a.wallet_address = c.wallet_address left join lpr_token_price_trends b on UNIX_TIMESTAMP(date_format(FROM_UNIXTIME(a.create_time), '%Y-%m-%d')) = b.time " +
		"where a.order_type = 'market_order' and b.coin_name = '" + string(currency) + "' " +
		"group by c.dex order by " + types.IndicatorToStr(indicator) + " desc"
	s.Db.Raw(sql).Offset((pageIndex - 1) * pageSize).Limit(pageSize).Scan(&res)
	return
}

func (s *RdsService) CountDexs() (res int) {
	s.Db.Model(&Dex{}).Select("count(DISTINCT wallet_address)").Count(&res)
	return
}

func (s *RdsService) GetTokensByRelay(currency types.Currency, relay string) (res []types.TokenFill) {
	sql := "select f.token, f.symbol, sum(f.trade) trade, sum(f.volume) volume, sum(f.fee) fee from " +
		"(select  @r:=@r+1 id, if(@r > 5, 'others', d.token) token, if(@r > 5, 'others', d.symbol) symbol, d.trade, d.volume, d.fee from " +
		"(select token_b token, symbol_b symbol, count(*) trade, " +
		"sum(a.token_amount_cal*b.coin_amount) volume, " +
		"sum(a.lrc_cal*b.coin_amount) fee " +
		"from lpr_full_fill_events a left join lpr_token_price_trends b on UNIX_TIMESTAMP(date_format(FROM_UNIXTIME(a.create_time), '%Y-%m-%d')) = b.time " +
		"where a.order_type = 'market_order' " +
		"and a.miner in (select miner from lpr_relays where relay = '" + relay + "') " +
		"and b.coin_name = '" + string(currency) + "' " +
		"group by token, symbol order by volume desc) d, (select @r:=0) e) f " +
		"group by f.token, f.symbol  order by case when f.token != 'others' then 0 else 1 end, volume desc"
	s.Db.Raw(sql).Scan(&res)
	return
}

func (s *RdsService) GetTradeDetails(delegateAddress string, ringIndex, fillIndex int64) (res []FullFillEvent) {
	sql := "select id, " +
		"contract_address, delegate_address, owner, ring_index, fill_index, create_time, ring_hash, tx_hash, order_hash, " +
		"symbol_s token_s, symbol_b token_b, concat('0x', conv(amount_s, 10, 16)) amount_s, concat('0x', conv(amount_b, 10, 16)) amount_b,  concat('0x', conv(lrc_fee, 10, 16)) lrc_fee, market, side, miner, wallet_address " +
		"from lpr_full_fill_events " +
		"where ring_index = " + strconv.FormatInt(ringIndex, 10) + " and fill_index = " + strconv.FormatInt(fillIndex, 10) + " "
	if len(delegateAddress) > 0 {
		sql += "and delegate_address = '" + delegateAddress + "'"
	}
	s.Db.Raw(sql).Scan(&res)
	return
}

func (s *RdsService) AddFullFills(fills []*FullFillEvent) {
	sql := "insert into lpr_full_fill_events(`contract_address`, `delegate_address`, `owner`, `ring_index`, `fill_index`, `create_time`, `ring_hash`, `tx_hash`, `order_hash`, `token_s`, `token_b`, `symbol_s`, `symbol_b`, `amount_s`, `amount_b`, `lrc_fee`, `market`, `side`, `miner`, `wallet_address`, `lrc_cal`, `token_amount_cal`, `order_type`, `amount_b_cal`) values "
	i := 0
	for _, fill := range fills {
		sql += "('" + fill.Protocol + "','" + fill.DelegateAddress + "','" + fill.Owner + "'," + strconv.FormatInt(fill.RingIndex, 10) + "," + strconv.FormatInt(fill.FillIndex, 10) + "," + strconv.FormatInt(fill.CreateTime, 10) + ",'" + fill.RingHash + "','" + fill.TxHash + "','" + fill.OrderHash + "'," +
			"'" + fill.TokenS + "','" + fill.TokenB + "','" + fill.SymbolS + "','" + fill.SymbolB + "','" + fill.AmountS + "','" + fill.AmountB + "','" + fill.LrcFee + "','" + fill.Market + "','" + fill.Side + "','" + fill.Miner + "','" + fill.WalletAddress + "'," + strconv.FormatFloat(fill.LrcCal, 'f', 10, 64) + ", " +
			strconv.FormatFloat(fill.TokenAmountCal, 'f', 10, 64) + ",'" + fill.OrderType + "', " + strconv.FormatFloat(fill.AmountBCal, 'f', 10, 64) + ")"
		if i != len(fills)-1 {
			sql += ","
			i++
		}
	}
	s.Db.Exec(sql)
}
