package userservice

import (
	"GalaxyEmpireWeb/logger"
	"GalaxyEmpireWeb/models"
	"GalaxyEmpireWeb/utils"
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	r "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type UserService struct {
	DB  *gorm.DB
	RDB *r.Client
}

var userServiceInstance *UserService
var log = logger.GetLogger()
var rolePrefix = "role_"

func NewService(db *gorm.DB, rdb *r.Client) *UserService {
	return &UserService{
		DB:  db,
		RDB: rdb,
	}
}

func InitService(db *gorm.DB, rdb *r.Client) error {
	if userServiceInstance != nil {
		return errors.New("UserService is already initialized")
	}
	userServiceInstance = NewService(db, rdb)
	return nil
}

func GetService(ctx context.Context) (*UserService, error) { // TODO:
	traceID := utils.TraceIDFromContext(ctx)
	log.Info("[service]GetService", zap.String("traceID", traceID))

	if userServiceInstance == nil {
		log.DPanic("[service]UserService is not initialized", zap.String("traceID", traceID))
		return nil, errors.New("UserService is not initialized")
	}
	return userServiceInstance, nil
}

func (service *UserService) Create(ctx context.Context, user *models.User) error {
	traceID := utils.TraceIDFromContext(ctx)
	log.Info("[service]Create",
		zap.String("traceID", traceID),
		zap.String("username", user.Username),
	)
	err := service.DB.Create(user).Error
	if err != nil {
		log.Error("[service]Create user failed",
			zap.String("uuid", traceID),
			zap.Error(err),
		)
	}
	return err
}
func (service *UserService) Update(ctx context.Context, user *models.User) error {
	traceID := utils.TraceIDFromContext(ctx)
	log.Info("[service]Update user",
		zap.String("uuid", traceID),
		zap.String("username", user.Username),
	)
	err := service.DB.Save(user).Error
	if err != nil {
		log.Error("[service]Update user failed",
			zap.String("uuid", traceID),
			zap.Error(err),
		)
	}
	return err
}

func (service *UserService) Delete(ctx context.Context, id uint) error {
	traceID := utils.TraceIDFromContext(ctx)
	log.Info("[service]Delete user", zap.String("uuid", traceID), zap.Uint("id", id))
	result := service.DB.Delete(&models.User{}, id)
	if result.Error != nil {
		log.Error("[service]Delete user failed", zap.String("uuid", traceID), zap.Error(result.Error))
		return result.Error
	}
	if result.RowsAffected == 0 {
		log.Warn("[service]Delete user failed - user not found")
		return fmt.Errorf("no user found with id %d", id)
	}
	return nil
}

func (service *UserService) GetAllUsers(ctx context.Context) ([]*models.User, error) {
	traceID := utils.TraceIDFromContext(ctx)
	log.Info("GetAllUsers",
		zap.String("uuid", traceID),
	)
	var users []*models.User
	err := service.DB.Find(&users).Error
	if err != nil {
		log.Error("[service]Get all users failed",
			zap.String("uuid", traceID),
			zap.Error(err),
		)
	}
	return users, err
}

func (service *UserService) GetById(ctx context.Context, id uint, fields []string) (*models.User, error) {
	traceID := utils.TraceIDFromContext(ctx)
	log.Info("[service]GetById",
		zap.String("uuid", traceID),
		zap.Uint("id", id),
	)
	var user models.User
	cur := service.DB
	for _, field := range fields {
		cur.Preload(field)
	}
	err := cur.Where("id = ?", id).First(&user).Error
	if err != nil {
		log.Error("[service]Get user by id failed",
			zap.String("uuid", traceID),
			zap.Error(err),
		)
	}
	return &user, err
}
func (service *UserService) UpdateBalance(ctx context.Context, user *models.User) error {
	traceID := utils.TraceIDFromContext(ctx)
	log.Info("[service]UpdateBalance",
		zap.String("uuid", traceID),
		zap.String("username", user.Username),
		zap.Int("balance", user.Balance),
	)
	result := service.DB.
		Model(&models.User{}).
		Where("id = ?", user.ID).
		Update("balance", user.Balance)

	if result.RowsAffected == 0 {
		log.Warn("[service]Update balance failed - record not found",
			zap.String("uuid", traceID),
		)
		return gorm.ErrRecordNotFound
	}

	if result.Error != nil {
		log.Error("[service]Update balance failed",
			zap.String("uuid", traceID),
			zap.Error(result.Error),
		)
	}

	return result.Error
}

func (service *UserService) GetUserRole(ctx context.Context, userID uint) int {
	traceID := utils.TraceIDFromContext(ctx)
	log.Info("[service]GetUserRole",
		zap.String("traceID", traceID),
		zap.Uint("userID", userID),
	)
	key := fmt.Sprintf("%s%d", rolePrefix, userID)
	roleStr, err := service.RDB.Get(ctx, key).Result()

	// 如果在Redis中找到了数据，将其转换为int并返回
	if err == nil {
		role, err := strconv.Atoi(roleStr)
		if err == nil {
			log.Info("[service]GetUserRole from redis",
				zap.String("traceID", traceID),
				zap.Uint("userID", userID),
				zap.Int("role", role),
			)
			return role
		}
		log.Warn("[service]GetUserRole parse to uint failed",
			zap.String("traceID", traceID),
			zap.Uint("userID", userID),
			zap.Error(err),
		)
	}

	// 如果Redis中没有数据，从数据库查询
	user, err := service.GetById(ctx, userID, []string{})
	if err != nil {
		log.Error("[service]GetUserRole from db failed",
			zap.String("traceID", traceID),
			zap.Uint("userID", userID),
			zap.Error(err),
		)
		return -1
	}
	role := user.Role

	// 将结果存储回Redis
	service.RDB.Set(ctx, key, role, 4*time.Hour)
	log.Info("[service]GetUserRole from db",
		zap.String("traceID", traceID),
		zap.Uint("userID", userID),
		zap.Int("role", role),
	)

	return role
}
