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
	libdao "github.com/Loopring/relay-lib/dao"
	"github.com/Loopring/relay-lib/log"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

type RdsService struct {
	libdao.RdsServiceImpl
}

func NewDb(options *libdao.MysqlOptions) RdsService {
	var s RdsService

	s.RdsServiceImpl = libdao.NewRdsService(options)

	tables := []interface{}{}
	tables = append(tables, &Block{})
	s.SetTables(tables)
	if err := s.CreateTables(); err != nil {
		log.Fatalf(err.Error())
	}

	return s
}
