package main

import (
	"context"
	"fmt"
	"os"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/sirupsen/logrus"

	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/joho/godotenv"
	"github.com/kryptomind/BidBox-Trades/controllers"
)

var server = controllers.Server{}

func Run() {
	err := godotenv.Load()
	log := logrus.New()
	// file, err := os.OpenFile("output.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// // Create a multi-writer that writes to both the file and stdout
	// writer := io.MultiWriter(os.Stdout, file)

	// // Set the log output to the multi-writer
	// log.SetOutput(writer)

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
	server.Run(":8080")
}

var (
	apiKey    = "00489718282f193987342a3c748f15193821e981d06de54988008f972fb42796"
	secretKey = "6479cecf6661c4073f11583306e41d486dfb320c2d4cc544817de61166bb82ba"
)

func main() {
	futures.UseTestnet = true
	BinanceClient := futures.NewClient(apiKey, secretKey)

	order, err := BinanceClient.NewCreateOrderService().Symbol("XRPUSDT").
		Side(futures.SideTypeBuy).Type(futures.OrderTypeLimit).
		TimeInForce(futures.TimeInForceTypeGTC).Quantity("100").
		NewOrderResponseType(futures.NewOrderRespTypeRESULT).
		Price("0.2").Do(context.Background())
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(order)

	Run()
}
