package controllers

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"

	helpers "github.com/ahmed-023/bitget-helpers"
	"github.com/kryptomind/BidBox-Trades/models"
	"github.com/kryptomind/BidBox-Trades/response"
)

func (server *Server) PlaceOrder(w http.ResponseWriter, r *http.Request, email string) {
	var order models.OrderRequest
	var saveOrder models.Order

	err := json.NewDecoder(r.Body).Decode(&order)
	if err != nil {
		response.ERROR(w, http.StatusBadRequest, err)
		return
	}

	//get keys by user id
	key := models.Key{}
	keys, err := key.FindKeysByEmail(server.DB, email)
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
				log.Error(err)
				return
			}
			secret_key, err = helpers.DecryptStrings(v.SecretKey)
			if err != nil {
				log.Error(err)
				return
			}
			passphrase, err = helpers.DecryptStrings(v.Passphrase)
			if err != nil {
				log.Error(err)
				return
			}
		}
	}
	if api_key == "" {
		response.ERROR(w, http.StatusNoContent, errors.New("api key not found"))
		return
	}
	var orderResp models.OrderResponse
	err = order.Validate()
	if err != nil {
		response.ERROR(w, http.StatusUnprocessableEntity, err)
		return
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
	json_val, _ := json.Marshal(res)

	saveOrder.Initialize(order, email, orderResp.Data.ClientOid, orderResp.Data.OrderID)

	saveOrder.SaveOrder(server.DB)
	w.Header().Set("Content-Type", "application/json")
	w.Write(json_val)
}

func GenerateBitgetSignature(apiSecret string, apiKey string, passphrase string, method string, uri string, timestamp string, requestBody string) string {

	message := ""
	if method == "GET" {
		message = fmt.Sprintf("%s%s%s", timestamp, method, uri)
	} else if method == "POST" {
		message = fmt.Sprintf("%s%s%s%s", timestamp, method, uri, requestBody)
	}

	// Calculate HMAC-SHA256 signature
	hmac := hmac.New(sha256.New, []byte(apiSecret))
	hmac.Write([]byte(message))
	signature := base64.StdEncoding.EncodeToString(hmac.Sum(nil))

	return signature
}

func NewOrder(api_key string, secret_key string, passphrase string, order *models.OrderRequest) (string, error) {

	host := "https://api.bitget.com"
	path := "/api/mix/v1/order/placeOrder"
	url := host + path

	method := "POST"
	client := &http.Client{}

	jsonVal, err := json.Marshal(order)
	if err != nil {
		log.Error(err)
		return "", err
	}

	server_time := helpers.GetBitgetServerTimeStamp()
	signatures := GenerateBitgetSignature(secret_key, api_key, passphrase, "POST", path, server_time, string(jsonVal))

	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonVal))
	req.Header.Add("ACCESS-KEY", api_key)
	req.Header.Add("ACCESS-SIGN", signatures)
	req.Header.Add("ACCESS-TIMESTAMP", server_time)
	req.Header.Add("ACCESS-PASSPHRASE", passphrase)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("local", "zh-CN")

	if err != nil {
		log.Error(err)
		return "", err
	}

	res, err := client.Do(req)
	if err != nil {
		log.Error(err)
		return "", err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Error(err)
		return "", err
	}
	return string(body), nil
}
