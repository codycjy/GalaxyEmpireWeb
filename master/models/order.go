package models

import "gorm.io/gorm"

type Order struct {
	gorm.Model
	Price     int     `json:"price"`
	UserID    uint    `json:"user_id"`
	AccountID uint    `json:"account_id"`
	User      User    `gorm:"foreignKey:UserID" json:"user"`
	Account   Account `gorm:"foreignKey:AccountID" json:"account"`
	PaymentID uint    `json:"payment_id"`
}
