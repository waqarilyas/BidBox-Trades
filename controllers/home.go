package controllers

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	helpers "github.com/ahmed-023/bitget-helpers"
	"github.com/asaskevich/govalidator"
	"github.com/kryptomind/BidBox-Trades/models"
	"github.com/kryptomind/BidBox-Trades/response"
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

func InitCoinPiars() ([]string, error) {

	var coin_pairs []string

	// Open the CSV file
	file, err := os.Open("test.csv")
	if err != nil {
		fmt.Println("Error:", err)
		return []string{}, err
	}
	defer file.Close()

	// Create a new CSV reader
	reader := csv.NewReader(file)

	// Read all the CSV records
	records, err := reader.ReadAll()
	if err != nil {
		fmt.Println("Error:", err)
		return []string{}, err
	}

	// Iterate over the records and print the values
	for _, record := range records {
		for i, value := range record {
			if i == 1 {
				coin_pairs = append(coin_pairs, value)
			}
		}
	}
	return coin_pairs, nil
}

func AiStub(orders int) ([]string, []string, error) {
	rand.Seed(time.Now().UnixNano()) // initialize the random number generator with the current time

	// define an array of integers
	coins, err := InitCoinPiars()

	if err != nil {
		return []string{}, []string{}, err
	}

	// create a slice to hold the selected elements
	list := make([]string, orders)

	// generate the random indices and select the corresponding elements
	for i := 0; i < orders; i++ {
		index := rand.Intn(len(coins)) // generate a random index within the range of the array
		list[i] = coins[index]         // select the element at the random index and add it to the selected slice
	}

	middle := orders / 2
	return list[middle:], list[:middle], nil
}

type User struct {
	Capital      int
	Trade_amount float64
	First_order  float64
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

	long, short, err := AiStub(cond.Positions)
	if err != nil {
		response.ERROR(w, http.StatusInternalServerError, err)
		return
	}

	log.Println(long)
	log.Println(short)
	response.JSON(w, http.StatusOK, user)

}

func checkBalance(val float64) error {
	if val < 200 {
		return errors.New("Balance should be at least 200 USDT")
	}
	return nil
}
