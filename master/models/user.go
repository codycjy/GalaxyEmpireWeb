package models

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username string `gorm:"unique;not null"`
	// WARNING: USERNAME MAY BE NOT UNIQUE! RECHECK THIS!
	// NOTE: Checked in db, DO api check
	Password string `gorm:"not null"`
	Balance  int
	Role     int       // 0: normal user, 1: admin
	Accounts []Account `gorm:"foreignKey:UserID"`
}

func (user *User) ToDTO() *UserDTO {
	accountDTOs := make([]AccountDTO, len(user.Accounts))
	for i, account := range user.Accounts {
		accountDTOs[i] = *account.ToDTO()
	}
	return &UserDTO{
		ID:       user.ID,
		Username: user.Username,
		Balance:  user.Balance,
		Accounts: accountDTOs,
	}
}

type UserDTO struct {
	ID       uint         `json:"id"`
	Username string       `json:"username"`
	Accounts []AccountDTO `json:"accounts"`
	Balance  int          `json:"balance"`
}

func (userDTO *UserDTO) ToModel() *User {
	accountModels := make([]Account, len(userDTO.Accounts))
	for i, accountDTO := range userDTO.Accounts {
		accountModels[i] = *accountDTO.ToModel()
	}
	return &User{
		Model: gorm.Model{
			ID: userDTO.ID,
		},
		Username: userDTO.Username,
		Balance:  userDTO.Balance,
		Accounts: accountModels,
	}
}
