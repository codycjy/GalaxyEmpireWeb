package models

import (
	"GalaxyEmpireWeb/repositories/mysql"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username string `gorm:"unique;not null"`
	// WARNING: USERNAME MAY BE NOT UNIQUE! RECHECK THIS!
	Password string `gorm:"not null"`
	Balance  int    `gorm:"not null"`
}

func (u *User) GetBalance() int {
	return u.Balance
}

func (u *User) Create() error {
	db := mysql.GetDB()
	err := db.Create(u)
	if err != nil {
		return err.Error
	}
	return nil
}

func (u *User) Update() error {
	db := mysql.GetDB()
	err := db.Save(u)
	if err != nil {
		return err.Error
	}
	return nil
}

func (u *User) Delete() error {
	db := mysql.GetDB()
	err := db.Delete(u)
	if err != nil {
		return err.Error
	}
	return nil
}

func (u *User) GetByUsername(username string) error {
	db := mysql.GetDB()
	err := db.Where("username = ?", username).First(u)
	if err != nil {
		return err.Error
	}
	return nil
}

func (u *User) GetByID(id uint) error {
	db := mysql.GetDB()
	err := db.First(u, id)
	if err != nil {
		return err.Error
	}
	return nil
}

func GetAllUsers(users *[]User) error {
	db := mysql.GetDB()
	err := db.Find(&users)
	if err != nil {
		return err.Error
	}
	return nil
}
