package market

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Loopring/relay-cluster/dao"
	"github.com/Loopring/relay-lib/cache"
	"github.com/Loopring/relay-lib/log"
	util "github.com/Loopring/relay-lib/marketutil"
	"github.com/Loopring/relay-lib/sns"
	"github.com/Loopring/relay-lib/types"
	"github.com/Loopring/relay-lib/zklock"
	gocache "github.com/patrickmn/go-cache"
	"github.com/robfig/cron"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	USD                         = "USD"
	CNY                         = "CNY"
	ETH                         = "ETH"
	LRC                         = "LRC"
	USDT                        = "USDT"
	TUSD                        = "TUSD"
	WETH                        = "WETH"
	TICKER_SOURCE_COINMARKETCAP = "coinmarketcap"

	MARKET_OF_WHITELIST = "whitelist"
	MARKET_OF_BLACKLIST = "blacklist"
	MARKET_OF_HIDELIST  = "hidelist"

	VOL_RANK100_MODE = "rank"

	SPLIT_MARK                = "-"
	marketTickerCachePreKey   = "COINMARKETCAP_TICKER_NEW_"
	marketTickerLocalCacheKey = "COINMARKETCAP_TICKER_LOCAL"
	CUSTOM_TOKENS_MARKETCAP   = "custom_tokens_marketcap_new_"
	allCustomTokens           = "ALLCT"

	tickerManagerCronJobZkLock = "tickerManagerZkLock"
	tickerManagerLockFailedMsg = "ticker manager try lock failed"
)

var wethMarkets = make(map[string]string)
var lrcMarkets = make(map[string]string)
var usdtMarkets = make(map[string]string)
var tusdMarkets = make(map[string]string)
var displayMarkets = make(map[string]string)
var blacklistMarkets = make(map[string]string)

var marketConvert = []string{USD, CNY, ETH, LRC, USDT, TUSD}

type TickerResp struct {
	Market    string  `json:"market"`
	Exchange  string  `json:"exchange"`
	Intervals string  `json:"interval"`
	Amount    float64 `json:"amount"`
	Vol       float64 `json:"vol"`
	Open      float64 `json:"open"`
	Close     float64 `json:"close"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
	Last      float64 `json:"last"`
	Buy       float64 `json:"buy"`
	Sell      float64 `json:"sell"`
	Change    string  `json:"change"`
	Label     string  `json:"label"`
	Decimals  int     `json:"decimals"`
}

type TickerUpdateMsg struct {
	TickerSource string `json:"tickerSource"`
}

type TickerManager interface {
	GetTickerBySource(tickerSource string) ([]TickerResp, error)
	getCMCMarketTicker() ([]TickerResp, error)
	GetTickerByMarket(market string) (Ticker, error)
	Start()
}

type GetTickerImpl struct {
	trendManager TrendManager
	cron         *cron.Cron
	localCache   *gocache.Cache
	rds          *dao.RdsService
}

func NewTickManager(rds *dao.RdsService, trendManager TrendManager) *GetTickerImpl {
	rst := &GetTickerImpl{trendManager: trendManager, cron: cron.New(), localCache: gocache.New(10*time.Second, 10*time.Minute), rds: rds}
	return rst
}

func (c *GetTickerImpl) Start() {
	go func() {
		refreshMarkets()
		c.updateTokenTickerCache()
		c.cron.AddFunc("1 0/10 * * * *", refreshMarkets)
		if zklock.TryLock(tickerManagerCronJobZkLock) == nil {
			c.cron.AddFunc("@every 10m", c.updateTokenTickerCache)
			log.Info("start ticker manager cron jobs......... ")
			c.cron.Start()
		} else {
			err := sns.PublishSns(tickerManagerLockFailedMsg, tickerManagerLockFailedMsg)
			if err != nil {
				log.Error(err.Error())
			}
		}
	}()
}

func refreshMarkets() {
	for _, mkt := range util.AllMarkets {
		market := strings.Split(mkt, SPLIT_MARK)
		if market[1] == WETH {
			wethMarkets[util.AliasToSource(market[0])] = mkt
		} else if market[1] == LRC {
			lrcMarkets[util.AliasToSource(market[0])] = mkt
		} else if market[1] == USDT {
			usdtMarkets[util.AliasToSource(market[0])] = mkt
		} else if market[1] == TUSD {
			tusdMarkets[util.AliasToSource(market[0])] = mkt
		}
	}

	for _, dpmkt := range util.DisplayMarkets {
		if MARKET_OF_WHITELIST == dpmkt.ListType {
			for _, market := range dpmkt.MarketPairs {
				displayMarkets[market] = dpmkt.ListType
			}
		} else if MARKET_OF_BLACKLIST == dpmkt.ListType {
			for _, market := range dpmkt.MarketPairs {
				blacklistMarkets[market] = dpmkt.ListType
			}
		}
	}
}

func (c *GetTickerImpl) updateTokenTickerCache() {
	for _, v := range marketConvert {
		err := c.syncTokenTickerFromDB(v)
		if err != nil {
			log.Errorf("update token ticker cache of "+v+"err:%s", err.Error())
		}

	}
}

func (c *GetTickerImpl) syncTokenTickerFromDB(market string) error {
	customTokens, _ := util.GetCustomTokensFromRedis(allCustomTokens)
	if tickerlist, err := c.rds.GetTokenTickerByMarket(market); err == nil || len(tickerlist) > 0 {
		tickerData := [][]byte{}
		customTokensQuote := [][]byte{}
		for _, v := range tickerlist {
			var cmcTicker types.CMCTicker
			v.ConvertUp(&cmcTicker)
			fmt.Println(cmcTicker.Price)
			if data, err2 := json.Marshal(cmcTicker); nil != err2 {
				log.Errorf("err:%s", err2.Error())
				return err2
			} else {
				tickerData = append(tickerData, []byte(cmcTicker.WebsiteSlug), data)
				if _, exists := customTokens[cmcTicker.Symbol]; exists {
					customTokensQuote = append(customTokensQuote, []byte(cmcTicker.Symbol), data)
				}
			}
		}

		// hmset tickerData in redis
		if len(tickerData) > 0 {
			err := cache.HMSet(marketTickerCachePreKey+market, int64(43200), tickerData...)
			if nil != err {
				log.Errorf("hmset market ticker err:%s", err.Error())
				return err
			}
		}

		//batch set customer's tokens quoteData
		if len(customTokensQuote) > 0 {
			err := cache.HMSet(CUSTOM_TOKENS_MARKETCAP+market, int64(43200), customTokensQuote...)
			if nil != err {
				log.Errorf("get custom tokens priceQuto err:%s", err.Error())
				return err
			}
		}
	}
	return nil
}

func (c *GetTickerImpl) GetTickerBySource(tickerSource string, mode string) (tickerResp []TickerResp, err error) {
	//select by tickerSource
	marketTicker := make([]TickerResp, 0)
	if tickerSource == TICKER_SOURCE_COINMARKETCAP {
		marketTicker, err = c.getCMCMarketTicker()
	} else {
		var tickers []Ticker
		tickers, err = c.trendManager.GetTicker()
		marketTicker = getDefaultTicker(tickers)
	}

	//select by mode
	mkts := make([]TickerResp, 0)
	if VOL_RANK100_MODE == mode {
		mkts = RankMode(marketTicker)
	} else {
		mkts = DefaultMode(marketTicker)
	}

	// sort by market
	tickerMap := make(map[string]TickerResp)
	markets := make([]string, 0)
	for _, t := range mkts {
		tickerMap[t.Market] = t
		markets = append(markets, t.Market)
	}
	sort.Strings(markets)
	tickerResp = make([]TickerResp, 0)
	for _, m := range markets {
		tickerResp = append(tickerResp, tickerMap[m])
	}

	return tickerResp, err
}

func (c *GetTickerImpl) GetTickerByMarket(market string) (Ticker, error) {
	ticker := Ticker{}
	if "" == market {
		return ticker, errors.New("market is not null.")
	}

	localData, ok := c.localCache.Get(marketTickerLocalCacheKey)
	if ok {
		tickerResp := localData.([]TickerResp)
		for _, v := range tickerResp {
			if market == v.Market {
				ticker.Market = market
				ticker.Vol = v.Vol
				ticker.Last = v.Last
				ticker.Change = v.Change
				ticker.Exchange = TICKER_SOURCE_COINMARKETCAP
				return ticker, nil
			}
		}
	}

	mkt := strings.Split(market, SPLIT_MARK)[1]
	marketPairs := make(map[string]string)
	if mkt == WETH {
		mkt = ETH
		marketPairs = wethMarkets
	} else if mkt == LRC {
		marketPairs = lrcMarkets
	} else if mkt == USDT {
		marketPairs = usdtMarkets
	} else if mkt == TUSD {
		marketPairs = tusdMarkets
	}

	tickers, _ := getTickersFromRedis(marketPairs, mkt)
	if len(tickers) > 0 {
		for _, v := range tickers {
			if market == v.Market {
				ticker.Market = market
				ticker.Vol = v.Vol
				ticker.Last = v.Last
				ticker.Change = v.Change
				ticker.Exchange = TICKER_SOURCE_COINMARKETCAP
				return ticker, nil
			}
		}
	}
	return ticker, nil
}

func RankMode(tickers []TickerResp) []TickerResp {
	// slice up  by weth/lrc/usdt
	mkts := make([]TickerResp, 0)
	wethmkt := make([]TickerResp, 0)
	lrcmkt := make([]TickerResp, 0)
	usdtmkt := make([]TickerResp, 0)
	tusdmkt := make([]TickerResp, 0)
	for _, v := range tickers {
		market := strings.Split(v.Market, SPLIT_MARK)
		if market[1] == WETH {
			wethmkt = append(wethmkt, v)
		} else if market[1] == LRC {
			lrcmkt = append(lrcmkt, v)
		} else if market[1] == USDT {
			usdtmkt = append(usdtmkt, v)
		} else if market[1] == TUSD {
			tusdmkt = append(tusdmkt, v)
		}
	}
	if len(wethmkt) > 0 {
		mkts = append(mkts, rankByVol(wethmkt)...)
	}
	if len(lrcmkt) > 0 {
		mkts = append(mkts, rankByVol(lrcmkt)...)
	}
	if len(usdtmkt) > 0 {
		mkts = append(mkts, rankByVol(usdtmkt)...)
	}
	if len(tusdmkt) > 0 {
		mkts = append(mkts, rankByVol(tusdmkt)...)
	}

	return mkts
}

func rankByVol(tickers []TickerResp) []TickerResp {
	mkts := make([]TickerResp, 0)
	SortMarketTicker(tickers, func(p, q *TickerResp) bool {
		return q.Vol < p.Vol //  desc sort
	})

	for k, v := range tickers {
		if k <= 100 {
			v.Label = MARKET_OF_WHITELIST
		} else {
			v.Label = MARKET_OF_HIDELIST
		}

		//filter market in blacklist
		if _, exists := blacklistMarkets[v.Market]; !exists {
			mkts = append(mkts, v)
		}

	}
	return mkts
}

func DefaultMode(tickers []TickerResp) []TickerResp {
	mkts := make([]TickerResp, 0)
	for _, resp := range tickers {
		if listType, exists := displayMarkets[resp.Market]; exists {
			resp.Label = listType
		} else {
			resp.Label = MARKET_OF_HIDELIST
		}

		//filter market in blacklist
		if _, exists := blacklistMarkets[resp.Market]; !exists {
			mkts = append(mkts, resp)
		}
	}
	return mkts
}

func getDefaultTicker(tickers []Ticker) []TickerResp {
	tickerResp := make([]TickerResp, 0)
	if len(tickers) > 0 {
		for _, data := range tickers {
			marketTicker := TickerResp{}
			marketTicker.Market = data.Market
			marketTicker.Exchange = data.Exchange
			marketTicker.Intervals = data.Intervals
			marketTicker.Amount = data.Amount
			marketTicker.Vol = data.Vol
			marketTicker.Open = data.Open
			marketTicker.Close = data.Close
			marketTicker.High = data.High
			marketTicker.Low = data.Low
			marketTicker.Last = data.Last
			marketTicker.Buy = data.Buy
			marketTicker.Sell = data.Sell
			marketTicker.Change = data.Change
			if marketDecimal, exists := util.MarketsDecimal[data.Market]; exists {
				marketTicker.Decimals = marketDecimal.Decimals
			} else {
				marketTicker.Decimals = 8
			}

			tickerResp = append(tickerResp, marketTicker)
		}
	}
	return tickerResp
}

func (c *GetTickerImpl) getCMCMarketTicker() (tickers []TickerResp, err error) {
	localData, ok := c.localCache.Get(marketTickerLocalCacheKey)
	if ok {
		return localData.([]TickerResp), nil
	}
	tickers = make([]TickerResp, 0)
	wethTicker, _ := getTickersFromRedis(wethMarkets, ETH)
	lrcTicker, _ := getTickersFromRedis(lrcMarkets, LRC)
	usdtTicker, _ := getTickersFromRedis(usdtMarkets, USDT)
	tusdTicker, _ := getTickersFromRedis(tusdMarkets, TUSD)
	if len(wethTicker) > 0 {
		tickers = append(tickers, wethTicker...)
	}

	if len(lrcTicker) > 0 {
		tickers = append(tickers, lrcTicker...)
	}

	if len(usdtTicker) > 0 {
		tickers = append(tickers, usdtTicker...)
	}

	if len(tusdTicker) > 0 {
		tickers = append(tickers, tusdTicker...)
	}

	c.localCache.Set(marketTickerLocalCacheKey, tickers, 5*time.Second)
	return tickers, nil
}

func getTickersFromRedis(marketPairs map[string]string, market string) (tickers []TickerResp, err error) {
	tickers = make([]TickerResp, 0)
	tickerMap := make(map[string]*types.CMCTicker)
	if ticketData, err := cache.HGetAll(marketTickerCachePreKey + market); nil != err {
		log.Debug(">>>>>>>> get ticker data from redis error " + err.Error())
		return tickers, err
	} else {
		if len(ticketData) > 0 {
			idx := 0
			for idx < len(ticketData) {
				ticker := &types.CMCTicker{}
				if err := json.Unmarshal(ticketData[idx+1], ticker); nil != err {
					log.Errorf("get marketcap of ticker data err:%s", err.Error())
					return nil, err
				} else {
					tickerMap[string(ticketData[idx])] = ticker
				}
				idx = idx + 2
			}
		}
	}

	for k, v := range marketPairs {
		ticker := TickerResp{}
		ticker.Market = v
		if priceQuote, exists := tickerMap[k]; exists {
			price := priceQuote.Price
			vol := priceQuote.Volume24H
			change24H := priceQuote.PercentChange24H
			ticker.Last, _ = strconv.ParseFloat(fmt.Sprintf("%0.8f", price), 64)
			ticker.Vol, _ = strconv.ParseFloat(fmt.Sprintf("%0.8f", vol), 64)
			ticker.Change = fmt.Sprintf("%.2f%%", change24H)

		}

		if marketDecimal, exists := util.MarketsDecimal[v]; exists {
			ticker.Decimals = marketDecimal.Decimals
		} else {
			ticker.Decimals = 8
		}

		tickers = append(tickers, ticker)
	}
	return tickers, nil
}
