package accountservice

import (
	"GalaxyEmpireWeb/logger"
	"GalaxyEmpireWeb/models"
	"GalaxyEmpireWeb/services/userservice"
	"GalaxyEmpireWeb/utils"
	"context"
	"errors"
	"fmt"
	"time"

	r "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type AccountService struct {
	DB  *gorm.DB
	RDB *r.Client
}

var accountServiceInstance *AccountService
var log = logger.GetLogger()
var accountListPrefix = "user_account_"

func NewService(db *gorm.DB, rdb *r.Client) *AccountService {
	return &AccountService{
		DB:  db,
		RDB: rdb,
	}
}
func InitService(db *gorm.DB, rdb *r.Client) error {
	if accountServiceInstance != nil {
		return errors.New("AccountService is already initialized")
	}
	accountServiceInstance = NewService(db, rdb)
	return nil
}
func GetService(ctx context.Context) (*AccountService, error) {
	traceID := utils.TraceIDFromContext(ctx)
	if accountServiceInstance == nil {
		log.DPanic("[service]AccountService is not initialized",
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
	userservice, _ := userservice.GetService(ctx)
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

func (service *AccountService) isUserAllowed(ctx context.Context, accountID uint) (bool, error) {
	traceID := utils.TraceIDFromContext(ctx)
	userID := ctx.Value("userID").(uint)
	log.Info("[service]Check User Permission",
		zap.Uint("userID", userID),
		zap.Uint("accountID", accountID),
		zap.String("traceID", traceID),
	)

	key := fmt.Sprintf("%s%d", accountListPrefix, userID)

	// 首先检查键是否存在
	exists, err := service.RDB.Exists(ctx, key).Result()
	if err != nil {
		log.Error("[service]Check User Permission failed",
			zap.String("traceID", traceID),
			zap.String("redis_key", key),
			zap.Error(err),
		)
		return false, err
	}
	if exists == 0 {
		log.Info("[service]Check User Permission - Not in Redis. Retrieving",
			zap.String("traceID", traceID),
		)
		accounts, err := service.GetByUserId(ctx, userID, []string{})
		if err != nil {
			return false, err
		}
		var accountIDs = make([]uint, len(*accounts))
		// NOTE: Could be optimized by early return
		for i, account := range *accounts {
			accountIDs[i] = account.ID
		}
		err = service.cacheUserAccounts(ctx, userID, accountIDs)
		if err != nil {
			return false, err
		}

	}

	// 如果键存在，检查集合中是否包含特定元素
	isMember, err := service.RDB.SIsMember(ctx, key, accountID).Result()
	if err != nil {
		log.Error("[service]Check User Permission failed",
			zap.String("traceID", traceID),
			zap.String("redis_key", key),
			zap.Error(err),
		)
		return false, err
	}
	log.Info("[service]Check User Permission - Success",
		zap.String("traceID", traceID),
		zap.Bool("isMember", isMember),
	)
	return isMember, nil
}

func (service *AccountService) cacheUserAccounts(ctx context.Context, userID uint, accountIDs []uint) error {
	key := fmt.Sprintf("%s%d", accountListPrefix, userID)

	// 使用pipeline批量添加元素以提高性能
	pipe := service.RDB.Pipeline()
	for _, accountID := range accountIDs {
		pipe.SAdd(ctx, key, accountID)
	}

	pipe.Expire(ctx, key, 4*time.Hour)

	// 执行pipeline
	_, err := pipe.Exec(ctx)
	if err != nil {
		log.Error("[service]Error caching user accounts",
			zap.Uint("userID", userID),
			zap.Error(err),
		)
		return err
	}

	log.Info("[service]User accounts cached successfully",
		zap.Uint("userID", userID),
		zap.Int("accountsCount", len(accountIDs)),
	)

	return nil
}
