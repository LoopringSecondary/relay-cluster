package dao

import (
	"github.com/Loopring/relay-lib/types"
)

type TokenTicker struct {
	TokenId           int64   `gorm:"column:token_id"`
	TokenName         string  `gorm:"column:token_name;type:varchar(60)"`
	Symbol            string  `gorm:"column:symbol;type:varchar(40)"`
	WebsiteSlug       string  `gorm:"column:website_slug;type:varchar(60)"`
	Market            string  `gorm:"column:market"`
	CmcRank           int64   `gorm:"column:cmc_rank"`
	CirculatingSupply float64 `gorm:"column:circulating_supply"`
	TotalSupply       float64 `gorm:"column:total_supply"`
	MaxSupply         float64 `gorm:"column:max_supply"`
	Price             float64 `gorm:"column:price"`
	Volume24H         float64 `gorm:"column:volume_24h"`
	MarketCap         float64 `gorm:"column:market_cap"`
	PercentChange1H   float64 `gorm:"column:percent_change_1h"`
	PercentChange24H  float64 `gorm:"column:percent_change_24h"`
	PercentChange7D   float64 `gorm:"column:percent_change_7d"`
	LastUpdated       int64   `gorm:"column:last_updated"`
}

//根据market来查询对应的token's tickers
func (s *RdsService) GetTokenTickerByMarket(market string) ([]TokenTicker, error) {
	var tickers []TokenTicker
	err := s.Db.Where("market=?", market).Find(&tickers).Error
	return tickers, err
}

func (t *TokenTicker) ConvertUp(ticker *types.CMCTicker) error {

	ticker.TokenId = t.TokenId
	ticker.TokenName = t.TokenName
	ticker.Symbol = t.Symbol
	ticker.WebsiteSlug = t.WebsiteSlug
	ticker.Market = t.Market
	ticker.CmcRank = t.CmcRank
	ticker.CirculatingSupply = t.CirculatingSupply
	ticker.TotalSupply = t.TotalSupply
	ticker.MaxSupply = t.MaxSupply
	ticker.Price = t.Price
	ticker.Volume24H = t.Volume24H
	ticker.MarketCap = t.MarketCap
	ticker.PercentChange1H = t.PercentChange1H
	ticker.PercentChange24H = t.PercentChange24H
	ticker.PercentChange7D = t.PercentChange7D
	ticker.LastUpdated = t.LastUpdated
	return nil
}
