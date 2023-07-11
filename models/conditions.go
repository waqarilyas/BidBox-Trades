package models

import (
	"errors"

	"github.com/jinzhu/gorm"
)

type ConditionsV2 struct {
	Capital    int
	FirstOrder float64
	Leverage   int
	Coins      int
}

func (u *ConditionsV2) FindAllConditions(db *gorm.DB) (*[]ConditionsV2, error) {
	conditions := []ConditionsV2{}
	err := db.Model(ConditionsV2{}).Find(conditions).Error
	if err != nil {
		return &[]ConditionsV2{}, err
	}
	return &conditions, nil
}

func (u *ConditionsV2) FindCondition(db *gorm.DB, capital int) (*ConditionsV2, error) {
	cond := ConditionsV2{}
	err := db.Model(ConditionsV2{}).Where("capital = ?", capital).Take(&cond).Error
	if err != nil {
		return &ConditionsV2{}, err
	}
	if gorm.IsRecordNotFoundError(err) {
		return &ConditionsV2{}, errors.New("Key not found")
	}
	return &cond, nil
}
