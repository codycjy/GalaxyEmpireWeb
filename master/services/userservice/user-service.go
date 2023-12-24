package userservice

import (
	"GalaxyEmpireWeb/models"
	"errors"

	"gorm.io/gorm"
)

type UserService struct {
	DB *gorm.DB
}

var userServiceInstance *UserService

func NewService(db *gorm.DB) *UserService {
	return &UserService{
		DB: db,
	}
}

func InitService(db *gorm.DB) error {
	if userServiceInstance != nil {
		return errors.New("UserService is already initialized")
	}
	userServiceInstance = NewService(db)
	return nil
}

func GetService() (*UserService, error) {
	if userServiceInstance == nil {
		return nil, errors.New("UserService is not initialized")
	}
	return userServiceInstance, nil
}

func (service *UserService) GetById(id uint, fields []string) (*models.User, error) {
	var user models.User
	cur := service.DB
	for _, field := range fields {
		cur.Preload(field)
	}
	err := cur.Where("id = ?", id).First(&user).Error
	return &user, err
}

func (service *UserService) Create(user *models.User) error {
	return service.DB.Create(user).Error
}

func (service *UserService) Update(user *models.User) error {
	return service.DB.Save(user).Error
}

func (service *UserService) Delete(id uint) error {
	return service.DB.Delete(&models.User{}, id).Error
}

func (service *UserService) GetAllUsers() (users []models.User, err error) {
	err = service.DB.Find(&users).Error
	return users, err
}

func (service *UserService) UpdateBalance(user *models.User) error {
	result := service.DB.
		Model(&models.User{}).
		Where("id = ?", user.ID).
		Update("balance", user.Balance)

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return result.Error

}
