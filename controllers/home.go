package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"

	"github.com/kryptomind/BidBox-Trades/models"
	"github.com/kryptomind/BidBox-Trades/response"
	"github.com/kryptomind/BidBox-Trades/utils"
	log "github.com/sirupsen/logrus"
)

var modes = []string{"conservative", "aggressive"}
var strategies = []string{"cycle", "single", "stop make", "stop long", "stop short"}

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
	CoinPair   string  `json:"coin_pair"`
	OpenPrice  float64 `json:"open_value"`
	ClosePrice float64 `json:"close_value"`
}

func (s *Server) handleUpdate(key *models.Key, val int, pos string) {
	key, err := key.ChangePositions(s.DB, val, pos)
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Println(key)
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
			if v.Service == "bitget" {
				api_key, secret_key, passphrase := utils.DecryptKeys(v.ApiKey, v.SecretKey, v.Passphrase)
				order := models.OrderRequest{}
				orderResp := models.OrderResponse{}
				if trade_req.ClosePrice > trade_req.OpenPrice {
					if v.OpenLong <= 0 {
						return
					}
					order.Side = "open_long"
				} else {
					if v.OpenShort <= 0 {
						return
					}
					order.Side = "open_short"
				}
				val := int(math.Floor(float64(v.TradeAmount)/100) * 100)
				cond := models.Conditions{}
				c, err := cond.FindCondition(s.DB, val)
				if err != nil {
					log.Fatal(err)
					return
				}
				first_order := float64(v.TradeAmount) * 0.08 / float64(c.Positions)
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
						response.ERROR(w, http.StatusInternalServerError, err)
						return
					}

					log.Println(str)

					if err = json.Unmarshal([]byte(str), &orderResp); err != nil {
						response.ERROR(w, http.StatusBadRequest, err)
						return
					}

					if orderResp.Code != "00000" {
						response.ERROR(w, http.StatusExpectationFailed, errors.New(orderResp.Msg))
						return
					}
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
			}
		}(v, i)
	}

	response.JSON(w, http.StatusOK, "trades made")

}

func (t *TradeRequest) Validate() error {
	if t.ClosePrice == 0 {
		return errors.New("close price is required")
	}
	if t.OpenPrice == 0 {
		return errors.New("open price is required")
	}
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

	if service != "bitget" {
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
