package userservice

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

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type userService struct { // change to private for factory
	DB       *gorm.DB
	Enforcer casbinservice.Enforcer
}

var userServiceInstance *userService
var log = logger.GetLogger()
var rolePrefix = consts.UserRolePrefix
var expireTime = consts.ProdExpire

const READ = 1 // TODO: change it later
const WRITE = 2

func NewService(db *gorm.DB, enforcer casbinservice.Enforcer) *userService {
	return &userService{
		DB:       db,
		Enforcer: enforcer,
	}
}

func InitService(db *gorm.DB, enforcer casbinservice.Enforcer) error {
	if db == nil || enforcer == nil {
		log.Fatal("db, rdb, enforcer is nil")
	}
	if userServiceInstance != nil {
		return errors.New("UserService is already initialized")
	}
	if os.Getenv("ENV") == "test" {
		expireTime = consts.TestExipre
	}
	userServiceInstance = NewService(db, enforcer)
	return nil
}

func GetService(ctx context.Context) (*userService, error) { // TODO:
	traceID := utils.TraceIDFromContext(ctx)
	log.Info("[service]GetService", zap.String("traceID", traceID))

	if userServiceInstance == nil {
		log.DPanic("[service]UserService is not initialized", zap.String("traceID", traceID))
		return nil, errors.New("UserService is not initialized")
	}
	return userServiceInstance, nil
}

func (service *userService) Create(ctx context.Context, user *models.User) *utils.ServiceError {
	traceID := utils.TraceIDFromContext(ctx)
	log.Info("[service]Create",
		zap.String("traceID", traceID),
		zap.String("username", user.Username),
	)

	// Start transaction
	tx := service.DB.Begin()
	if tx.Error != nil {
		return utils.NewServiceError(http.StatusInternalServerError, "failed to start transaction", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create user
	if err := tx.Create(user).Error; err != nil {
		tx.Rollback()
		log.Error("[service]Create user failed",
			zap.String("traceID", traceID),
			zap.Error(err),
		)
		return utils.NewServiceError(http.StatusInternalServerError, "failed create user", err)
	}

	userID := strconv.Itoa(int(user.ID))
	obj := fmt.Sprintf("%s%d", user.GetEntityPrefix(), user.ID)

	// Add policies in batch
	policies := [][]string{
		{userID, obj, "read"},
		{userID, obj, "write"},
		{userID, "user", "read"}, // Add basic user role permissions
		{userID, fmt.Sprintf("%s*", user.GetEntityPrefix()), "read"}, // Allow user to read their own resources
	}

	// Log the exact policies being added
	log.Info("[service]Adding policies",
		zap.String("traceID", traceID),
		zap.Any("policies", policies),
	)

	if _, err := service.Enforcer.AddPolicies(ctx, tx, policies); err != nil {
		tx.Rollback()
		log.Error("[service]Add policies failed",
			zap.String("traceID", traceID),
			zap.Error(err),
			zap.Any("policies", policies),
		)
		return utils.NewServiceError(http.StatusInternalServerError, "failed to add policies", err)
	}

	// Add user to group
	if _, err := service.Enforcer.AddUserToGroup(ctx, tx, userID, "user"); err != nil {
		tx.Rollback()
		log.Error("[service]Add user to group failed",
			zap.String("traceID", traceID),
			zap.Error(err),
		)
		return utils.NewServiceError(http.StatusInternalServerError, "failed to add user to group", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		log.Error("[service]Failed to commit transaction",
			zap.String("traceID", traceID),
			zap.Error(err),
		)
		return utils.NewServiceError(http.StatusInternalServerError, "failed to commit transaction", err)
	}

	log.Info("[service]Successfully created user with policies",
		zap.String("traceID", traceID),
		zap.String("username", user.Username),
		zap.Uint("userID", user.ID),
		zap.Any("policies", policies),
	)
	go service.Enforcer.ReloadPolicy()

	return nil
}

func (service *userService) Update(ctx context.Context, user *models.User) *utils.ServiceError {
	traceID := utils.TraceIDFromContext(ctx)
	log.Info("[service]Update user",
		zap.String("traceID", traceID),
		zap.String("username", user.Username),
	)
	obj := fmt.Sprintf("%s%d", user.GetEntityPrefix(), user.ID)
	allowed, _ := service.IsUserAllowed(ctx, obj, READ|WRITE)
	if !allowed {
		log.Info("[service]Get Update By ID - Not allowed",
			zap.String("traceID", traceID),
		)
		return utils.NewServiceError(http.StatusUnauthorized, "User Not allowed", nil)

	}
	err := service.DB.Save(user).Error
	if err != nil {
		log.Error("[service]Update user failed",
			zap.String("traceID", traceID),
			zap.Error(err),
		)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.NewServiceError(http.StatusNotFound, "User Not Found", err)
		}
		return utils.NewServiceError(http.StatusInternalServerError, "Failed to Update User", err)

	}
	return nil
}

func (service *userService) Delete(ctx context.Context, id uint) *utils.ServiceError {
	traceID := utils.TraceIDFromContext(ctx)
	log.Info("[service]Delete user", zap.String("traceID", traceID), zap.Uint("id", id))
	result := service.DB.Delete(&models.User{}, id)
	err := result.Error
	if err != nil {
		log.Error("[service]Delete user failed", zap.String("traceID", traceID), zap.Error(result.Error))
		return utils.NewServiceError(http.StatusInternalServerError, "failed to delete user", err)
	}
	if result.RowsAffected == 0 {
		log.Warn("[service]Delete user failed - user not found")
		return utils.NewServiceError(http.StatusNotFound, "User Not Found", nil)
	}
	return nil
}

func (service *userService) GetAllUsers(ctx context.Context) ([]*models.User, *utils.ServiceError) {
	traceID := utils.TraceIDFromContext(ctx)
	log.Info("GetAllUsers",
		zap.String("traceID", traceID),
	)
	var users []*models.User
	allowed, _ := service.IsUserAllowed(ctx, "all", READ)
	if !allowed {
		log.Info("[service]Get All Users - Not allowed",
			zap.String("traceID", traceID),
		)
		return nil, utils.NewServiceError(http.StatusUnauthorized, "User Not allowed", nil)
	}
	err := service.DB.Find(&users).Error
	if err != nil {
		log.Error("[service]Get all users failed",
			zap.String("traceID", traceID),
			zap.Error(err),
		)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.NewServiceError(http.StatusNotFound, "User Not Found", err)
		}
		return nil, utils.NewServiceError(http.StatusInternalServerError, "Failed To Find All User", err)
	}
	return users, nil
}

func (service *userService) GetById(ctx context.Context, id uint, fields []string) (*models.User, *utils.ServiceError) {
	traceID := utils.TraceIDFromContext(ctx)
	log.Info("[service]Get User by id",
		zap.String("traceID", traceID),
		zap.Uint("userID", id),
	)
	var user models.User
	cur := service.DB
	for _, field := range fields {
		cur.Preload(field)
	}
	obj := fmt.Sprintf("%s%d", models.User{}.GetEntityPrefix(), id)
	allowed, err := service.IsUserAllowed(ctx, obj, READ)
	if err != nil {
		log.Error("[service]Get By ID - failed to validate user permission",
			zap.String("traceID", traceID),
			zap.Uint("userID", id),
			zap.Error(err),
		)
		return nil, utils.NewServiceError(http.StatusInternalServerError, "Failed to Validate User Permission", err)
	}
	if !allowed {
		log.Info("[service]Get By ID - Not allowed",
			zap.String("traceID", traceID),
			zap.Uint("userID", id),
		)
		return nil, utils.NewServiceError(http.StatusUnauthorized, "User Not allowed", nil)
	}

	err = cur.Where("id = ?", id).First(&user).Error
	if err != nil {
		log.Error("[service]Get user by id failed",
			zap.String("traceID", traceID),
			zap.Error(err),
		)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.NewServiceError(http.StatusNotFound, "User Not Found", err)
		}
		return nil, utils.NewServiceError(http.StatusInternalServerError, "Failed To Find User By ID", err)
	}
	log.Info("[serviec]User got",
		zap.String("traceID", traceID),
		zap.Uint("UserID", user.ID),
	)
	return &user, nil
}
func (service *userService) getById(ctx context.Context, id uint, fields []string) (*models.User, *utils.ServiceError) {
	traceID := utils.TraceIDFromContext(ctx)
	log.Info("[service]Get User by id Unsafe",
		zap.String("traceID", traceID),
		zap.Uint("userID", id),
	)
	cur := service.DB
	var user models.User

	err := cur.Where("id = ?", id).First(&user).Error
	if err != nil {
		log.Error("[service]Get user by id failed",
			zap.String("traceID", traceID),
			zap.Error(err),
		)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Warn("[service]Get user by id unsafe- user not found",
				zap.String("traceID", traceID),
				zap.Uint("userID", id),
			)
			return nil, utils.NewServiceError(http.StatusNotFound, "User Not Found", err)
		}
		log.Error("[service]Get user by id unsafe- failed to find user",
			zap.String("traceID", traceID),
			zap.Uint("userID", id),
			zap.Error(err),
		)
		return nil, utils.NewServiceError(http.StatusInternalServerError, "Failed To Find User By ID", err)
	}
	log.Info("[service]Get User by id Unsafe Success",
		zap.String("traceID", traceID),
		zap.Uint("userID", id),
		zap.Int("userRole", user.Role),
	)

	return &user, nil
}

func (service *userService) UpdateBalance(ctx context.Context, user *models.User) *utils.ServiceError {
	traceID := utils.TraceIDFromContext(ctx)
	log.Info("[service]UpdateBalance",
		zap.String("traceID", traceID),
		zap.String("username", user.Username),
		zap.Int("balance", user.Balance),
	)
	result := service.DB.
		Model(&models.User{}).
		Where("id = ?", user.ID).
		Update("balance", user.Balance)

	if result.RowsAffected == 0 {
		log.Warn("[service]Update balance failed - record not found",
			zap.String("traceID", traceID),
		)
		return utils.NewServiceError(http.StatusNotFound, "User Not Found", nil)
	}

	if result.Error != nil {
		log.Error("[service]Update balance failed",
			zap.String("traceID", traceID),
			zap.Error(result.Error),
		)
		return utils.NewServiceError(http.StatusInternalServerError, "Failed to Update Balance", result.Error)
	}
	return nil
}

func (service *userService) GetUserRole(ctx context.Context, userID uint) int {
	traceID := utils.TraceIDFromContext(ctx)
	log.Info("[service]GetUserRole",
		zap.String("traceID", traceID),
		zap.Uint("userID", userID),
	)

	// 如果Redis中没有数据，从数据库查询
	user, err1 := service.getById(ctx, userID, []string{})
	if err1 != nil {
		log.Error("[service]GetUserRole from db failed",
			zap.String("traceID", traceID),
			zap.Uint("userID", userID),
			zap.Error(err1),
		)
		return -1
	}
	role := user.Role

	return role
}

// Prepared for more complicated cases
// Seem Useless currently lol
func (service *userService) IsUserAllowed(ctx context.Context, obj string, rw int) (allowed bool, err error) {
	traceID := utils.TraceIDFromContext(ctx)
	role := ctx.Value("role")
	if role == nil {
		log.Error("[service]Check User Permission - No role in context",
			zap.String("traceID", traceID),
		)
		return false, errors.New("role not found in context")
	}

	userID := ctx.Value("userID")
	if userID == nil {
		log.Error("[service]Check User Permission - No userID in context",
			zap.String("traceID", traceID),
		)
		return false, errors.New("userID not found in context")
	}

	roleInt := role.(int)
	ctxUserID := userID.(uint)

	log.Info(
		"[service]Check user Permission",
		zap.String("traceID", traceID),
		zap.Int("role", roleInt),
		zap.String("obj", obj),
		zap.Uint("requestUserID", ctxUserID),
	)

	userIDStr := strconv.Itoa(int(ctxUserID))
	opt := "read"
	if rw&2 == 2 {
		opt = "write"
	}

	// Add debug logging to show exact values being checked
	log.Info("[service]Checking permission",
		zap.String("traceID", traceID),
		zap.String("userID", userIDStr),
		zap.String("obj", obj),
		zap.String("act", opt),
	)

	allowed, err = service.Enforcer.Enforce(ctx, userIDStr, obj, opt)
	if err != nil {
		log.Error("[service]IsUserAllowed failed to validate user permission",
			zap.String("traceID", traceID),
			zap.Int("role", roleInt),
			zap.Uint("requestUserID", ctxUserID),
			zap.Error(err),
		)
		return false, utils.NewServiceError(http.StatusInternalServerError, "Failed to Validate User Permission", err)
	}

	log.Info("[service]Permission check result",
		zap.String("traceID", traceID),
		zap.String("userID", userIDStr),
		zap.String("obj", obj),
		zap.String("act", opt),
		zap.Bool("allowed", allowed),
	)

	return allowed, nil
}

func (service *userService) LoginUser(ctx context.Context, user *models.User) *utils.ServiceError {
	traceID := utils.TraceIDFromContext(ctx)
	username := user.Username
	password := user.Password
	log.Info("[service]LoginUser",
		zap.String("traceID", traceID),
		zap.String("username", username),
	)
	// 检查用户密码是否匹配
	err1 := service.DB.Where("username = ? AND password = ?", username, password).First(&user).Error
	if err1 != nil {
		log.Warn("[service]LoginUser failed - wrong password",
			zap.String("traceID", traceID),
			zap.String("username", username),
		)
		return utils.NewServiceError(http.StatusUnauthorized, "Wrong Password", err1)

	}
	return nil
}
