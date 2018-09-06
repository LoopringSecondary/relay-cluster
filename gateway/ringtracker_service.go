package gateway

import (
	"github.com/Loopring/relay-cluster/ringtrackermanager/viewer"
	"github.com/Loopring/relay-cluster/ringtrackermanager/types"
	"github.com/Loopring/relay-cluster/market"
	"github.com/Loopring/relay-cluster/dao"
)

type RingTrackerServiceImpl struct {
	ringTrackerViewer viewer.RingTrackerViewer
	market            market.GlobalMarket
}

func NewRingTrackerService(ringTrackerViewer viewer.RingTrackerViewer, market market.GlobalMarket) *RingTrackerServiceImpl {
	r := &RingTrackerServiceImpl{}
	r.ringTrackerViewer = ringTrackerViewer
	r.market = market
	return r
}

func (r *RingTrackerServiceImpl) GetAmount() types.AmountResp {
	return r.ringTrackerViewer.GetAmount()
}

func (r *RingTrackerServiceImpl) GetRingTrackerTrend(req types.TrendReq) types.TrendRsp {
	return r.ringTrackerViewer.GetTrend(types.StrToTrendDuration(req.Duration), types.StrToTrendType(req.Type), req.Keyword, req.Len, types.StrToCurrency(req.Currency))
}

func (r *RingTrackerServiceImpl) GetEcosystemTrend(req types.EcoTrendReq) []types.EcoTrendRsp {
	return r.ringTrackerViewer.GetEcosystemTrend(types.StrToEcoTrendDuration(req.Duration), types.StrToTrendType(req.Type), types.StrToIndicator(req.Indicator), types.StrToCurrency(req.Currency))
}

func (r *RingTrackerServiceImpl) GetTrades(req types.QueryReq) dao.PageResult {
	return r.ringTrackerViewer.GetTrades(types.StrToCurrency(req.Currency), types.StrToTrendType(req.Type), req.Keyword, req.Search, req.PageIndex, req.PageSize)
}

func (r *RingTrackerServiceImpl) GetTradeDetails(req types.QueryReq) []dao.FullFillEvent {
	return r.ringTrackerViewer.GetTradeDetails(req.DelegateAddress, req.RingIndex, req.FillIndex)
}

func (r *RingTrackerServiceImpl) GetAllTokens(req types.QueryReq) dao.PageResult {
	return r.ringTrackerViewer.GetAllTokens(types.StrToCurrency(req.Currency), types.StrToSort(req.Sort), req.PageIndex, req.PageSize)
}

func (r *RingTrackerServiceImpl) GetAllRelays(req types.QueryReq) dao.PageResult {
	return r.ringTrackerViewer.GetAllRelays(types.StrToCurrency(req.Currency), types.StrToSort(req.Sort), req.PageIndex, req.PageSize)
}

func (r *RingTrackerServiceImpl) GetAllDexs(req types.QueryReq) dao.PageResult {
	return r.ringTrackerViewer.GetAllDexs(types.StrToCurrency(req.Currency), types.StrToSort(req.Sort), req.PageIndex, req.PageSize)
}

func (r *RingTrackerServiceImpl) GetTokensByRelay(req types.QueryReq) ([]types.TokenFill, error) {
	return r.ringTrackerViewer.GetTokensByRelay(types.StrToCurrency(req.Currency), req.Relay)
}

func (r *RingTrackerServiceImpl) SetLegalTender() {
	r.ringTrackerViewer.SetEthTokenPrice()
	r.ringTrackerViewer.SetFullFills()
	//r.ringTrackerViewer.SetTokenPrices()
	viewer.ClearRingTrackerCache()
}
