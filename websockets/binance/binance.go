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

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	for {
		conn, err := connectWebSocket()
		if err != nil {
			log.Println("WebSocket connection error:", err)
			time.Sleep(5 * time.Second) // Wait for 5 seconds before reconnecting
			continue
		}

		err = subscribeToMarketEvents(conn, paramsList)
		if err != nil {
			log.Println("WebSocket subscribe error:", err)
			conn.Close()
			time.Sleep(5 * time.Second) // Wait for 5 seconds before reconnecting
			continue
		}

		go handleWebSocketMessages(conn)

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

func connectWebSocket() (*websocket.Conn, error) {
	url := "wss://fstream.binance.com/ws/"
	dialer := &websocket.Dialer{}

	conn, _, err := dialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func subscribeToMarketEvents(conn *websocket.Conn, paramsList []string) error {
	subscribeRequest := struct {
		Method string   `json:"method"`
		Params []string `json:"params"`
		ID     int      `json:"id"`
	}{
		Method: "SUBSCRIBE",
		Params: paramsList,
		ID:     1,
	}

	err := conn.WriteJSON(subscribeRequest)
	if err != nil {
		return err
	}

	return nil
}

func handleWebSocketMessages(conn *websocket.Conn) {
	defer conn.Close()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("WebSocket message receiving error:", err)
			return
		}

		var eventData MarketEvent

		err = json.Unmarshal(message, &eventData)
		if err != nil {
			log.Println("WebSocket message parsing error:", err)
			continue
		}

		go handleMarketUpdate(eventData.Symbol, eventData.MarketPrice)
	}
}

func handleMarketUpdate(coinPair string, marketPrice string) {
	fmt.Println("---- coin pair ----", coinPair, " ---- ", marketPrice)
}
