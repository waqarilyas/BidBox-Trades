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
	Long     int    `json:"long"`
	Exchange string `json:"exchange"`
}

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

func (s *Server) StartTrade(w http.ResponseWriter, r *http.Request) {
	res := make(map[string]string)

	trade_req := &TradeRequest{}
	err := json.NewDecoder(r.Body).Decode(trade_req)
	if err != nil {
		response.ERROR(w, http.StatusBadRequest, err)
		return
	}

	err = trade_req.Validate()
	if err != nil {
		response.ERROR(w, http.StatusBadRequest, err)
		return
	}
	fmt.Println("started")
	//	stop_loss := 0.8 * float64(300)
	//take_profit := 1.1 * float64(300)
	for i, v := range app_data.Keys_list { // for each user key
		// fmt.Println("user: ", v.UserEmail)
		go func(v models.Key, i int) {

			if trade_req.Long == 1 && v.Prev == "long" {
				return
			}
			if trade_req.Long == 0 && v.Prev == "short" {
				return
			}
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

			// fmt.Println("condition: ", c.Positions)
			first_order := float64(v.TradeAmount) * 0.08 / float64(c.Positions)
			if trade_req.Exchange == "bitget" {
				// fmt.Println("trade req for bitget")
				if v.Service == "bitget" {
					fmt.Println("trade service for bitget")
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
						_, err = new_order.SaveOrder(s.DB)
						if err != nil {
							log.Error("err saving order: " + err.Error() + " for " + v.UserEmail)
							return
						}
						if order.Side == "open_short" {
							app_data.Keys_list[i].OpenShort = v.OpenShort - 1
							app_data.Keys_list[i].Prev = "short"
							go s.handleUpdate(&v, app_data.Keys_list[i].OpenShort, "short")
						} else if order.Side == "open_long" {
							app_data.Keys_list[i].OpenLong = v.OpenLong - 1
							app_data.Keys_list[i].Prev = "long"
							go s.handleUpdate(&v, app_data.Keys_list[i].OpenLong, "long")
						}
						res["client_id"] = orderResp.Data.ClientOid
						res["order_id"] = orderResp.Data.OrderID
						// fmt.Println("res order placed for " + v.UserEmail)
						// fmt.Println(res)
					}()
				} else {
					// fmt.Println("trade service for bitget not found")
				}
			} else if trade_req.Exchange == "binance" {
				// fmt.Println("trade req for binance")
				if v.Service == "binance" {
					fmt.Println("trade service for binance")
					futures.UseTestnet = true
					api_key, secret_key, _, err := utils.DecryptKeys(v.ApiKey, v.SecretKey, v.Passphrase, "binance")
					if err != nil {
						fmt.Println("error in decryptkeys: ", err)
						return
					}
					// api_key, secret_key, _ := utils.DecryptKeys(v.ApiKey, v.SecretKey, v.Passphrase, "binance")
					BinanceClient := futures.NewClient(api_key, secret_key)
					// symbol := strings.Split(trade_req.CoinPair, "_")[0]
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
					if trade_req.Long == 1 {
						side = futures.SideTypeBuy
					} else {
						side = futures.SideTypeSell
					}
					// rounded := math.Round(x)
					// str := strconv.Itoa(int(rounded))

					// fmt.Println(str)
					fmt.Println("get size: ", str)
					order, err := BinanceClient.NewCreateOrderService().Symbol(symbol).
						Side(side).Type(futures.OrderTypeMarket).
						Quantity(x).
						NewOrderResponseType(futures.NewOrderRespTypeRESULT).
						Do(context.Background())
					if err != nil {
						log.Error("error in order: " + err.Error() + " for " + v.UserEmail)
						fmt.Println("error in create order", err)
						return
					}
					log.Println(order)
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
						app_data.Keys_list[i].OpenShort = v.OpenShort - 1
						app_data.Keys_list[i].Prev = "short"
						go s.handleUpdate(&v, app_data.Keys_list[i].OpenShort, "short")
					} else if side == futures.SideTypeBuy {
						app_data.Keys_list[i].OpenLong = v.OpenLong - 1
						app_data.Keys_list[i].Prev = "long"
						go s.handleUpdate(&v, app_data.Keys_list[i].OpenLong, "long")
					}
					response.JSON(w, http.StatusOK, "trades made")
					// fmt.Println("trade for binance made")
				} else {
					response.JSON(w, http.StatusBadRequest, "trades are not allowed for this user")
				}
			} else if trade_req.Exchange == "okex" {
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
							app_data.Keys_list[i].OpenShort = v.OpenShort - 1
							app_data.Keys_list[i].Prev = "short"
							go s.handleUpdate(&v, app_data.Keys_list[i].OpenShort, "short")
						} else if side == okex.OrderBuy {
							app_data.Keys_list[i].OpenLong = v.OpenLong - 1
							app_data.Keys_list[i].Prev = "long"
							go s.handleUpdate(&v, app_data.Keys_list[i].OpenLong, "long")
						}

					} else {
						log.Error(res.Msg)
					}
				} else {
					fmt.Println("trade service for okex not allowed")
				}

			} else if trade_req.Exchange == "bybit" {
				// fmt.Println("trade req for bybit")
				if v.Service == "bybit" {

					fmt.Println("trade service for bybit")
					// fmt.Println("api secret")
					api_key, secret_key, passphrase, err := utils.DecryptKeys(v.ApiKey, v.SecretKey, v.Passphrase, "bybit")
					if err != nil {
						fmt.Println("error in decryptkeys: ", err)
						return
					}
					// api_key, secret_key, passphrase := utils.DecryptKeys(v.ApiKey, v.SecretKey, v.Passphrase, "bybit")
					order := models.BybitOrderRequest{}
					orderResp := models.BybitResponse{}
					// fmt.Println("models")
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
							// fmt.Println("failed order: " + err.Error() + " for " + v.UserEmail)
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
						_, err = new_order.SaveOrder(s.DB)
						if err != nil {
							log.Error("err saving order: " + err.Error() + " for " + v.UserEmail)
							fmt.Println("error in saving order: ", err)
							return
						}
						if order.Side == "Buy" {
							app_data.Keys_list[i].OpenShort = v.OpenShort - 1
							app_data.Keys_list[i].Prev = "short"
							go s.handleUpdate(&v, app_data.Keys_list[i].OpenShort, "short")
						} else if order.Side == "Sell" {
							app_data.Keys_list[i].OpenLong = v.OpenLong - 1
							app_data.Keys_list[i].Prev = "long"
							go s.handleUpdate(&v, app_data.Keys_list[i].OpenLong, "long")
						}
						// res["client_id"] = orderResp.Result.OrderID
						// res["order_id"] = orderResp.Result.OrderLinkId
						// fmt.Println("res order placed for " + v.UserEmail)
						// fmt.Println(res)
					}()
				} else {
					// fmt.Println("trade service for bybit not found")
				}
			}
		}(v, i)
	}

	response.JSON(w, http.StatusOK, "trades made successfully")

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
