package market

import (
	"sort"
)

type TickerRespWrapper struct {
	tickerResp []TickerResp
	by         func(p, q *TickerResp) bool
}

type SortBy func(p, q *TickerResp) bool

func (trw TickerRespWrapper) Len() int { // overwrite Len()
	return len(trw.tickerResp)
}

func (trw TickerRespWrapper) Swap(i, j int) { // overwrite Swap()
	trw.tickerResp[i], trw.tickerResp[j] = trw.tickerResp[j], trw.tickerResp[i]
}

func (trw TickerRespWrapper) Less(i, j int) bool { // overwrite Less()
	return trw.by(&trw.tickerResp[i], &trw.tickerResp[j])
}

func SortMarketTicker(tickerResp []TickerResp, by SortBy) {
	sort.Sort(TickerRespWrapper{tickerResp, by})
}
