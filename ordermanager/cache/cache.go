package cache

import (
	"github.com/Loopring/relay-cluster/dao"
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
