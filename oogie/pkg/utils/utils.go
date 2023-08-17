package utils

import (
	"encoding/json"
	"errors"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/hirokisan/bybit/v2"
)

func SelectRandomElement() string {
	rand.Seed(time.Now().UnixNano())
	index := rand.Intn(len(AllowedCoins))
	return AllowedCoins[index]
}

type SymbolDetails struct {
	MarkPrice string `json:"markPrice"`
}

type ResultData struct {
	List []SymbolDetails `json:"list"`
}

type IndexPriceReqBybit struct {
	RetCode    int         `json:"retCode"`
	RetMsg     string      `json:"retMsg"`
	Result     ResultData  `json:"result"`
	RetExtInfo interface{} `json:"retExtInfo"`
	Time       int64       `json:"time"`
}

func GetSizeBybit(symbol string, first_order float64) (float64, float64, error) {
	url := "https://api-testnet.bybit.com/v5/market/tickers?category=inverse&symbol=" + symbol

	client := http.Client{}

	res, err := client.Get(url)
	if err != nil {
		return 0, 0, err
	}
	defer res.Body.Close()

	price := IndexPriceReqBybit{}
	if err := json.NewDecoder(res.Body).Decode(&price); err != nil {
		return 0, 0, err
	}

	if price.RetCode != 0 {
		return 0, 0, errors.New(price.RetMsg)
	}

	fprice, err := strconv.ParseFloat(price.Result.List[0].MarkPrice, 64)
	if err != nil {
		return 0, 0, err
	}

	return first_order / fprice, fprice, nil
}

func GetPositionsTest() (*bybit.V5GetPositionInfoResponse, error) {

	coin := bybit.CoinUSDT
	params := bybit.V5GetPositionInfoParam{
		Category:   bybit.CategoryV5Linear,
		SettleCoin: &coin,
	}

	client := bybit.NewTestClient().
		WithAuth(Apikey, Secretkey)

	resp, err := client.V5().Position().GetPositionInfo(params)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func GetPositions(api_key string, secret_key string, symbol string) (string, string) {

	params := bybit.V5GetPositionInfoParam{
		Category: bybit.CategoryV5Linear,
		Symbol:   (*bybit.SymbolV5)(&symbol),
	}

	client := bybit.NewTestClient().
		WithAuth(api_key, secret_key)

	resp, err := client.V5().Position().GetPositionInfo(params)
	if err != nil {
		log.Fatal(err)
	}

	size := make([]string, 0)
	for _, v := range resp.Result.List {
		size = append(size, v.Size)
	}

	return size[0], size[1]
}
