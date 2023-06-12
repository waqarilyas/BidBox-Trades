package binance

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
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


