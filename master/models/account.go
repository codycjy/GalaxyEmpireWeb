package models

import (
	"time"

	"gorm.io/gorm"
)

// Account represents a user account in the system.
// It includes fields for the username, password, email, server, and related tasks.
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

// ToDTO converts an Account to an AccountDTO.
func (account *Account) ToDTO() *AccountDTO {
	return &AccountDTO{
		ID:         account.ID,
		Username:   account.Username,
		Email:      account.Email,
		Server:     account.Server,
		RouteTasks: account.RouteTasks,
	}
}

// AccountDTO is a data transfer object for Account.
// It is used when interacting with external systems.
type AccountDTO struct {
	ID         uint        `json:"id"`
	Username   string      `json:"username"`
	Email      string      `json:"email"`
	Server     string      `json:"server"`
	RouteTasks []RouteTask `json:"route_tasks"`
}

// ToModel converts an AccountDTO to an Account.
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
