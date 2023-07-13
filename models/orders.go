package models

import (
	"errors"
	"fmt"

	"github.com/jinzhu/gorm"
)

type BybitOrderRequest struct {
	Category       string `json:"category"`
	Symbol         string `json:"symbol"`
	Side           string `json:"side"`
	OrderType      string `json:"orderType"`
	Qty            string `json:"qty"`
	TimeInForce    string `json:"timeInForce"`
	ReduceOnly     bool   `json:"reduce_only"`
	CloseOnTrigger bool   `json:"closeOnTrigger"`
	TakeProfit     string `json:"takeProfit"`
	StopLoss       string `json:"stopLoss"`
	SlTriggerBy    string `json:"slTriggerBy"`
	TpTriggerBy    string `json:"tpTriggerBy"`
	TpslMode       string `json:"tpslMode"`
	SlOrderType    string `json:"slOrderType"`
	SlLimitPrice   string `json:"slLimitPrice"`
}
type BybitResponse struct {
	RetCode int    `json:"retCode"`
	RetMsg  string `json:"retMsg"`
	Result  struct {
		OrderID     string `json:"orderId"`
		OrderLinkId string `json:"orderLinkId"`
	} `json:"result"`
	RetExtInfo map[string]interface{} `json:"retExtInfo"`
	Time       int64                  `json:"time"`
}

// type BybitResponse struct {
// 	RetCode    int                    `json:"retCode"`
// 	RetMsg     string                 `json:"retMsg"`
// 	Result     struct {
// 		OrderID     string `json:"order_id"`
// 		OrderLinkId string `json:"orderLinkId"`
// 	}
// 	RetExtInfo map[string]interface{} `json:"retExtInfo"`
// 	Time       int64                  `json:"time"`
// }

type OrderRequest struct {
	Symbol     string `json:"symbol"`
	MarginCoin string `json:"marginCoin"`
	Size       string `json:"size"`
	Side       string `json:"side"`
	OrderType  string `json:"orderType"`
	StopLoss   string `json:"presetStopLossPrice"`
	TakeProfit string `json:"presetTakeProfitPrice"`
}

type TriggerType string
type OrderSide string

const (
	FillPrice   TriggerType = "fill_price"
	MarketPrice TriggerType = "market_price"
)

const (
	OpenLong   OrderSide = "open_long"
	OpenShort  OrderSide = "open_short"
	CloseLong  OrderSide = "close_long"
	CloseShort OrderSide = "close_short"
	BuySingle  OrderSide = "buy_single"
	SellSingle OrderSide = "sell_single"
)

type TrailingStopOrderRequest struct {
	Symbol       string      `json:"symbol"`
	MarginCoin   string      `json:"marginCoin"`
	TriggerPrice string      `json:"triggerPrice"`
	TriggerType  TriggerType `json:"triggerType"`
	Size         string      `json:"size"`
	Side         string      `json:"side"`
	RangeRate    string      `json:"rangeRate"`
}

type BybitTrailingStopOrderRequest struct {
	Category     string `json:"category"`
	Symbol       string `json:"symbol"`
	TakeProfit   string `json:"takeProfit"`
	StopLoss     string `json:"stopLoss"`
	TrailingStop string `json:"trailingStop"`
	ActivePrice  string `json:"activePrice"`
	TpslMode     string `json:"tpslMode"`
	TpSize       string `json:"tpSize"`
	SlSize       string `json:"slSize"`
	TpOrderType  string `json:"tpOrderType"`
	SlOrderType  string `json:"slOrderType"`
	TpTriggerBy  string `json:"tpTriggerBy"`
	SlTriggerBy  string `json:"slTriggerBy"`
	TpLimitPrice string `json:"tpLimitPrice"`
	SlLimitPrice string `json:"slLimitPrice"`
	PositionIdx  int    `json:"positionIdx"`
}

type OrderResponse struct {
	Code        string `json:"code"`
	Msg         string `json:"msg"`
	RequestTime int64  `json:"requestTime"`
	Data        struct {
		ClientOid string `json:"clientOid"`
		OrderID   string `json:"orderId"`
	} `json:"data"`
}

type Order struct {
	Email       string
	Symbol      string
	MarginCoin  string
	Size        string
	Side        string
	OrderType   string
	Service     string
	QuoteAmount float64
	Profit      float64
	PositionId  int     `json:"position_id"`
	OrderPrice  string  `json:"order_price"`
	Fee         float64 `json:"fee"`
	OrderId     string  `json:"order_id"`
}

func (o *Order) Initialize(order OrderRequest, email string, client_id string, order_id string) {
	o.MarginCoin = order.MarginCoin
	o.Side = order.Side
	o.Symbol = order.Symbol
	o.Size = order.Size
	o.OrderType = order.OrderType
	o.Email = email
}

func (o *OrderRequest) Validate() error {
	if o.MarginCoin == "" {
		return errors.New("margin coin is required")
	}
	if o.OrderType == "" {
		return errors.New("ordertype is required")
	}
	if o.Side == "" {
		return errors.New("side is required")
	}
	if o.Size == "" {
		return errors.New("size is required")
	}
	if o.Symbol == "" {
		return errors.New("symbol is required")
	}
	return nil
}

func (o *Order) SaveOrder(db *gorm.DB) (*Order, error) {
	err := db.Create(&o).Error
	if err != nil {
		fmt.Println("error in saving func")
		return &Order{}, err
	}
	return o, nil
}
