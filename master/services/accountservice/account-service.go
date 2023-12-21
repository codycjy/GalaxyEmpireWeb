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

var accountService *AccountService
func NewService(db *gorm.DB) *AccountService {
	return &AccountService{
		DB: db,
	}
}
func InitService(db *gorm.DB) error {
	if accountService != nil {
		return errors.New("AccountService is already initialized")
	}
	accountService = NewService(db)
	return nil
}
func GetService() (*AccountService,error) {
	if accountService == nil {
		return nil, errors.New("AccountService is not initialized")
	}
	return accountService, nil
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

func (service *AccountService) GetByUserId(userId uint) (*[]models.Account, error) {
	userservice := userservice.NewService(service.DB)
	user, err := userservice.GetById(userId, []string{"Accounts"})
	if err != nil {
		return nil, err
	}
	return &user.Accounts, nil
}

// TODO: Create
