package pnl

import (
	"fmt"
	"log"
	"strconv"

	"github.com/hirokisan/bybit/v2"
	"github.com/kryptomind/BidBox-Trades/oogie/pkg/averaging"
	"github.com/kryptomind/BidBox-Trades/oogie/pkg/utils"
)

type PnlCron struct{}

func NewPnlCron() *PnlCron {
	return &PnlCron{}
}

func (pc *PnlCron) Run() {

	log.Println("--- PnL Cron started ---")
	client := bybit.NewTestClient().
		WithAuth(utils.Apikey, utils.Secretkey)

	resp, resp_map, err := GetBybitPositions(client)
	if err != nil {
		log.Fatal(err)
	}

	for _, v := range *resp {

		op_side := "long"
		if v.Side == bybit.SideBuy {
			op_side = "short"
		}

		op_pos, ok := resp_map[string(v.Symbol)+"_"+op_side]

		if !ok {
			continue
		}

		log.Println(string(v.Symbol) + " - " + v.UnrealisedPnl + " - " + op_pos.UnrealisedPnl)
		upnl, err := strconv.ParseFloat(v.UnrealisedPnl, 64)
		if err != nil {
			log.Fatal(err)
		}

		op_upnl, err := strconv.ParseFloat(op_pos.UnrealisedPnl, 64)
		if err != nil {
			log.Fatal(err)
		}

		markp, err := strconv.ParseFloat(v.MarkPrice, 64)
		if err != nil {
			log.Fatal(err)
		}

		size1, err := strconv.ParseFloat(v.Size, 64)
		if err != nil {
			log.Fatal(err)
		}

		size2, err := strconv.ParseFloat(op_pos.Size, 64)
		if err != nil {
			log.Fatal(err)
		}

		val1 := 0.005 * (size1 * markp)
		val2 := 0.005 * (size2 * markp)

		if upnl > val1 && size1 <= size2 {
			log.Println("long average")
			log.Println("Val: " + fmt.Sprintf("%f", val1) + " - U_pnl: " + v.UnrealisedPnl)
			averaging.HandleLongAveraging(string(v.Symbol), v.Size)
			continue
		} else if op_upnl > val2 && size1 < size2 {
			log.Println("long close")
			log.Println("Val: " + fmt.Sprintf("%f", -val2) + " - U_pnl: " + v.UnrealisedPnl)
			averaging.HandleCancel(string(v.Symbol), v.Size)
			continue
		}
		if op_upnl > val2 && size1 >= size2 {
			log.Println("short average")
			log.Println("Val: " + fmt.Sprintf("%f", val1) + " - U_pnl: " + v.UnrealisedPnl)
			averaging.HandleShortAveraging(string(v.Symbol), v.Size)
			continue
		} else if upnl > val1 && size1 > size2 {
			log.Println("short close")
			log.Println("Val: " + fmt.Sprintf("%f", -val2) + " - U_pnl: " + v.UnrealisedPnl)
			averaging.HandleCancel(string(v.Symbol), v.Size)
			continue
		}
	}

	log.Println("--- PnL Cron ended ---")

}

func GetBybitPositions(client *bybit.Client) (*[]bybit.V5GetPositionInfoItem, map[string]bybit.V5GetPositionInfoItem, error) {

	coin := bybit.CoinUSDT
	params := bybit.V5GetPositionInfoParam{
		Category:   bybit.CategoryV5Linear,
		SettleCoin: &coin,
	}

	resp, err := client.V5().Position().GetPositionInfo(params)
	if err != nil {
		return nil, nil, err
	}

	pos_map := make(map[string]bybit.V5GetPositionInfoItem, 0)
	list := make([]bybit.V5GetPositionInfoItem, 0)

	for _, v := range resp.Result.List {
		var side string
		if v.Side == bybit.SideBuy {
			side = "long"
			list = append(list, v)
		} else {
			side = "short"
		}
		val := string(v.Symbol) + "_" + side
		pos_map[val] = v
	}

	return &list, pos_map, nil
}

func CloseTrade(pos bybit.V5GetPositionInfoItem) {

	client := bybit.NewTestClient().
		WithAuth(utils.Apikey, utils.Secretkey)

	if pos.Side == bybit.SideBuy {
		// Stop Long Position
		reduce := true
		trigger := false
		time_to_force := bybit.TimeInForceGoodTillCancel

		pos_idx := bybit.PositionIdxHedgeBuy

		param1 := bybit.V5CreateOrderParam{
			Category:       bybit.CategoryV5Linear,
			Symbol:         bybit.SymbolV5(pos.Symbol),
			Side:           bybit.SideSell,
			OrderType:      bybit.OrderTypeMarket,
			Qty:            pos.Size,
			TimeInForce:    &time_to_force,
			ReduceOnly:     &reduce,
			CloseOnTrigger: &trigger,
			PositionIdx:    &pos_idx,
		}

		resp, err := client.V5().Order().CreateOrder(param1)
		if err != nil {
			log.Println(err)
		}

		log.Println(resp)
	} else if pos.Side == bybit.SideSell {

		reduce := true
		trigger := false
		time_to_force := bybit.TimeInForceGoodTillCancel

		pos_idx := bybit.PositionIdxHedgeSell

		param2 := bybit.V5CreateOrderParam{
			Category:       bybit.CategoryV5Linear,
			Symbol:         bybit.SymbolV5(pos.Symbol),
			Side:           bybit.SideBuy,
			OrderType:      bybit.OrderTypeMarket,
			Qty:            pos.Size,
			TimeInForce:    &time_to_force,
			ReduceOnly:     &reduce,
			CloseOnTrigger: &trigger,
			PositionIdx:    &pos_idx,
		}

		resp, err := client.V5().Order().CreateOrder(param2)
		if err != nil {
			log.Println(err)
		}

		log.Println(resp)

	}

}
