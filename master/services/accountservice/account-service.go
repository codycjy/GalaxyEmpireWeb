package accountservice

import (
	"GalaxyEmpireWeb/models"
	"GalaxyEmpireWeb/services/userservice"
	"errors"

	"gorm.io/gorm"
)

type AccountService struct {
	DB *gorm.DB
}

var accountServiceInstance *AccountService

func NewService(db *gorm.DB) *AccountService {
	return &AccountService{
		DB: db,
	}
}
func InitService(db *gorm.DB) error {
	if accountServiceInstance != nil {
		return errors.New("AccountService is already initialized")
	}
	accountServiceInstance = NewService(db)
	return nil
}
func GetService() (*AccountService, error) {
	if accountServiceInstance == nil {
		return nil, errors.New("AccountService is not initialized")
	}
	return accountServiceInstance, nil
}

func (service *AccountService) GetById(id uint, fields []string) (*models.Account, error) {
	var account models.Account
	cur := service.DB
	for _, field := range fields {
		cur.Preload(field)
	}
	err := cur.Where("id = ?", id).First(&account).Error
	return &account, err
}

func (service *AccountService) GetByUserId(userId uint,fields []string) (*[]models.Account, error) {
	userservice := userservice.NewService(service.DB)
	fields = append(fields, "Accounts")
	user, err := userservice.GetById(userId, fields)
	if err != nil {
		return nil, err
	}
	return &user.Accounts, nil
}

func (service *AccountService) Update(account *models.Account) error {
	return service.DB.Save(account).Error
}

func (service *AccountService) Delete(ID uint) error {
	return service.DB.Delete(&models.Account{}, ID).Error

}
