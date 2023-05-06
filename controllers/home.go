package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	helpers "github.com/ahmed-023/bitget-helpers"
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

func (s *Server) Home(w http.ResponseWriter, r *http.Request, email string) {
	//get keys by user id
	key := models.Key{}
	keys, err := key.FindKeysByEmail(s.DB, email)
	if err != nil {
		response.JSON(w, http.StatusBadRequest, errors.New("User not found"))
		return
	}

	var api_key string
	var secret_key string
	var passphrase string
	for _, v := range *keys {
		if strings.ToLower(v.Service) != "bitget" {
			response.JSON(w, http.StatusNoContent, errors.New("Exchange coming soon"))
			return
		} else {
			api_key, err = helpers.DecryptStrings(v.ApiKey)
			if err != nil {
				log.Fatal(err)
				return
			}
			secret_key, err = helpers.DecryptStrings(v.SecretKey)
			if err != nil {
				log.Fatal(err)
				return
			}
			passphrase, err = helpers.DecryptStrings(v.Passphrase)
			if err != nil {
				log.Fatal(err)
				return
			}
		}
	}

	res, err := helpers.GetAccountDetailsList(secret_key, api_key, passphrase)
	if err != nil {
		response.ERROR(w, http.StatusInternalServerError, err)
		return
	}

	acc := Account{}
	err = json.Unmarshal([]byte(res), &acc)

	if err != nil {
		response.ERROR(w, http.StatusInternalServerError, err)
		return
	}

	val, err := strconv.ParseFloat(acc.Data[0].Available, 64)

	if err != nil {
		response.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	val = 300
	err = utils.CheckBalance(val)

	if err != nil {
		response.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	log.Println(val)
	log.Println(int(math.Floor(val)))

	conds := models.Conditions{}
	cond, err := conds.FindKeyById(s.DB, int(math.Floor(val/100)*100))

	if err != nil {
		response.ERROR(w, http.StatusInternalServerError, err)
		return
	}

	user := User{}
	user.Capital = int(val)
	user.Trade_amount = 0.08 * float64(cond.Capital)
	stop_loss := 0.8 * float64(cond.Capital)
	take_profit := 0.1 * float64(cond.Capital)
	user.First_order = user.Trade_amount / float64(cond.Positions)

	long, short, err := utils.AiStub(cond.Positions)
	if err != nil {
		response.ERROR(w, http.StatusInternalServerError, err)
		return
	}

	for i := 0; i < cond.Positions; i++ {
		order := models.OrderRequest{}
		orderResp := models.OrderResponse{}
		order.Symbol = "SETHSUSDT_SUMCBL"
		order.MarginCoin = "SUSDT"
		order.Size = "0.01"
		order.OrderType = "market"
		order.StopLoss = fmt.Sprintf("%F", stop_loss)
		order.TakeProfit = fmt.Sprintf("%F", take_profit)
		if i%2 == 0 {
			order.Side = "open_long"
		} else {
			order.Side = "open_short"
		}
		str, err := NewOrder(api_key, secret_key, passphrase, &order)
		if err != nil {
			response.ERROR(w, http.StatusInternalServerError, err)
			return
		}

		err = json.Unmarshal([]byte(str), &orderResp)
		if err != nil {
			response.ERROR(w, http.StatusBadRequest, err)
			return
		}
		res := map[string]string{"client_id": orderResp.Data.ClientOid, "order_id": orderResp.Data.OrderID}
		log.Println(res)

	}

	log.Println(long)
	log.Println(short)
	response.JSON(w, http.StatusOK, user)

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

func (s *Server) StartTrade(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

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

	key := models.Key{}
	keys, err := key.FindAllKeys(s.DB)
	if err != nil {
		response.ERROR(w, http.StatusInternalServerError, err)
		return
	}

	//	stop_loss := 0.8 * float64(300)
	//take_profit := 1.1 * float64(300)

	for _, v := range *keys {
		api_key, secret_key, passphrase := utils.DecryptKeys(v.ApiKey, v.SecretKey, v.Passphrase)
		if v.Service == "bitget" {
			order := models.OrderRequest{}
			orderResp := models.OrderResponse{}
			if trade_req.ClosePrice > trade_req.OpenPrice {
				if v.OpenLong <= 0 {
					continue
				}
				go func() {
					key, err := v.ChangePositions(s.DB, v.OpenLong-1, "long")
					if err != nil {
						response.ERROR(w, http.StatusInternalServerError, err)
						return
					}
					log.Println(key)
				}()
				order.Side = "open_long"
			} else {
				if v.OpenShort <= 0 {
					continue
				}
				go func() {
					key, err := v.ChangePositions(s.DB, v.OpenShort-1, "short")
					if err != nil {
						response.ERROR(w, http.StatusInternalServerError, err)
						return
					}
					log.Println(key)
				}()
				order.Side = "open_short"
			}
			order.Symbol = trade_req.CoinPair
			order.MarginCoin = "SUSDT"
			order.Size = "0.01"
			order.OrderType = "market"
			//order.StopLoss = fmt.Sprintf("%F", stop_loss)
			//order.TakeProfit = fmt.Sprintf("%F", take_profit)
			go func() {
				str, err := NewOrder(api_key, secret_key, passphrase, &order)
				if err != nil {
					response.ERROR(w, http.StatusInternalServerError, err)
					return
				}

				log.Println(str)

				err = json.Unmarshal([]byte(str), &orderResp)
				if err != nil {
					response.ERROR(w, http.StatusBadRequest, err)
					return
				}
				res := map[string]string{"client_id": orderResp.Data.ClientOid, "order_id": orderResp.Data.OrderID}
				log.Println(res)
			}()

		}
	}
	response.JSON(w, http.StatusOK, "trades made")

	elapsed := time.Since(start)
	fmt.Printf("Time taken: %s\n", elapsed)
}
