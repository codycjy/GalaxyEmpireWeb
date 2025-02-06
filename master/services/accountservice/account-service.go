package accountservice

import (
	"GalaxyEmpireWeb/consts"
	"GalaxyEmpireWeb/logger"
	"GalaxyEmpireWeb/models"
	"GalaxyEmpireWeb/services/casbinservice"
	"GalaxyEmpireWeb/utils"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type accountService struct {
	DB       *gorm.DB
	Enforcer casbinservice.Enforcer
}

var (
	accountServiceInstance *accountService
	log                    = logger.GetLogger()
	accountListPrefix      = consts.UserAccountPrefix
	expireTime             = consts.ProdExpire
)

const (
	READ  = 1
	WRITE = 2
)

func NewService(db *gorm.DB, enforcer casbinservice.Enforcer) *accountService {
	return &accountService{
		DB:       db,
		Enforcer: enforcer,
	}
}

func InitService(db *gorm.DB, enforcer casbinservice.Enforcer) error {
	if accountServiceInstance != nil {
		return errors.New("AccountService is already initialized")
	}
	if os.Getenv("ENV") == "test" {
		expireTime = consts.TestExipre
	}
	accountServiceInstance = NewService(db, enforcer)
	log.Info("[service] Account service Initialized")
	return nil
}

func GetService() *accountService {
	if accountServiceInstance == nil {
		log.Fatal("[service] Account service is not initialized")
	}
	return accountServiceInstance
}

func (service *accountService) GetById(ctx context.Context, id uint) (*models.Account, *utils.ServiceError) {
	traceID := utils.TraceIDFromContext(ctx)
	log.Info("[service]Get Account By ID",
		zap.Uint("id", id),
		zap.String("traceID", traceID),
	)

	allowed, serviceErr := service.isUserAllowed(ctx, id, casbinservice.READ)
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
	err := cur.Where("id = ?", id).Preload("Tasks").
		Preload("Tasks.Targets").
		Preload("Tasks.Fleet").
		First(&account).Error
	if err != nil {
		log.Error("[service]Get Account By ID failed",
			zap.String("traceID", traceID),
			zap.Error(err),
		)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.NewServiceError(http.StatusNotFound, "Account Not found", err)
		}
		return nil, utils.NewServiceError(http.StatusInternalServerError, "Service Error", err)
	}
	return &account, nil
}

func (service *accountService) GetByUserId(ctx context.Context,
	userId uint,
) (*[]models.Account, *utils.ServiceError) {
	traceID := utils.TraceIDFromContext(ctx)
	log.Info("[service]Get Account By User ID",
		zap.Uint("userId", userId),
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

	// Start transaction
	tx := service.DB.Begin()
	if tx.Error != nil {
		log.Error("[service]Failed to start database transaction",
			zap.String("traceID", traceID),
			zap.Error(tx.Error),
		)
		return utils.NewServiceError(http.StatusInternalServerError, "Failed to start database transaction", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Error("[service]Panic recovered in Create Account",
				zap.String("traceID", traceID),
				zap.Any("panic", r),
			)
		}
	}()

	if account.UserID != 0 && account.UserID != userID {
		log.Warn("[service]Create Account - User ID not match",
			zap.String("traceID", traceID),
			zap.Uint("userID", userID),
			zap.Uint("accountUserID", account.UserID),
		)
		return utils.NewServiceError(http.StatusUnauthorized, "User ID not match", nil)
	}
	account.UserID = userID // If not set, set to current user

	account.ExpireAt = time.Now()
	if err := tx.Create(account).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			log.Info("[service]Create Account failed - Account already exists",
				zap.String("traceID", traceID),
			)
			return utils.NewServiceError(http.StatusConflict, "Account already exists", err)
		}
		log.Error("[service]Create Account failed",
			zap.String("traceID", traceID),
			zap.Uint("userId", userID),
			zap.Error(err),
		)
		return utils.NewServiceError(http.StatusInternalServerError, "failed create account", err)
	}

	// Add policies in batch using the same transaction
	policies := [][]string{
		{strconv.Itoa(int(userID)), account.GetEntityPrefix() + fmt.Sprint(account.ID), "write"},
		{strconv.Itoa(int(userID)), account.GetEntityPrefix() + fmt.Sprint(account.ID), "read"},
	}

	if _, err := service.Enforcer.AddPolicies(ctx, tx, policies); err != nil {
		tx.Rollback()
		log.Error("[service]Create Account failed - Add Policies",
			zap.String("traceID", traceID),
			zap.Error(err),
			zap.Any("policies", policies),
		)
		return utils.NewServiceError(http.StatusInternalServerError, "Failed to add policies", err)
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		log.Error("[service]Failed to commit transaction",
			zap.String("traceID", traceID),
			zap.Error(err),
		)
		return utils.NewServiceError(http.StatusInternalServerError, "Failed to commit transaction", err)
	}

	// Reload policies after successful creation
	go service.Enforcer.ReloadPolicy()

	return nil
}

func (service *accountService) Update(ctx context.Context, account *models.Account) *utils.ServiceError {
	traceID := utils.TraceIDFromContext(ctx)
	log.Info("[service]Update Account Info",
		zap.Uint("accountID", account.ID),
		zap.String("username", account.Username),
		zap.String("traceID", traceID),
	)

	allowed, serviceErr := service.isUserAllowed(ctx, account.ID, casbinservice.WRITE)
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

	// Start transaction
	tx := service.DB.Begin()
	if tx.Error != nil {
		log.Error("[service]Failed to start database transaction",
			zap.String("traceID", traceID),
			zap.Error(tx.Error),
		)
		return utils.NewServiceError(http.StatusInternalServerError, "Failed to start database transaction", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Error("[service]Panic recovered in Update Account",
				zap.String("traceID", traceID),
				zap.Any("panic", r),
			)
		}
	}()

	// Perform update within transaction
	if err := tx.Save(account).Error; err != nil {
		tx.Rollback()
		log.Error("[service]Update Account failed",
			zap.String("traceID", traceID),
			zap.Error(err),
		)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.NewServiceError(http.StatusNotFound, "Account Not found", err)
		}
		return utils.NewServiceError(http.StatusInternalServerError, "Failed to Update Account", err)
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		log.Error("[service]Failed to commit transaction",
			zap.String("traceID", traceID),
			zap.Error(err),
		)
		return utils.NewServiceError(http.StatusInternalServerError, "Failed to commit transaction", err)
	}

	return nil
}

func (service *accountService) Delete(ctx context.Context, ID uint) *utils.ServiceError {
	traceID := utils.TraceIDFromContext(ctx)
	log.Info("[service]Delete Account Info",
		zap.Uint("accountID", ID),
		zap.String("traceID", traceID),
	)

	allowed, serviceErr := service.isUserAllowed(ctx, ID, casbinservice.WRITE)
	if serviceErr != nil {
		log.Error("[service]Delete Account Info - Check Permission failed",
			zap.String("traceID", traceID),
			zap.Error(serviceErr),
		)
		return serviceErr
	}
	if !allowed {
		log.Info("[service]Delete Account Info - Not allowed",
			zap.String("traceID", traceID),
			zap.Uint("accountID", ID),
			zap.Uint("userID", utils.UserIDFromContext(ctx)),
		)
		return utils.NewServiceError(http.StatusUnauthorized, "User has no Permission", nil)
	}

	result := service.DB.Unscoped().Delete(&models.Account{}, ID)
	if result.Error != nil {
		log.Info("[service]Delete Account failed",
			zap.String("traceID", traceID),
			zap.Error(result.Error),
			zap.Uint("accountID", ID),
		)
		return utils.NewServiceError(http.StatusInternalServerError, "Failed to delete user", result.Error)
	}

	if result.RowsAffected == 0 {
		log.Warn("[server]Delete Account failed - no such user",
			zap.String("traceID", traceID),
		)
		return utils.NewServiceError(http.StatusNotFound, "Account not found", nil)
	}

	return nil
}

func (service *accountService) IsAccountAllowed(ctx context.Context, accountID uint, rw int) (bool, *utils.ServiceError) {
	return service.isUserAllowed(ctx, accountID, rw)
}

// ________________________________
// |  Private Functions         |
func (service *accountService) isUserAllowed(ctx context.Context, accountID uint, rw int) (bool, *utils.ServiceError) {
	userID, ok := ctx.Value("userID").(uint)
	if !ok {
		log.Warn("[service]Check User Permission - No userID in context",
			zap.String("traceID", utils.TraceIDFromContext(ctx)),
		)
		return false, utils.NewServiceError(http.StatusInternalServerError, "No userID in context", nil)
	}
	log.Info("[AccountService]Check User Permission",
		zap.String("traceID", utils.TraceIDFromContext(ctx)),
		zap.Uint("accountID", accountID),
		zap.String("userID", fmt.Sprint(userID)),
		zap.Int("rw", rw),
	)

	obj := fmt.Sprintf("%s%d", models.Account{}.GetEntityPrefix(), accountID)
	opt := "read"
	if rw&WRITE == WRITE {
		opt = "write"
	}

	allowed, err := service.Enforcer.Enforce(ctx, fmt.Sprint(userID), obj, opt)
	if err != nil {
		return false, utils.NewServiceError(http.StatusInternalServerError, "Failed to check permission", err)
	}

	return allowed, nil
}
