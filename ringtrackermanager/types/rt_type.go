package types

type AmountResp struct {
	Trades int64 `gorm:"column:trades;type:int" json:"trades"`
	Tokens int64 `gorm:"column:tokens;type:int" json:"tokens"`
	Relays int64 `gorm:"column:relays;type:int" json:"relays"`
	Dexs   int64 `gorm:"column:dexs;type:int" json:"dexs"`
	Rings  int64 `gorm:"column:rings;type:int" json:"rings"`
}

type TrendReq struct {
	Duration string `json:"duration"`
	Type     string `json:"type"`
	Currency string `json:"currency"`
	Keyword  string `json:"keyword"`
	Len      int    `json:"len"`
}

type EcoTrendReq struct {
	Duration  string `json:"duration"`
	Type      string `json:"type"`
	Indicator string `json:"indicator"`
	Currency  string `json:"currency"`
}

type TrendRsp struct {
	Trends      []TrendFill `json:"trends"`
	TotalFee    float64     `json:"totalFee"`
	TotalVolume float64     `json:"totalVolume"`
	TotalTrade  int64       `json:"totalTrade"`
}

type EcoTrendRsp struct {
	Type      string         `json:"type"`
	Indicator []IndicatorRsp `json:"indicator"`
}

type IndicatorRsp struct {
	Name string             `json:"name"`
	Data []IndicatorDataRsp `json:"data"`
}

type IndicatorDataRsp struct {
	Name  string  `json:"name"`
	Rate  float64 `json:"rate"`
	Value float64 `json:"value"`
}

type QueryReq struct {
	PageIndex       int    `json:"pageIndex"`
	PageSize        int    `json:"pageSize"`
	Currency        string `json:"currency"`
	TrendType       string `json:"trendType"`
	Keyword         string `json:"keyword"`
	Search          string `json:"search"`
	Sort            string `json:"sort"`
	Relay           string `json:"relay"`
	DelegateAddress string `json:"delegateAddress"`
	RingIndex       int64  `json:"ringIndex"`
	FillIndex       int64  `json:"fillIndex"`
}

type EcoStatFill struct {
	Name  string  `gorm:"column:name;type:varchar(42)"`
	Value float64 `gorm:"column:value;type:float"`
}

type TrendFill struct {
	Fee    float64 `json:"fee"`
	Volume float64 `json:"volume"`
	Trade  int64   `json:"trade"`
	Date   int64   `json:"date"`
}

type TokenFill struct {
	Token       string  `gorm:"column:token;type:varchar(42)" json:"token"`
	Symbol      string  `gorm:"column:symbol;type:varchar(42)" json:"symbol"`
	LastPrice   float64 `json:"lastPrice"`
	Trade       int64   `gorm:"column:trade;type:int" json:"trade"`
	Volume      float64 `gorm:"column:volume;type:float" json:"volume"`
	TokenVolume float64 `gorm:"column:token_volume;type:float" json:"tokenVolume"`
	Fee         float64 `gorm:"column:fee;type:float" json:"fee"`
}

type RelayFill struct {
	Relay   string  `gorm:"column:relay;type:varchar(42)" json:"relay"`
	Website string  `gorm:"column:website;type:text" json:"website"`
	Trade   int64   `gorm:"column:trade;type:int" json:"trade"`
	Volume  float64 `gorm:"column:volume;type:float" json:"volume"`
	Fee     float64 `gorm:"column:fee;type:float" json:"fee"`
}

type DexFill struct {
	Dex     string  `gorm:"column:dex;type:varchar(42)" json:"dex"`
	Website string  `gorm:"column:website;type:text" json:"website"`
	Trade   int64   `gorm:"column:trade;type:int" json:"trade"`
	Volume  float64 `gorm:"column:volume;type:float" json:"volume"`
	Fee     float64 `gorm:"column:fee;type:float" json:"fee"`
}

type RingFill struct {
	Ring   string  `gorm:"column:ring;type:varchar(42)" json:"ring"`
	Trade  int64   `gorm:"column:trade;type:int" json:"trade"`
	Volume float64 `gorm:"column:volume;type:float" json:"volume"`
	Fee    float64 `gorm:"column:fee;type:float" json:"fee"`
}
