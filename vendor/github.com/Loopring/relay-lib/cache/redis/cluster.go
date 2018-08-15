/*

  Copyright 2017 Loopring Project Ltd (Loopring Foundation).

  Licensed under the Apache License, Version 2.0 (the "License");
  you may not use this file except in compliance with the License.
  You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

  Unless required by applicable law or agreed to in writing, software
  distributed under the License is distributed on an "AS IS" BASIS,
  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
  See the License for the specific language governing permissions and
  limitations under the License.

*/

package redis

import (
	"fmt"
	"github.com/Loopring/relay-lib/log"
	"github.com/chasex/redis-go-cluster"
)

type ClusterOptions struct {
	StartNodes   []string
	ConnTimeout  int
	ReadTimeout  int
	WriteTimeout int
	KeepAlive    int
	AliveTime    int
}

type ClusterCacheImpl struct {
	options redis.Options
	cluster *redis.Cluster
}

func (impl *ClusterCacheImpl) Initialize(cfg interface{}) {
	options := cfg.(redis.Options)
	impl.options = options

	cluster, err := redis.NewCluster(&impl.options)

	if err != nil {
		log.Fatal(err.Error())
	}

	impl.cluster = cluster
}

func (impl *ClusterCacheImpl) Get(key string) ([]byte, error) {

	//log.Info("[REDIS-Get] key : " + key)

	cluster := impl.cluster
	defer cluster.Close()

	reply, err := cluster.Do("get", key)

	if nil != err {
		log.Errorf(" key:%s, err:%s", key, err.Error())
		return []byte{}, err
	} else if nil == reply {
		if nil == err {
			err = fmt.Errorf("no this key:%s", key)
		}
		return []byte{}, err
	} else {
		return reply.([]byte), err
	}
}

func (impl *ClusterCacheImpl) Exists(key string) (bool, error) {

	//log.Info("[REDIS-Exists] key : " + key)

	cluster := impl.cluster
	defer cluster.Close()

	reply, err := cluster.Do("exists", key)

	if err != nil {
		log.Errorf(" key:%s, err:%s", key, err.Error())
		return false, err
	} else {
		exists := reply.(int64)
		if exists == 1 {
			return true, nil
		} else {
			return false, nil
		}
	}
}

func (impl *ClusterCacheImpl) Set(key string, value []byte, ttl int64) error {

	//log.Info("[REDIS-Set] key : " + key)

	cluster := impl.cluster
	defer cluster.Close()

	if _, err := cluster.Do("set", key, value); err != nil {
		log.Errorf(" key:%s, err:%s", key, err.Error())
		return err
	}

	if ttl > 0 {
		if _, err := cluster.Do("expire", key, ttl); err != nil {
			log.Errorf(" key:%s, err:%s", key, err.Error())
			return err
		}
	}
	return nil
}

func (impl *ClusterCacheImpl) Del(key string) error {

	//log.Info("[REDIS-Del] key : " + key)

	cluster := impl.cluster
	defer cluster.Close()

	_, err := cluster.Do("del", key)
	if nil != err {
		log.Errorf(" key:%s, err:%s", key, err.Error())
	}
	return err
}

func (impl *ClusterCacheImpl) Dels(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("redis dels args empty")
	}

	//log.Info("[REDIS-Dels]")

	cluster := impl.cluster
	defer cluster.Close()

	var list []interface{}

	for _, v := range args {
		list = append(list, v)
	}

	num, err := cluster.Do("del", list...)
	if err != nil {
		log.Debugf("delete multi keys error:%s", err.Error())
	} else {
		log.Debugf("delete %d keys", num.(int64))
	}

	return nil
}

func (impl *ClusterCacheImpl) Keys(keyFormat string) ([][]byte, error) {

	//log.Info("[REDIS-Keys]")

	cluster := impl.cluster
	defer cluster.Close()

	vs := []interface{}{}
	vs = append(vs, keyFormat)
	reply, err := cluster.Do("keys", vs...)
	res := [][]byte{}
	if nil != err {
		log.Errorf(" key:%s, err:%s", keyFormat, err.Error())
	} else if nil == err && nil != reply {
		rs := reply.([]interface{})
		for _, r := range rs {
			if nil == r {
				res = append(res, []byte{})
			} else {
				res = append(res, r.([]byte))
			}
		}
	}
	return res, err
}

func (impl *ClusterCacheImpl) HMSet(key string, ttl int64, args ...[]byte) error {
	if len(args) == 0 {
		return fmt.Errorf("redis hmset args empty")
	}
	//log.Info("[REDIS-ZAdd] key : " + key)

	cluster := impl.cluster
	defer cluster.Close()

	if len(args)%2 != 0 {
		return fmt.Errorf("the length of `args` must be even")
	}

	vs := []interface{}{}
	vs = append(vs, key)
	for _, v := range args {
		vs = append(vs, v)
	}
	_, err := cluster.Do("hmset", vs...)
	if nil != err {
		log.Errorf(" key:%s, err:%s", key, err.Error())
	}
	if ttl > 0 {
		if _, err := cluster.Do("expire", key, ttl); err != nil {
			log.Errorf(" key:%s, err:%s", key, err.Error())
			return err
		}
	}
	return err
}

func (impl *ClusterCacheImpl) ZAdd(key string, ttl int64, args ...[]byte) error {
	if len(args) == 0 {
		return fmt.Errorf("redis zadd args empty")
	}
	//log.Info("[REDIS-ZAdd] key : " + key)

	cluster := impl.cluster
	defer cluster.Close()

	if len(args)%2 != 0 {
		return fmt.Errorf("the length of `args` must be even")
	}
	vs := []interface{}{}
	vs = append(vs, key)
	for _, v := range args {
		vs = append(vs, v)
	}
	_, err := cluster.Do("zadd", vs...)
	if nil != err {
		log.Errorf(" key:%s, err:%s", key, err.Error())
	}
	if ttl > 0 {
		if _, err := cluster.Do("expire", key, ttl); err != nil {
			log.Errorf(" key:%s, err:%s", key, err.Error())
			return err
		}
	}
	return err
}

func (impl *ClusterCacheImpl) HMGet(key string, fields ...[]byte) ([][]byte, error) {
	if len(fields) == 0 {
		return [][]byte{}, fmt.Errorf("redis hmget fields empty")
	}
	//log.Info("[REDIS-HMGet] key : " + key)

	cluster := impl.cluster
	defer cluster.Close()

	vs := []interface{}{}
	vs = append(vs, key)
	for _, v := range fields {
		vs = append(vs, v)
	}
	reply, err := cluster.Do("hmget", vs...)

	res := [][]byte{}
	if nil != err {
		log.Errorf(" key:%s, err:%s", key, err.Error())
	} else if nil == err && nil != reply {
		rs := reply.([]interface{})
		for _, r := range rs {
			if nil == r {
				res = append(res, []byte{})
			} else {
				res = append(res, r.([]byte))
			}
		}
	}
	return res, err
}

func (impl *ClusterCacheImpl) ZRange(key string, start, stop int64, withScores bool) ([][]byte, error) {

	//log.Info("[REDIS-ZRange] key : " + key)

	cluster := impl.cluster
	defer cluster.Close()

	vs := []interface{}{}
	vs = append(vs, key, start, stop)
	if withScores {
		vs = append(vs, []byte("WITHSCORES"))
	}
	reply, err := cluster.Do("ZRANGE", vs...)

	res := [][]byte{}
	if nil != err {
		log.Errorf(" key:%s, err:%s", key, err.Error())
	} else if nil == err && nil != reply {
		rs := reply.([]interface{})
		for _, r := range rs {
			if nil == r {
				res = append(res, []byte{})
			} else {
				res = append(res, r.([]byte))
			}
		}
	}
	return res, err
}

func (impl *ClusterCacheImpl) HDel(key string, fields ...[]byte) (int64, error) {
	if len(fields) == 0 {
		return 0, fmt.Errorf("redis hdel fields empty")
	}
	//log.Info("[REDIS-HDel] key : " + key)

	cluster := impl.cluster
	defer cluster.Close()

	vs := []interface{}{}
	vs = append(vs, key)
	for _, v := range fields {
		vs = append(vs, v)
	}
	reply, err := cluster.Do("hdel", vs...)

	if err != nil {
		log.Errorf(" key:%s, err:%s", key, err.Error())
		return 0, err
	} else {
		res := reply.(int64)
		return res, err
	}
}
func (impl *ClusterCacheImpl) SCard(key string) (int64, error) {

	//log.Info("[REDIS-SCARD] key : " + key)

	cluster := impl.cluster
	defer cluster.Close()

	vs := []interface{}{}
	vs = append(vs, key)
	reply, err := cluster.Do("scard", vs...)

	if err != nil {
		log.Errorf(" key:%s, err:%s", key, err.Error())
		return 0, err
	} else {
		res := reply.(int64)
		return res, err
	}
}

func (impl *ClusterCacheImpl) ZRemRangeByScore(key string, start, stop int64) (int64, error) {

	//log.Info("[REDIS-ZRemRangeByScore] key : " + key)

	cluster := impl.cluster
	defer cluster.Close()

	vs := []interface{}{}
	vs = append(vs, key, start, stop)

	reply, err := cluster.Do("ZREMRANGEBYSCORE", vs...)

	if err != nil {
		log.Errorf(" key:%s, err:%s", key, err.Error())
		return 0, err
	} else {
		res := reply.(int64)
		return res, err
	}
}

func (impl *ClusterCacheImpl) SRem(key string, members ...[]byte) (int64, error) {
	if len(members) == 0 {
		return 0, fmt.Errorf("redis srem members empty")
	}
	//log.Info("[REDIS-SRem] key : " + key)

	cluster := impl.cluster
	defer cluster.Close()

	vs := []interface{}{}
	vs = append(vs, key)
	for _, v := range members {
		vs = append(vs, v)
	}
	reply, err := cluster.Do("srem", vs...)

	if err != nil {
		log.Errorf(" key:%s, err:%s", key, err.Error())
		return 0, err
	} else {
		res := reply.(int64)
		return res, err
	}
}

func (impl *ClusterCacheImpl) SIsMember(key string, member []byte) (bool, error) {

	//log.Info("[REDIS-SIsMember] key : " + key)

	cluster := impl.cluster
	defer cluster.Close()

	reply, err := cluster.Do("sismember", key, member)
	if err != nil {
		log.Errorf("key:%s, err:%s", key, err.Error())
		return false, err
	} else {
		return reply.(int64) > 0, nil
	}
}

func (impl *ClusterCacheImpl) HGetAll(key string) ([][]byte, error) {

	//log.Info("[REDIS-HGetAll] key : " + key)

	cluster := impl.cluster
	defer cluster.Close()

	reply, err := cluster.Do("hgetall", key)

	res := [][]byte{}
	if nil != err {
		log.Errorf(" key:%s, err:%s", key, err.Error())
	} else if nil == err && nil != reply {
		rs := reply.([]interface{})
		for _, r := range rs {
			res = append(res, r.([]byte))
		}
	}
	return res, err
}
func (impl *ClusterCacheImpl) HVals(key string) ([][]byte, error) {

	//log.Info("[REDIS-HVals] key : " + key)

	cluster := impl.cluster
	defer cluster.Close()

	//todo:test nil result
	reply, err := cluster.Do("hvals", key)

	res := [][]byte{}
	if nil != err {
		log.Errorf(" key:%s, err:%s", key, err.Error())
	} else if nil == err && nil != reply {
		rs := reply.([]interface{})
		for _, r := range rs {
			res = append(res, r.([]byte))
		}
	}
	return res, err
}

func (impl *ClusterCacheImpl) HExists(key string, field []byte) (bool, error) {

	//log.Info("[REDIS-HExists] key : " + key)

	cluster := impl.cluster
	defer cluster.Close()

	reply, err := cluster.Do("hexists", key, field)
	if nil != err {
		log.Errorf(" key:%s, err:%s", key, err.Error())
	} else if nil == err && nil != reply {
		exists := reply.(int64)
		return exists > 0, nil
	}

	return false, err
}

func (impl *ClusterCacheImpl) SAdd(key string, ttl int64, members ...[]byte) error {
	if len(members) == 0 {
		return fmt.Errorf("redis sadd members empty")
	}
	//log.Info("[REDIS-SAdd] key : " + key)

	cluster := impl.cluster
	defer cluster.Close()

	vs := []interface{}{}
	vs = append(vs, key)
	for _, v := range members {
		vs = append(vs, v)
	}
	_, err := cluster.Do("sadd", vs...)
	if nil != err {
		log.Errorf(" key:%s, err:%s", key, err.Error())
	}
	if ttl > 0 {
		if _, err := cluster.Do("expire", key, ttl); err != nil {
			log.Errorf(" key:%s, err:%s", key, err.Error())
			return err
		}
	}
	return err
}

func (impl *ClusterCacheImpl) SMembers(key string) ([][]byte, error) {

	//log.Info("[REDIS-SMembers] key : " + key)

	cluster := impl.cluster
	defer cluster.Close()

	reply, err := cluster.Do("smembers", key)

	res := [][]byte{}
	if nil != err {
		log.Errorf(" key:%s, err:%s", key, err.Error())
	} else if nil == err && nil != reply {
		rs := reply.([]interface{})
		for _, r := range rs {
			res = append(res, r.([]byte))
		}
	}
	return res, err
}
