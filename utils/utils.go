package utils

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	helpers "github.com/WAQAR5/bitget-helpers"
)

func initCoinPiars() ([]string, error) {

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
	coins := []string{"SETHSUSDT_SUMCBL"}
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

type IndexPriceReq struct {
	Code string `json:"code"`
	Data struct {
		Symbol    string `json:"symbol"`
		Index     string `json:"index"`
		Timestamp string `json:"timestamp"`
	} `json:"data"`
	Msg         string `json:"msg"`
	RequestTime int64  `json:"requestTime"`
}

func GetSize(symbol string, first_order float64) (float64, float64, error) {
	url := "https://api.bitget.com/api/mix/v1/market/index?symbol=" + symbol

	client := http.Client{}

	res, err := client.Get(url)
	if err != nil {
		return 0, 0, err
	}
	defer res.Body.Close()

	price := IndexPriceReq{}
	if err := json.NewDecoder(res.Body).Decode(&price); err != nil {
		return 0, 0, err
	}

	if price.Code != "00000" {
		return 0, 0, errors.New(price.Msg)
	}

	fprice, err := strconv.ParseFloat(price.Data.Index, 64)
	if err != nil {
		return 0, 0, err
	}

	tradeAmount := first_order / fprice

	decimalMultiplier := math.Pow(10, float64(3))
	fixedAmount := math.Round(tradeAmount*decimalMultiplier) / decimalMultiplier

	return fixedAmount, fprice, nil
}
func GetSizeBybit(symbol string, first_order float64) (float64, float64, error) {
	url := "https://api.bitget.com/api/mix/v1/market/index?symbol=" + symbol

	client := http.Client{}

	res, err := client.Get(url)
	if err != nil {
		return 0, 0, err
	}
	defer res.Body.Close()

	price := IndexPriceReq{}
	if err := json.NewDecoder(res.Body).Decode(&price); err != nil {
		return 0, 0, err
	}

	if price.Code != "00000" {
		return 0, 0, errors.New(price.Msg)
	}

	fprice, err := strconv.ParseFloat(price.Data.Index, 64)
	if err != nil {
		return 0, 0, err
	}

	return first_order / fprice, fprice, nil
}

func CheckBalance(val int) error {
	if val < 200 {
		return errors.New("Balance should be at least 200 USDT")
	}
	return nil
}

func DecryptKeys(api_key string, secret_key string, passphrase string, service string) (string, string, string, error) {
	api_key, err := helpers.DecryptStrings(api_key)
	fmt.Println("in decrypt keys function")
	if err != nil {
		log.Fatal(err)
		return "", "", "", err
	}
	secret_key, err = helpers.DecryptStrings(secret_key)
	if err != nil {
		log.Fatal(err)
		return "", "", "", err
	}
	if service == "bitget" || service == "okx" {
		passphrase, err = helpers.DecryptStrings(passphrase)
		if err != nil {
			log.Fatal(err)
			return "", "", "", err
		}
	} else if service == "binance" {
		passphrase = ""
	}

	return api_key, secret_key, passphrase, nil
}
