package dao_test

import (
	"github.com/Loopring/relay-cluster/dao"
	"github.com/Loopring/relay-cluster/test"
	"testing"
)

func TestCreate(t *testing.T) {
	rds := test.Rds()
	data := &dao.RingMinedEvent{TxHash: "0x123"}
	if err := rds.Db.Create(data).Error; err != nil {
		t.Fatalf(err.Error())
	} else {
		t.Log(data.ID)
	}
}
