package binance

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
	"fmt"
)

const (
	BinanceAPIEndpoint = "https://testnet.binancefuture.com"
)

func GetBinanceAccountDetails(apiKey string, secret string) (*AccountsResponse, error) {
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	params := map[string]string{
		"timestamp":  strconv.FormatInt(timestamp, 10),
		"recvWindow": "5000",
	}
	accSignature := GenerateBinanceSignature(params, secret)
	finalURL := BinanceAPIEndpoint + "/fapi/v2/account?recvWindow=5000&timestamp=" + strconv.FormatInt(timestamp, 10) + "&signature=" + accSignature

	req, err := http.NewRequest("GET", finalURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-MBX-APIKEY", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		var errorResponse BinanceErrorResponse
		err = json.Unmarshal(body, &errorResponse)
		if err != nil {
			return nil, err
		}
		return nil, errors.New(errorResponse.Msg)
	}

	var response AccountsResponse

	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}


func GetBinanceTradeLimit(coin_pair string) (float64, error) {
	url := "https://www.binance.com/fapi/v1/exchangeInfo?showall=true"
	method := "GET"
	
	client := &http.Client {
	}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return 0.0, err
	}

	// req.Header.Add("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return 0.0, err
	}
	defer res.Body.Close()
	symbols := BinanceLimitResponse{}
		if err := json.NewDecoder(res.Body).Decode(&symbols); err != nil {
			return 0.0, err
		}
	var minPrice string
	for _, symbol := range symbols.SymbolStruct {
		if symbol.Symbol == coin_pair {
			minPrice = symbol.Filters[1].StepSize
			break
		}
	}
	// Check if MinPrice was found
	if minPrice != "" {
		fmt.Println("MinPrice for " , coin_pair, minPrice)
	} else {
		fmt.Println("BTCUSDT not found")
	}
	minPriceFloat , err := strconv.ParseFloat(minPrice, 64)
	return minPriceFloat, nil
}
	
