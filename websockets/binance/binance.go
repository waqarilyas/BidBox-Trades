package binance_websockets

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/jinzhu/gorm"
	"github.com/kryptomind/BidBox-Trades/models"
)

type Server struct {
	DB     *gorm.DB
	Router *mux.Router
}

type MarketEvent struct {
	Symbol      string `json:"s"`
	MarketPrice string `json:"p"`
}

func (s *Server) WebsocketTest() {

	var paramsList []string

	coinPair := models.CoinPair{}
	coinPairs, err := coinPair.GetAllCoins(s.DB)
	if err != nil {
		fmt.Println("---- error fetching coins ----", err)
		return
	}

	for _, coinPair := range *coinPairs {
		result := strings.ReplaceAll(coinPair.Coin, "/", "")
		eventString := strings.ToLower(result)
		eventString = fmt.Sprintf("%s%s", eventString, "@markPrice")
		paramsList = append(paramsList, eventString)
	}

	url := "wss://fstream.binance.com/ws/"
	dialer := &websocket.Dialer{}

	conn, _, err := dialer.Dial(url, nil)
	if err != nil {
		log.Fatal("WebSocket connection error:", err)
	}
	defer conn.Close()

	go func() {
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("WebSocket message receiving error:", err)
				break
			}

			var eventData MarketEvent

			err = json.Unmarshal(message, &eventData)
			if err != nil {
				return
			}
			go handleMarketUpdate(eventData.Symbol, eventData.MarketPrice)

		}
	}()

	subscribeRequest := struct {
		Method string   `json:"method"`
		Params []string `json:"params"`
		ID     int      `json:"id"`
	}{
		Method: "SUBSCRIBE",
		Params: paramsList,
		ID:     1,
	}

	err = conn.WriteJSON(subscribeRequest)
	if err != nil {
		log.Println("WebSocket message sending error:", err)
		return
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	for {
		select {
		case <-interrupt:
			log.Println("Received interrupt signal. Closing WebSocket connection...")
			err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("WebSocket close message sending error:", err)
			}
			time.Sleep(1 * time.Second) // Wait for the server to close the connection
			return
		}
	}
}

func handleMarketUpdate(coinPair string, marketPrice string) {
	fmt.Println("---- coin pair ----", coinPair, " ---- ", marketPrice)

}
