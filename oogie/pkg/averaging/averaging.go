package averaging

import (
	"fmt"
	"log"
	"strconv"
	"sync"

	"github.com/hirokisan/bybit/v2"
	"github.com/kryptomind/BidBox-Trades/oogie/pkg/utils"
)

func HandleLongAveraging(symbol string, size string) {

	// Stop Long Position
	reduce := true
	trigger := false
	time_to_force := bybit.TimeInForceGoodTillCancel

	pos_idx := bybit.PositionIdxHedgeBuy

	client := bybit.NewTestClient().
		WithAuth(utils.Apikey, utils.Secretkey)

	param1 := bybit.V5CreateOrderParam{
		Category:       bybit.CategoryV5Linear,
		Symbol:         bybit.SymbolV5(symbol),
		Side:           bybit.SideSell,
		OrderType:      bybit.OrderTypeMarket,
		Qty:            size,
		TimeInForce:    &time_to_force,
		ReduceOnly:     &reduce,
		CloseOnTrigger: &trigger,
		PositionIdx:    &pos_idx,
	}

	resp, err := client.V5().Order().CreateOrder(param1)
	if err != nil {
		log.Println(err)
		return
	}

	log.Println(resp)

	var wg sync.WaitGroup

	wg.Add(2)

	// Start Long Position
	go func() {
		defer wg.Done()
		pos_idx := bybit.PositionIdxHedgeBuy
		reduce := false

		param2 := bybit.V5CreateOrderParam{
			Category:       bybit.CategoryV5Linear,
			Symbol:         bybit.SymbolV5(symbol),
			Side:           bybit.SideBuy,
			OrderType:      bybit.OrderTypeMarket,
			Qty:            size,
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
	}()

	// Start Short Position
	go func() {
		defer wg.Done()
		pos_idx = bybit.PositionIdxHedgeSell
		reduce = false

		fl_size, err := strconv.ParseFloat(size, 64)
		if err != nil {
			log.Fatal(err)
		}

		u_size := 0.2 * fl_size

		param3 := bybit.V5CreateOrderParam{
			Category:       bybit.CategoryV5Linear,
			Symbol:         bybit.SymbolV5(symbol),
			Side:           bybit.SideSell,
			OrderType:      bybit.OrderTypeMarket,
			Qty:            fmt.Sprintf("%f", u_size),
			TimeInForce:    &time_to_force,
			ReduceOnly:     &reduce,
			CloseOnTrigger: &trigger,
			PositionIdx:    &pos_idx,
		}

		resp, err = client.V5().Order().CreateOrder(param3)
		if err != nil {
			log.Println(err)
		}

		log.Println(resp)
	}()

	wg.Wait()
}

func HandleShortAveraging(symbol string, size string) {
	// Stop Short Position
	reduce := true
	trigger := false
	time_to_force := bybit.TimeInForceGoodTillCancel

	pos_idx := bybit.PositionIdxHedgeSell

	client := bybit.NewTestClient().
		WithAuth(utils.Apikey, utils.Secretkey)

	param1 := bybit.V5CreateOrderParam{
		Category:       bybit.CategoryV5Linear,
		Symbol:         bybit.SymbolV5(symbol),
		Side:           bybit.SideBuy,
		OrderType:      bybit.OrderTypeMarket,
		Qty:            size,
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

	var wg sync.WaitGroup

	wg.Add(2)

	// Start Short Position

	go func() {
		defer wg.Done()
		pos_idx := bybit.PositionIdxHedgeSell
		reduce := false

		param2 := bybit.V5CreateOrderParam{
			Category:       bybit.CategoryV5Linear,
			Symbol:         bybit.SymbolV5(symbol),
			Side:           bybit.SideSell,
			OrderType:      bybit.OrderTypeMarket,
			Qty:            size,
			TimeInForce:    &time_to_force,
			ReduceOnly:     &reduce,
			CloseOnTrigger: &trigger,
			PositionIdx:    &pos_idx,
		}

		resp, err = client.V5().Order().CreateOrder(param2)
		if err != nil {
			log.Println(err)
		}

		log.Println(resp)
	}()

	// Start Long Position

	go func() {
		defer wg.Done()

		pos_idx := bybit.PositionIdxHedgeBuy
		reduce := false

		fl_size, err := strconv.ParseFloat(size, 64)
		if err != nil {
			log.Fatal(err)
		}

		u_size := 0.2 * fl_size

		param3 := bybit.V5CreateOrderParam{
			Category:       bybit.CategoryV5Linear,
			Symbol:         bybit.SymbolV5(symbol),
			Side:           bybit.SideBuy,
			OrderType:      bybit.OrderTypeMarket,
			Qty:            fmt.Sprintf("%f", u_size),
			TimeInForce:    &time_to_force,
			ReduceOnly:     &reduce,
			CloseOnTrigger: &trigger,
			PositionIdx:    &pos_idx,
		}

		resp, err := client.V5().Order().CreateOrder(param3)
		if err != nil {
			log.Println(err)
		}

		log.Println(resp)
	}()

	wg.Wait()
}

func HandleCancel(symbol string, size string) {

	buy, sell := utils.GetPositions(utils.Apikey, utils.Secretkey, symbol)
	reduce := true
	trigger := false
	time_to_force := bybit.TimeInForceGoodTillCancel

	client := bybit.NewTestClient().
		WithAuth(utils.Apikey, utils.Secretkey)

	var wg sync.WaitGroup

	wg.Add(2)

	// Stop Long Position
	go func() {
		defer wg.Done()

		pos_idx := bybit.PositionIdxHedgeBuy

		param1 := bybit.V5CreateOrderParam{
			Category:       bybit.CategoryV5Linear,
			Symbol:         bybit.SymbolV5(symbol),
			Side:           bybit.SideSell,
			OrderType:      bybit.OrderTypeMarket,
			Qty:            buy,
			TimeInForce:    &time_to_force,
			ReduceOnly:     &reduce,
			CloseOnTrigger: &trigger,
			PositionIdx:    &pos_idx,
		}

		resp1, err := client.V5().Order().CreateOrder(param1)
		if err != nil {
			log.Println(err)
		}

		log.Println(resp1)
	}()

	// Stop Short Position
	go func() {
		defer wg.Done()
		pos_idx := bybit.PositionIdxHedgeSell

		param2 := bybit.V5CreateOrderParam{
			Category:       bybit.CategoryV5Linear,
			Symbol:         bybit.SymbolV5(symbol),
			Side:           bybit.SideBuy,
			OrderType:      bybit.OrderTypeMarket,
			Qty:            sell,
			TimeInForce:    &time_to_force,
			ReduceOnly:     &reduce,
			CloseOnTrigger: &trigger,
			PositionIdx:    &pos_idx,
		}

		resp2, err := client.V5().Order().CreateOrder(param2)
		if err != nil {
			log.Println(err)
		}

		log.Println(resp2)
	}()

	wg.Wait()
}
