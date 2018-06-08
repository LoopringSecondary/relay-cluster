package cache

import (
	"github.com/Loopring/relay-cluster/dao"
	"github.com/Loopring/relay-lib/cache"
	"github.com/Loopring/relay-lib/types"
	"github.com/ethereum/go-ethereum/common"
	omtyp "github.com/Loopring/relay-cluster/ordermanager/types"
)

func BaseInfo(rds *dao.RdsService, orderhash common.Hash) (*types.OrderState, error) {
	state := &types.OrderState{}
	model, err := rds.GetOrderByHash(orderhash)
	if err != nil {
		return nil, err
	}

	model.ConvertUp(state)

	// todo(fuk):set in redis
	return state, nil
}

// miner & order owner
func HasOrderPermission(rds *dao.RdsService, owner common.Address) bool {
	ttl := int64(86400 * 10)

	key := "om_order_permission_" + owner.Hex()
	if ok, _ := cache.Exists(key); ok {
		return true
	}

	if !rds.IsOrderOwner(owner) && !rds.IsMiner(owner) {
		return false
	}

	cache.Set(key, []byte(""), ttl)
	return true
}

// todo
func SetPendingOrders(owner common.Address, orderhash common.Hash) error {
	return nil
}

// todo
func GetPendingOrders(owner common.Address) []common.Hash {
	var list []common.Hash

	return list
}

// todo
func GetOrderPendingTx(orderhash common.Hash) []omtyp.OrderTx {
	var list []omtyp.OrderTx
	return list
}
