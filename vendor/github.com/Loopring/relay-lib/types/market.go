package types

type Market struct {
	ListType    string   `json:"ListType"`
	MarketPairs []string `json:"MarketPairs"`
}

type MarketDecimal struct {
	Market   string `json:"Market"`
	Decimals int    `json:"Decimals"`
}
