package bitget

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
	"github.com/kryptomind/BidBox-Trades/exchange/bitget"
)

type ArgsData struct {
	APIKey     string `json:"apiKey"`
	Passphrase string `json:"passphrase"`
	Timestamp  string `json:"timestamp"`
	Sign       string `json:"sign"`
}

type LoginReq struct {
	Op   string     `json:"op"`
	Args []ArgsData `json:"args"`
}

func readLoop(conn *websocket.Conn) {
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("read error:", err)
			return
		}
		fmt.Println("Received message:", string(message))
	}
}

func TestBitget() {
	url := "wss://ws.bitget.com/mix/v1/stream"

	dialer := websocket.DefaultDialer

	conn, _, err := dialer.Dial(url, nil)
	if err != nil {
		log.Fatal("dial error:", err)
	}
	time := fmt.Sprintf("%d", time.Now().UnixMilli())
	api_key := "bg_1777318a8c54c7e6e773330ff92d3b7e"
	passphrase := "thisisapassphrase"
	api_secret := "3c2fe03569e49d2809bc9449cf443b7a86fa40eb8821e4ded783af996e063872"
	sign := bitget.GenerateBitgetSignature(api_secret, "GET", "/user/verify", time)

	login := LoginReq{
		Op: "login",
		Args: []ArgsData{
			{
				APIKey:     api_key,
				Passphrase: passphrase,
				Timestamp:  time,
				Sign:       sign,
			},
		},
	}

	b, _ := json.Marshal(&login)

	defer conn.Close()

	message := b
	err = conn.WriteMessage(websocket.TextMessage, message)
	if err != nil {
		log.Println("write error:", err)
	}

	go readLoop(conn)

	select {}
}
