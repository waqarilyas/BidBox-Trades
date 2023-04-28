package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	helpers "github.com/ahmed-023/bitget-helpers"
	"github.com/asaskevich/govalidator"
	"github.com/kryptomind/BidBox-Trades/models"
	"github.com/kryptomind/BidBox-Trades/response"
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

func (s *Server) Home(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	if email == "" {
		response.ERROR(w, http.StatusBadRequest, errors.New("email is required"))
		return
	}

	if !govalidator.IsEmail(email) {
		response.ERROR(w, http.StatusBadRequest, errors.New("invalid email address"))
		return
	}

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

	err = checkBalance(val)

	if err != nil {
		response.ERROR(w, http.StatusInternalServerError, err)
		return
	}

	response.JSON(w, http.StatusOK, val)

}

func checkBalance(val float64) error {
	if val < 200 {
		return errors.New("Balance should be at least 200 USDT")
	}
	return nil
}
