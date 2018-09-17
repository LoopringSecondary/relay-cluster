package market

import (
	"encoding/json"
	"fmt"
	"github.com/Loopring/relay-lib/types"
	"github.com/ethereum/go-ethereum/common"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"strings"
	"testing"
)

var MarketBaseOrder = map[string]uint8{"BAR": 5, "LRC": 10, "WETH": 20, "USDT": 30, "TUSD": 40}

var (
	SupportTokens  map[string]types.Token // token symbol to entity
	AllTokens      map[string]types.Token
	SupportMarkets map[string]types.Token // token symbol to contract hex address
	AllMarkets     []string
	SymbolTokenMap map[common.Address]string
)

type token struct {
	Protocol string `json:"Protocol"`
	Symbol   string `json:"Symbol"`
	Name     string `json:"Name"`
	Source   string `json:"Source"`
	Deny     bool   `json:"Deny"`
	Decimals int    `json:"Decimals"`
	IsMarket bool   `json:"IsMarket"`
	IcoPrice string `json:"IcoPrice"`
}

func getDisplayMarketsFromDB(marketfile string) (displayMarkets []types.Market) {
	fn, err := os.Open(marketfile)
	if err != nil {
		log.Fatalf("market util load markets failed:%s", err.Error())
	}
	bs, err := ioutil.ReadAll(fn)
	if err != nil {
		log.Fatalf("market util read markets json file failed:%s", err.Error())
	}
	if err := json.Unmarshal(bs, &displayMarkets); err != nil {
		log.Fatalf("market util unmarshal tokens failed:%s", err.Error())
	}
	return
}

func TestAllMarkets(t *testing.T) {

	var list []types.Market
	list = getDisplayMarketsFromDB("/Users/michelle/gopath/src/github.com/Loopring/relay-cluster/config/markets.json")
	for _, v := range list {
		jsonstr, _ := json.Marshal(v)
		fmt.Println(string(jsonstr))
	}
}

func TestAllTokens(t *testing.T) {
	SupportTokens, SupportMarkets, AllTokens, AllMarkets, SymbolTokenMap = getTokenAndMarketFromDB("/Users/michelle/Desktop/tokens.json")
	fmt.Println(len(AllTokens))
	fmt.Println(len(AllMarkets))
}

func getTokenAndMarketFromDB(tokenfile string) (
	supportTokens map[string]types.Token,
	supportMarkets map[string]types.Token,
	allTokens map[string]types.Token,
	allMarkets []string,
	symbolTokenMap map[common.Address]string) {

	supportTokens = make(map[string]types.Token)
	supportMarkets = make(map[string]types.Token)
	allTokens = make(map[string]types.Token)
	allMarkets = make([]string, 0)
	symbolTokenMap = make(map[common.Address]string)

	var list []token
	fn, err := os.Open(tokenfile)
	if err != nil {
		log.Fatalf("market util load tokens failed:%s", err.Error())
	}
	bs, err := ioutil.ReadAll(fn)
	if err != nil {
		log.Fatalf("market util read tokens json file failed:%s", err.Error())
	}
	if err := json.Unmarshal(bs, &list); err != nil {
		log.Fatalf("market util unmarshal tokens failed:%s", err.Error())
	}

	for _, v := range list {
		if v.Deny == false {
			t := v.convert()
			if t.IsMarket == true {
				supportMarkets[t.Symbol] = t
			} else {
				supportTokens[t.Symbol] = t
			}
		}
	}

	// set all tokens
	for k, v := range supportTokens {
		allTokens[k] = v
		symbolTokenMap[v.Protocol] = v.Symbol
	}
	for k, v := range supportMarkets {
		allTokens[k] = v
		symbolTokenMap[v.Protocol] = v.Symbol
	}

	// set all markets
	for k := range allTokens { // lrc,omg
		for kk := range supportMarkets { //eth
			o, ok := MarketBaseOrder[k]
			if ok {
				baseOrder := MarketBaseOrder[kk]
				if o < baseOrder {
					allMarkets = append(allMarkets, k+"-"+kk)
				}
			} else {
				allMarkets = append(allMarkets, k+"-"+kk)
			}
		}
	}

	return
}

func (t *token) convert() types.Token {
	var dst types.Token

	dst.Protocol = common.HexToAddress(t.Protocol)
	dst.Symbol = strings.ToUpper(t.Symbol)
	dst.Name = strings.ToUpper(t.Name)
	dst.Source = t.Source
	dst.Deny = t.Deny
	dst.Decimals = new(big.Int)
	dst.Decimals.SetString("1"+strings.Repeat("0", t.Decimals), 0)
	dst.IsMarket = t.IsMarket
	if "" != t.IcoPrice {
		dst.IcoPrice = new(big.Rat)
		dst.IcoPrice.SetString(t.IcoPrice)
	}

	return dst
}
