package bitget

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/kryptomind/BidBox-Trades/helpers"
	shared_utils "github.com/kryptomind/BidBox-Trades/utils"
	// "github.com/kryptomind/bidboxapi/KeyService/helpers"
	// "github.com/kryptomind/bidboxapi/AccountsService/api/helpers"
)

func GetBitgetAccountData(apiKey string, apiSecret string, passphrase string) (*AccountData, error) {
	expires := helpers.GetBitgetServerTimeStamp()

	host := "https://api.bitget.com"
	path := "/api/mix/v1/account/accounts?productType=" + shared_utils.PRODUCT_TYPE
	url := host + path

	signature := GenerateBitgetSignature(apiSecret, "GET", path, expires)

	fmt.Println(passphrase)
	fmt.Println(apiSecret)
	fmt.Println(apiKey)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("ACCESS-KEY", apiKey)
	req.Header.Add("ACCESS-PASSPHRASE", passphrase)
	req.Header.Add("ACCESS-TIMESTAMP", expires)
	req.Header.Add("ACCESS-SIGN", signature)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Print(err.Error())

		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {

		return nil, err
	}

	var accountData AccountData

	err = json.Unmarshal(body, &accountData)
	if err != nil {
		return nil, err
	}

	return &accountData, nil
}
