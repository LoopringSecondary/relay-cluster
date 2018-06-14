package gateway

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Loopring/relay-cluster/dao"
	"github.com/Loopring/relay-cluster/market"
	txtyp "github.com/Loopring/relay-cluster/txmanager/types"
	"github.com/Loopring/relay-lib/eth/loopringaccessor"
	"github.com/Loopring/relay-lib/kafka"
	"github.com/Loopring/relay-lib/log"
	util "github.com/Loopring/relay-lib/marketutil"
	"github.com/Loopring/relay-lib/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/googollee/go-socket.io"
	"github.com/robfig/cron"
	"gopkg.in/googollee/go-engine.io.v1"
	"net"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"time"
)

const (
	EventPostfixReq         = "_req"
	EventPostfixRes         = "_res"
	EventPostfixEnd         = "_end"
	DefaultCronSpec3Second  = "0/3 * * * * *"
	DefaultCronSpec5Second  = "0/5 * * * * *"
	DefaultCronSpec10Second = "0/10 * * * * *"
	DefaultCronSpec5Minute  = "0 */5 * * * *"
	DefaultCronSpec10Hour   = "0 0 */10 * * *"
	DefaultCronSpec30Day    = "0 0 0 */30 * *"
)

const (
	emitTypeByEvent = 1
	emitTypeByCron  = 2
	priceQuoteCNY   = "CNY"
	priceQuoteUSD   = "USD"
)

const Kafka_Topic_SocketIO_Order_Transfer = "Kafka_Topic_SocketIO_Order_Transfer"
const Kafka_Topic_SocketIO_Scan_Login = "Kafka_Topic_SocketIO_Scan_Login"
const Kafka_Topic_SocketIO_Notify_Circulr = "Kafka_Topic_SocketIO_Notify_Circulr"

type Server struct {
	socketio.Server
}

type SocketIOJsonResp struct {
	Error string      `json:"error"`
	Code  string      `json:"code"`
	Data  interface{} `json:"data"`
}

func NewServer(s socketio.Server) Server {
	return Server{s}
}

func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	OriginList := r.Header["Origin"]
	Origin := ""
	if len(OriginList) > 0 {
		Origin = OriginList[0]
	}
	w.Header().Add("Access-Control-Allow-Origin", Origin)
	w.Header().Add("Access-Control-Allow-Credentials", "true")
	w.Header().Add("Access-Control-Allow-Headers", "accept, origin, content-type")
	w.Header().Add("Access-Control-Allow-Methods", "PUT,POST,GET,DELETE,OPTIONS")
	s.Server.ServeHTTP(w, r)
}

type InvokeInfo struct {
	MethodName  string
	Query       interface{}
	isBroadcast bool
	emitType    int
	spec        string
}

const (
	eventKeyTickers           = "tickers"
	eventKeyLoopringTickers   = "loopringTickers"
	eventKeyTrends            = "trends"
	eventKeyMarketCap         = "marketcap"
	eventKeyBalance           = "balance"
	eventKeyTransaction       = "transaction"
	eventKeyLatestTransaction = "latestTransaction"
	eventKeyPendingTx         = "pendingTx"
	eventKeyDepth             = "depth"
	eventKeyOrderBook         = "orderBook"
	eventKeyTrades            = "trades"
	eventKeyOrders            = "orders"
	eventKeyOrderTracing      = "orderTracing"
	eventKeyEstimatedGasPrice = "estimatedGasPrice"

	eventKeyGlobalTicker       = "globalTicker"
	eventKeyGlobalTrend        = "globalTrend"
	eventKeyGlobalMarketTicker = "globalMarketTicker"

	eventKeyOrderTransfer = "authorization"
	eventKeyScanLogin     = "addressUnlock"
	eventKeyCirculrNotify = "circulrNotify"
)

type SocketIOService interface {
	Start(port string)
	Stop()
}

type SocketIOServiceImpl struct {
	port           string
	walletService  WalletServiceImpl
	connIdMap      *sync.Map
	cron           *cron.Cron
	consumer       *kafka.ConsumerRegister
	eventTypeRoute map[string]InvokeInfo
}

type SocketMsgHandler struct {
	Data    interface{}
	Handler func(data interface{}) error
}

func NewSocketIOService(port string, walletService WalletServiceImpl, brokers []string) *SocketIOServiceImpl {
	so := &SocketIOServiceImpl{}
	so.port = port
	so.walletService = walletService
	so.connIdMap = &sync.Map{}
	so.cron = cron.New()
	so.consumer = &kafka.ConsumerRegister{}
	so.consumer.Initialize(brokers)

	var topicList = map[string]SocketMsgHandler{
		kafka.Kafka_Topic_SocketIO_Loopring_Ticker_Updated: {market.TrendUpdateMsg{}, so.broadcastLoopringTicker},
		//kafka.Kafka_Topic_SocketIO_Tickers_Updated:         {nil, so.broadcastTpTickers},
		kafka.Kafka_Topic_SocketIO_Trades_Updated: {dao.FillEvent{}, so.broadcastTrades},
		kafka.Kafka_Topic_SocketIO_Trends_Updated: {market.TrendUpdateMsg{}, so.broadcastTrends},

		kafka.Kafka_Topic_SocketIO_Order_Updated: {types.OrderState{}, so.handleOrderUpdate},
		kafka.Kafka_Topic_SocketIO_Cutoff:        {types.CutoffEvent{}, so.handleCutOff},
		kafka.Kafka_Topic_SocketIO_Cutoff_Pair:   {types.CutoffPairEvent{}, so.handleCutOffPair},

		kafka.Kafka_Topic_SocketIO_BalanceUpdated:      {types.BalanceUpdateEvent{}, so.handleBalanceUpdate},
		kafka.Kafka_Topic_SocketIO_Transaction_Updated: {txtyp.TransactionView{}, so.handleTransactionUpdate},
		Kafka_Topic_SocketIO_Order_Transfer:            {OrderTransfer{}, so.handleOrderTransfer},
		Kafka_Topic_SocketIO_Scan_Login:                {LoginInfo{}, so.handleScanLogin},
		Kafka_Topic_SocketIO_Notify_Circulr:            {NotifyCirculrBody{}, so.handleCirculrNotify},
	}

	so.eventTypeRoute = map[string]InvokeInfo{
		eventKeyTickers:           {"GetTickers", SingleMarket{}, true, emitTypeByCron, DefaultCronSpec5Second},
		eventKeyLoopringTickers:   {"GetTicker", nil, true, emitTypeByEvent, DefaultCronSpec5Second},
		eventKeyTrends:            {"GetTrend", TrendQuery{}, true, emitTypeByEvent, DefaultCronSpec10Second},
		eventKeyMarketCap:         {"GetPriceQuote", PriceQuoteQuery{}, true, emitTypeByCron, DefaultCronSpec5Minute},
		eventKeyDepth:             {"GetDepth", DepthQuery{}, true, emitTypeByEvent, DefaultCronSpec10Second},
		eventKeyOrderBook:         {"GetUnmergedOrderBook", DepthQuery{}, true, emitTypeByEvent, DefaultCronSpec10Second},
		eventKeyTrades:            {"GetLatestFills", FillQuery{}, true, emitTypeByEvent, DefaultCronSpec10Second},
		eventKeyEstimatedGasPrice: {"GetEstimateGasPrice", nil, true, emitTypeByEvent, DefaultCronSpec5Minute},

		eventKeyBalance:           {"GetBalance", CommonTokenRequest{}, false, emitTypeByEvent, DefaultCronSpec10Second},
		eventKeyTransaction:       {"GetTransactions", TransactionQuery{}, false, emitTypeByEvent, DefaultCronSpec10Second},
		eventKeyLatestTransaction: {"GetLatestTransactions", TransactionQuery{}, false, emitTypeByEvent, DefaultCronSpec10Second},
		eventKeyPendingTx:         {"GetPendingTransactions", SingleOwner{}, false, emitTypeByEvent, DefaultCronSpec10Second},
		eventKeyOrders:            {"GetLatestOrders", LatestOrderQuery{}, false, emitTypeByEvent, DefaultCronSpec10Second},
		eventKeyOrderTracing:      {"GetOrderByHash", OrderQuery{}, false, emitTypeByEvent, DefaultCronSpec3Second},

		eventKeyGlobalTicker:       {"GetGlobalTicker", SingleToken{}, true, emitTypeByEvent, DefaultCronSpec5Second},
		eventKeyGlobalTrend:        {"GetGlobalTrend", SingleToken{}, true, emitTypeByEvent, DefaultCronSpec10Second},
		eventKeyGlobalMarketTicker: {"GetGlobalMarketTicker", SingleToken{}, true, emitTypeByEvent, DefaultCronSpec10Hour},
		eventKeyOrderTransfer:      {"GetOrderTransfer", OrderTransferQuery{}, true, emitTypeByEvent, DefaultCronSpec5Second},
		eventKeyScanLogin:          {"", nil, true, emitTypeByEvent, DefaultCronSpec30Day},
		eventKeyCirculrNotify:      {"", nil, true, emitTypeByEvent, DefaultCronSpec30Day},
	}

	var groupId string
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		groupId = "DefaultSocketioTopic" + time.Now().String()
	} else {
		groupId = addrs[1].String()
	}

	for k, v := range topicList {
		err = so.consumer.RegisterTopicAndHandler(k, groupId, v.Data, v.Handler)
		if err != nil {
			log.Fatalf("Failed init socketio consumer, %s", err.Error())
		}
	}
	return so
}

func (so *SocketIOServiceImpl) Start() {
	server, err := socketio.NewServer(&engineio.Options{
		PingInterval: time.Second * 60 * 60,
		PingTimeout:  time.Second * 60 * 60,
	})
	if err != nil {
		log.Fatalf(err.Error())
	}
	server.OnConnect("/", func(s socketio.Conn) error {
		so.connIdMap.Store(s.ID(), s)
		return nil
	})
	server.OnEvent("/", "test", func(s socketio.Conn, msg string) {
		fmt.Println("test:", msg)
		s.Emit("reply", "pong relay msg : "+msg)
		fmt.Println("emit message finished...")
		fmt.Println(s.RemoteAddr())
	})

	for v, event := range so.eventTypeRoute {
		aliasOfV := v
		aliasOfEvent := event

		server.OnEvent("/", aliasOfV+EventPostfixReq, func(s socketio.Conn, msg string) {
			context := make(map[string]string)
			if s != nil && s.Context() != nil {
				context = s.Context().(map[string]string)
			}
			context[aliasOfV] = msg
			s.SetContext(context)
			so.connIdMap.Store(s.ID(), s)

			if len(aliasOfEvent.MethodName) != 0 {
				so.EmitNowByEventType(aliasOfV, s, msg)
			}
		})

		server.OnEvent("/", aliasOfV+EventPostfixEnd, func(s socketio.Conn, msg string) {
			if s != nil && s.Context() != nil {
				businesses := s.Context().(map[string]string)
				delete(businesses, aliasOfV)
				s.SetContext(businesses)
			}
		})
	}

	for k, events := range so.eventTypeRoute {
		copyOfK := k
		spec := events.spec

		switch k {
		case eventKeyTickers:
			so.cron.AddFunc(spec, func() { so.broadcastTpTickers(nil) })
		case eventKeyLoopringTickers:
			so.cron.AddFunc(spec, func() { so.broadcastLoopringTicker(nil) })
		case eventKeyDepth:
			so.cron.AddFunc(spec, func() { so.broadcastDepth(nil) })
		case eventKeyOrderBook:
			so.cron.AddFunc(spec, func() { so.broadcastOrderBook(nil) })
		case eventKeyTrades:
			so.cron.AddFunc(spec, func() { so.broadcastTrades(nil) })
		case eventKeyMarketCap:
			so.cron.AddFunc(spec, func() { so.broadcastMarketCap(nil) })
		case eventKeyEstimatedGasPrice:
			so.cron.AddFunc(spec, func() { so.broadcastGasPrice(nil) })
		case eventKeyGlobalTicker:
			so.cron.AddFunc(spec, func() { so.broadcastGlobalTicker(nil) })
		case eventKeyGlobalMarketTicker:
			so.cron.AddFunc(spec, func() { so.broadcastGlobalMarketTicker(nil) })
		case eventKeyGlobalTrend:
			so.cron.AddFunc(spec, func() { so.broadcastGlobalTrend(nil) })
		default:
			log.Infof("add cron emit %d ", events.emitType)

			if len(events.MethodName) == 0 {
				continue
			}

			so.cron.AddFunc(spec, func() {
				so.connIdMap.Range(func(key, value interface{}) bool {
					v := value.(socketio.Conn)
					if v.Context() != nil {
						businesses := v.Context().(map[string]string)
						eventContext, ok := businesses[copyOfK]
						if ok {
							//log.Infof("[SOCKETIO-EMIT]cron emit by key : %s, connId : %s", copyOfK, v.ID())
							so.EmitNowByEventType(copyOfK, v, eventContext)
						}
					}
					return true
				})
			})

		}
	}

	so.cron.Start()

	server.OnError("/", func(e error) {
		fmt.Println("meet error:", e)
		infos := strings.Split(e.Error(), "SOCKETFORLOOPRING")
		if len(infos) == 2 {
			so.connIdMap.Delete(infos[0])
		}

	})

	server.OnDisconnect("/", func(s socketio.Conn, msg string) {
		s.Close()
		so.connIdMap.Delete(s.ID())
		fmt.Println("closed", msg)
	})
	go server.Serve()
	defer server.Close()

	http.Handle("/socket.io/", NewServer(*server))
	log.Info("Serving at localhost: " + so.port)
	log.Fatal(http.ListenAndServe(":"+so.port, nil).Error())

}

func (so *SocketIOServiceImpl) EmitNowByEventType(bk string, v socketio.Conn, bv string) {
	if invokeInfo, ok := so.eventTypeRoute[bk]; ok {
		so.handleAfterEmit(bk, invokeInfo.Query, invokeInfo.MethodName, v, bv)
	}
}

func (so *SocketIOServiceImpl) handleWith(eventType string, query interface{}, methodName string, ctx string) string {

	results := make([]reflect.Value, 0)
	var err error

	if query == nil {
		results = reflect.ValueOf(&so.walletService).MethodByName(methodName).Call(nil)
	} else {
		queryType := reflect.TypeOf(query)
		queryClone := reflect.New(queryType)
		err = json.Unmarshal([]byte(ctx), queryClone.Interface())
		if err != nil {
			log.Info("unmarshal error " + err.Error())
			errJson, _ := json.Marshal(SocketIOJsonResp{Error: err.Error()})
			return string(errJson[:])

		}
		params := make([]reflect.Value, 1)
		params[0] = queryClone.Elem()
		results = reflect.ValueOf(&so.walletService).MethodByName(methodName).Call(params)
	}

	res := results[0]
	if results[1].Interface() == nil {
		err = nil
	} else {
		err = results[1].Interface().(error)
	}
	if err != nil {
		errJson, _ := json.Marshal(SocketIOJsonResp{Error: err.Error()})
		return string(errJson[:])
	} else {
		rst := SocketIOJsonResp{Data: res.Interface()}
		b, _ := json.Marshal(rst)
		return string(b[:])
	}
}

func (so *SocketIOServiceImpl) handleAfterEmit(eventType string, query interface{}, methodName string, conn socketio.Conn, ctx string) {
	result := so.handleWith(eventType, query, methodName, ctx)
	conn.Emit(eventType+EventPostfixRes, result)
}

func (so *SocketIOServiceImpl) broadcastTpTickers(input interface{}) (err error) {

	mkts, _ := so.walletService.GetSupportedMarket()

	tickerMap := make(map[string]SocketIOJsonResp)

	for _, mkt := range mkts {
		ticker, err := so.walletService.GetTickers(SingleMarket{mkt})
		resp := SocketIOJsonResp{}

		if err != nil {
			resp = SocketIOJsonResp{Error: err.Error()}
		} else {
			resp.Data = ticker
		}
		tickerMap[mkt] = resp
	}

	so.connIdMap.Range(func(key, value interface{}) bool {
		v := value.(socketio.Conn)
		if v.Context() != nil {
			businesses := v.Context().(map[string]string)
			ctx, ok := businesses[eventKeyLoopringTickers]
			if ok {
				var singleMarket SingleMarket
				err = json.Unmarshal([]byte(ctx), &singleMarket)
				if err != nil {
					return true
				}
				tks, ok := tickerMap[strings.ToUpper(singleMarket.Market)]
				if ok {
					v.Emit(eventKeyTickers+EventPostfixRes, tks)
				}
			}
		}
		return true
	})
	return nil
}

func (so *SocketIOServiceImpl) broadcastLoopringTicker(input interface{}) (err error) {

	resp := SocketIOJsonResp{}
	tickers, err := so.walletService.GetTicker()

	if err != nil {
		resp = SocketIOJsonResp{Error: err.Error()}
	} else {
		resp.Data = tickers
	}

	respJson, _ := json.Marshal(resp)

	so.connIdMap.Range(func(key, value interface{}) bool {
		v := value.(socketio.Conn)
		if v.Context() != nil {
			businesses := v.Context().(map[string]string)
			_, ok := businesses[eventKeyLoopringTickers]
			if ok {
				//log.Info("emit loopring ticker info")
				v.Emit(eventKeyLoopringTickers+EventPostfixRes, string(respJson[:]))
			}
		}
		return true
	})
	return nil
}

func (so *SocketIOServiceImpl) broadcastDepth(input interface{}) (err error) {
	markets := so.getConnectedMarketForDepth(eventKeyDepth, input)
	respMap := so.getDepthPushData(eventKeyDepth, markets)
	so.pushDepthData(eventKeyDepth, respMap)
	return nil
}

func (so *SocketIOServiceImpl) broadcastOrderBook(input interface{}) (err error) {
	markets := so.getConnectedMarketForDepth(eventKeyOrderBook, input)
	respMap := so.getDepthPushData(eventKeyOrderBook, markets)
	so.pushDepthData(eventKeyOrderBook, respMap)
	return nil
}

func (so *SocketIOServiceImpl) getDepthPushData(eventKey string, markets map[string]bool) map[string]string {
	respMap := make(map[string]string, 0)
	for mk := range markets {
		mktAndDelegate := strings.Split(mk, "_")
		delegate := mktAndDelegate[0]
		mkt := mktAndDelegate[1]
		resp := SocketIOJsonResp{}

		var data interface{}
		var err error
		if eventKeyDepth == eventKey {
			data, err = so.walletService.GetDepth(DepthQuery{delegate, mkt})
		} else {
			data, err = so.walletService.GetUnmergedOrderBook(DepthQuery{delegate, mkt})
		}

		if err == nil {
			resp.Data = data
		} else {
			resp = SocketIOJsonResp{Error: err.Error()}
		}
		respJson, _ := json.Marshal(resp)
		respMap[mk] = string(respJson[:])
	}
	return respMap
}

func (so *SocketIOServiceImpl) pushDepthData(eventKey string, respMap map[string]string) {
	so.connIdMap.Range(func(key, value interface{}) bool {
		v := value.(socketio.Conn)
		if v.Context() != nil {
			businesses := v.Context().(map[string]string)
			ctx, ok := businesses[eventKey]
			if ok {
				dQuery := &DepthQuery{}
				err := json.Unmarshal([]byte(ctx), dQuery)
				if err == nil && len(dQuery.DelegateAddress) > 0 && len(dQuery.Market) > 0 {
					depthKey := strings.ToLower(dQuery.DelegateAddress) + "_" + strings.ToLower(dQuery.Market)
					if len(respMap[depthKey]) > 0 {
						v.Emit(eventKey+EventPostfixRes, respMap[depthKey])
					}
				}
			}
		}
		return true
	})
}

func (so *SocketIOServiceImpl) broadcastTrades(input interface{}) (err error) {

	log.Infof("[SOCKETIO-RECEIVE-EVENT] loopring depth input.")
	markets := make(map[string]bool)

	if input != nil {
		fillEvent := input.(*dao.FillEvent)
		delegateAddress := fillEvent.DelegateAddress
		market := fillEvent.Market
		markets[strings.ToLower(delegateAddress)+"_"+strings.ToLower(market)] = true
	} else {
		markets = so.getConnectedMarketForFill()
	}

	respMap := make(map[string]string, 0)
	for mk := range markets {
		mktAndDelegate := strings.Split(mk, "_")
		delegate := mktAndDelegate[0]
		mkt := mktAndDelegate[1]
		resp := SocketIOJsonResp{}
		fills, err := so.walletService.GetLatestFills(FillQuery{DelegateAddress: delegate, Market: mkt, Side: util.SideSell})
		if err == nil {
			//log.Infof("fetch fill from wallet %d, %s", len(fills), mkt)
			resp.Data = fills
		} else {
			resp = SocketIOJsonResp{Error: err.Error()}
		}
		respJson, _ := json.Marshal(resp)
		respMap[mk] = string(respJson[:])
	}

	so.connIdMap.Range(func(key, value interface{}) bool {
		v := value.(socketio.Conn)
		if v.Context() != nil {
			businesses := v.Context().(map[string]string)
			ctx, ok := businesses[eventKeyTrades]
			if ok {
				fQuery := &FillQuery{}
				err := json.Unmarshal([]byte(ctx), fQuery)
				if err == nil && len(fQuery.DelegateAddress) > 0 && len(fQuery.Market) > 0 {
					fillKey := strings.ToLower(fQuery.DelegateAddress) + "_" + strings.ToLower(fQuery.Market)
					v.Emit(eventKeyTrades+EventPostfixRes, respMap[fillKey])
				}
			}
		}
		return true
	})
	return nil
}

func (so *SocketIOServiceImpl) broadcastMarketCap(input interface{}) (err error) {

	cnyResp := so.getPriceQuoteResp(priceQuoteCNY)
	usdResp := so.getPriceQuoteResp(priceQuoteUSD)

	so.connIdMap.Range(func(key, value interface{}) bool {
		v := value.(socketio.Conn)
		if v.Context() != nil {
			businesses := v.Context().(map[string]string)
			ctx, ok := businesses[eventKeyMarketCap]
			if ok {
				query := &PriceQuoteQuery{}
				err := json.Unmarshal([]byte(ctx), query)
				if err == nil && strings.ToLower(priceQuoteCNY) == strings.ToLower(query.Currency) {
					v.Emit(eventKeyMarketCap+EventPostfixRes, cnyResp)
				} else if err == nil && strings.ToLower(priceQuoteUSD) == strings.ToLower(query.Currency) {
					v.Emit(eventKeyMarketCap+EventPostfixRes, usdResp)
				}
			}
		}
		return true
	})
	return nil
}

func (so *SocketIOServiceImpl) broadcastGasPrice(input interface{}) (err error) {

	resp := SocketIOJsonResp{}
	gasPrice, err := so.walletService.GetEstimateGasPrice()

	if err != nil {
		resp = SocketIOJsonResp{Error: err.Error()}
	} else {
		resp.Data = gasPrice
	}

	respJson, _ := json.Marshal(resp)

	so.connIdMap.Range(func(key, value interface{}) bool {
		v := value.(socketio.Conn)
		if v.Context() != nil {
			businesses := v.Context().(map[string]string)
			_, ok := businesses[eventKeyEstimatedGasPrice]
			if ok {
				//log.Info("emit loopring gas price info")
				v.Emit(eventKeyEstimatedGasPrice+EventPostfixRes, string(respJson[:]))
			}
		}
		return true
	})
	return nil
}

func (so *SocketIOServiceImpl) broadcastGlobalTicker(input interface{}) (err error) {

	resp := SocketIOJsonResp{}
	tickers, err := so.walletService.GetGlobalTicker(SingleToken{})

	if err != nil {
		resp = SocketIOJsonResp{Error: err.Error()}
	} else {
		resp.Data = tickers
	}

	respJson, _ := json.Marshal(resp)

	so.connIdMap.Range(func(key, value interface{}) bool {
		v := value.(socketio.Conn)
		if v.Context() != nil {
			businesses := v.Context().(map[string]string)
			_, ok := businesses[eventKeyGlobalTicker]
			if ok {
				//log.Info("emit loopring gas price info")
				v.Emit(eventKeyGlobalTicker+EventPostfixRes, string(respJson[:]))
			}
		}
		return true
	})
	return nil
}

func (so *SocketIOServiceImpl) broadcastGlobalTrend(input interface{}) (err error) {

	resp := SocketIOJsonResp{}
	trendMap, err := so.walletService.GetGlobalTrend(SingleToken{})

	respMap := make(map[string]string)
	for k, v := range trendMap {
		if err != nil {
			resp = SocketIOJsonResp{Error: err.Error()}
		} else {
			resp.Data = v
		}

		respJson, _ := json.Marshal(resp)
		respMap[k] = string(respJson[:])
	}

	so.connIdMap.Range(func(key, value interface{}) bool {
		v := value.(socketio.Conn)
		if v.Context() != nil {
			businesses := v.Context().(map[string]string)
			ctx, ok := businesses[eventKeyGlobalTrend]
			if ok {
				query := &SingleToken{}
				err := json.Unmarshal([]byte(ctx), query)
				if err == nil && len(query.Token) > 0 {
					v.Emit(eventKeyGlobalTrend+EventPostfixRes, respMap[strings.ToUpper(query.Token)])
				}
			}
		}
		return true
	})
	return nil
}

func (so *SocketIOServiceImpl) broadcastGlobalMarketTicker(input interface{}) (err error) {

	resp := SocketIOJsonResp{}
	tickerMap, err := so.walletService.GetGlobalMarketTicker(SingleToken{})

	respMap := make(map[string]string)
	for k, v := range tickerMap {
		if err != nil {
			resp = SocketIOJsonResp{Error: err.Error()}
		} else {
			resp.Data = v
		}

		respJson, _ := json.Marshal(resp)
		respMap[k] = string(respJson[:])
	}

	so.connIdMap.Range(func(key, value interface{}) bool {
		v := value.(socketio.Conn)
		if v.Context() != nil {
			businesses := v.Context().(map[string]string)
			ctx, ok := businesses[eventKeyGlobalMarketTicker]
			if ok {
				query := &SingleToken{}
				err := json.Unmarshal([]byte(ctx), query)
				if err == nil && len(query.Token) > 0 {
					v.Emit(eventKeyGlobalMarketTicker+EventPostfixRes, respMap[strings.ToUpper(query.Token)])
				}
			}
		}
		return true
	})
	return nil
}

func (so *SocketIOServiceImpl) getPriceQuoteResp(currency string) string {
	resp := SocketIOJsonResp{}

	price, err := so.walletService.GetPriceQuote(PriceQuoteQuery{Currency: currency})
	if err != nil {
		log.Debug("query cny price error")
		resp = SocketIOJsonResp{Error: err.Error()}
	} else {
		resp.Data = price
	}
	respJson, _ := json.Marshal(resp)
	return string(respJson)
}

func (so *SocketIOServiceImpl) getConnectedMarketForDepth(eventKey string, input interface{}) map[string]bool {

	markets := make(map[string]bool, 0)

	if input != nil {
		depthQuery := input.(DepthQuery)
		markets[strings.ToLower(depthQuery.DelegateAddress)+"_"+strings.ToLower(depthQuery.Market)] = true
		return markets
	}

	count := 0
	so.connIdMap.Range(func(key, value interface{}) bool {
		v := value.(socketio.Conn)
		count++
		if v.Context() != nil {
			businesses := v.Context().(map[string]string)
			DCtx, ok := businesses[eventKey]
			if ok {
				dQuery := &DepthQuery{}
				err := json.Unmarshal([]byte(DCtx), dQuery)
				if err == nil && len(dQuery.DelegateAddress) > 0 && len(dQuery.Market) > 0 {
					markets[strings.ToLower(dQuery.DelegateAddress)+"_"+strings.ToLower(dQuery.Market)] = true
				}
			}
		}
		return true
	})
	return markets
}

func (so *SocketIOServiceImpl) getConnectedMarketForFill() map[string]bool {
	markets := make(map[string]bool, 0)
	so.connIdMap.Range(func(key, value interface{}) bool {
		v := value.(socketio.Conn)
		if v.Context() != nil {
			businesses := v.Context().(map[string]string)
			fCtx, ok := businesses[eventKeyTrades]
			if ok {
				fQuery := &FillQuery{}
				err := json.Unmarshal([]byte(fCtx), fQuery)
				if err == nil && len(fQuery.DelegateAddress) > 0 && len(fQuery.Market) > 0 {
					markets[strings.ToLower(fQuery.DelegateAddress)+"_"+strings.ToLower(fQuery.Market)] = true
				}
			}
		}
		return true
	})
	return markets
}

func (so *SocketIOServiceImpl) broadcastTrends(input interface{}) (err error) {

	//log.Infof("[SOCKETIO-RECEIVE-EVENT] trend input. %s", input)

	req := input.(*market.TrendUpdateMsg)
	trendQuery := TrendQuery{Market: req.Market, Interval: req.Interval}
	resp := SocketIOJsonResp{}
	trends, err := so.walletService.GetTrend(trendQuery)

	if err != nil {
		resp = SocketIOJsonResp{Error: err.Error()}
	} else {
		resp.Data = trends
	}

	respJson, _ := json.Marshal(resp)

	so.connIdMap.Range(func(key, value interface{}) bool {
		v := value.(socketio.Conn)
		if v.Context() != nil {
			businesses := v.Context().(map[string]string)
			ctx, ok := businesses[eventKeyTrends]

			if ok {
				trendQuery := &TrendQuery{}
				err = json.Unmarshal([]byte(ctx), trendQuery)
				if err != nil {
					log.Error("trend query unmarshal error, " + err.Error())
				} else if strings.ToUpper(trendQuery.Market) == strings.ToUpper(trendQuery.Market) &&
					strings.ToUpper(trendQuery.Interval) == strings.ToUpper(trendQuery.Interval) {
					log.Info("emit trend " + ctx)
					v.Emit(eventKeyTrends+EventPostfixRes, string(respJson[:]))
				}
			}
		}
		return true
	})
	return nil
}

func (so *SocketIOServiceImpl) handleBalanceUpdate(input interface{}) (err error) {

	log.Infof("[SOCKETIO-RECEIVE-EVENT] balance input.")

	balanceUpdateEvent := input.(*types.BalanceUpdateEvent)
	if len(balanceUpdateEvent.Owner) == 0 {
		return errors.New("owner can't be nil")
	}

	if common.IsHexAddress(balanceUpdateEvent.DelegateAddress) {
		so.notifyBalanceUpdateByDelegateAddress(balanceUpdateEvent.Owner, balanceUpdateEvent.DelegateAddress)
	} else {
		for k := range loopringaccessor.DelegateAddresses() {
			so.notifyBalanceUpdateByDelegateAddress(balanceUpdateEvent.Owner, k.Hex())
		}
	}
	return nil
}

func (so *SocketIOServiceImpl) notifyBalanceUpdateByDelegateAddress(owner, delegateAddress string) (err error) {

	if len(delegateAddress) == 0 || len(owner) == 0 {
		return nil
	}

	req := CommonTokenRequest{delegateAddress, owner}
	resp := SocketIOJsonResp{}
	balance, err := so.walletService.GetBalance(req)

	if err != nil {
		resp = SocketIOJsonResp{Error: err.Error()}
	} else {
		resp.Data = balance
	}

	respJson, _ := json.Marshal(resp)

	so.connIdMap.Range(func(key, value interface{}) bool {
		v := value.(socketio.Conn)
		if v.Context() != nil {
			businesses := v.Context().(map[string]string)
			ctx, ok := businesses[eventKeyBalance]
			if ok {
				query := &CommonTokenRequest{}
				err = json.Unmarshal([]byte(ctx), query)
				if err != nil {
					return true
				}

				if strings.ToLower(query.Owner) == strings.ToLower(req.Owner) && strings.ToLower(query.DelegateAddress) == strings.ToLower(req.DelegateAddress) {
					//log.Info("emit balance info")
					v.Emit(eventKeyBalance+EventPostfixRes, string(respJson[:]))
				}
			}
		}
		return true
	})
	return nil
}

func (so *SocketIOServiceImpl) handleTransactions(input interface{}) (err error) {

	log.Infof("[SOCKETIO-RECEIVE-EVENT] transactions input.")

	req := input.(*txtyp.TransactionView)
	owner := req.Owner.Hex()
	log.Infof("received owner is %s ", owner)
	so.connIdMap.Range(func(key, value interface{}) bool {
		v := value.(socketio.Conn)
		if v.Context() != nil {
			businesses := v.Context().(map[string]string)
			ctx, ok := businesses[eventKeyTransaction]
			log.Infof("cxt contains event key %b", ok)

			if ok {
				txQuery := &TransactionQuery{}
				log.Info("txQuery owner is " + txQuery.Owner)
				err = json.Unmarshal([]byte(ctx), txQuery)
				if err != nil {
					log.Error("tx query unmarshal error, " + err.Error())
				} else if strings.ToUpper(owner) == strings.ToUpper(txQuery.Owner) {
					log.Info("emit trend " + ctx)

					txs, err := so.walletService.GetTransactions(*txQuery)
					resp := SocketIOJsonResp{}

					if err != nil {
						resp = SocketIOJsonResp{Error: err.Error()}
					} else {
						resp.Data = txs
					}
					respJson, _ := json.Marshal(resp)
					v.Emit(eventKeyTransaction+EventPostfixRes, string(respJson[:]))
				}
			}
		}
		return true
	})

	return nil
}

func (so *SocketIOServiceImpl) handleLatestTransactions(input interface{}) (err error) {

	log.Infof("[SOCKETIO-RECEIVE-EVENT] latest transactions input")

	req := input.(*txtyp.TransactionView)
	owner := req.Owner.Hex()
	log.Infof("received owner is %s ", owner)
	so.connIdMap.Range(func(key, value interface{}) bool {
		v := value.(socketio.Conn)
		if v.Context() != nil {
			businesses := v.Context().(map[string]string)
			ctx, ok := businesses[eventKeyLatestTransaction]
			log.Infof("cxt contains event key %b", ok)

			if ok {
				txQuery := &TransactionQuery{}
				log.Info("txQuery owner is " + txQuery.Owner)
				err = json.Unmarshal([]byte(ctx), txQuery)
				if err != nil {
					log.Error("tx query unmarshal error, " + err.Error())
				} else if strings.ToUpper(owner) == strings.ToUpper(txQuery.Owner) {
					log.Info("emit trend " + ctx)

					txs, err := so.walletService.GetLatestTransactions(*txQuery)
					resp := SocketIOJsonResp{}

					if err != nil {
						resp = SocketIOJsonResp{Error: err.Error()}
					} else {
						resp.Data = txs
					}
					respJson, _ := json.Marshal(resp)
					v.Emit(eventKeyTransaction+EventPostfixRes, string(respJson[:]))
				}
			}
		}
		return true
	})

	return nil
}

func (so *SocketIOServiceImpl) handleTransactionUpdate(input interface{}) (err error) {
	so.handleTransactions(input)
	so.handlePendingTransaction(input)
	so.handleLatestTransactions(input)
	return nil
}

func (so *SocketIOServiceImpl) handlePendingTransaction(input interface{}) (err error) {

	log.Infof("[SOCKETIO-RECEIVE-EVENT] pending transaction input (for pending).")

	req := input.(*txtyp.TransactionView)
	owner := req.Owner.Hex()
	log.Infof("received owner is %s ", owner)
	so.connIdMap.Range(func(key, value interface{}) bool {
		v := value.(socketio.Conn)
		if v.Context() != nil {
			businesses := v.Context().(map[string]string)
			ctx, ok := businesses[eventKeyPendingTx]
			log.Infof("cxt contains event key %b", ok)

			if ok {
				txQuery := &SingleOwner{}
				err = json.Unmarshal([]byte(ctx), txQuery)
				log.Info("single owner is: " + txQuery.Owner)
				if err != nil {
					log.Error("tx query unmarshal error, " + err.Error())
				} else if strings.ToUpper(owner) == strings.ToUpper(txQuery.Owner) {
					log.Info("emit tx pending " + ctx)
					txs, err := so.walletService.GetPendingTransactions(SingleOwner{owner})
					resp := SocketIOJsonResp{}

					if err != nil {
						resp = SocketIOJsonResp{Error: err.Error()}
					} else {
						resp.Data = txs
					}
					respJson, _ := json.Marshal(resp)
					v.Emit(eventKeyPendingTx+EventPostfixRes, string(respJson[:]))
				}
			}
		}
		return true
	})

	return nil
}

func (so *SocketIOServiceImpl) handleOrdersUpdate(input interface{}) (err error) {

	log.Infof("[SOCKETIO-RECEIVE-EVENT] order update input.")

	req := input.(*types.OrderState)
	owner := req.RawOrder.Owner.Hex()
	log.Infof("received owner is %s ", owner)
	so.connIdMap.Range(func(key, value interface{}) bool {
		v := value.(socketio.Conn)
		if v.Context() != nil {
			businesses := v.Context().(map[string]string)
			ctx, ok := businesses[eventKeyOrders]
			log.Infof("cxt contains event key %b", ok)

			if ok {
				query := &LatestOrderQuery{}
				err = json.Unmarshal([]byte(ctx), query)
				log.Info("single owner is: " + query.Owner)
				if err != nil {
					log.Error("query unmarshal error, " + err.Error())
				} else if strings.ToUpper(owner) == strings.ToUpper(query.Owner) &&
					strings.ToLower(req.RawOrder.Market) == strings.ToLower(query.Market) &&
					strings.ToLower(req.RawOrder.OrderType) == strings.ToLower(query.OrderType) {
					log.Info("emit " + ctx)
					resp := SocketIOJsonResp{}
					resp.Data = orderStateToJson(*req)
					respJson, _ := json.Marshal(resp)
					v.Emit(eventKeyOrders+EventPostfixRes, string(respJson[:]))
				}
			}
		}
		return true
	})

	return nil
}

func (so *SocketIOServiceImpl) handleOrderTracing(input interface{}) (err error) {

	log.Infof("[SOCKETIO-RECEIVE-EVENT] order hash tracing input")

	req := input.(*types.OrderState)
	orderHash := req.RawOrder.Hash.Hex()
	log.Infof("received orderHash is %s ", orderHash)
	so.connIdMap.Range(func(key, value interface{}) bool {
		v := value.(socketio.Conn)
		if v.Context() != nil {
			businesses := v.Context().(map[string]string)
			ctx, ok := businesses[eventKeyOrderTracing]
			log.Infof("cxt contains event key %b", ok)

			if ok {
				query := &OrderQuery{}
				err = json.Unmarshal([]byte(ctx), query)
				log.Info("single owner is: " + query.Owner)
				if err != nil {
					log.Error("query unmarshal error, " + err.Error())
				} else if strings.ToUpper(orderHash) == strings.ToUpper(query.OrderHash) {
					log.Info("emit " + ctx)
					resp := SocketIOJsonResp{}
					resp.Data = orderStateToJson(*req)
					respJson, _ := json.Marshal(resp)
					v.Emit(eventKeyOrderTracing+EventPostfixRes, string(respJson[:]))
				}
			}
		}
		return true
	})

	return nil
}

func (so *SocketIOServiceImpl) handleOrderUpdate(input interface{}) (err error) {
	log.Infof("[SOCKETIO-RECEIVE-EVENT] order update.")
	order := input.(*types.OrderState)
	so.handleOrdersUpdate(order)
	so.handleOrderTracing(order)

	if order.RawOrder.OrderType == types.ORDER_TYPE_P2P {
		return nil
	}

	//TODO finish the depth cache.
	so.broadcastOrderBook(DepthQuery{DelegateAddress: order.RawOrder.DelegateAddress.Hex(), Market: order.RawOrder.Market})
	so.broadcastDepth(DepthQuery{DelegateAddress: order.RawOrder.DelegateAddress.Hex(), Market: order.RawOrder.Market})
	return nil
}

func (so *SocketIOServiceImpl) handleCutOff(input interface{}) (err error) {
	log.Infof("[SOCKETIO-RECEIVE-EVENT] order update.")
	//req := input.(*socketioutil.KafkaMsg)
	//owner := req.Data.(string)
	//so.e
	so.broadcastOrderBook(nil)
	so.broadcastDepth(nil)
	return nil
}

func (so *SocketIOServiceImpl) handleCutOffPair(input interface{}) (err error) {
	log.Infof("[SOCKETIO-RECEIVE-EVENT] order update.")
	cutoffPair := input.(*types.CutoffPairEvent)
	//so.handleOrdersUpdate(req.Data.(*types.CutoffPairEvent))
	market, err := util.WrapMarketByAddress(cutoffPair.Token1.Hex(), cutoffPair.Token2.Hex())
	if err != nil {
		return err
	}
	so.broadcastOrderBook(DepthQuery{DelegateAddress: cutoffPair.DelegateAddress.Hex(), Market: market})
	so.broadcastDepth(DepthQuery{DelegateAddress: cutoffPair.DelegateAddress.Hex(), Market: market})
	return nil
}

func (so *SocketIOServiceImpl) handleOrderTransfer(input interface{}) (err error) {

	log.Infof("[SOCKETIO-RECEIVE-EVENT] order transfer input.")

	ot := input.(*OrderTransfer)
	ot.Origin = ""
	log.Infof("received hash is %s ", ot.Hash)
	so.connIdMap.Range(func(key, value interface{}) bool {
		v := value.(socketio.Conn)
		if v.Context() != nil {
			businesses := v.Context().(map[string]string)
			ctx, ok := businesses[eventKeyOrderTransfer]
			log.Infof("cxt contains event key %b", ok)

			if ok {
				query := &OrderTransferQuery{}
				err = json.Unmarshal([]byte(ctx), query)
				if err != nil {
					log.Error("query unmarshal error, " + err.Error())
				} else if strings.ToLower(ot.Hash) == strings.ToLower(query.Hash) {
					log.Info("emit " + ctx)
					resp := SocketIOJsonResp{}
					resp.Data = ot
					respJson, _ := json.Marshal(resp)
					v.Emit(eventKeyOrderTransfer+EventPostfixRes, string(respJson[:]))
				}
			}
		}
		return true
	})

	return nil
}

func (so *SocketIOServiceImpl) handleScanLogin(input interface{}) (err error) {

	log.Infof("[SOCKETIO-RECEIVE-EVENT] scan login input.")

	ot := input.(*LoginInfo)
	log.Infof("received UUID is %s ", ot.UUID)
	so.connIdMap.Range(func(key, value interface{}) bool {
		v := value.(socketio.Conn)
		if v.Context() != nil {
			businesses := v.Context().(map[string]string)
			ctx, ok := businesses[eventKeyScanLogin]
			log.Infof("cxt contains event key %b", ok)

			if ok {
				query := &LoginInfo{}
				err = json.Unmarshal([]byte(ctx), query)
				if err != nil {
					log.Error("query unmarshal error, " + err.Error())
				} else if strings.ToLower(ot.UUID) == strings.ToLower(query.UUID) {
					log.Info("emit " + ctx)
					resp := SocketIOJsonResp{}
					resp.Data = ot
					respJson, _ := json.Marshal(resp)
					v.Emit(eventKeyScanLogin+EventPostfixRes, string(respJson[:]))
				}
			}
		}
		return true
	})

	return nil
}

func (so *SocketIOServiceImpl) handleCirculrNotify(input interface{}) (err error) {

	ot := input.(*NotifyCirculrBody)
	log.Infof("received owner is %s ", ot.Owner)
	so.connIdMap.Range(func(key, value interface{}) bool {
		v := value.(socketio.Conn)
		if v.Context() != nil {
			businesses := v.Context().(map[string]string)
			ctx, ok := businesses[eventKeyCirculrNotify]
			log.Infof("cxt contains event key %b", ok)

			if ok {
				query := &SingleOwner{}
				err = json.Unmarshal([]byte(ctx), query)
				if err != nil {
					log.Error("query unmarshal error, " + err.Error())
				} else if strings.ToLower(ot.Owner) == strings.ToLower(query.Owner) {
					log.Info("emit " + ctx)
					resp := SocketIOJsonResp{}
					resp.Data = ot
					respJson, _ := json.Marshal(resp)
					v.Emit(eventKeyCirculrNotify+EventPostfixRes, string(respJson[:]))
				}
			}
		}
		return true
	})

	return nil
}
