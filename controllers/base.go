package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres" //postgres database driver
	"github.com/kryptomind/BidBox-Trades/models"

	log "github.com/sirupsen/logrus"
)

type Server struct {
	DB     *gorm.DB
	Router *mux.Router
}

type AppData struct {
	Keys_list  []models.Key
	Conditions []models.Conditions
}

var app_data AppData

func (server *Server) Initialize(Dbdriver, DbUser, DbPassword, DbPort, DbHost, DbName string) {

	var err error
	DBURL := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=disable password=%s", DbHost, DbPort, DbUser, DbName, DbPassword)
	server.DB, err = gorm.Open(Dbdriver, DBURL)
	if err != nil {
		log.Info("Cannot connect to the %s database", Dbdriver)
		log.Fatal("This is the error:", err)
	} else {
		log.Info("Connected to the %s database", Dbdriver)
	}

	// server.DB.Debug().AutoMigrate(&models.Key{}) //database migration
	server.Router = mux.NewRouter()
	server.initializeRoutes()
	server.InitKeys()
	//server.InitConditions()
}

type IndexPriceReq struct {
	Code string `json:"code"`
	Data struct {
		Symbol    string `json:"symbol"`
		Index     string `json:"index"`
		Timestamp string `json:"timestamp"`
	} `json:"data"`
	Msg         string `json:"msg"`
	RequestTime int64  `json:"requestTime"`
}

func GetSize(symbol string, first_order float64) (float64, error) {
	url := "https://api.bitget.com/api/mix/v1/market/index?symbol=" + symbol

	client := http.Client{}

	res, err := client.Get(url)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()

	price := IndexPriceReq{}
	if err := json.NewDecoder(res.Body).Decode(&price); err != nil {
		return 0, err
	}

	if price.Code != "00000" {
		return 0, errors.New(price.Msg)
	}

	fprice, err := strconv.ParseFloat(price.Data.Index, 64)
	if err != nil {
		return 0, err
	}

	return first_order / fprice, nil
}

func (server *Server) InitKeys() {
	key := models.Key{}
	keys, err := key.FindAllKeys(server.DB)
	if err != nil {
		log.Fatal("error getting keys")
		app_data.Keys_list = []models.Key{}
		return
	}
	log.Info("retreived keys")
	app_data.Keys_list = *keys
}

func (server *Server) InitConditions() {
	cond := models.Conditions{}
	conds, err := cond.FindAllConditions(server.DB)
	if err != nil {
		log.Fatal("error getting keys")
		app_data.Conditions = []models.Conditions{}
		return
	}
	log.Info("retreived conditions")
	app_data.Conditions = *conds
}

func (server *Server) Run(addr string) {
	log.Info("Listening on port 8080")
	log.Fatal(http.ListenAndServe(addr, server.Router))

}
