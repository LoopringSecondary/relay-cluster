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

package node

import (
	"errors"
	"os"
	"reflect"

	"github.com/Loopring/extractor/extractor"
	"github.com/Loopring/relay-lib/cache/redis"
	"github.com/Loopring/relay-lib/cloudwatch"
	libdao "github.com/Loopring/relay-lib/dao"
	"github.com/Loopring/relay-lib/eth/accessor"
	lpraccssor "github.com/Loopring/relay-lib/eth/loopringaccessor"
	"github.com/Loopring/relay-lib/kafka"
	util "github.com/Loopring/relay-lib/marketutil"
	"github.com/Loopring/relay-lib/zklock"
	"github.com/naoina/toml"
	"go.uber.org/zap"
)

func LoadConfig(file string) *GlobalConfig {
	if "" == file {
		dir, _ := os.Getwd()
		file = dir + "/config/extractor.toml"
	}

	io, err := os.Open(file)
	if err != nil {
		panic(err)
	}
	defer io.Close()

	c := &GlobalConfig{}
	if err := toml.NewDecoder(io).Decode(c); err != nil {
		panic(err)
	}

	println(c.CloudWatch.Enabled, c.CloudWatch.Region)
	return c
}

type GlobalConfig struct {
	Title            string `required:"true"`
	Owner            string
	Log              zap.Config
	Mysql            libdao.MysqlOptions
	Redis            redis.RedisOptions
	ZkLock           zklock.ZkLockConfig
	Kafka            kafka.KafkaOptions
	Extractor        extractor.ExtractorOptions
	Accessor         accessor.AccessorOptions
	LoopringProtocol lpraccssor.LoopringProtocolOptions
	Market           util.MarketOptions
	CloudWatch       cloudwatch.CloudWatchConfig
}

func Validator(cv reflect.Value) (bool, error) {
	for i := 0; i < cv.NumField(); i++ {
		cvt := cv.Type().Field(i)

		if cv.Field(i).Type().Kind() == reflect.Struct {
			if res, err := Validator(cv.Field(i)); nil != err {
				return res, err
			}
		} else {
			if "true" == cvt.Tag.Get("required") {
				if !isSet(cv.Field(i)) {
					return false, errors.New("The field " + cvt.Name + " in config must be setted")
				}
			}
		}
	}

	return true, nil
}

func isSet(v reflect.Value) bool {
	switch v.Type().Kind() {
	case reflect.Invalid:
		return false
	case reflect.String:
		return v.String() != ""
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() != 0
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() != 0
	case reflect.Map:
		return len(v.MapKeys()) != 0
	}
	return true
}
