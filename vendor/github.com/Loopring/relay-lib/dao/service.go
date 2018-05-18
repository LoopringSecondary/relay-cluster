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

package dao

import (
	"github.com/Loopring/relay-lib/log"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"time"
)

type RdsServiceImpl struct {
	options *MysqlOptions
	tables  []interface{}
	Db      *gorm.DB
}

type MysqlOptions struct {
	Hostname           string
	Port               string
	User               string
	Password           string
	DbName             string
	TablePrefix        string
	MaxOpenConnections int
	MaxIdleConnections int
	ConnMaxLifetime    int
	Debug              bool
}

func NewRdsService(options *MysqlOptions) RdsServiceImpl {
	var impl RdsServiceImpl

	impl.options = options

	gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
		return options.TablePrefix + defaultTableName
	}

	url := options.User + ":" + options.Password + "@tcp(" + options.Hostname + ":" + options.Port + ")/" + options.DbName + "?charset=utf8&parseTime=True"
	db, err := gorm.Open("mysql", url)
	if err != nil {
		log.Fatalf("mysql connection error:%s", err.Error())
	}

	db.DB().SetConnMaxLifetime(time.Duration(options.ConnMaxLifetime) * time.Second)
	db.DB().SetMaxIdleConns(options.MaxIdleConnections)
	db.DB().SetMaxOpenConns(options.MaxOpenConnections)

	db.LogMode(options.Debug)

	impl.Db = db

	return impl
}
