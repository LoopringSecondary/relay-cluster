package cache

import (
	"github.com/Loopring/relay-lib/cache"
	"github.com/ethereum/go-ethereum/common"
)

const (
	FillOwnerPrefix = "txm_fill_owner_"
	FillOwnerTtl    = 864000 // todo 临时数据,只存储10分钟,系统性宕机后无法重启后丢失?
)

func SetFillOwnerCache(txhash common.Hash, owner common.Address) error {
	key := generateFillOwnerKey(txhash)
	field := []byte(owner.Hex())
	return cache.SAdd(key, FillOwnerTtl, field)
}

func ExistFillOwnerCache(txhash common.Hash, owner common.Address) (bool, error) {
	key := generateFillOwnerKey(txhash)
	field := []byte(owner.Hex())
	return cache.SIsMember(key, field)
}

func generateFillOwnerKey(txhash common.Hash) string {
	return FillOwnerPrefix + txhash.Hex()
}
