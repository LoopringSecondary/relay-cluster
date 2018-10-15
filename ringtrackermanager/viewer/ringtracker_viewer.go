package viewer

import (
	"github.com/Loopring/extractor/extractor"
	"github.com/Loopring/relay-cluster/dao"
	cache2 "github.com/Loopring/relay-cluster/ringtrackermanager/cache"
	"github.com/Loopring/relay-cluster/ringtrackermanager/types"
	"github.com/Loopring/relay-lib/cache"
	"github.com/Loopring/relay-lib/eth/accessor"
	"github.com/Loopring/relay-lib/eth/contract"
	"github.com/Loopring/relay-lib/eth/loopringaccessor"
	types2 "github.com/Loopring/relay-lib/eth/types"
	"github.com/Loopring/relay-lib/log"
	"github.com/Loopring/relay-lib/marketcap"
	"github.com/Loopring/relay-lib/marketutil"
	"github.com/Loopring/relay-lib/sns"
	types3 "github.com/Loopring/relay-lib/types"
	"github.com/Loopring/relay-lib/zklock"
	"github.com/bitly/go-simplejson"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/pkg/errors"
	"github.com/robfig/cron"
	"io/ioutil"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	START_TIME                  = int64(1524096000)
	TIME_INTERVAL               = 86400
	RING_TRACKER_PREFIX         = "ringtracker_"
	RING_TRACKER_DATA           = "data_"
	TOKEN_PRICE                 = "tokenprice_"
	RING_TRACKER_CronJob_ZkLock = "ringTrackerZkLock"
	RingTrackerLockFailedMsg    = "ringTracker viewer try lock failed"
)

type RingTrackerViewer interface {
	GetAmount() types.AmountResp
	SetEthTokenPrice()
	SetTokenPrices()
	SetFullFills()
	GetTrend(duration types.Duration, trendType types.TrendType, keyword string, len int, currency types.Currency) types.TrendRsp
	GetEcosystemTrend(duration types.Duration, trendType types.TrendType, indicator types.Indicator, currency types.Currency) []types.EcoTrendRsp
	GetTrades(currency types.Currency, trendType types.TrendType, keyword, search string, pageIndex, pageSize int) dao.PageResult
	GetTradeDetails(delegateAddress string, ringIndex, fillIndex int64) []dao.FullFillEvent
	GetAllTokens(currency types.Currency, sort types.Indicator, pageIndex, pageSize int) dao.PageResult
	GetAllRelays(currency types.Currency, sort types.Indicator, pageIndex, pageSize int) dao.PageResult
	GetAllDexs(currency types.Currency, sort types.Indicator, pageIndex, pageSize int) dao.PageResult
	GetTokensByRelay(currency types.Currency, relayer string) ([]types.TokenFill, error)
}

type RingTrackerViewerImpl struct {
	rds  *dao.RdsService
	mc   marketcap.MarketCapProvider
	cron *cron.Cron
}

func NewRingTrackerViewer(rds *dao.RdsService, mc marketcap.MarketCapProvider) *RingTrackerViewerImpl {
	var viewer RingTrackerViewerImpl
	viewer.rds = rds
	viewer.mc = mc
	viewer.cron = cron.New()
	go func() {
		if zklock.TryLock(RING_TRACKER_CronJob_ZkLock) == nil {
			viewer.cron.AddFunc("0 0 0 * * *", viewer.SetEthTokenPrice)
			viewer.cron.AddFunc("0 */30 * * * *", viewer.SetTokenPrices)
			log.Info("start ringTracker viewer cron jobs......... ")
			viewer.cron.Start()
			//go viewer.SetFullFills()
			//viewer.SetEthTokenPrice()
			ClearRingTrackerCache()
			viewer.SetTokenPrices()
		} else {
			err := sns.PublishSns(RingTrackerLockFailedMsg, RingTrackerLockFailedMsg)
			if err != nil {
				log.Error(err.Error())
			}
		}
	}()
	return &viewer
}

func (r *RingTrackerViewerImpl) GetAmount() types.AmountResp {
	return r.rds.GetAmount()
}

func (r *RingTrackerViewerImpl) SetEthTokenPrice() {
	log.Debugf("[SetEthTokenPrice] start...")
	currentTime := START_TIME
	if r.rds.CountPriceTrend() != 0 {
		currentTime = r.rds.GetPriceTrendMaxTime() + TIME_INTERVAL
	}
	for currentTime <= time.Now().Unix() {
		res := GetHisCryptoPrice("ETH", "ETH,CNY,USDT", currentTime)
		r.rds.Add(&dao.TokenPriceTrend{CoinName: "ETH", CoinAmount: res["ETH"], Time: currentTime})
		r.rds.Add(&dao.TokenPriceTrend{CoinName: "CNY", CoinAmount: res["CNY"], Time: currentTime})
		r.rds.Add(&dao.TokenPriceTrend{CoinName: "USDT", CoinAmount: res["USDT"], Time: currentTime})
		currentTime += TIME_INTERVAL
	}
}

func (r *RingTrackerViewerImpl) SetTokenPrices() {
	log.Debugf("[SetTokenPrices] start...")
	tokens := r.rds.GetTokenSymbols()
	for _, legal := range []string{"ETH", "CNY", "USDT"} {
		tokenPriceMap := getCurCryptoPrice(legal, tokens)
		for k, v := range tokenPriceMap {
			cache.Set(RING_TRACKER_PREFIX+TOKEN_PRICE+k+"-"+legal, []byte(strconv.FormatFloat(v, 'f', 10, 64)), 1800)
		}
	}
}

func (r *RingTrackerViewerImpl) SetFullFills() {
	log.Debugf("[SetFullFills] start...")
	var event extractor.EventData
	GetEvent(&event)
	for _, ring := range r.rds.GetAllRings() {
		fills := r.rds.GetAllFills(ring.Miner, ring.TxHash)
		if len(fills) < 2 {
			fills = GetFills(&ring, &event)
			if len(fills) == 0 {
				r.rds.Add(&dao.FailFill{TxHash: ring.TxHash})
				continue
			}
		}
		FormatFullFills(&ring, fills)
		if len(fills) > 0 {
			ClearRingTrackerCache()
		}
		r.rds.AddFullFills(fills)
	}
}

func ClearRingTrackerCache() {
	keys, _ := cache.Keys(RING_TRACKER_PREFIX + RING_TRACKER_DATA + "*")
	for _, key := range keys {
		log.Debugf("[clear cache] key: " + string(key))
		cache.Del(string(key))
	}
}

func (r *RingTrackerViewerImpl) GetTrend(duration types.Duration, trendType types.TrendType, keyword string, l int, currency types.Currency) types.TrendRsp {
	if l > 50 || l <= 0 {
		l = 50
	}
	if duration == types.DURATION_1M && (l > 12 || l <= 0) {
		l = 12
	}
	res := types.TrendRsp{TotalVolume: 0, TotalFee: 0, TotalTrade: 0}
	sql, start, end := getDuration(duration, l)
	trendMap := make(map[int64]types.TrendFill)
	for _, trend := range r.rds.GetTrend(duration, trendType, keyword, currency, sql, start.Unix()) {
		res.TotalVolume += trend.Volume
		res.TotalTrade += trend.Trade
		res.TotalFee += trend.Fee
		trendMap[trend.Date] = trend
	}

	tmp := make([]types.TrendFill, 0)
	for l > 0 {
		if _, ok := trendMap[end.Unix()]; !ok {
			tmp = append(tmp, types.TrendFill{Fee: 0, Volume: 0, Trade: 0, Date: end.Unix()})
		} else {
			tmp = append(tmp, trendMap[end.Unix()])
		}
		end = subInterval(duration, end)
		l--
	}
	res.Trends = tmp
	return res
}

func (r *RingTrackerViewerImpl) GetEcosystemTrend(duration types.Duration, trendType types.TrendType, indicator types.Indicator, currency types.Currency) []types.EcoTrendRsp {
	res := make([]types.EcoTrendRsp, 0)
	key := RING_TRACKER_PREFIX + RING_TRACKER_DATA + "ecosystemtrend_" + strings.Join([]string{"duration:" + string(duration), "trendType:" + string(trendType), "indicator:" + string(indicator), "currency:" + string(currency)}, ",")
	if cache2.GetFromCache(key, &res) {
		return res
	}

	var (
		trends     = []types.TrendType{trendType}
		indicators = []types.Indicator{indicator}
	)
	if trendType == types.ALL_TREND {
		trends = []types.TrendType{types.TOKEN, types.RELAY, types.DEX}
	}
	indicators = []types.Indicator{indicator}
	if indicator == types.ALL_INDICATOR {
		indicators = []types.Indicator{types.VOLUME, types.FEE, types.TRADE}
	}

	startTime := time.Now().Unix() - 365*24*3600
	switch duration {
	case types.DURATION_24H:
		startTime = time.Now().Unix() - 24*3600
	case types.DURATION_7D:
		startTime = time.Now().Unix() - 7*24*3600
	case types.DURATION_1M:
		startTime = time.Now().Unix() - 30*24*3600
	}

	for _, tr := range trends {
		indicatorRsps := make([]types.IndicatorRsp, 0)
		for _, in := range indicators {
			indicatorData := make([]types.IndicatorDataRsp, 0)
			fills := r.rds.GetEcoStatFill(currency, tr, in, startTime)
			total := 0.0
			for _, fill := range fills {
				total += fill.Value
			}
			for _, fill := range fills {
				indicatorData = append(indicatorData, types.IndicatorDataRsp{Name: fill.Name, Rate: fill.Value / total, Value: fill.Value})
			}
			indicatorRsps = append(indicatorRsps, types.IndicatorRsp{Name: types.IndicatorToStr(in), Data: indicatorData})
		}
		res = append(res, types.EcoTrendRsp{Type: types.TrendTypeToStr(tr), Indicator: indicatorRsps})
	}

	cache2.SaveCache(key, res)
	return res
}

func (r *RingTrackerViewerImpl) GetTrades(currency types.Currency, trendType types.TrendType, keyword, search string, pageIndex, pageSize int) (res dao.PageResult) {
	data := make([]interface{}, 0)
	if pageIndex <= 0 {
		pageIndex = 1
	}
	if pageSize == 0 {
		pageSize = 10
	}

	key := RING_TRACKER_PREFIX + RING_TRACKER_DATA + "trades_" + strings.Join([]string{"trendType:" + string(trendType), "currency:" + string(currency), "keyword:" + keyword, "page:" + strconv.Itoa(pageIndex), "size:" + strconv.Itoa(pageSize)}, ",")
	if len(search) == 0 && cache2.GetFromCache(key, &res) {
		return res
	}
	res = dao.PageResult{PageIndex: pageIndex, PageSize: pageSize, Total: r.rds.CountFullFills(trendType, keyword, search)}
	for _, fill := range r.rds.GetAllFullFills(currency, trendType, keyword, search, pageIndex, pageSize) {
		data = append(data, fill)
	}
	res.Data = data

	if len(search) == 0 {
		cache2.SaveCache(key, res)
	}
	return res
}

func (r *RingTrackerViewerImpl) GetTradeDetails(delegateAddress string, ringIndex, fillIndex int64) (res []dao.FullFillEvent) {
	key := RING_TRACKER_PREFIX + RING_TRACKER_DATA + "tradedetails_" + strings.Join([]string{"delegate:" + delegateAddress, "ringIndex:" + strconv.FormatInt(ringIndex, 10), "fillIndex:" + strconv.FormatInt(fillIndex, 10)}, ",")
	if cache2.GetFromCache(key, &res) {
		return res
	}
	res = r.rds.GetTradeDetails(delegateAddress, ringIndex, fillIndex)
	cache2.SaveCache(key, res)
	return res
}

func (r *RingTrackerViewerImpl) GetAllTokens(currency types.Currency, sort types.Indicator, pageIndex, pageSize int) (res dao.PageResult) {
	if pageIndex <= 0 {
		pageIndex = 1
	}
	if pageSize == 0 {
		pageSize = 10
	}
	tokens := r.rds.GetTokenSymbols()
	res = dao.PageResult{Total: len(tokens), PageIndex: pageIndex, PageSize: pageSize}
	data := make([]interface{}, 0)
	fills := r.rds.GetFillsByToken(currency, sort, pageIndex, pageSize)
	for _, fill := range fills {
		if lastPrice, err := cache.Get(RING_TRACKER_PREFIX + TOKEN_PRICE + fill.Symbol + "-" + string(currency)); err == nil {
			price, _ := strconv.ParseFloat(string(lastPrice), 64)
			fill.LastPrice = price
		}
		data = append(data, fill)
	}
	res.Data = data
	return res
}

func (r *RingTrackerViewerImpl) GetAllRelays(currency types.Currency, sort types.Indicator, pageIndex, pageSize int) (res dao.PageResult) {
	if pageIndex <= 0 {
		pageIndex = 1
	}
	if pageSize == 0 {
		pageSize = 10
	}
	key := RING_TRACKER_PREFIX + RING_TRACKER_DATA + "relays_" + strings.Join([]string{"currency:" + string(currency), "sort:" + string(sort), "page:" + strconv.Itoa(pageIndex), "size:" + strconv.Itoa(pageSize)}, ",")
	if cache2.GetFromCache(key, &res) {
		return res
	}
	fills := r.rds.GetFillsByRelay(currency, sort, pageIndex, pageSize)
	res = dao.PageResult{Total: r.rds.CountRelays(), PageIndex: pageIndex, PageSize: pageSize}
	data := make([]interface{}, 0)
	for _, fill := range fills {
		data = append(data, fill)
	}
	res.Data = data
	cache2.SaveCache(key, res)
	return res
}

func (r *RingTrackerViewerImpl) GetAllDexs(currency types.Currency, sort types.Indicator, pageIndex, pageSize int) (res dao.PageResult) {
	if pageIndex <= 0 {
		pageIndex = 1
	}
	if pageSize == 0 {
		pageSize = 10
	}
	key := RING_TRACKER_PREFIX + RING_TRACKER_DATA + "dexs_" + strings.Join([]string{"currency:" + string(currency), "sort:" + string(sort), "page:" + strconv.Itoa(pageIndex), "size:" + strconv.Itoa(pageSize)}, ",")
	if cache2.GetFromCache(key, &res) {
		return res
	}
	fills := r.rds.GetFillsByDex(currency, sort, pageIndex, pageSize)
	res = dao.PageResult{Total: r.rds.CountDexs(), PageIndex: pageIndex, PageSize: pageSize}
	data := make([]interface{}, 0)
	for _, fill := range fills {
		data = append(data, fill)
	}
	res.Data = data
	cache2.SaveCache(key, res)
	return res
}

func (r *RingTrackerViewerImpl) GetTokensByRelay(currency types.Currency, relay string) ([]types.TokenFill, error) {
	if len(relay) == 0 {
		return nil, errors.New("relay can't be empty")
	}
	res := make([]types.TokenFill, 0)
	key := RING_TRACKER_PREFIX + RING_TRACKER_DATA + "tokensbyrelay_" + strings.Join([]string{"currency:" + string(currency), "relay:" + relay}, ",")
	if cache2.GetFromCache(key, &res) {
		return res, nil
	}
	res = r.rds.GetTokensByRelay(currency, relay)
	cache2.SaveCache(key, res)
	return res, nil
}

func GetHisCryptoPrice(srcSymbol string, desSymbols string, currentTime int64) map[string]float64 {
	if strings.ToUpper(srcSymbol) == "WETH" {
		srcSymbol = "ETH"
	}
	url := "https://min-api.cryptocompare.com/data/pricehistorical?fsym=" + srcSymbol + "&tsyms=" + desSymbols + "&ts=" + strconv.FormatInt(currentTime, 10)
	log.Debugf(url)
	resp, _ := http.Get(url)
	body, _ := ioutil.ReadAll(resp.Body)
	js, _ := simplejson.NewJson(body)
	res := make(map[string]float64)
	for _, desToken := range strings.Split(desSymbols, ",") {
		res[desToken] = js.Get(srcSymbol).Get(desToken).MustFloat64()
	}
	return res
}

func getCurCryptoPrice(srcSymbol string, desSymbols []string) map[string]float64 {
	if srcSymbol == "WETH" {
		srcSymbol = "ETH"
	}
	for i := 0; i < len(desSymbols); i++ {
		if desSymbols[i] == "WETH" {
			desSymbols[i] = "ETH"
			break
		}
	}
	url := "https://min-api.cryptocompare.com/data/price?fsym=" + srcSymbol + "&tsyms=" + strings.Join(desSymbols, ",")
	log.Debugf(url)
	resp, _ := http.Get(url)
	body, _ := ioutil.ReadAll(resp.Body)
	js, _ := simplejson.NewJson(body)
	res := make(map[string]float64)
	for _, desToken := range desSymbols {
		price := js.Get(desToken).MustFloat64()
		if desToken == "ETH" {
			desToken = "WETH"
		}
		if price != 0 {
			res[desToken] = 1.0 / price
		} else {
			res[desToken] = 0
		}
	}
	return res
}

func getDuration(duration types.Duration, len int) (sql string, start time.Time, end time.Time) {
	now := time.Now()
	switch duration {
	case types.DURATION_1H:
		mm, _ := time.ParseDuration("-1" + strconv.Itoa(len) + "h")
		end, _ = time.ParseInLocation("2006-01-02 15", now.Format("2006-01-02 15"), time.UTC)
		return "UNIX_TIMESTAMP(date_format(FROM_UNIXTIME(create_time), '%Y-%m-%d %H'))", end.Add(mm), end
	case types.DURATION_7D:
		end, _ := time.ParseInLocation("2006-01-02", now.AddDate(0, 0, -1*(int(time.Now().Weekday())-1)).Format("2006-01-02"), time.UTC)
		return "UNIX_TIMESTAMP(str_to_date(concat(yearweek(FROM_UNIXTIME(create_time)), 'Monday'), '%X%V %W'))", end.AddDate(0, 0, -7*len), end
	case types.DURATION_1M:
		currentYear, currentMonth, _ := now.Date()
		end = time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, time.UTC)
		return "UNIX_TIMESTAMP(date_format(FROM_UNIXTIME(create_time), '%Y-%m-01'))", end.AddDate(0, -1*len, 0), end
	case types.DURATION_24H:
		end, _ = time.ParseInLocation("2006-01-02", now.Format("2006-01-02"), time.UTC)
		return "UNIX_TIMESTAMP(date_format(FROM_UNIXTIME(create_time), '%Y-%m-%d'))", end.AddDate(0, 0, -1*len), end
	}
	return
}

func subInterval(duration types.Duration, currentTime time.Time) time.Time {
	var res time.Time
	switch duration {
	case types.DURATION_1H:
		h, _ := time.ParseDuration("-1h")
		res = currentTime.Add(h)
	case types.DURATION_24H:
		res = currentTime.AddDate(0, 0, -1)
	case types.DURATION_7D:
		res = currentTime.AddDate(0, 0, -7)
	case types.DURATION_1M:
		res = currentTime.AddDate(0, -1, 0)
	}
	return res
}

func FormatFullFills(ring *dao.RingMinedEvent, fills []*dao.FullFillEvent) {
	lrcPrice := GetHisCryptoPrice("LRC", "ETH", ring.Time)["ETH"]
	for _, fill := range fills {
		fill.SymbolS = marketutil.AddressToAlias(fill.TokenS)
		fill.SymbolB = marketutil.AddressToAlias(fill.TokenB)
		fill.AmountBCal, _ = marketutil.StringToFloat(fill.TokenB, fill.AmountB)
		token, amount := fill.TokenB, fill.AmountB
		if !marketutil.IsSupportedMarket(fill.SymbolB) {
			token, amount = fill.TokenS, fill.AmountS
		}
		tokenFloat, _ := marketutil.StringToFloat(token, amount)
		lrc, _ := marketutil.StringToFloat(marketutil.AliasToAddress("LRC").Hex(), fill.LrcFee)
		fill.LrcCal = lrcPrice * lrc
		switch marketutil.AddressToAlias(token) {
		case "WETH":
			fill.TokenAmountCal = tokenFloat
		case "LRC":
			fill.TokenAmountCal = lrcPrice * tokenFloat
		default:
			fill.TokenAmountCal = GetHisCryptoPrice(marketutil.AddressToAlias(token), "ETH", fill.CreateTime)["ETH"] * tokenFloat
		}
	}
}

func GetFills(ring *dao.RingMinedEvent, event *extractor.EventData) []*dao.FullFillEvent {
	var (
		tx        types2.Transaction
		recipient types2.TransactionReceipt
		block     types2.Block
		list      []*dao.FullFillEvent
	)
	retry := 10
	for i := 0; i < retry; i++ {
		if err := accessor.GetTransactionReceipt(&recipient, ring.TxHash, strconv.FormatInt(ring.BlockNumber, 10)); err != nil {
			log.Errorf("retry to get transaction recipient, retry count:%d", i+1)
		} else {
			break
		}
	}
	for i := 0; i < retry; i++ {
		if err := accessor.GetTransactionByHash(&tx, ring.TxHash, strconv.FormatInt(ring.BlockNumber, 10)); err != nil {
			log.Errorf("retry to get transaction, retry count:%d", i+1)
		} else {
			break
		}
	}

	if len(recipient.Logs) < 1 {
		return list
	}
	var (
		evtLog        types2.Log
		decodedValues [][]byte
	)

	for _, v := range recipient.Logs {
		if common.HexToHash(v.Topics[0]) == event.Id {
			evtLog = v
		}
	}

	if len(evtLog.Data) == 0 {
		return list
	}
	data := hexutil.MustDecode(evtLog.Data)
	for _, topic := range evtLog.Topics {
		decodeBytes := hexutil.MustDecode(topic)
		decodedValues = append(decodedValues, decodeBytes)
	}
	event.Abi.UnpackEvent(event.Event, event.Name, data, decodedValues)
	src := event.Event.(*contract.RingMinedEvent)
	_, fills, err := src.ConvertDown()
	if err != nil {
		log.Errorf(err.Error())
		return list
	}

	submitRingData := hexutil.MustDecode("0x" + tx.Input[10:])
	var submitRing contract.SubmitRingMethodInputs
	submitRing.Protocol = common.HexToAddress(ring.Protocol)
	event.Abi.UnpackMethod(&submitRing, "submitRing", submitRingData)
	submitRingEvent, err := submitRing.ConvertDown(submitRing.Protocol, common.HexToAddress(ring.DelegateAddress))
	if err != nil {
		log.Errorf("%+v", err)
	}

	accessor.GetBlockByNumber(&block, big.NewInt(ring.BlockNumber), true)
	txinfo := setTxInfo(&tx, recipient.GasUsed.BigInt(), block.Timestamp.BigInt(), "submitRing")
	orderType := "market_order"
	for _, v := range fills {
		v.TxInfo = txinfo
		var fill dao.FullFillEvent
		fill.ConvertToFullFill(v, ring.Miner)
		fill.Side = marketutil.GetSide(fill.TokenS, fill.TokenB)
		fill.Market, _ = marketutil.WrapMarketByAddress(fill.TokenS, fill.TokenB)
		for _, order := range submitRingEvent.OrderList {
			if order.Hash.Hex() == fill.OrderHash {
				fill.WalletAddress = order.WalletAddress.Hex()
				break
			}
		}
		if fill.Owner == fill.Miner {
			orderType = "p2p_order"
		}
		list = append(list, &fill)
	}
	for _, v := range list {
		v.OrderType = orderType
	}
	return list
}

func GetEvent(event *extractor.EventData) {
	for name, e := range loopringaccessor.ProtocolImplAbi().Events {
		switch name {
		case contract.EVENT_RING_MINED:
			event.Id = e.Id()
			event.Name = e.Name
			event.Abi = loopringaccessor.ProtocolImplAbi()
			event.Event = &contract.RingMinedEvent{}
		}
	}
	log.Debugf("extractor,contract event name:%s -> key:%s", event.Name, event.Id.Hex())
}

func setTxInfo(tx *types2.Transaction, gasUsed, blockTime *big.Int, methodName string) types3.TxInfo {
	var txinfo types3.TxInfo

	txinfo.Protocol = common.HexToAddress(tx.To)
	txinfo.From = common.HexToAddress(tx.From)
	txinfo.To = common.HexToAddress(tx.To)

	if impl, ok := loopringaccessor.ProtocolAddresses()[txinfo.To]; ok {
		txinfo.DelegateAddress = impl.DelegateAddress
	} else {
		txinfo.DelegateAddress = types3.NilAddress
	}

	txinfo.BlockNumber = tx.BlockNumber.BigInt()
	txinfo.BlockTime = blockTime.Int64()
	txinfo.BlockHash = common.HexToHash(tx.BlockHash)
	txinfo.TxHash = common.HexToHash(tx.Hash)
	txinfo.TxIndex = tx.TransactionIndex.Int64()
	txinfo.Value = tx.Value.BigInt()

	txinfo.GasLimit = tx.Gas.BigInt()
	txinfo.GasUsed = gasUsed
	txinfo.GasPrice = tx.GasPrice.BigInt()
	txinfo.Nonce = tx.Nonce.BigInt()

	txinfo.Identify = methodName

	return txinfo
}
