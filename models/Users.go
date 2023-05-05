package models

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

type User struct {
	Id            uuid.UUID `gorm:"primary_key;auto_increment" json:"id"`
	Email         string    `gorm:"size:100;not null;unique" json:"email"`
	Password      string    `gorm:"size:100;not null;" json:"password"`
	Confirmed     bool      `gorm:"default:false" json:"confirmed"`
	CreatedAt     time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt     time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
	OpenPositions int       `json:"openPositions"`
}

func (u *User) GetAllUsers(db *gorm.DB) (*[]User, error) {
	Users := []User{}
	err := db.Debug().Model(&Key{}).Limit(100).Find(&Users).Error
	if err != nil {
		return &[]User{}, err
	}
	return &Users, nil
}

func (u *User) FindUserById(db *gorm.DB, uid uint32) (*User, error) {
	err := db.Debug().Model(User{}).Where("id = ?", uid).Take(&u).Error
	if err != nil {
		return &User{}, err
	}
	if gorm.IsRecordNotFoundError(err) {
		return &User{}, errors.New("User not found")
	}
	return u, nil
}

func (u *User) FindUserByEmail(db *gorm.DB, email string) (*User, error) {
	err := db.Debug().Model(User{}).Where("email = ?", email).Take(&u).Error
	if err != nil {
		return &User{}, err
	}
	if gorm.IsRecordNotFoundError(err) {
		return &User{}, errors.New("User not found")
	}
	return u, nil
}

func (u *User) ChangePositions(db *gorm.DB, pos int) (*User, error) {
	db = db.Debug().Model(&User{}).Where("email = ?", u.Email).Take(&User{}).UpdateColumns(
		map[string]interface{}{
			"open_positions": pos,
			"updated_at":     time.Now(),
		},
	)
	if db.Error != nil {
		return &User{}, db.Error
	}
	// This is the display the updated user
	err := db.Debug().Model(&User{}).Where("email = ?", u.Email).Take(&u).Error
	if err != nil {
		return &User{}, err
	}
	return u, nil
}
