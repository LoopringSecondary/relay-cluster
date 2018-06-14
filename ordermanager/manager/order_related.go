package manager

import (
	"fmt"
	"github.com/Loopring/relay-cluster/dao"
	omcm "github.com/Loopring/relay-cluster/ordermanager/common"
	notify "github.com/Loopring/relay-cluster/util"
	"github.com/Loopring/relay-lib/eventemitter"
	"github.com/Loopring/relay-lib/log"
	util "github.com/Loopring/relay-lib/marketutil"
	"github.com/Loopring/relay-lib/types"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

// 所有来自gateway的订单都是新订单
func HandleGatewayOrder(state *types.OrderState) error {
	model, err := NewOrderEntity(state, nil)
	if err != nil {
		return fmt.Errorf("order manager,handle gateway order:%s error", state.RawOrder.Hash.Hex())
	}

	if err = rds.Add(model); err != nil {
		log.Errorf(err.Error())
		return err
	}

	log.Debugf("order manager,handle gateway order,order.hash:%s amountS:%s", state.RawOrder.Hash.Hex(), state.RawOrder.AmountS.String())

	notify.NotifyOrderUpdate(state)

	return nil
}

func HandleSubmitRingMethodEvent(event *types.SubmitRingMethodEvent) error {

	if event.Status != types.TX_STATUS_PENDING && event.Status != types.TX_STATUS_FAILED {
		log.Errorf("order manager, submitRingHandler, tx:%s, txstatus:%s invalid", event.TxHash.Hex(), types.StatusStr(event.Status))
		return nil
	}

	// validate and save event
	var (
		model = &dao.RingMinedEvent{}
		err   error
	)

	//if IsEventExist(event.Status, model.Status, err) {
	//	log.Debugf("order manager, submitRingHandler, tx:%s, txstatus:%s already exist", event.TxHash.Hex(), types.StatusStr(event.Status))
	//	return nil
	//}

	//if eventStatus == types.TX_STATUS_PENDING && findModelErr == nil {
	//	return true
	//}
	//if eventStatus != types.TX_STATUS_PENDING && findModelErr == nil && uint8(eventStatus) == modelStatus {
	//	return true
	//}

	if model, err = rds.FindRingMined(event.TxHash.Hex()); err != nil {
		model.FromSubmitRingMethod(event)
		err = rds.Add(model)
	} else {
		if event.Status == types.TX_STATUS_PENDING {
			return fmt.Errorf("")
		}
		model.FromSubmitRingMethod(event)
		err = rds.Save(model)
	}

	if err != nil {
		return err
	}

	log.Debugf("order manager, submitRingHandler, tx:%s, txstatus:%s", event.TxHash.Hex(), types.StatusStr(event.Status))

	for _, v := range event.OrderList {
		txhandler := FullOrderTxHandler(event.TxInfo, v.Hash, types.ORDER_PENDING)
		txhandler.HandlerOrderRelatedTx()
	}

	return nil
}

func HandleRingMinedEvent(event *types.RingMinedEvent) error {
	if event.Status != types.TX_STATUS_SUCCESS {
		return nil
	}

	model, err := rds.FindRingMined(event.TxHash.Hex())
	if IsEventExist(event.Status, model.Status, err); err != nil {
		log.Debugf("order manager, ringMinedHandler, tx:%s, txstatus:%s, err:%s", event.TxHash.Hex(), types.StatusStr(event.Status), err.Error())
		return nil
	}

	log.Debugf("order manager, ringMinedHandler, tx:%s, txstatus:%s", event.TxHash.Hex(), types.StatusStr(event.Status))
	model.ConvertDown(event)
	if err != nil {
		return rds.Add(model)
	} else {
		return rds.Save(model)
	}
}

func HandleOrderFilledEvent(event *types.OrderFilledEvent) error {
	// save fill event
	_, err := rds.FindFillEvent(event.TxHash.Hex(), event.FillIndex.Int64())
	if err == nil {
		return fmt.Errorf("order manager fillHandler, tx:%s, fillIndex:%s, orderhash:%s, err:fill already exist", event.TxHash.Hex(), event.FillIndex.String(), event.OrderHash.Hex())
	}

	// get rds.Order and types.OrderState
	state := &types.OrderState{UpdatedBlock: event.BlockNumber}
	model, err := rds.GetOrderByHash(event.OrderHash)
	if err != nil {
		return err
	}
	if err := model.ConvertUp(state); err != nil {
		return err
	}

	newFillModel := &dao.FillEvent{}
	newFillModel.ConvertDown(event)
	newFillModel.Fork = false
	newFillModel.OrderType = state.RawOrder.OrderType
	newFillModel.Side = util.GetSide(util.AddressToAlias(event.TokenS.Hex()), util.AddressToAlias(event.TokenB.Hex()))
	newFillModel.Market, _ = util.WrapMarketByAddress(event.TokenB.Hex(), event.TokenS.Hex())

	if err := rds.Add(newFillModel); err != nil {
		return err
	}

	// judge order status
	if state.Status == types.ORDER_CUTOFF || state.Status == types.ORDER_FINISHED || state.Status == types.ORDER_UNKNOWN {
		return fmt.Errorf("order manager fillHandler, tx:%s, fillIndex:%s, orderhash:%s, err:order status(%d) invalid", event.TxHash.Hex(), event.FillIndex.String(), event.OrderHash.Hex(), state.Status)
	}

	// calculate dealt amount
	state.UpdatedBlock = event.BlockNumber
	state.DealtAmountS = new(big.Int).Add(state.DealtAmountS, event.AmountS)
	state.DealtAmountB = new(big.Int).Add(state.DealtAmountB, event.AmountB)
	state.SplitAmountS = new(big.Int).Add(state.SplitAmountS, event.SplitS)
	state.SplitAmountB = new(big.Int).Add(state.SplitAmountB, event.SplitB)

	// update order status
	SettleOrderStatus(state, false)

	// update rds.Order
	if err := model.ConvertDown(state); err != nil {
		return err
	}
	if err := rds.UpdateOrderWhileFill(state.RawOrder.Hash, state.Status, state.DealtAmountS, state.DealtAmountB, state.SplitAmountS, state.SplitAmountB, state.UpdatedBlock); err != nil {
		return err
	}

	// update orderTx
	txhandler := FullOrderTxHandler(event.TxInfo, state.RawOrder.Hash, types.ORDER_PENDING)
	txhandler.HandlerOrderRelatedTx()

	log.Debugf("order manager fillHandler, tx:%s, fillIndex:%s, orderhash:%s, dealAmountS:%s, dealtAmountB:%s", event.TxHash.Hex(), event.FillIndex.String(), event.OrderHash.Hex(), state.DealtAmountS.String(), state.DealtAmountB.String())

	notify.NotifyOrderFilled(newFillModel)

	return nil
}

func HandleOrderCancelledEvent(event *types.OrderCancelledEvent) error {
	eventModel, noRecordErr := rds.GetCancelEvent(event.TxHash)
	if err := ValidateDuplicateEvent(event.Status, eventModel.Status, noRecordErr); err != nil {
		return fmt.Errorf("order manager orderCancelHandler, tx:%s, orderhash:%s, txstatus:%s, err:%s", event.TxHash.Hex(), event.OrderHash.Hex(), types.StatusStr(event.Status), err.Error())
	}

	log.Debugf("order manager orderCancelHandler, tx:%s, orderhash:%s, txstatus:%s", event.TxHash.Hex(), event.OrderHash.Hex(), types.StatusStr(event.Status))

	eventModel.ConvertDown(event)
	eventModel.Fork = false

	if noRecordErr != nil {
		rds.Add(&eventModel)
	} else {
		rds.Save(&eventModel)
	}

	if event.Status == types.TX_STATUS_SUCCESS {
		// get rds.Order and types.OrderState
		state := &types.OrderState{}
		model, err := rds.GetOrderByHash(event.OrderHash)
		if err != nil {
			return err
		}
		if err := model.ConvertUp(state); err != nil {
			return err
		}

		// calculate remainAmount and cancelled amount should be saved whether order is finished or not
		if state.RawOrder.BuyNoMoreThanAmountB {
			state.CancelledAmountB = new(big.Int).Add(state.CancelledAmountB, event.AmountCancelled)
			log.Debugf("order manager orderCancelHandler, tx:%s, orderhash:%s, cancelledAmountB:%s", event.TxHash.Hex(), event.OrderHash.Hex(), state.CancelledAmountB.String())
		} else {
			state.CancelledAmountS = new(big.Int).Add(state.CancelledAmountS, event.AmountCancelled)
			log.Debugf("order manager orderCancelHandler, tx:%s, orderhash:%s, cancelledAmountS:%s", event.TxHash.Hex(), event.OrderHash.Hex(), state.CancelledAmountS.String())
		}

		// update order status
		SettleOrderStatus(state, true)
		state.UpdatedBlock = event.BlockNumber

		// update rds.Order
		if err := model.ConvertDown(state); err != nil {
			return err
		}
		if err := rds.UpdateOrderWhileCancel(state.RawOrder.Hash, state.Status, state.CancelledAmountS, state.CancelledAmountB, state.UpdatedBlock); err != nil {
			return err
		}

		notify.NotifyOrderUpdate(state)
	}

	// 原则上不允许订单状态干扰到其他动作
	txhandler := FullOrderTxHandler(event.TxInfo, event.OrderHash, types.ORDER_CANCELLING)
	return txhandler.HandlerOrderRelatedTx()
}

func HandleCutoffEvent(event *types.CutoffEvent) error {
	var orderhashList []common.Hash

	model, noRecordErr := rds.GetCutoffEvent(event.TxHash)
	if err := ValidateDuplicateEvent(event.Status, model.Status, noRecordErr); err != nil {
		return fmt.Errorf("order manager, CutoffHandler, tx:%s, err:%s", event.TxHash.Hex(), err.Error())
	}

	// insert
	if noRecordErr != nil {
		orders, _ := rds.GetCutoffOrders(event.Owner, event.Cutoff, omcm.ValidCutoffStatus)
		for _, v := range orders {
			var state types.OrderState
			v.ConvertUp(&state)
			orderhashList = append(orderhashList, state.RawOrder.Hash)
		}
		model.Fork = false
		event.OrderHashList = orderhashList
		model.ConvertDown(event)
		rds.Add(&model)
	} else {
		orderhashList = dao.UnmarshalStrToHashList(model.OrderHashList)
		event.OrderHashList = orderhashList
		model.ConvertDown(event)
		rds.Save(&model)
	}

	log.Debugf("order manager, CutoffHandler, tx:%s, owner:%s, cutofftime:%s, txstatus:%s", event.TxHash.Hex(), event.Owner.Hex(), event.Cutoff.String(), types.StatusStr(event.Status))

	if event.Status == types.TX_STATUS_SUCCESS {
		// 首次存储到缓存，lastCutoff == currentCutoff
		lastCutoff := cutoffcache.GetCutoff(event.Protocol, event.Owner)
		if event.Cutoff.Cmp(lastCutoff) < 0 {
			return fmt.Errorf("order manager, CutoffHandler, tx:%s, lastCutofftime:%s > currentCutoffTime:%s", event.TxHash.Hex(), lastCutoff.String(), event.Cutoff.String())
		}

		cutoffcache.UpdateCutoff(event.Protocol, event.Owner, event.Cutoff)
		rds.SetCutOffOrders(orderhashList, event.BlockNumber)

		notify.NotifyCutoff(event)

		for _, orderhash := range orderhashList {
			txhandler := FullOrderTxHandler(event.TxInfo, orderhash, types.ORDER_CUTOFFING)
			txhandler.HandlerOrderRelatedTx()
		}
	}

	return nil
}

func HandleCutoffPair(event *types.CutoffPairEvent) error {
	var orderhashlist []common.Hash

	model, noRecordErr := rds.GetCutoffPairEvent(event.TxHash)
	if err := ValidateDuplicateEvent(event.Status, model.Status, noRecordErr); err != nil {
		return fmt.Errorf("order manager cutoffPairHandler, tx:%s, err:%s", event.TxHash.Hex(), err.Error())
	}

	if noRecordErr != nil {
		orders, _ := rds.GetCutoffPairOrders(event.Owner, event.Token1, event.Token2, event.Cutoff, omcm.ValidCutoffStatus)
		for _, v := range orders {
			var state types.OrderState
			v.ConvertUp(&state)
			orderhashlist = append(orderhashlist, state.RawOrder.Hash)
		}
		model.Fork = false
		event.OrderHashList = orderhashlist
		model.ConvertDown(event)
		rds.Add(&model)
	} else {
		orderhashlist = dao.UnmarshalStrToHashList(model.OrderHashList)
		event.OrderHashList = orderhashlist
		model.ConvertDown(event)
		rds.Save(&model)
	}

	log.Debugf("order manager cutoffPairHandler, tx:%s, owner:%s, token1:%s, token2:%s, cutoffTimestamp:%s, txstatus:%s", event.TxHash.Hex(), event.Owner.Hex(), event.Token1.Hex(), event.Token2.Hex(), event.Cutoff.String(), types.StatusStr(event.Status))

	if event.Status == types.TX_STATUS_SUCCESS {
		// 首次存储到缓存，lastCutoffPair == currentCutoffPair
		lastCutoffPair := cutoffcache.GetCutoffPair(event.Protocol, event.Owner, event.Token1, event.Token2)
		if event.Cutoff.Cmp(lastCutoffPair) < 0 {
			return fmt.Errorf("order manager cutoffPairHandler, tx:%s, lastCutoffPairTime:%s > currentCutoffPairTime:%s", event.TxHash.Hex(), lastCutoffPair.String(), event.Cutoff.String())
		}

		cutoffcache.UpdateCutoffPair(event.Protocol, event.Owner, event.Token1, event.Token2, event.Cutoff)
		rds.SetCutOffOrders(orderhashlist, event.BlockNumber)

		notify.NotifyCutoffPair(event)
	}

	for _, orderhash := range orderhashlist {
		txhandler := FullOrderTxHandler(event.TxInfo, orderhash, types.ORDER_CUTOFFING)
		txhandler.HandlerOrderRelatedTx()
	}

	return nil
}
