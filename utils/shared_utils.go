package utils

import (
	"fmt"
)

var IS_TESTNET bool = true
var PRODUCT_TYPE string = "UMCBL"
var MARGIN_COIN string = "USDT"

func InitSharedData(IS_TESTNET bool) {

	if IS_TESTNET {
		PRODUCT_TYPE = "SUMCBL"
		MARGIN_COIN = "SUSDT"
	} else {
		PRODUCT_TYPE = "UMCBL"
		MARGIN_COIN = "USDT"
	}
	fmt.Println("🚀 ~ file: shared_utils.go:27 ~ funcInitSharedData ~ MARGIN_COIN:", MARGIN_COIN)

}
