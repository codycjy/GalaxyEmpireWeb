package accountservice

import (
	"GalaxyEmpireWeb/logger"
	"GalaxyEmpireWeb/models"
	"GalaxyEmpireWeb/services/userservice"
	"GalaxyEmpireWeb/utils"
	"context"
	"errors"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type AccountService struct {
	DB *gorm.DB
}

var accountServiceInstance *AccountService
var log = logger.GetLogger()

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
func GetService(ctx context.Context) (*AccountService, error) {
	traceID := utils.TraceIDFromContext(ctx)
	if accountServiceInstance == nil {
		log.Fatal("[service]AccountService is not initialized",
			zap.String("traceID", traceID),
		)
		return nil, errors.New("AccountService is not initialized")
	}
	return accountServiceInstance, nil
}

func (service *AccountService) GetById(ctx context.Context, id uint, fields []string) (*models.Account, error) {
	traceID := utils.TraceIDFromContext(ctx)
	log.Info("[service]Get Account By ID",
		zap.Uint("id", id),
		zap.Strings("fields", fields),
		zap.String("traceID", traceID),
	)
	var account models.Account
	cur := service.DB
	for _, field := range fields {
		cur.Preload(field)
	}
	err := cur.Where("id = ?", id).First(&account).Error
	if err != nil {
		log.Error("[service]Get Account By ID failed",
			zap.String("traceID", traceID),
			zap.Error(err),
		)
	}
	return &account, err
}

func (service *AccountService) GetByUserId(ctx context.Context, userId uint, fields []string) (*[]models.Account, error) {
	traceID := utils.TraceIDFromContext(ctx)
	log.Info("[service]Get Account By User ID",
		zap.Uint("userId", userId),
		zap.Strings("fields", fields),
		zap.String("traceID", traceID),
	)
	userservice := userservice.NewService(service.DB)
	fields = append(fields, "Accounts")
	user, err := userservice.GetById(ctx, userId, fields)

	if err != nil {
		log.Error("[service]Get Account By User ID failed",
			zap.String("traceID", traceID),
			zap.Error(err),
		)
		return nil, err
	}
	return &user.Accounts, nil
}

func (service *AccountService) Update(ctx context.Context, account *models.Account) error {
	traceID := utils.TraceIDFromContext(ctx)
	log.Info("[service]Update Account Info",
		zap.Uint("userId", account.ID),
		zap.String("username", account.Username),
		zap.String("traceID", traceID),
	)
	err := service.DB.Save(account).Error
	if err != nil {
		log.Error("[service]Update Account failed",
			zap.String("traceID", traceID),
			zap.Error(err),
		)
		return err
	}
	return nil
}

func (service *AccountService) Delete(ctx context.Context, ID uint) error {
	traceID := utils.TraceIDFromContext(ctx)
	log.Info("[service]Update Account Info",
		zap.Uint("userId", ID),
		zap.String("traceID", traceID),
	)
	err := service.DB.Delete(&models.Account{}, ID).Error
	if err != nil {
		log.Info("[service]Delete Account failed",
			zap.String("traceID", traceID),
		)
		return err

	}
	return nil

}
