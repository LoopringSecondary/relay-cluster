package cache

import (
	"github.com/Loopring/relay-cluster/dao"
	"github.com/Loopring/relay-lib/cache"
	"github.com/Loopring/relay-lib/types"
	"github.com/ethereum/go-ethereum/common"
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

	key := "om_order_owner_" + owner.Hex()
	if ok, _ := cache.Exists(key); ok {
		return true
	}

	if !rds.IsOrderOwner(owner) && !rds.IsMiner(owner) {
		return false
	}

	cache.Set(key, []byte(""), ttl)
	return true
}
