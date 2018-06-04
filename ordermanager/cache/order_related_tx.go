package cache

import "github.com/ethereum/go-ethereum/common"

// 该模块缓存管理当前拥有pending状态订单用户的nonce值
// 该模块存储的操作包含两种情况:
// 1.与订单相关的tx
// submitRing/cancel/cutoff/cutoffPair pending时添加
// 上述event failed/success时删除
// 2.与订单不相关的tx
// transfer/approve...事件过来时,如果用户有pending的tx,
// 判断该pending tx是否需要设置为failed

type OrderRelatedTx struct {
	Owner     common.Address `json:"owner"`
	TxHash    common.Hash    `json:"tx_hash"`
	OrderHash common.Hash    `json:"order_hash"`
	Nonce     int64          `json:"nonce"`
}

func AddPendingOrderRelatedTx(tx OrderRelatedTx) error {
	return nil
}

func GetPendingOrderRelatedTx(owner common.Address, nonce int64) []OrderRelatedTx {
	return nil
}

func DelPendingOrderRelatedTx() error {
	return nil
}
