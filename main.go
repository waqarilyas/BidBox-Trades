package main

import (
	"os"

	"github.com/sirupsen/logrus"

	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/joho/godotenv"
	"github.com/kryptomind/BidBox-Trades/controllers"
	"github.com/kryptomind/BidBox-Trades/utils"
)

var server = controllers.Server{}

// var binance_WS = binance_websockets.Server{}

func Run() {
	err := godotenv.Load()
	log := logrus.New()

	log.SetFormatter(&nested.Formatter{
		HideKeys:    true,
		FieldsOrder: []string{"file", "function"},
	})
	if err != nil {
		log.WithFields(logrus.Fields{
			"file":     "main.go",
			"function": "Run",
		}).Fatal("Error getting env")
	} else {
		log.WithFields(logrus.Fields{
			"file":     "main.go",
			"function": "Run",
		}).Info("Getting Values")
	}

	server.Initialize(os.Getenv("DB_DRIVER"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_PORT"), os.Getenv("DB_HOST"), os.Getenv("DB_NAME"))

	utils.InitSharedData(server.DB)
	// binance_WS.DB = server.DB
	// bitget.TestBitget()

	server.Run(":8080")
}

func main() {
	// c := cron.New()

	// trade_cron := trade.TradeCron{}
	// c.AddFunc("@every 20s", trade_cron.Run)

	// pnl_cron := pnl.PnlCron{}
	// c.AddFunc("@every 5s", pnl_cron.Run)

	// c.Start()

	Run()
}
