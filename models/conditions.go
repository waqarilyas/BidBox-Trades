package models

import (
	"errors"

	"github.com/jinzhu/gorm"
)

type Conditions struct {
	Capital   int
	Positions int
	Leverage  int
	StopLoss  int
}

func (u *Conditions) FindKeyById(db *gorm.DB, capital int) (*Conditions, error) {
	err := db.Debug().Model(Key{}).Where("capital = ?", capital).Take(&u).Error
	if err != nil {
		return &Conditions{}, err
	}
	if gorm.IsRecordNotFoundError(err) {
		return &Conditions{}, errors.New("Key not found")
	}
	return u, nil
}
