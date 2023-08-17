package trade

import (
	"fmt"
	"log"

	"github.com/hirokisan/bybit/v2"
	"github.com/kryptomind/BidBox-Trades/oogie/pkg/utils"
)

type TradeCron struct{}

func NewTradeCron() *TradeCron {
	return &TradeCron{}
}

func (tc *TradeCron) Run() {
	log.Println("trade cron started")

	client := bybit.NewTestClient().
		WithAuth(utils.Apikey, utils.Secretkey)

	res, err := utils.GetPositionsTest()

	if err != nil {
		log.Fatal(err)
	}

	if len(res.Result.List)/2 >= utils.CoinLimit {
		log.Println("coin limit reached")
		return
	}

	symbol := utils.SelectRandomElement()

	first_order := utils.Amount

	size, _, err := utils.GetSizeBybit(string(symbol), first_order)
	fmt.Println(size)
	if err != nil {
		log.Fatal(err)
		return
	}

	if err := SwitchPositionMode(client); err != nil {
		log.Println(err)
		return
	}

	if err := SetLeverage(client, symbol); err != nil {
		log.Println(err)
	}

	var long_order_id string
	var short_order_id string
	var param1 bybit.V5CreateOrderParam
	var param2 bybit.V5CreateOrderParam
	reduce := false
	trigger := false
	time_to_force := bybit.TimeInForceGoodTillCancel
	var pos_idx bybit.PositionIdx

	// Long Position
	fmt.Println("long")
	pos_idx = bybit.PositionIdxHedgeBuy

	param1 = bybit.V5CreateOrderParam{
		Category:       bybit.CategoryV5Linear,
		Symbol:         bybit.SymbolV5(symbol),
		Side:           bybit.SideBuy,
		OrderType:      bybit.OrderTypeMarket,
		Qty:            fmt.Sprintf("%.3f", size),
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

	long_order_id = resp.Result.OrderID

	log.Println(long_order_id)

	// Short Position
	fmt.Println("short")
	pos_idx = bybit.PositionIdxHedgeSell
	param2 = bybit.V5CreateOrderParam{
		Category:       bybit.CategoryV5Linear,
		Symbol:         bybit.SymbolV5(symbol),
		Side:           bybit.SideSell,
		OrderType:      bybit.OrderTypeMarket,
		Qty:            fmt.Sprintf("%.3f", size),
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

	short_order_id = resp.Result.OrderID
	log.Println(short_order_id)

	log.Println("trade cron ended")

}

func SetLeverage(client *bybit.Client, symbol string) error {
	lev_param := bybit.V5SetLeverageParam{
		Category:     bybit.CategoryV5Linear,
		Symbol:       bybit.SymbolV5(symbol),
		BuyLeverage:  "20",
		SellLeverage: "20",
	}

	lev_resp, err := client.V5().Position().SetLeverage(lev_param)
	if err != nil {
		return err
	}

	log.Println(" Leverage Response - " + lev_resp.RetMsg)
	return nil
}

func SwitchPositionMode(client *bybit.Client) error {
	coin := bybit.CoinUSDT
	pos_params := bybit.V5SwitchPositionModeParam{
		Category: bybit.CategoryV5Linear,
		Mode:     bybit.PositionModeBothSides,
		Coin:     &coin,
	}
	res, err := client.V5().Position().SwitchPositionMode(pos_params)

	if err != nil {
		return err
	}
	log.Println(" Switch Position Mode Response: " + res.RetMsg)

	return nil
}
