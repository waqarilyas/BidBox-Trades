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

	requests "github.com/amir-the-h/okex/requests/rest/trade"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/amir-the-h/okex"
	"github.com/amir-the-h/okex/api"
	"github.com/kryptomind/BidBox-Trades/models"
	"github.com/kryptomind/BidBox-Trades/response"
	"github.com/kryptomind/BidBox-Trades/utils"
	log "github.com/sirupsen/logrus"
)

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
}

func (s *Server) handleUpdate(key *models.Key, val int, pos string) {
	key, err := key.ChangePositions(s.DB, val, pos)
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Println(key)
}

func (server *Server) Home(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, "Trade Service")
}

func (s *Server) StartTrade(w http.ResponseWriter, r *http.Request) {

	res := make(map[string]string)

	trade_req := &TradeRequest{}
	err := json.NewDecoder(r.Body).Decode(trade_req)
	if err != nil {
		response.ERROR(w, http.StatusBadRequest, err)
		return
	}
	log.Println(trade_req)
	err = trade_req.Validate()
	if err != nil {
		response.ERROR(w, http.StatusBadRequest, err)
		return
	}

	//	stop_loss := 0.8 * float64(300)
	//take_profit := 1.1 * float64(300)
	for i, v := range app_data.Keys_list {

		go func(v models.Key, i int) {

			if v.OpenLong <= 0 {
				return
			}
			val := int(math.Floor(float64(v.TradeAmount)/100) * 100)
			cond := models.Conditions{}
			c, err := cond.FindCondition(s.DB, val)
			if err != nil {
				log.Fatal(err)
				return
			}

			first_order := float64(v.TradeAmount) * 0.08 / float64(c.Positions)
			if v.Service == "bitget" {
				api_key, secret_key, passphrase := utils.DecryptKeys(v.ApiKey, v.SecretKey, v.Passphrase, "bitget")
				order := models.OrderRequest{}
				orderResp := models.OrderResponse{}

				order.Side = "open_long"

				log.Println(v.UserEmail)

				order.Symbol = trade_req.CoinPair
				order.MarginCoin = "SUSDT"
				size, err := utils.GetSize(order.Symbol, first_order)
				if err != nil {
					log.Fatal(err)
					return
				}
				order.Size = fmt.Sprintf("%f", size)
				log.Println(order.Size)
				order.OrderType = "market"
				// order.StopLoss = fmt.Sprintf("%F", (size * 0.8))
				// fmt.Println(order.StopLoss)
				// order.TakeProfit = fmt.Sprintf("%F", (size * 1.1))
				// fmt.Println(order.TakeProfit)
				go func() {
					str, err := NewOrder(api_key, secret_key, passphrase, &order)
					if err != nil {
						log.Error("failed order: " + err.Error() + " for " + v.UserEmail)
						return
					}

					if err = json.Unmarshal([]byte(str), &orderResp); err != nil {
						log.Error("parse fail: " + err.Error() + " for " + v.UserEmail)
						return
					}

					if orderResp.Code != "00000" {
						log.Error(orderResp.Msg + " : " + v.UserEmail)
						return
					}
					new_order := models.Order{
						Email:      v.UserEmail,
						Symbol:     order.Symbol,
						Size:       order.Size,
						Side:       order.Side,
						MarginCoin: order.MarginCoin,
						OrderType:  order.OrderType,
						OrderID:    orderResp.Data.OrderID,
						ClientID:   orderResp.Data.ClientOid,
					}
					new_order.SaveOrder(s.DB)
					if order.Side == "open_short" {
						app_data.Keys_list[i].OpenShort = v.OpenShort - 1
						go s.handleUpdate(&v, app_data.Keys_list[i].OpenShort, "short")
					} else if order.Side == "open_long" {
						app_data.Keys_list[i].OpenLong = v.OpenLong - 1
						go s.handleUpdate(&v, app_data.Keys_list[i].OpenLong, "long")
					}
					res["client_id"] = orderResp.Data.ClientOid
					res["order_id"] = orderResp.Data.OrderID
					log.Println(res)
				}()
			} else if v.Service == "binance" {
				futures.UseTestnet = true

				api_key, secret_key, _ := utils.DecryptKeys(v.ApiKey, v.SecretKey, v.Passphrase, "binance")
				BinanceClient := futures.NewClient(api_key, secret_key)

				symbol := strings.Split(trade_req.CoinPair, "_")[0]

				x, err := utils.GetSize(trade_req.CoinPair, first_order)

				if err != nil {
					log.Error(err)
					return
				}

				var str string

				switch symbol {
				case "SETHSUSDT":
					symbol = "ETHUSDT"
					str = fmt.Sprintf("%.3f", x)
				case "SEOSSUSDT":
					symbol = "EOSUSDT"
					str = strconv.Itoa(int(math.Round(x)))
				case "SXRPSUSDT":
					symbol = "XRPUSDT"
					str = strconv.Itoa(int(math.Round(x)))
				case "SBTCSUSDT":
					symbol = "BTCUSDT"
					str = fmt.Sprintf("%.4f", x)
				default:
					symbol = "XRPUSDT"
					str = "5"
				}

				// rounded := math.Round(x)
				// str := strconv.Itoa(int(rounded))

				fmt.Println(str)

				order, err := BinanceClient.NewCreateOrderService().Symbol(symbol).
					Side(futures.SideTypeBuy).Type(futures.OrderTypeMarket).
					Quantity(str).
					NewOrderResponseType(futures.NewOrderRespTypeRESULT).
					Do(context.Background())
				if err != nil {
					log.Error(err)
					return
				}
				log.Println(order)
				app_data.Keys_list[i].OpenLong = v.OpenLong - 1
				go s.handleUpdate(&v, app_data.Keys_list[i].OpenLong, "long")

			} else if v.Service == "okx" {

				dest := okex.DemoServer // The main API server
				ctx := context.Background()

				api_key, secret_key, passphrase := utils.DecryptKeys(v.ApiKey, v.SecretKey, v.Passphrase, "okx")
				c, err := api.NewClient(ctx, api_key, secret_key, passphrase, dest)
				if err != nil {
					log.Fatalln(err)
				}
				symbol := strings.Split(trade_req.CoinPair, "_")[0]

				x, err := utils.GetSize(trade_req.CoinPair, first_order)

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

				req := []requests.PlaceOrder{
					{
						InstID:  symbol,
						TdMode:  okex.TradeCashMode,
						Side:    okex.OrderBuy,
						OrdType: okex.OrderMarket,
						Sz:      x,
					},
				}
				res, _ := c.Rest.Trade.PlaceOrder(req)

				if res.Code == 0 {
					log.Println("Trade made for " + symbol + " : " + v.UserEmail)
					app_data.Keys_list[i].OpenLong = v.OpenLong - 1
					go s.handleUpdate(&v, app_data.Keys_list[i].OpenLong, "long")

				} else {
					log.Error(res.Msg)
				}
			}
		}(v, i)
	}

	response.JSON(w, http.StatusOK, "trades made")

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
