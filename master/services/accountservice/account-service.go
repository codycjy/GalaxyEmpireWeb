package accountservice

import (
	"GalaxyEmpireWeb/consts"
	"GalaxyEmpireWeb/logger"
	"GalaxyEmpireWeb/models"
	"GalaxyEmpireWeb/utils"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"

	r "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type accountService struct {
	DB  *gorm.DB
	RDB *r.Client
}

var accountServiceInstance *accountService
var log = logger.GetLogger()
var accountListPrefix = consts.UserAccountPrefix
var expireTime = consts.ProdExpire

func NewService(db *gorm.DB, rdb *r.Client) *accountService {
	return &accountService{
		DB:  db,
		RDB: rdb,
	}
}
func InitService(db *gorm.DB, rdb *r.Client) error {
	if accountServiceInstance != nil {
		return errors.New("AccountService is already initialized")
	}
	if os.Getenv("ENV") == "test" {
		expireTime = consts.TestExipre
	}
	accountServiceInstance = NewService(db, rdb)
	log.Info("[service] Account service Initialized")
	return nil
}
func GetService(ctx context.Context) (*accountService, error) {
	traceID := utils.TraceIDFromContext(ctx)
	if accountServiceInstance == nil {
		log.DPanic("[service]AccountService is not initialized",
			zap.String("traceID", traceID),
		)
		return nil, errors.New("AccountService is not initialized")
	}
	return accountServiceInstance, nil
}

func (service *accountService) GetById(ctx context.Context, id uint, fields []string) (*models.Account, *utils.ServiceError) {
	traceID := utils.TraceIDFromContext(ctx)
	log.Info("[service]Get Account By ID",
		zap.Uint("id", id),
		zap.Strings("fields", fields),
		zap.String("traceID", traceID),
	)

	allowed, serviceErr := service.isUserAllowed(ctx, id)
	if serviceErr != nil {
		return nil, serviceErr
	}
	if !allowed {
		log.Info("[service]Get Account By ID - Not allowed",
			zap.String("traceID", traceID),
		)
		return nil, utils.NewServiceError(
			http.StatusForbidden,
			"Account Not allowed",
			nil,
		)
	}
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
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.NewServiceError(http.StatusNotFound, "Account Not found", err)
		}
		return nil, utils.NewServiceError(http.StatusInternalServerError, "SQL Server Error", err)
	}
	return &account, nil
}

func (service *accountService) GetByUserId(ctx context.Context,
	userId uint, fields []string) (*[]models.Account, *utils.ServiceError) {
	traceID := utils.TraceIDFromContext(ctx)
	log.Info("[service]Get Account By User ID",
		zap.Uint("userId", userId),
		zap.Strings("fields", fields),
		zap.String("traceID", traceID),
	)
	var accounts []models.Account
	result := service.DB.Model(&models.Account{}).Where("user_id = ?", userId).Find(&accounts)
	err := result.Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error("[service]Get Account By User ID failed - Not found",
				zap.String("traceID", traceID),
				zap.Error(err),
			)
			return nil, utils.NewServiceError(http.StatusNotFound, "Account Not found", err)

		}
		log.Error("[service]Get Account By User ID failed",
			zap.String("traceID", traceID),
			zap.Error(err),
		)
		return nil, utils.NewServiceError(http.StatusInternalServerError, "SQL Service Error", err)
	}
	log.Info("[service]Successfully get accounts",
		zap.String("traceID", traceID),
		zap.Int("accounts count", len(accounts)))
	return &accounts, nil
}

func (service *accountService) Create(ctx context.Context, account *models.Account) *utils.ServiceError {
	traceID := utils.TraceIDFromContext(ctx)
	userID := ctx.Value("userID").(uint)
	log.Info("[service]Create Account ",
		zap.Uint("userId", userID),
		zap.String("username", account.Username),
		zap.String("traceID", traceID),
	)
	account.UserID = userID
	err := service.DB.Create(account).Error
	if err != nil {
		log.Error("[service]Create Account failed",
			zap.String("traceID", traceID),
			zap.Uint("userId", userID),
			zap.Error(err),
		)
		return utils.NewServiceError(http.StatusInternalServerError, "failed create account", err)
	}
	return nil
}

func (service *accountService) Update(ctx context.Context, account *models.Account) *utils.ServiceError {
	traceID := utils.TraceIDFromContext(ctx)
	log.Info("[service]Update Account Info",
		zap.Uint("accountID", account.ID),
		zap.String("username", account.Username),
		zap.String("traceID", traceID),
	)

	allowed, serviceErr := service.isUserAllowed(ctx, account.ID)
	if serviceErr != nil {
		return serviceErr
	}
	if !allowed {
		log.Info("[service]Update Account Info - Not allowed",
			zap.String("traceID", traceID),
		)
		return utils.NewServiceError(
			http.StatusUnauthorized,
			"Account Not allowed",
			nil,
		)
	}

	err := service.DB.Save(account).Error
	if err != nil {
		log.Error("[service]Update Account failed",
			zap.String("traceID", traceID),
			zap.Error(err),
		)
		if err == gorm.ErrRecordNotFound {
			return utils.NewServiceError(http.StatusNotFound, "Account Not found", err)
		}
		return utils.NewServiceError(http.StatusInternalServerError, "Failed to Update Account", err)
	}
	return nil
}

func (service *accountService) Delete(ctx context.Context, ID uint) *utils.ServiceError {
	traceID := utils.TraceIDFromContext(ctx)
	log.Info("[service]Delete Account Info",
		zap.Uint("userId", ID),
		zap.String("traceID", traceID),
	)
	allowed, serviceErr := service.isUserAllowed(ctx, ID)
	if serviceErr != nil {
		return serviceErr
	}
	if !allowed {
		log.Info("[service]Delete Account Info - Not allowed",
			zap.String("traceID", traceID),
		)
		return utils.NewServiceError(http.StatusUnauthorized, "User has no Permission", nil)
	}
	result := service.DB.Delete(&models.Account{}, ID)
	err := result.Error
	if result.RowsAffected == 0 {
		return utils.NewServiceError(http.StatusNotFound, "Account Not found", err)
	}
	if err != nil {
		log.Info("[service]Delete Account failed",
			zap.String("traceID", traceID),
		)
		return utils.NewServiceError(http.StatusInternalServerError, "Failed to delete user", err)

	}
	if result.RowsAffected == 0 {
		log.Warn("[server]Delete Account failed - no such user",
			zap.String("traceID", traceID),
		)
		return utils.NewServiceError(http.StatusNotFound, "Account not found", nil)
	}
	return nil

}

func (service *accountService) isUserAllowed(ctx context.Context, accountID uint) (bool, *utils.ServiceError) {
	traceID := utils.TraceIDFromContext(ctx)
	userID1 := ctx.Value("userID")
	if userID1 == nil {
		log.Warn("[service]Check User Permission - No userID in context",
			zap.String("traceID", traceID),
		)
		return false, utils.NewServiceError(http.StatusInternalServerError, "No userID in context", nil)
	}
	userID := userID1.(uint)
	log.Info("[service]Check User Permission",
		zap.Uint("userID", userID),
		zap.Uint("accountID", accountID),
		zap.String("traceID", traceID),
	)
	if userID == 0 {
		log.Warn("[service]Check User Permission - No userID in context",
			zap.String("traceID", traceID),
		)
		return false, utils.NewServiceError(http.StatusInternalServerError, "No userID in context", nil)
	}
	key := fmt.Sprintf("%s%d", accountListPrefix, userID)

	// 首先检查键是否存在
	exists, err := service.RDB.Exists(ctx, key).Result()
	if err != nil {
		log.Warn("[service]Check User Permission failed - redis check exists",
			zap.String("traceID", traceID),
			zap.String("redis_key", key),
			zap.Error(err),
		)
	}
	if exists == 0 {
		log.Info("[service]Check User Permission - Not in Redis. Retrieving",
			zap.String("traceID", traceID),
		)
		accounts, err := service.GetByUserId(ctx, userID, []string{})
		if err != nil {
			return false, utils.NewServiceError(http.StatusInternalServerError, "Service Error", err)
		}
		var accountIDs = make([]uint, len(*accounts))
		// NOTE: Could be optimized by early return
		for i, account := range *accounts {
			accountIDs[i] = account.ID
		}
		serviceErr := service.cacheUserAccounts(ctx, userID, accountIDs)
		if serviceErr != nil {
			return false, serviceErr
		}

	}

	// 如果键存在，检查集合中是否包含特定元素
	isMember, err := service.RDB.SIsMember(ctx, key, accountID).Result()
	// 首先检查错误
	if err != nil {
		log.Error("[service]Check User Permission - failed",
			zap.String("traceID", traceID),
			zap.String("redis_key", key),
			zap.Error(err),
		)
		return false, utils.NewServiceError(http.StatusInternalServerError, "redis retrieve error", err)
	}

	log.Info("[service]Check User Permission - Success",
		zap.String("traceID", traceID),
		zap.Bool("isMember", isMember),
	)
	return isMember, nil

}

func (service *accountService) cacheUserAccounts(ctx context.Context, userID uint, accountIDs []uint) *utils.ServiceError {
	key := fmt.Sprintf("%s%d", accountListPrefix, userID)

	// 使用pipeline批量添加元素以提高性能
	pipe := service.RDB.Pipeline()
	for _, accountID := range accountIDs {
		pipe.SAdd(ctx, key, accountID)
	}

	pipe.Expire(ctx, key, expireTime)

	// 执行pipeline
	_, err := pipe.Exec(ctx)
	if err != nil {
		log.Error("[service]Error caching user accounts",
			zap.Uint("userID", userID),
			zap.Error(err),
		)
		return utils.NewServiceError(http.StatusInternalServerError, "redis cache error", err)
	}

	log.Info("[service]User accounts cached successfully",
		zap.Uint("userID", userID),
		zap.Int("accountsCount", len(accountIDs)),
	)

	return nil
}
