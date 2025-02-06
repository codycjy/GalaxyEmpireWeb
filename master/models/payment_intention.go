package models

import (
	"time"

	"gorm.io/gorm"
)

type PaymentIntention struct {
	gorm.Model
	UserID        uint `gorm:"index;not null"`
	User          User `gorm:"foreignKey:UserID"`
	Amount        int64
	Currency      string
	Status        string // pending, completed, failed // TODO: new type later
	SessionID     string `gorm:"index"`
	PaymentIntent string
	CompletedAt   *time.Time
}
