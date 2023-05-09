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

func (u *Conditions) FindAllConditions(db *gorm.DB) (*[]Conditions, error) {
	conditions := []Conditions{}
	err := db.Debug().Model(Conditions{}).Limit(100).Take(conditions).Error
	if err != nil {
		return &[]Conditions{}, err
	}
	return &conditions, nil
}

func (u *Conditions) FindCondition(db *gorm.DB, capital int) (*Conditions, error) {
	err := db.Debug().Model(Conditions{}).Where("capital = ?", capital).Take(&u).Error
	if err != nil {
		return &Conditions{}, err
	}
	if gorm.IsRecordNotFoundError(err) {
		return &Conditions{}, errors.New("Key not found")
	}
	return u, nil
}
