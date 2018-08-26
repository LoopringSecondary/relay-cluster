package dao

import (
	"fmt"
)

type TokenPriceTrend struct {
	ID         int     `gorm:"column:id;primary_key;"`
	CoinName   string  `gorm:"column:coin_name;type:varchar(20)"`
	CoinAmount float64 `gorm:"column:coin_amount;type:float"`
	Time       int64   `gorm:"column:time;type:bigint"`
}

func (s *RdsService) CountPriceTrend() int64 {
	var count int64
	s.Db.Model(&TokenPriceTrend{}).Count(&count)
	return count
}

func (s *RdsService) AddTokenPriceTrends(trends []TokenPriceTrend) error {
	values := ""
	for i, trend := range trends {
		values += fmt.Sprintf("('%s', %f, %d)", trend.CoinName, trend.CoinAmount, trend.Time)
		if i != len(trends)-1 {
			values += ","
		}
	}
	return s.Db.Exec("INSERT IGNORE INTO `lpr_token_price_trends` (`coin_name`,`coin_amount`,`time`) VALUES " + values).Error
}

func (s *RdsService) GetPriceTrendMaxTime() int64 {
	var time int64
	s.Db.Model(&TokenPriceTrend{}).Select("max(time)").Row().Scan(&time)
	return time
}
