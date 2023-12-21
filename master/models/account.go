package models

import (
	"time"

	"gorm.io/gorm"
)

type Account struct {
	gorm.Model
	Username   string `gorm:"not null;uniqueIndex:idx_username_server"`
	Password   string `gorm:"not null"` // MD5 hash
	Email      string `gorm:"not null"`
	Server     string `gorm:"not null;uniqueIndex:idx_username_server"`
	ExpireAt   time.Time
	UserID     uint
	RouteTasks []RouteTask `gorm:"foreignKey:AccountID"`
}

func (account *Account) ToDTO() *AccountDTO {
	return &AccountDTO{
		ID:         account.ID,
		Username:   account.Username,
		Email:      account.Email,
		Server:     account.Server,
		RouteTasks: account.RouteTasks,
	}
}

type AccountDTO struct {
	ID         uint        `json:"id"`
	Username   string      `json:"username"`
	Email      string      `json:"email"`
	Server     string      `json:"server"`
	RouteTasks []RouteTask `json:"route_tasks"`
}

func (accountDTO *AccountDTO) ToModel() *Account {
	return &Account{
		Model: gorm.Model{
			ID: accountDTO.ID,
		},
		Username:   accountDTO.Username,
		Email:      accountDTO.Email,
		Server:     accountDTO.Server,
		RouteTasks: accountDTO.RouteTasks,
	}
}
