package gateway

import (
	"encoding/json"
	"errors"
	"fmt"
	txtyp "github.com/Loopring/relay-cluster/txmanager/types"
	socketioutil "github.com/Loopring/relay-cluster/util"
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
	DefaultCronSpec5Second  = "0/5 * * * * *"
	DefaultCronSpec10Second = "0/10 * * * * *"
	DefaultCronSpec5Minute  = "0 */5 * * * *"
)

const (
	emitTypeByEvent = 1
	emitTypeByCron  = 2
	priceQuoteCNY   = "CNY"
	priceQuoteUSD   = "USD"
)

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
	MethodName   string
	Query        interface{}
	isBroadcast  bool
	emitType     int
	spec         string
	eventHandler func(event interface{}) error
}

const (
	eventKeyTickers         = "tickers"
	eventKeyLoopringTickers = "loopringTickers"
	eventKeyTrends          = "trends"
	eventKeyMarketCap       = "marketcap"
	eventKeyBalance         = "balance"
	eventKeyTransaction     = "transaction"
	eventKeyPendingTx       = "pendingTx"
	eventKeyDepth           = "depth"
	eventKeyOrderBook       = "orderBook"
	eventKeyTrades          = "trades"
	eventKeyMarketOrders    = "marketOrders"
	eventKeyP2POrders       = "p2pOrders"
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

func NewSocketIOService(port string, walletService WalletServiceImpl, brokers []string) *SocketIOServiceImpl {
	so := &SocketIOServiceImpl{}
	so.port = port
	so.walletService = walletService
	so.connIdMap = &sync.Map{}
	so.cron = cron.New()
	so.consumer = &kafka.ConsumerRegister{}
	so.consumer.Initialize(brokers)

	so.eventTypeRoute = map[string]InvokeInfo{
		eventKeyTickers:         {"GetTickers", SingleMarket{}, true, emitTypeByCron, DefaultCronSpec5Second, so.broadcastTpTickers},
		eventKeyLoopringTickers: {"GetTicker", nil, true, emitTypeByEvent, DefaultCronSpec5Second, so.broadcastLoopringTicker},
		eventKeyTrends:          {"GetTrend", TrendQuery{}, true, emitTypeByEvent, DefaultCronSpec10Second, so.broadcastTrends},
		eventKeyMarketCap:       {"GetPriceQuote", PriceQuoteQuery{}, true, emitTypeByCron, DefaultCronSpec5Minute, so.broadcastMarketCap},
		eventKeyDepth:           {"GetDepth", DepthQuery{}, true, emitTypeByEvent, DefaultCronSpec10Second, so.broadcastDepth},
		eventKeyOrderBook:       {"GetUnmergedOrderBook", DepthQuery{}, true, emitTypeByEvent, DefaultCronSpec10Second, so.broadcastOrderBook},
		eventKeyTrades:          {"GetLatestFills", FillQuery{}, true, emitTypeByEvent, DefaultCronSpec10Second, so.broadcastTrades},

		eventKeyBalance:      {"GetBalance", CommonTokenRequest{}, false, emitTypeByEvent, DefaultCronSpec10Second, so.handleBalanceUpdate},
		eventKeyTransaction:  {"GetLatestTransactions", TransactionQuery{}, false, emitTypeByEvent, DefaultCronSpec10Second, so.handleTransactionUpdate},
		eventKeyPendingTx:    {"GetPendingTransactions", SingleOwner{}, false, emitTypeByEvent, DefaultCronSpec10Second, so.handlePendingTransaction},
		eventKeyMarketOrders: {"GetLatestMarketOrders", LatestOrderQuery{}, false, emitTypeByEvent, DefaultCronSpec10Second, so.handleMarketOrdersUpdate},
		//eventKeyP2POrders:    {"GetLatestP2POrders", LatestOrderQuery{}, false, emitTypeByEvent, DefaultCronSpec10Second},
	}

	var topic string
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		topic = "DefaultSocketioTopic" + time.Now().String()
	} else {
		topic = addrs[0].String()
	}

	for k, v := range so.eventTypeRoute {
		err = so.consumer.RegisterTopicAndHandler(k, topic, socketioutil.KafkaMsg{}, v.eventHandler)
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

	for v := range so.eventTypeRoute {
		aliasOfV := v

		server.OnEvent("/", aliasOfV+EventPostfixReq, func(s socketio.Conn, msg string) {
			context := make(map[string]string)
			if s != nil && s.Context() != nil {
				context = s.Context().(map[string]string)
			}
			context[aliasOfV] = msg
			s.SetContext(context)
			so.connIdMap.Store(s.ID(), s)
			so.EmitNowByEventType(aliasOfV, s, msg)
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
		default:
			log.Infof("add cron emit %d ", events.emitType)
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

	//so.cron.AddFunc("0/10 * * * * *", func() {
	//
	//	for _, v := range so.connIdMap {
	//		if v.Context() == nil {
	//			continue
	//		} else {
	//			businesses := v.Context().(map[string]string)
	//			if businesses != nil {
	//				for bk, bv := range businesses {
	//					so.EmitNowByEventType(bk, v, bv)
	//				}
	//			}
	//		}
	//	}
	//})
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
	//log.Infof("[SOCKETIO-RECEIVE-EVENT] loopring depth input. %s", input)
	markets := so.getConnectedMarketForDepth(eventKeyDepth)
	respMap := so.getDepthPushData(eventKeyDepth, markets)
	so.pushDepthData(eventKeyDepth, respMap)
	return nil
}

func (so *SocketIOServiceImpl) broadcastOrderBook(input interface{}) (err error) {
	markets := so.getConnectedMarketForDepth(eventKeyOrderBook)
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
					v.Emit(eventKey+EventPostfixRes, respMap[depthKey])
				}
			}
		}
		return true
	})
}

func (so *SocketIOServiceImpl) broadcastTrades(input interface{}) (err error) {

	//log.Infof("[SOCKETIO-RECEIVE-EVENT] loopring depth input. %s", input)

	markets := so.getConnectedMarketForFill()

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

func (so *SocketIOServiceImpl) getPriceQuoteResp(currency string) string {
	resp := SocketIOJsonResp{}

	price, err := so.walletService.GetPriceQuote(PriceQuoteQuery{Currency: currency})
	if err != nil {
		log.Debug("query cny price error")
		resp.Data = price
	} else {
		resp = SocketIOJsonResp{Error: err.Error()}
	}
	respJson, _ := json.Marshal(resp)
	return string(respJson)
}

func (so *SocketIOServiceImpl) getConnectedMarketForDepth(eventKey string) map[string]bool {
	markets := make(map[string]bool, 0)
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

	req := input.(TrendQuery)
	resp := SocketIOJsonResp{}
	trends, err := so.walletService.GetTrend(req)

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
				} else if strings.ToUpper(req.Market) == strings.ToUpper(trendQuery.Market) &&
					strings.ToUpper(req.Interval) == strings.ToUpper(trendQuery.Interval) {
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

	log.Infof("[SOCKETIO-RECEIVE-EVENT] balance input. %s", input)

	req := input.(types.BalanceUpdateEvent)
	if len(req.Owner) == 0 {
		return errors.New("owner can't be nil")
	}

	if common.IsHexAddress(req.DelegateAddress) {
		so.notifyBalanceUpdateByDelegateAddress(req.Owner, req.DelegateAddress)
	} else {
		for k := range loopringaccessor.DelegateAddresses() {
			so.notifyBalanceUpdateByDelegateAddress(req.Owner, k.Hex())
		}
	}
	return nil
}

func (so *SocketIOServiceImpl) notifyBalanceUpdateByDelegateAddress(owner, delegateAddress string) (err error) {
	req := CommonTokenRequest{owner, delegateAddress}
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
			_, ok := businesses[eventKeyBalance]
			if ok {
				//log.Info("emit balance info")
				v.Emit(eventKeyBalance+EventPostfixRes, string(respJson[:]))
			}
		}
		return true
	})
	return nil
}

//func (so *SocketIOServiceImpl) broadcastDepth(input interface{}) (err error) {
//
//	log.Infof("[SOCKETIO-RECEIVE-EVENT] depth input. %s", input)
//
//	req := input.(types.DepthUpdateEvent)
//	resp := SocketIOJsonResp{}
//	depths, err := so.walletService.GetDepth(DepthQuery{req.DelegateAddress, req.DelegateAddress})
//
//	if err != nil {
//		resp = SocketIOJsonResp{Error: err.Error()}
//	} else {
//		resp.Data = depths
//	}
//
//	respJson, _ := json.Marshal(resp)
//
//	so.connIdMap.Range(func(key, value interface{}) bool {
//		v := value.(socketio.Conn)
//		if v.Context() != nil {
//			businesses := v.Context().(map[string]string)
//			ctx, ok := businesses[eventKeyDepth]
//
//			if ok {
//				depthQuery := &DepthQuery{}
//				err = json.Unmarshal([]byte(ctx), depthQuery)
//				if err != nil {
//					log.Error("depth query unmarshal error, " + err.Error())
//				} else if strings.ToUpper(req.DelegateAddress) == strings.ToUpper(depthQuery.DelegateAddress) &&
//					strings.ToUpper(req.Market) == strings.ToUpper(depthQuery.Market) {
//					log.Info("emit trend " + ctx)
//					v.Emit(eventKeyDepth+EventPostfixRes, string(respJson[:]))
//				}
//			}
//		}
//		return true
//	})
//	return nil
//}

func (so *SocketIOServiceImpl) handleTransactionUpdate(input interface{}) (err error) {

	log.Infof("[SOCKETIO-RECEIVE-EVENT] transaction input. %s", input)

	req := input.(*txtyp.TransactionView)
	owner := req.Owner.Hex()
	log.Infof("received owner is %s ", owner)
	fmt.Println(so.connIdMap)
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

func (so *SocketIOServiceImpl) handlePendingTransaction(input interface{}) (err error) {

	log.Infof("[SOCKETIO-RECEIVE-EVENT] transaction input (for pending). %s", input)

	req := input.(*txtyp.TransactionView)
	owner := req.Owner.Hex()
	log.Infof("received owner is %s ", owner)
	so.connIdMap.Range(func(key, value interface{}) bool {
		v := value.(socketio.Conn)
		fmt.Println(key)
		fmt.Println(value)
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

func (so *SocketIOServiceImpl) handleMarketOrdersUpdate(input interface{}) (err error) {

	log.Infof("[SOCKETIO-RECEIVE-EVENT] market order input (for pending). %s", input)

	req := input.(*types.OrderState)
	owner := req.RawOrder.Owner.Hex()
	log.Infof("received owner is %s ", owner)
	so.connIdMap.Range(func(key, value interface{}) bool {
		v := value.(socketio.Conn)
		fmt.Println(key)
		fmt.Println(value)
		if v.Context() != nil {
			businesses := v.Context().(map[string]string)
			ctx, ok := businesses[eventKeyMarketOrders]
			log.Infof("cxt contains event key %b", ok)

			if ok {
				query := &LatestOrderQuery{}
				err = json.Unmarshal([]byte(ctx), query)
				log.Info("single owner is: " + query.Owner)
				if err != nil {
					log.Error("query unmarshal error, " + err.Error())
				} else if strings.ToUpper(owner) == strings.ToUpper(query.Owner) && strings.ToLower(req.RawOrder.Market) == strings.ToLower(query.Market) {
					log.Info("emit " + ctx)
					resp := SocketIOJsonResp{}
					resp.Data = orderStateToJson(*req)
					respJson, _ := json.Marshal(resp)
					v.Emit(eventKeyMarketOrders+EventPostfixRes, string(respJson[:]))
				}
			}
		}
		return true
	})

	return nil
}
