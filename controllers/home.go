package controllers

import (
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"strconv"
	"strings"

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
