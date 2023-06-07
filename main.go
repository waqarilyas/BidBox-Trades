package main

import (
	"os"

	"github.com/sirupsen/logrus"

	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/joho/godotenv"
	"github.com/kryptomind/BidBox-Trades/controllers"
	binance_websockets "github.com/kryptomind/BidBox-Trades/websockets/binance"
)

var server = controllers.Server{}
var binance_WS = binance_websockets.Server{}

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
	binance_WS.DB = server.DB
	binance_WS.WebsocketTest()
	server.Run(":8080")
}

func main() {
	Run()
}
