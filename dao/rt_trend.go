package dao

import "fmt"

type RtTrend struct {
	Id     int     `gorm:"column:id;primary_key;"`
	Volume float64 `gorm:"column:volume;type:float;"`
	Fee    float64 `gorm:"column:fee;type:float;"`
	Trade  int     `gorm:"column:trade;type:int(11);"`
	Time   int64   `gorm:"column:time;type:bigint"`
}

func (s *RdsService) CountRtTrend() int64 {
	var count int64
	s.Db.Model(&RtTrend{}).Count(&count)
	return count
}

func (s *RdsService) AddRtTrends(trends []RtTrend) error {
	values := ""
	for i, trend := range trends {
		values += fmt.Sprintf("(%f, %f, %d, %d)", trend.Volume, trend.Fee, trend.Trade, trend.Time)
		if i != len(trends)-1 {
			values += ","
		}
	}
	return s.Db.Exec("INSERT IGNORE INTO `lpr_rt_trends`(`volume`, `fee`, `trade`, `time`) VALUES " + values).Error
}

func (s *RdsService) GetRtTrendMaxTime() int64 {
	var time int64
	s.Db.Model(&RtTrend{}).Select("max(time)").Row().Scan(&time)
	return time
}
