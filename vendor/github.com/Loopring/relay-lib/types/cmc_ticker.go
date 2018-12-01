package types

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

type CMCTicker struct {
	TokenId           int64          `json:"TokenId"`
	Address           common.Address `json:"-"`
	TokenName         string         `json:"TokenName"`
	Symbol            string         `json:"Symbol"`
	WebsiteSlug       string         `json:"WebsiteSlug"`
	Market            string         `json:"Market"`
	CmcRank           int64          `json:"CmcRank"`
	CirculatingSupply float64        `json:"CirculatingSupply"`
	TotalSupply       float64        `json:"TotalSupply"`
	MaxSupply         float64        `json:"MaxSupply"`
	Price             float64        `json:"Price"`
	Volume24H         float64        `json:"Valume24H"`
	MarketCap         float64        `json:"MarketCap"`
	PercentChange1H   float64        `json:"PercentChange1H"`
	PercentChange24H  float64        `json:"PercentChange24H"`
	PercentChange7D   float64        `json:"PercentChange7D"`
	LastUpdated       int64          `json:"LastUpdated"`
	Decimals          *big.Int
}
