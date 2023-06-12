package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"

	requests "github.com/amir-the-h/okex/requests/rest/trade"
	"github.com/jinzhu/gorm"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/amir-the-h/okex"
	"github.com/amir-the-h/okex/api"
	"github.com/kryptomind/BidBox-Trades/models"
	"github.com/kryptomind/BidBox-Trades/response"
	"github.com/kryptomind/BidBox-Trades/utils"
	log "github.com/sirupsen/logrus"
)

var mutex sync.Mutex

type Account struct {
	Code        string `json:"code"`
	Msg         string `json:"msg"`
	RequestTime int64  `json:"requestTime"`
	Data        []struct {
		MarginCoin        string `json:"marginCoin"`
		Locked            string `json:"locked"`
		Available         string `json:"available"`
		CrossMaxAvailable string `json:"crossMaxAvailable"`
		FixedMaxAvailable string `json:"fixedMaxAvailable"`
		MaxTransferOut    string `json:"maxTransferOut"`
		Equity            string `json:"equity"`
		UsdtEquity        string `json:"usdtEquity"`
		BtcEquity         string `json:"btcEquity"`
		CrossRiskRate     string `json:"crossRiskRate"`
		UnrealizedPL      string `json:"unrealizedPL"`
		Bonus             string `json:"bonus"`
	} `json:"data"`
}

type User struct {
	Capital      int
	Trade_amount float64
	First_order  float64
}

type TradeRequest struct {
	CoinPair string `json:"coin_pair"`
	Long     int    `json:"long"`
	Exchange string `json:"exchange"`
}

const (
	TAKE_PROFIT_PERCENTAGE = 80
	STOP_LOSS_PERCENTAGE   = 80
)

func (s *Server) handleUpdate(key *models.Key, val int, pos string) {
	key, err := key.ChangePositions(s.DB, val, pos)
	if err != nil {
		log.Error(err)
		fmt.Println("error while updating positions")
		fmt.Println(err)
		return
	}
	log.Println(key)
	// fmt.Println("succes updated positions")
}

func (server *Server) Home(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, "Trade Service")
}
func GetExchangeSpecificKeys(server *Server, service string) []models.Key {
	key := models.Key{}
	keys, err := key.FindKeysByService(server.DB, service)
	if err != nil {
		log.Fatal("error getting keys")
		emptyKeys := []models.Key{}
		return emptyKeys
	}

	log.Info("retrieved keys")
	return *keys
}
func (s *Server) StartTrade(w http.ResponseWriter, r *http.Request) {
	res := make(map[string]string)

	trade_req := &TradeRequest{}
	err := json.NewDecoder(r.Body).Decode(trade_req)
	if err != nil {
		response.ERROR(w, http.StatusBadRequest, err)
		return
	}
	keys := GetExchangeSpecificKeys(s, trade_req.Exchange)
	fmt.Println("keys", keys)
	err = trade_req.Validate()
	if err != nil {
		response.ERROR(w, http.StatusBadRequest, err)
		return
	}
	fmt.Println("started")
	//	stop_loss := 0.8 * float64(300)
	//take_profit := 1.1 * float64(300)
	for i, v := range keys { // for each user key
		go func(v models.Key, i int) {
			if v.Service == trade_req.Exchange {
				fmt.Println("user: ", v.UserEmail)

				if trade_req.Long == 1 && v.OpenLong <= 0 {
					return
				}
				if trade_req.Long == 0 && v.OpenShort <= 0 {
					return
				}
				val := int(math.Floor(float64(v.TradeAmount)/100) * 100)
				cond := models.Conditions{}
				c, err := cond.FindCondition(s.DB, val)
				if err != nil {
					log.Error(err)
					return
				}
				first_order := float64(v.TradeAmount) * 0.08 / float64(c.Positions)
				switch trade_req.Exchange {
				case "bitget":
					go bitgetTrade(s, v, i, trade_req, first_order, res, keys)
				case "binance":
					go binanceTrade(s, v, i, trade_req, first_order, res, w, keys)
				case "okex":
					go okexTrade(s, v, i, trade_req, first_order, res, keys)
				case "bybit":
					go bybitTrade(s, v, i, trade_req, first_order, res, keys)
				default:
					log.Println("Invalid exchange specified")
				}
			}
		}(v, i)
	}

	response.JSON(w, http.StatusOK, "trades made successfully")

}

func bitgetTrade(s *Server, v models.Key, i int, trade_req *TradeRequest, first_order float64, res map[string]string, keys []models.Key) {
	api_key, secret_key, passphrase, err := utils.DecryptKeys(v.ApiKey, v.SecretKey, v.Passphrase, "bitget")
	if err != nil {
		fmt.Println("error in decryptkeys: ", err)
		return
	}

	order := models.OrderRequest{}
	orderResp := models.OrderResponse{}

	if trade_req.Long == 1 {
		order.Side = "open_long"
	} else {
		order.Side = "open_short"
	}

	order.Symbol = trade_req.CoinPair
	order.MarginCoin = "SUSDT"
	size, p, err := utils.GetSize(order.Symbol, first_order)
	if err != nil {
		log.Fatal(err)
		// fmt.Println(err, " for ", v.UserEmail)
		return
	}
	order.Size = fmt.Sprintf("%f", size)
	order.OrderType = "market"
	// order.StopLoss = fmt.Sprintf("%F", (size * 0.8))
	// fmt.Println("order size : " + order.Size)
	// fmt.Println(order)
	// order.TakeProfit = fmt.Sprintf("%F", (size * 1.1))
	// // fmt.Println(order.TakeProfit)
	go func() {
		str, err := NewOrder(api_key, secret_key, passphrase, &order)
		if err != nil {
			log.Error("failed order: " + err.Error() + " for " + v.UserEmail)
			fmt.Println("failed order: ", err, " for ", v.UserEmail)
			return
		}

		if err = json.Unmarshal([]byte(str), &orderResp); err != nil {
			log.Error("parse fail: " + err.Error() + " for " + v.UserEmail)
			return
		}

		if orderResp.Code != "00000" {
			log.Error(orderResp.Msg + " :: " + v.UserEmail)
			return
		}

		go fetchAndUpdateBitgetPosition(order, v, s.DB, api_key, secret_key, passphrase)

		sp := strings.Split(order.Side, "_")

		new_order := models.Order{
			Email:       v.UserEmail,
			Symbol:      order.Symbol,
			Size:        order.Size,
			Side:        sp[1],
			MarginCoin:  order.MarginCoin,
			OrderType:   order.OrderType,
			Service:     "bitget",
			QuoteAmount: p,
		}
		_, error := new_order.SaveOrder(s.DB)
		if err != nil {
			log.Error("err saving order: " + error.Error() + " for " + v.UserEmail)
			return
		}
		mutex.Lock()
		if order.Side == "open_short" {
			keys[i].OpenShort = v.OpenShort - 1
			keys[i].Prev = "short"
			go s.handleUpdate(&v, keys[i].OpenShort, "short")
		} else if order.Side == "open_long" {
			keys[i].OpenLong = v.OpenLong - 1
			keys[i].Prev = "long"
			go s.handleUpdate(&v, keys[i].OpenLong, "long")
		}
		mutex.Unlock()
		res["client_id"] = orderResp.Data.ClientOid
		res["order_id"] = orderResp.Data.OrderID
		// fmt.Println("res order placed for " + v.UserEmail)
		// fmt.Println(res)
	}()

}

func fetchAndUpdateBitgetPosition(order models.OrderRequest, v models.Key, db *gorm.DB, api_key string, secret_key string, passphrase string) {
	positionsResponse, err := utils.PerformBitgetPositionQuery(api_key, secret_key, passphrase, order.Symbol)
	if err != nil {
		fmt.Println("---error getting position data ---", err)
	}

	strLeverage := fmt.Sprintf("%d", positionsResponse.Leverage)
	sp := strings.Split(order.Side, "_")

	userPosition := models.Positions{
		Symbol:       order.Symbol,
		Leverage:     strLeverage,
		OpenPrice:    positionsResponse.AverageOpenPrice,
		LiqPrice:     positionsResponse.LiquidationPrice,
		TakeProfit:   order.TakeProfit,
		StopLoss:     order.StopLoss,
		UnrealizedPl: positionsResponse.UnrealizedPL,
		MarkPrice:    positionsResponse.MarketPrice,
		Side:         sp[1],
		Size:         order.Size,
		Margin:       positionsResponse.Margin,
		UserEmail:    v.UserEmail,
		Status:       "opened",
		Exchange:     "bitget",
	}

	posResponse, createErr := userPosition.UpdateOrCreatePosition(db)
	if createErr != nil {
		fmt.Println(userPosition, "---- error creating new position in database ----", createErr)
		return
	}
	fmt.Println("---- position saved successfully---", posResponse)
}

func binanceTrade(s *Server, v models.Key, i int, trade_req *TradeRequest, first_order float64, res map[string]string, w http.ResponseWriter, keys []models.Key) {
	if v.UserEmail != "kmtester@yopmail.com" {
		return
	}

	futures.UseTestnet = true
	api_key, secret_key, _, err := utils.DecryptKeys(v.ApiKey, v.SecretKey, v.Passphrase, "binance")
	if err != nil {
		fmt.Println("error in decryptkeys: ", err)
		return
	}

	BinanceClient := futures.NewClient(api_key, secret_key)
	symbol2 := trade_req.CoinPair

	switch symbol2 {
	case "BTCUSDT":
		symbol2 = "SBTCSUSDT_SUMCBL"
	case "EOSUSDT":
		symbol2 = "SEOSSUSDT_SUMCBL"
	case "XRPUSDT":
		symbol2 = "SXRPSUSDT_SUMCBL"
	case "ETHUSDT":
		symbol2 = "SETHSUSDT_SUMCBL"
	default:
		symbol2 = "SXRPSUSDT_SUMCBL"
	}

	x, p, err := utils.GetSize(symbol2, first_order)

	if err != nil {
		log.Error(err)
		return
	}

	var str string
	symbol := trade_req.CoinPair
	switch symbol {
	case "ETHUSDT":
		symbol = "ETHUSDT"
		str = fmt.Sprintf("%.3f", x)
	case "EOSUSDT":
		symbol = "EOSUSDT"
		str = strconv.Itoa(int(math.Round(x)))
	case "XRPUSDT":
		symbol = "XRPUSDT"
		str = strconv.Itoa(int(math.Round(x)))
	case "BTCUSDT":
		symbol = "BTCUSDT"
		str = fmt.Sprintf("%.4f", x)
		// fmt.Println("for coin BTCUSDT",str)
	default:
		symbol = "XRPUSDT"
		str = "5"
	}

	var side futures.SideType

	var positionSide futures.PositionSideType

	stopLossValue := 0.0
	takeProfitValue := 0.0

	if trade_req.Long == 1 {
		side = futures.SideTypeBuy
		positionSide = "LONG"

		stopLossValue = p * (1 - STOP_LOSS_PERCENTAGE/100)
		takeProfitValue = p * (1 + TAKE_PROFIT_PERCENTAGE/100)

	} else {
		side = futures.SideTypeSell
		positionSide = "SHORT"

		stopLossValue = p * (1 + STOP_LOSS_PERCENTAGE/100)
		takeProfitValue = p * (1 - TAKE_PROFIT_PERCENTAGE/100)
	}

	strValue := fmt.Sprintf("%f", x)
	strStopLoss := fmt.Sprintf("%f", stopLossValue)
	strTakeProfit := fmt.Sprintf("%f", takeProfitValue)

	order, err := BinanceClient.NewCreateOrderService().
		Symbol(symbol).
		StopPrice(strStopLoss).
		Side(side).
		Type(futures.OrderTypeStopMarket).
		ActivationPrice(strTakeProfit).
		Quantity(strValue).
		NewOrderResponseType(futures.NewOrderRespTypeRESULT).
		Do(context.Background())

	if err != nil {
		log.Error("error in order: " + err.Error() + " for " + v.UserEmail)
		fmt.Println("error in create order", err)
		return
	}
	log.Println(order)

	go fetchAndUpdateBinancePosition(order, v, s.DB, api_key, secret_key, positionSide)

	new_order := models.Order{
		Email:       v.UserEmail,
		Symbol:      order.Symbol,
		Size:        str,
		Side:        string(order.Side),
		MarginCoin:  "USDT",
		OrderType:   "market",
		Service:     "binance",
		QuoteAmount: p,
	}
	_, err = new_order.SaveOrder(s.DB)
	if err != nil {
		log.Error(err)
		return
	}

	if side == futures.SideTypeSell {
		keys[i].OpenShort = v.OpenShort - 1
		keys[i].Prev = "short"
		go s.handleUpdate(&v, keys[i].OpenShort, "short")
	} else if side == futures.SideTypeBuy {
		keys[i].OpenLong = v.OpenLong - 1
		keys[i].Prev = "long"
		go s.handleUpdate(&v, keys[i].OpenLong, "long")
	}
	response.JSON(w, http.StatusOK, "trades made")
	// fmt.Println("trade for binance made")

}

func fetchAndUpdateBinancePosition(order *futures.CreateOrderResponse, v models.Key, db *gorm.DB, api_key string, secret_key string, positionSide futures.PositionSideType) {
	positionsResponse, err := utils.GetBinanceAccountOpenPositions(api_key, secret_key, order.Symbol)
	if err != nil {
		fmt.Println("---error getting position data ---", err)
	}

	side := strings.ToLower(string(positionSide))

	userPosition := models.Positions{
		Symbol:       order.Symbol,
		Leverage:     positionsResponse.Leverage,
		OpenPrice:    positionsResponse.EntryPrice,
		LiqPrice:     positionsResponse.LiquidationPrice,
		TakeProfit:   "",
		StopLoss:     "",
		UnrealizedPl: positionsResponse.UnrealizedProfit,
		MarkPrice:    positionsResponse.MarkPrice,
		Side:         side,
		Size:         order.OrigQuantity,
		Margin:       positionsResponse.InitialMargin,
		UserEmail:    v.UserEmail,
		Status:       "opened",
		Exchange:     "binance",
	}

	posResponse, createErr := userPosition.UpdateOrCreatePosition(db)
	if createErr != nil {
		fmt.Println(userPosition, "---- error creating new position in database ----", createErr)
		return
	}
	fmt.Println("---- position saved successfully---", posResponse)
}

func okexTrade(s *Server, v models.Key, i int, trade_req *TradeRequest, first_order float64, res map[string]string, keys []models.Key) {
	// fmt.Println("trade req for okex")
	if v.Service == "okx" {
		// fmt.Println("trade service for okex")
		dest := okex.DemoServer // The main API server
		ctx := context.Background()
		api_key, secret_key, passphrase, err := utils.DecryptKeys(v.ApiKey, v.SecretKey, v.Passphrase, "okx")
		if err != nil {
			fmt.Println("error in decryptkeys: ", err)
			return
		}
		// api_key, secret_key, passphrase := utils.DecryptKeys(v.ApiKey, v.SecretKey, v.Passphrase, "okx")
		c, err := api.NewClient(ctx, api_key, secret_key, passphrase, dest)
		if err != nil {
			log.Fatalln(err)
		}
		symbol := strings.Split(trade_req.CoinPair, "_")[0]

		x, _, err := utils.GetSize(trade_req.CoinPair, first_order)

		if err != nil {
			log.Error(err)
			return
		}

		switch symbol {
		case "SETHSUSDT":
			symbol = "ETH-USDT"
		case "SEOSSUSDT":
			symbol = "EOS-USDT"
		case "SXRPSUSDT":
			symbol = "XRP-USDT"
		case "SBTCSUSDT":
			symbol = "BTC-USDT"
		default:
			symbol = "XRP-USDT"
		}

		var side okex.OrderSide
		if trade_req.Long == 1 {
			side = okex.OrderBuy
		} else {
			side = okex.OrderSell
		}
		req := []requests.PlaceOrder{
			{
				InstID:  symbol,
				TdMode:  okex.TradeCashMode,
				Side:    side,
				OrdType: okex.OrderMarket,
				Sz:      x,
			},
		}
		res, _ := c.Rest.Trade.PlaceOrder(req)

		if res.Code == 0 {
			log.Println("Trade made for " + symbol + " : " + v.UserEmail)
			if side == okex.OrderSell {
				keys[i].OpenShort = v.OpenShort - 1
				keys[i].Prev = "short"
				go s.handleUpdate(&v, keys[i].OpenShort, "short")
			} else if side == okex.OrderBuy {
				keys[i].OpenLong = v.OpenLong - 1
				keys[i].Prev = "long"
				go s.handleUpdate(&v, keys[i].OpenLong, "long")
			}

		} else {
			log.Error(res.Msg)
		}
	} else {
		fmt.Println("trade service for okex not allowed")
	}
}

func bybitTrade(s *Server, v models.Key, i int, trade_req *TradeRequest, first_order float64, res map[string]string, keys []models.Key) {

	if v.UserEmail != "kmtester@yopmail.com" {
		return
	}

	api_key, secret_key, passphrase, err := utils.DecryptKeys(v.ApiKey, v.SecretKey, v.Passphrase, "bybit")
	if err != nil {
		fmt.Println("error in decryptkeys: ", err)
		return
	}

	order := models.BybitOrderRequest{}
	orderResp := models.BybitResponse{}

	if trade_req.Long == 1 {
		order.Side = "Buy"
	} else {
		order.Side = "Sell"
	}
	order.Symbol = trade_req.CoinPair
	symbol := trade_req.CoinPair
	order.Category = "linear"
	order.OrderType = "Market"
	order.TimeInForce = "GoodTillCancel"
	order.ReduceOnly = false
	order.CloseOnTrigger = false
	order.OrderType = "Market"
	switch symbol {
	case "BTCUSDT":
		symbol = "SBTCSUSDT_SUMCBL"
	case "EOSUSDT":
		symbol = "SEOSSUSDT_SUMCBL"
	case "XRPUSDT":
		symbol = "SXRPSUSDT_SUMCBL"
	case "ETHUSDT":
		symbol = "SETHSUSDT_SUMCBL"
	default:
		symbol = "SXRPSUSDT_SUMCBL"
	}

	size, p, err := utils.GetSize(symbol, first_order)

	if err != nil {
		log.Fatal(err)
		fmt.Println(err, " in get size for ", v.UserEmail)
		return
	}
	order.Qty = fmt.Sprintf("%f", size)
	go func() {
		str, err := BybitNewOrder2(api_key, secret_key, passphrase, &order)
		if err != nil {
			log.Error("failed order: " + err.Error() + " for " + v.UserEmail)
			return
		}
		if str == "" {
			log.Error("empty response for " + v.UserEmail)
			fmt.Println("empty response for " + v.UserEmail)
			return
		}
		if err = json.Unmarshal([]byte(str), &orderResp); err != nil {
			log.Error("parse fail: " + err.Error() + " for " + v.UserEmail)
			fmt.Println("parse fail: " + err.Error() + " for " + v.UserEmail)
			return
		}

		if orderResp.RetCode != 0 {
			log.Error(orderResp.RetMsg + " :: " + v.UserEmail)
			fmt.Println("error in retCode")
			return
		}

		new_order := models.Order{
			Email:       v.UserEmail,
			Symbol:      order.Symbol,
			Size:        order.Qty,
			Side:        order.Side,
			MarginCoin:  "USDT",
			OrderType:   order.OrderType,
			Service:     "bybit",
			QuoteAmount: p,
		}

		side := "long"
		if order.Side == "Sell" {
			side = "short"
		}

		go fetchAndUpdateBybitPosition(order, v, s.DB, api_key, secret_key, side)

		_, err = new_order.SaveOrder(s.DB)
		if err != nil {
			log.Error("err saving order: " + err.Error() + " for " + v.UserEmail)
			fmt.Println("error in saving order: ", err)
			return
		}

		if order.Side == "Buy" {
			keys[i].OpenShort = v.OpenShort - 1
			keys[i].Prev = "short"
			go s.handleUpdate(&v, keys[i].OpenShort, "short")
		} else if order.Side == "Sell" {
			keys[i].OpenLong = v.OpenLong - 1
			keys[i].Prev = "long"
			go s.handleUpdate(&v, keys[i].OpenLong, "long")
		}
		// res["client_id"] = orderResp.Result.OrderID
		// res["order_id"] = orderResp.Result.OrderLinkId
		// fmt.Println("res order placed for " + v.UserEmail)
		// fmt.Println(res)
	}()
}

func fetchAndUpdateBybitPosition(order models.BybitOrderRequest, v models.Key, db *gorm.DB, api_key string, secret_key string, positionSide string) {

	positionsResponse, err := utils.GetBybitAccountPositions(api_key, secret_key, order.Symbol)
	if err != nil {
		fmt.Println("---error getting position data ---", err)
	}

	side := strings.ToLower(string(positionSide))

	userPosition := models.Positions{
		Symbol:       order.Symbol,
		Leverage:     positionsResponse.Leverage,
		OpenPrice:    positionsResponse.AvgPrice,
		LiqPrice:     positionsResponse.LiqPrice,
		TakeProfit:   "",
		StopLoss:     "",
		UnrealizedPl: positionsResponse.UnrealisedPnl,
		MarkPrice:    positionsResponse.MarkPrice,
		Side:         side,
		Size:         order.Qty,
		Margin:       positionsResponse.PositionMM,
		UserEmail:    v.UserEmail,
		Status:       "opened",
		Exchange:     "bybit",
	}

	posResponse, createErr := userPosition.UpdateOrCreatePosition(db)
	if createErr != nil {
		fmt.Println(userPosition, "---- error creating new position in database ----", createErr)
		return
	}
	fmt.Println("---- position saved successfully---", posResponse)
}

func (t *TradeRequest) Validate() error {
	if t.CoinPair == "" {
		return errors.New("coin pair is required")
	}
	return nil
}

func (s *Server) UpdateAmount(w http.ResponseWriter, r *http.Request, email string) {

	service := r.URL.Query().Get("service")
	if service == "" {
		response.ERROR(w, http.StatusBadRequest, errors.New("service required"))
		return
	}

	if service != "bitget" && service != "binance" && service != "okx" {
		response.ERROR(w, http.StatusBadRequest, errors.New("service not supported"))
		return
	}

	key := models.Key{}
	key.Service = service
	key.UserEmail = email

	var data map[string]int

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		response.ERROR(w, http.StatusInternalServerError, err)
		return
	}

	if data["trade_amount"] == 0 {
		response.ERROR(w, http.StatusBadRequest, errors.New("trade amount is required"))
		return
	}

	if err := utils.CheckBalance(data["trade_amount"]); err != nil {
		response.ERROR(w, http.StatusBadRequest, errors.New("minimum amount must be 200"))
		return
	}

	keys, err := key.ChangeTradeAmount(s.DB, data["trade_amount"])
	if err != nil {
		response.ERROR(w, http.StatusInternalServerError, err)
		return
	}

	go func() {
		new_keys, err := key.FindAllKeys(s.DB)
		if err != nil {
			log.Fatal("error getting keys")
			app_data.Keys_list = []models.Key{}
			return
		}
		log.Info("retreived keys")
		app_data.Keys_list = *new_keys
	}()

	response.JSON(w, http.StatusOK, keys)
}
