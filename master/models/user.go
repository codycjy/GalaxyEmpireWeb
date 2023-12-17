package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username string `gorm:"unique;not null"`
	// WARNING: USERNAME MAY BE NOT UNIQUE! RECHECK THIS!
	Password string `gorm:"not null"`
	Balance  int    `gorm:"not null"`
}
