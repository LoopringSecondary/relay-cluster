package types

import "strings"

type Duration string
type TrendType string
type Currency string
type Indicator string

const (
	DURATION_1H  Duration = "1h"
	DURATION_24H Duration = "24h"
	DURATION_7D  Duration = "7d"
	DURATION_1M  Duration = "1m"
	DURATION_1Y  Duration = "1y"
)

func StrToTrendDuration(duration string) Duration {
	switch strings.ToLower(duration) {
	case "1h":
		return DURATION_1H
	case "24h":
		return DURATION_24H
	case "7d":
		return DURATION_7D
	case "1m":
		return DURATION_1M
	default:
		return DURATION_24H
	}
}
func StrToEcoTrendDuration(duration string) Duration {
	switch strings.ToLower(duration) {
	case "24h":
		return DURATION_24H
	case "7d":
		return DURATION_7D
	case "1m":
		return DURATION_1M
	case "1y":
		return DURATION_1Y
	default:
		return DURATION_1Y
	}
}

const (
	ALL_TREND TrendType = "all"
	TOKEN     TrendType = "token"
	RELAY   TrendType = "relay"
	DEX       TrendType = "dex"
	RING      TrendType = "ring"
)

func StrToTrendType(trendType string) TrendType {
	switch strings.ToLower(trendType) {
	case "", "all":
		return ALL_TREND
	case "token":
		return TOKEN
	case "relay":
		return RELAY
	case "dex":
		return DEX
	case "ring":
		return RING
	default:
		return ALL_TREND
	}
}

func TrendTypeToStr(trendType TrendType) string {
	switch trendType {
	case TOKEN:
		return "token"
	case RELAY:
		return "relay"
	case DEX:
		return "dex"
	case RING:
		return "ring"
	case ALL_TREND:
		return "all"
	default:
		return "all"
	}
}

const (
	ETH  Currency = "ETH"
	CNY  Currency = "CNY"
	USDT Currency = "USDT"
)

func StrToCurrency(trendType string) Currency {
	switch strings.ToUpper(trendType) {
	case "ETH", "WETH":
		return ETH
	case "CNY":
		return CNY
	case "USDT":
		return USDT
	default:
		return USDT
	}
}

func CurrencyToStr(currency Currency) string {
	switch currency {
	case ETH:
		return "ETH"
	case CNY:
		return "CNY"
	case USDT:
		return "USDT"
	default:
		return "USDT"
	}
}

const (
	ALL_INDICATOR Indicator = "all"
	VOLUME        Indicator = "volume"
	TRADE         Indicator = "trade"
	FEE           Indicator = "fee"
)

func StrToIndicator(indicator string) Indicator {
	switch strings.ToLower(indicator) {
	case "":
		return ALL_INDICATOR
	case "volume":
		return VOLUME
	case "trade":
		return TRADE
	case "fee":
		return FEE
	default:
		return VOLUME
	}
}

func StrToSort(sort string) Indicator {
	switch strings.ToLower(sort) {
	case "volume":
		return VOLUME
	case "trade":
		return TRADE
	case "fee":
		return FEE
	default:
		return VOLUME
	}
}

func IndicatorToStr(indicator Indicator) string {
	switch indicator {
	case ALL_INDICATOR:
		return "all"
	case VOLUME:
		return "volume"
	case TRADE:
		return "trade"
	case FEE:
		return "fee"
	default:
		return "volume"
	}
}
