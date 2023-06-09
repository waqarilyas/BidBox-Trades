package helpers

import (
	"encoding/json"
	// "errors"
	"fmt"
	"io/ioutil"
	"net/http"
)
type bitgetServerTimeStampResponse struct {
	Code        string `json:"code"`
	Msg         string `json:"msg"`
	RequestTime int    `json:"requestTime"`
	Data        string `json:"data"`
}


func GetBitgetServerTimeStamp() string {

	url := "https://api.bitget.com/api/spot/v1/public/time"
	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		fmt.Print("-----error in request---", err.Error())
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Print(err.Error())
	}

	defer res.Body.Close()
	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		fmt.Print("---read error---", err.Error())
	}

	var response bitgetServerTimeStampResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		fmt.Print("---json unmarshal error---", err.Error())
	}

	return response.Data

}