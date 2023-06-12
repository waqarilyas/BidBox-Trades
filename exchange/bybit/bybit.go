package bybit

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	// "github.com/kryptomind/bidboxapi/KeyService/internal/shared"
	"github.com/kryptomind/BidBox-Trades/shared"
)

const (
	BybitAPIEndpoint = "https://api-testnet.bybit.com"
)

func GetBybitApiKeyInfo(apiKey, secret string) (*APIKeyInfoResponse, error) {
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	queryParameters := fmt.Sprintf("api_key=%s&timestamp=%d", apiKey, timestamp)
	signature := GenerateBybitV2Signature(queryParameters, secret)

	url := fmt.Sprintf("%s/v2/private/account/api-key?%s&sign=%s", BybitAPIEndpoint, queryParameters, signature)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var response APIKeyInfoResponse

	err = json.Unmarshal(body, &response)

	if err != nil {
		return nil, err
	}

	return &response, nil
}

func GetBybitAccountDetails(apiKey string, secret string) (*AccountInfo, error) {
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	accSignature := GenerateBybitSignature(apiKey, secret, 50000, timestamp, "")

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/v5/account/info", BybitAPIEndpoint), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-BAPI-SIGN-TYPE", "2")
	req.Header.Set("X-BAPI-SIGN", accSignature)
	req.Header.Set("X-BAPI-API-KEY", apiKey)
	req.Header.Set("X-BAPI-TIMESTAMP", strconv.FormatInt(timestamp, 10))
	req.Header.Set("X-BAPI-RECV-WINDOW", "50000")
	req.Header.Set("Content-Type", "application/json")

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

	var response AccountInfo

	err = json.Unmarshal(body, &response)

	if err != nil {
		return nil, err
	}

	return &response, nil

}

func GetBybitAccountBalance(apiKey string, secret string) (*shared.AccountData, error) {
	accountDetails, accountDetailsError := GetBybitAccountDetails(apiKey, secret)
	if accountDetailsError != nil {
		return nil, accountDetailsError
	}

	accountType := "UNIFIED"
	IS_UNIFIED := true
	if accountDetails.Result.UnifiedMarginStatus == 1 {
		accountType = "CONTRACT"
		IS_UNIFIED = false
	}

	queryString := "accountType=" + accountType
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	accSignature := GenerateBybitSignature(apiKey, secret, 50000, timestamp, queryString)
	finalURL := BybitAPIEndpoint + "/v5/account/wallet-balance?" + queryString

	req, err := http.NewRequest("GET", finalURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-BAPI-SIGN-TYPE", "2")
	req.Header.Set("X-BAPI-SIGN", accSignature)
	req.Header.Set("X-BAPI-API-KEY", apiKey)
	req.Header.Set("X-BAPI-TIMESTAMP", strconv.FormatInt(timestamp, 10))
	req.Header.Set("X-BAPI-RECV-WINDOW", "50000")
	req.Header.Set("Content-Type", "application/json")

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
		var errorResponse ErrorResponse
		err = json.Unmarshal(body, &errorResponse)
		if err != nil {
			return nil, err
		}
		return nil, errors.New(errorResponse.RetMsg)
	}

	var response AccountBalanceResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	var accountData shared.AccountData

	if IS_UNIFIED {
		accountData = TransformUnifiedAccountBalance(response)
	} else {
		accountData = TransformContractAccountBalance(response)
	}

	return &accountData, nil
}

func GetBybitAccountPositions(apiKey string, secret string) (*PositionsResponse, error) {
	queryString := "settleCoin=USDT&category=linear"
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	accSignature := GenerateBybitSignature(apiKey, secret, 50000, timestamp, queryString)
	finalURL := BybitAPIEndpoint + "/v5/position/list?" + queryString

	req, err := http.NewRequest("GET", finalURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-BAPI-SIGN-TYPE", "2")
	req.Header.Set("X-BAPI-SIGN", accSignature)
	req.Header.Set("X-BAPI-API-KEY", apiKey)
	req.Header.Set("X-BAPI-TIMESTAMP", strconv.FormatInt(timestamp, 10))
	req.Header.Set("X-BAPI-RECV-WINDOW", "50000")
	req.Header.Set("Content-Type", "application/json")

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
		var errorResponse ErrorResponse
		err = json.Unmarshal(body, &errorResponse)
		if err != nil {
			return nil, err
		}
		return nil, errors.New(errorResponse.RetMsg)
	}

	var response PositionsResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}
