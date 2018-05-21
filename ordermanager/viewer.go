package ordermanager

import (
	"github.com/Loopring/relay-cluster/dao"
	"github.com/Loopring/relay-cluster/usermanager"
	"github.com/Loopring/relay-lib/log"
	"github.com/Loopring/relay-lib/marketcap"
	"github.com/Loopring/relay-lib/types"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

type OrderViewer interface {
	MinerOrders(delegate, tokenS, tokenB common.Address, length int, reservedTime, startBlockNumber, endBlockNumber int64, filterOrderHashLists ...*types.OrderDelayList) []*types.OrderState
	GetOrderBook(protocol, tokenS, tokenB common.Address, length int) ([]types.OrderState, error)
	GetOrders(query map[string]interface{}, statusList []types.OrderStatus, pageIndex, pageSize int) (dao.PageResult, error)
	GetLatestOrders(query map[string]interface{}, length int) ([]types.OrderState, error)
	GetOrderByHash(hash common.Hash) (*types.OrderState, error)
	UpdateBroadcastTimeByHash(hash common.Hash, bt int) error
	FillsPageQuery(query map[string]interface{}, pageIndex, pageSize int) (dao.PageResult, error)
	GetLatestFills(query map[string]interface{}, limit int) ([]dao.FillEvent, error)
	FindFillsByRingHash(ringHash common.Hash) (result []dao.FillEvent, err error)
	RingMinedPageQuery(query map[string]interface{}, pageIndex, pageSize int) (dao.PageResult, error)
	IsOrderCutoff(protocol, owner, token1, token2 common.Address, validsince *big.Int) bool
	GetFrozenAmount(owner common.Address, token common.Address, statusSet []types.OrderStatus, delegateAddress common.Address) (*big.Int, error)
	GetFrozenLRCFee(owner common.Address, statusSet []types.OrderStatus) (*big.Int, error)
}

type OrderViewerImpl struct {
	um          usermanager.UserManager
	mc          marketcap.MarketCapProvider
	rds         *dao.RdsService
	cutoffCache *CutoffCache
}

func NewOrderViewer(options *OrderManagerOptions,
	rds *dao.RdsService,
	market marketcap.MarketCapProvider,
	userManager usermanager.UserManager) *OrderViewerImpl {

	var viewer OrderViewerImpl
	viewer.um = userManager
	viewer.mc = market
	viewer.rds = rds
	viewer.cutoffCache = NewCutoffCache(options.CutoffCacheCleanTime)

	return &viewer
}

func (om *OrderViewerImpl) MinerOrders(delegate, tokenS, tokenB common.Address, length int, reservedTime, startBlockNumber, endBlockNumber int64, filterOrderHashLists ...*types.OrderDelayList) []*types.OrderState {
	var list []*types.OrderState

	var (
		modelList    []*dao.Order
		err          error
		filterStatus = []types.OrderStatus{types.ORDER_FINISHED, types.ORDER_CUTOFF, types.ORDER_CANCEL}
	)

	for _, orderDelay := range filterOrderHashLists {
		orderHashes := []string{}
		for _, hash := range orderDelay.OrderHash {
			orderHashes = append(orderHashes, hash.Hex())
		}
		if len(orderHashes) > 0 && orderDelay.DelayedCount != 0 {
			if err = om.rds.MarkMinerOrders(orderHashes, orderDelay.DelayedCount); err != nil {
				log.Debugf("order manager,provide orders for miner error:%s", err.Error())
			}
		}
	}

	// 从数据库获取订单
	if modelList, err = om.rds.GetOrdersForMiner(delegate.Hex(), tokenS.Hex(), tokenB.Hex(), length, filterStatus, reservedTime, startBlockNumber, endBlockNumber); err != nil {
		log.Errorf("err:%s", err.Error())
		return list
	}

	for _, v := range modelList {
		state := &types.OrderState{}
		v.ConvertUp(state)
		if om.um.InWhiteList(state.RawOrder.Owner) {
			list = append(list, state)
		} else {
			log.Debugf("order manager,owner:%s not in white list", state.RawOrder.Owner.Hex())
		}
	}

	return list
}

func (om *OrderViewerImpl) GetOrderBook(protocol, tokenS, tokenB common.Address, length int) ([]types.OrderState, error) {
	var list []types.OrderState
	models, err := om.rds.GetOrderBook(protocol, tokenS, tokenB, length)
	if err != nil {
		return list, err
	}

	for _, v := range models {
		var state types.OrderState
		if err := v.ConvertUp(&state); err != nil {
			continue
		}
		list = append(list, state)
	}

	return list, nil
}

func (om *OrderViewerImpl) GetOrders(query map[string]interface{}, statusList []types.OrderStatus, pageIndex, pageSize int) (dao.PageResult, error) {
	var (
		pageRes dao.PageResult
	)
	sL := make([]int, 0)
	for _, s := range statusList {
		sL = append(sL, int(s))
	}
	tmp, err := om.rds.OrderPageQuery(query, sL, pageIndex, pageSize)

	if err != nil {
		return pageRes, err
	}
	pageRes.PageIndex = tmp.PageIndex
	pageRes.PageSize = tmp.PageSize
	pageRes.Total = tmp.Total

	for _, v := range tmp.Data {
		var state types.OrderState
		model := v.(dao.Order)
		if err := model.ConvertUp(&state); err != nil {
			log.Debug("convertUp error occurs " + err.Error())
			continue
		}
		pageRes.Data = append(pageRes.Data, state)
	}
	return pageRes, nil
}

func (om *OrderViewerImpl) GetLatestOrders(query map[string]interface{}, length int) (rst []types.OrderState, err error) {
	tmp, err := om.rds.GetLatestOrders(query, length)
	if err != nil {
		return nil, err
	}

	rst = make([]types.OrderState, 0)

	for _, v := range tmp {
		var state types.OrderState
		if err := v.ConvertUp(&state); err != nil {
			log.Debug("convertUp error occurs " + err.Error())
			continue
		}
		rst = append(rst, state)
	}
	return rst, nil
}

func (om *OrderViewerImpl) GetOrderByHash(hash common.Hash) (orderState *types.OrderState, err error) {
	var result types.OrderState
	order, err := om.rds.GetOrderByHash(hash)
	if err != nil {
		return nil, err
	}

	if err := order.ConvertUp(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (om *OrderViewerImpl) UpdateBroadcastTimeByHash(hash common.Hash, bt int) error {
	return om.rds.UpdateBroadcastTimeByHash(hash.Hex(), bt)
}

func (om *OrderViewerImpl) FillsPageQuery(query map[string]interface{}, pageIndex, pageSize int) (result dao.PageResult, err error) {
	return om.rds.FillsPageQuery(query, pageIndex, pageSize)
}

func (om *OrderViewerImpl) GetLatestFills(query map[string]interface{}, limit int) (result []dao.FillEvent, err error) {
	return om.rds.GetLatestFills(query, limit)
}

func (om *OrderViewerImpl) FindFillsByRingHash(ringHash common.Hash) (result []dao.FillEvent, err error) {
	return om.rds.FindFillsByRingHash(ringHash)
}

func (om *OrderViewerImpl) RingMinedPageQuery(query map[string]interface{}, pageIndex, pageSize int) (result dao.PageResult, err error) {
	return om.rds.RingMinedPageQuery(query, pageIndex, pageSize)
}

func (om *OrderViewerImpl) IsOrderCutoff(protocol, owner, token1, token2 common.Address, validsince *big.Int) bool {
	return om.cutoffCache.IsOrderCutoff(protocol, owner, token1, token2, validsince)
}

func (om *OrderViewerImpl) GetFrozenAmount(owner common.Address, token common.Address, statusSet []types.OrderStatus, delegateAddress common.Address) (*big.Int, error) {
	orderList, err := om.rds.GetFrozenAmount(owner, token, statusSet, delegateAddress)
	if err != nil {
		return nil, err
	}

	totalAmount := big.NewInt(0)

	if len(orderList) == 0 {
		return totalAmount, nil
	}

	for _, v := range orderList {
		var state types.OrderState
		if err := v.ConvertUp(&state); err != nil {
			continue
		}
		rs, _ := state.RemainedAmount()
		totalAmount.Add(totalAmount, rs.Num())
	}

	return totalAmount, nil
}

func (om *OrderViewerImpl) GetFrozenLRCFee(owner common.Address, statusSet []types.OrderStatus) (*big.Int, error) {
	orderList, err := om.rds.GetFrozenLrcFee(owner, statusSet)
	if err != nil {
		return nil, err
	}

	totalAmount := big.NewInt(0)

	if len(orderList) == 0 {
		return totalAmount, nil
	}

	for _, v := range orderList {
		lrcFee, _ := new(big.Int).SetString(v.LrcFee, 0)
		totalAmount.Add(totalAmount, lrcFee)
	}

	return totalAmount, nil
}
