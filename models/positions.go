package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

type Position struct {
	Id           int       `gorm:"primary_key:autoIncrement" json:"id"`
	CreatedAt    time.Time `gorm:"type:timestamptz;default:now()" json:"created_at"`
	UpdatedAt    time.Time `gorm:"type:timestamptz:;default:now()" json:"updated_at"`
	Symbol       string    `json:"symbol"`
	Leverage     string    `json:"leverage"`
	OpenPrice    float64   `json:"open_price"`
	LiqPrice     float64   `json:"liq_price"`
	TakeProfit   float64   `json:"take_profit"`
	StopLoss     float64   `json:"stop_loss"`
	UnrealizedPl float32   `json:"unrealized_pl"`
	Markprice    float64   `json:"mark_price"`
	Side         string    `json:"side"`
	Size         string    `json:"size"`
	Margin       string    `json:"margin"`
	UserEmail    string    `gorm:"not null" json:"user_email"`
	Status       string    `gorm:"default:opened" json:"status"`
}

func (position *Position) CreateNewPosition(db *gorm.DB) (*Position, error) {
	err := db.Create(&position).Error
	if err != nil {
		return &Position{}, err
	}
	return position, nil
}

func (u *Key) GetAllPositions(db *gorm.DB) (*[]Key, error) {
	Keys := []Key{}
	err := db.Model(&Key{}).Limit(100).Find(&Keys).Error
	if err != nil {
		return &[]Key{}, err
	}
	return &Keys, nil
}
