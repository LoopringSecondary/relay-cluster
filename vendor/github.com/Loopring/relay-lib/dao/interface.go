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

import "fmt"

type RdsService interface {
	// create tables
	SetTables(tables []interface{})
	CreateTables() error

	// base functions
	Add(item interface{}) error
	Del(item interface{}) error
	First(item interface{}) error
	Last(item interface{}) error
	Save(item interface{}) error
	FindAll(item interface{}) error
}

// add single item
func (s *RdsServiceImpl) Add(item interface{}) error {
	return s.Db.Create(item).Error
}

// del single item
func (s *RdsServiceImpl) Del(item interface{}) error {
	return s.Db.Delete(item).Error
}

// select first item order by primary key asc
func (s *RdsServiceImpl) First(item interface{}) error {
	return s.Db.First(item).Error
}

// select the last item order by primary key asc
func (s *RdsServiceImpl) Last(item interface{}) error {
	return s.Db.Last(item).Error
}

// update single item
func (s *RdsServiceImpl) Save(item interface{}) error {
	return s.Db.Save(item).Error
}

// find all items in table where primary key > 0
func (s *RdsServiceImpl) FindAll(item interface{}) error {
	return s.Db.Table("lpr_orders").Find(item, s.Db.Where("id > ", 0)).Error
}

func (s *RdsServiceImpl) SetTables(tables []interface{}) {
	s.tables = tables
}

func (s *RdsServiceImpl) CreateTables() error {
	var tables []interface{}

	for _, t := range s.tables {
		if ok := s.Db.HasTable(t); !ok {
			if err := s.Db.CreateTable(t).Error; err != nil {
				return fmt.Errorf("create mysql table error:%s", err.Error())
			}
		}
	}

	// auto migrate to keep schema update to date
	// AutoMigrate will ONLY create tables, missing columns and missing indexes,
	// and WON'T change existing column's type or delete unused columns to protect your data
	return s.Db.AutoMigrate(tables...).Error
}
