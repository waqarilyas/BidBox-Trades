package utils

import (
	"fmt"

	"github.com/jinzhu/gorm"
	"github.com/kryptomind/bidboxapi/AccountsService/api/models/admin"
)

var IS_TESTNET bool = true
var PRODUCT_TYPE string = "UMCBL"
var MARGIN_COIN string = "USDT"

func InitSharedData(db *gorm.DB) {
	var set admin.Settings
	settings, err := set.GetSettings(db)
	if err != nil {

		return
	}
	IS_TESTNET = settings.IsTestnet

	if IS_TESTNET {
		PRODUCT_TYPE = "SUMCBL"
		MARGIN_COIN = "SUSDT"
	} else {
		PRODUCT_TYPE = "UMCBL"
		MARGIN_COIN = "USDT"
	}
	fmt.Println("🚀 ~ file: shared_utils.go:27 ~ funcInitSharedData ~ MARGIN_COIN:", MARGIN_COIN)

}
