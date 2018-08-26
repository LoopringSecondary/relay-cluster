package cache

import (
	"encoding/json"
	"github.com/Loopring/relay-lib/log"
	"github.com/Loopring/relay-lib/cache"
)

func GetFromCache(key string, res interface{}) bool {
	if data, err := cache.Get(key); err == nil && len(data) > 0 {
		json.Unmarshal(data, res)
		log.Debugf("[get from cache]key:" + key + "")
		return true
	}
	return false
}

func SaveCache(key string, res interface{}) {
	if data, err := json.Marshal(res); err == nil && len(data) > 0 {
		cache.Set(key, data, 7*24*3600)
		log.Debugf("[save cache]key:" + key + "")
	}
}
