package models

import (
	"errors"

	"github.com/jinzhu/gorm"
)

type OrderRequest struct {
	Symbol     string `json:"symbol"`
	MarginCoin string `json:"marginCoin"`
	Size       string `json:"size"`
	Side       string `json:"side"`
	OrderType  string `json:"orderType"`
	StopLoss   string `json:"presetStopLossPrice"`
	TakeProfit string `json:"presetTakeProfitPrice"`
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
	Email      string
	Symbol     string
	MarginCoin string
	Size       string
	Side       string
	OrderType  string
	Service    string
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
		return &Order{}, err
	}
	return o, nil
}
