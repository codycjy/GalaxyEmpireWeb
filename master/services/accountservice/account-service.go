package accountservice

import (
	"GalaxyEmpireWeb/consts"
	"GalaxyEmpireWeb/logger"
	"GalaxyEmpireWeb/models"
	"GalaxyEmpireWeb/services/casbinservice"
	"GalaxyEmpireWeb/services/taskservice"
	"GalaxyEmpireWeb/utils"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"

	r "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type accountService struct {
	DB       *gorm.DB
	RDB      *r.Client
	Enforcer casbinservice.Enforcer
}

var accountServiceInstance *accountService
var log = logger.GetLogger()
var accountListPrefix = consts.UserAccountPrefix
var expireTime = consts.ProdExpire

func NewService(db *gorm.DB, rdb *r.Client, enforcer casbinservice.Enforcer) *accountService {
	return &accountService{
		DB:       db,
		RDB:      rdb,
		Enforcer: enforcer,
	}
}
func InitService(db *gorm.DB, rdb *r.Client, enforcer casbinservice.Enforcer) error {
	if accountServiceInstance != nil {
		return errors.New("AccountService is already initialized")
	}
	if os.Getenv("ENV") == "test" {
		expireTime = consts.TestExipre
	}
	accountServiceInstance = NewService(db, rdb, enforcer)
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
	_, err = service.Enforcer.AddPolicy(ctx, strconv.Itoa(int(userID)), account.GetEntityPrefix()+fmt.Sprint(account.ID), "write")
	if err != nil {
		log.Error("[service]Create Account failed - Add Policy",
			zap.String("traceID", traceID),
			zap.Error(err),
		)
		return utils.NewServiceError(http.StatusInternalServerError, "Failed to add policy", err)
	}
	_, err = service.Enforcer.AddPolicy(ctx, strconv.Itoa(int(userID)), account.GetEntityPrefix()+fmt.Sprint(account.ID), "read")
	if err != nil {
		log.Error("[service]Create Account failed - Add Policy",
			zap.String("traceID", traceID),
			zap.Error(err),
		)
		return utils.NewServiceError(http.StatusInternalServerError, "Failed to add policy", err)
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

	allowed, serviceErr := service.isUserAllowed(ctx, ID, casbinservice.WRITE)
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
	if result.Error != nil {
		log.Info("[service]Delete Account failed",
			zap.String("traceID", traceID),
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

func (service *accountService) RequestCheckingAccountLogin(ctx context.Context, account *models.Account) (string, *utils.ServiceError) {
	traceID := utils.TraceIDFromContext(ctx)
	log.Info("[service]Check Account Available",
		zap.String("traceID", traceID),
		zap.String("username", account.Username),
	)
	taskservice := taskservice.GetService()
	uuid, err := taskservice.CheckAccountLogin(ctx, account)
	if err != nil {
		log.Error("[service]Check Account Available failed",
			zap.String("traceID", traceID),
			zap.Error(err),
		)
		return "", err
	}
	return uuid, nil
}

func (serveice *accountService) GetLoginInfo(ctx context.Context, uuid string) bool {
	traceID := utils.TraceIDFromContext(ctx)
	log.Info("[service]Get Login Info",
		zap.String("traceID", traceID),
		zap.String("uuid", uuid),
	)
	taskservice := taskservice.GetService()
	return taskservice.GetLoginInfo(ctx, uuid)
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
	if rw&2 == 2 {
		opt = "write"
	}

	allowed, err := service.Enforcer.Enforce(ctx, fmt.Sprint(userID), obj, opt)
	if err != nil {
		return false, utils.NewServiceError(http.StatusInternalServerError, "Failed to check permission", err)
	}
	log.Info("[AccountService]Check User Permission",
		zap.String("traceID", utils.TraceIDFromContext(ctx)),
		zap.String("userID", fmt.Sprint(userID)),
		zap.Uint("accountID", accountID),
		zap.Bool("allowed", allowed),
	)

	return allowed, nil
}
