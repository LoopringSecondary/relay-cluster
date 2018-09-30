package gateway

import (
	"github.com/Loopring/relay-cluster/ordermanager/viewer"
	cache "github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"strconv"
	"strings"
	"time"
)

type ContestRoundType uint8

const (
	Round1 ContestRoundType = 1
	Round2 ContestRoundType = 2
	Round3 ContestRoundType = 3
)

const contestRankCacheKey = "contestRank"

type ContestRoundRange struct {
	Start int64
	End   int64
}

type ContestRankReq struct {
	Head      int    `json: "head"`
	PageIndex int    `json: "pageIndex"`
	PageSize  int    `json: "pageSize"`
	Owner     string `json: "owner"`
	Round     uint8  `json: "round"`
}

type ContestRankItem struct {
	Owner        string `json: "owner"`
	TradeCount   int    `json: "tradeCount"`
	Rank         int    `json: "rank"`
	RewardAmount int    `json: "rewardAmount"`
}

type ContestRankServiceImpl struct {
	roundRange  map[ContestRoundType]ContestRoundRange
	orderViewer viewer.OrderViewer
	local       *cache.Cache
}

func NewContestRankService(orderViewer viewer.OrderViewer) *ContestRankServiceImpl {
	roundRange := make(map[ContestRoundType]ContestRoundRange)
	roundRange[Round1] = ContestRoundRange{1538323200, 1538755200}
	r := &ContestRankServiceImpl{orderViewer: orderViewer, roundRange: roundRange, local: cache.New(1*time.Minute, 5*time.Minute)}
	return r
}

func (c *ContestRankServiceImpl) GetContestRankByOwner(req ContestRankReq) (item ContestRankItem, err error) {

	if len(req.Owner) == 0 {
		return item, errors.New("owner can't be null")
	}

	if req.Round < 1 || req.Round > 3 {
		return item, errors.New("round must be 1~3")
	}

	items, err := c.GetAllItems(ContestRoundType(req.Round))
	if err != nil {
		return item, err
	}
	for _, v := range items {
		if strings.ToLower(v.Owner) == strings.ToLower(req.Owner) {
			return v, nil
		}
	}
	return item, errors.New("can't found rank information by owner " + req.Owner)
}

func (c *ContestRankServiceImpl) GetHeadContestRanks(req ContestRankReq) (items []ContestRankItem, err error) {

	if req.Round < 1 || req.Round > 3 {
		return items, errors.New("round must be 1~3")
	}

	if req.Head < 0 || req.Head > 500 {
		return items, errors.New("head must be 1~500")
	}

	items, err = c.GetAllItems(ContestRoundType(req.Round))
	if err != nil {
		return items, err
	}

	head := len(items)
	if req.Head < head {
		head = req.Head
	}

	return items[0:head], nil
}

func (c *ContestRankServiceImpl) GetPagedContestRanks(req ContestRankReq) (rst PageResult, err error) {

	if req.Round < 1 || req.Round > 3 {
		return rst, errors.New("round must be 1~3")
	}

	if req.PageIndex < 1 {
		req.PageIndex = 1
	}

	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 10
	}

	start := (req.PageIndex - 1) * req.PageSize
	end := req.PageSize + start

	items, err := c.GetAllItems(ContestRoundType(req.Round))
	if err != nil {
		return rst, err
	}

	if start > len(items) {
		start = len(items) - 1
	}

	if end > len(items) {
		end = len(items)
	}

	data := make([]interface{}, 0)
	itemResult := items[start:end]
	for _, v := range itemResult {
		data = append(data, v)
	}
	return PageResult{PageIndex: req.PageIndex, PageSize: req.PageSize, Total: len(items), Data: data}, err
}

func (c *ContestRankServiceImpl) GetItemsFromCache(round ContestRoundType) (items []ContestRankItem, ok bool) {
	itemsFromCache, ok := c.local.Get(getCacheKeyByRound(round))
	if ok {
		return itemsFromCache.([]ContestRankItem), ok
	} else {
		return items, ok
	}
}

func (c *ContestRankServiceImpl) SetItemsCache(round ContestRoundType, items []ContestRankItem) {
	c.local.Set(getCacheKeyByRound(round), items, 1*time.Minute)
}

func (c *ContestRankServiceImpl) GetAllItems(round ContestRoundType) (items []ContestRankItem, err error) {
	items, ok := c.GetItemsFromCache(round)
	if !ok {
		items, err = c.GetAllItemsFromDB(round)
		c.SetItemsCache(round, items)
	}
	return items, err
}

func (c *ContestRankServiceImpl) GetAllItemsFromDB(round ContestRoundType) (items []ContestRankItem, err error) {
	rankDOList := c.orderViewer.GetAllTradeByRank(c.roundRange[round].Start, c.roundRange[round].End)
	items = make([]ContestRankItem, 0)
	for k, v := range rankDOList {
		items = append(items, ContestRankItem{Rank: k + 1, Owner: v.Owner, TradeCount: v.TradeCount})
	}
	return items, err
}

func getCacheKeyByRound(round ContestRoundType) string {
	return contestRankCacheKey + strconv.FormatInt(int64(round), 10)
}
