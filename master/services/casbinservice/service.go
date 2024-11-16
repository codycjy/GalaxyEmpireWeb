package casbinservice

import (
	"context"
	"errors"
	"fmt"

	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"gorm.io/gorm"
)

// CasbinService 结构体
type Enforcer interface {
    Enforce(ctx context.Context, sub, obj, act string) (bool, error)
    AddPolicy(ctx context.Context, sub, obj, act string) (bool, error)
    AddUserToGroup(ctx context.Context, user, group string) (bool, error)
}

type CasbinService struct {
    enforcer *casbin.Enforcer
}

var casbinEnforcer Enforcer

// casbinService 单例
func NewCasbinEnforcer(db *gorm.DB, modelPath string) (Enforcer, error) {
	// 初始化 Casbin 适配器
	adapter, err := gormadapter.NewAdapterByDB(db)
	if err != nil {
		return nil, fmt.Errorf("failed to create adapter: %w", err)
	}

	// 加载模型和策略
	enforcer, err := casbin.NewEnforcer(modelPath, adapter)
	if err != nil {
		return nil, fmt.Errorf("failed to create enforcer: %w", err)
	}

	// 从数据库中加载策略
	if err = enforcer.LoadPolicy(); err != nil {
		return nil, fmt.Errorf("failed to load policy: %w", err)
	}

	return &CasbinService{
		enforcer: enforcer,
	}, nil
}

func InitCasbinService(db *gorm.DB, modelPath string) {
	var err error
	casbinEnforcer, err = NewCasbinEnforcer(db, modelPath)
	if err != nil {
		panic(err)
	}
}

// GetCasbinService 获取 CasbinService 单例
func GetCasbinService() Enforcer {
	if casbinEnforcer == nil {
		panic(errors.New("casbinService not init"))
	}
	return casbinEnforcer
}


// AddPolicy 添加策略
func (s *CasbinService) AddPolicy(ctx context.Context, sub, obj, act string) (bool, error) {
	ok, err := s.enforcer.AddPolicy(sub, obj, act)
	if err != nil {
		return false, fmt.Errorf("failed to add policy: %w", err)
	}
	return ok, nil
}

// RemovePolicy 删除策略
func (s *CasbinService) RemovePolicy(ctx context.Context, sub, obj, act string) (bool, error) {
	ok, err := s.enforcer.RemovePolicy(sub, obj, act)
	if err != nil {
		return false, fmt.Errorf("failed to remove policy: %w", err)
	}
	return ok, nil
}

// AddUserToGroup 将用户添加到组
func (s *CasbinService) AddUserToGroup(ctx context.Context, user, group string) (bool, error) {
	ok, err := s.enforcer.AddGroupingPolicy(user, group)
	if err != nil {
		return false, fmt.Errorf("failed to add user to group: %w", err)
	}
	return ok, nil
}

// RemoveUserFromGroup 将用户从组中删除
func (s *CasbinService) RemoveUserFromGroup(ctx context.Context, user, group string) (bool, error) {
	ok, err := s.enforcer.RemoveGroupingPolicy(user, group)
	if err != nil {
		return false, fmt.Errorf("failed to remove user from group: %w", err)
	}
	return ok, nil
}

func (s *CasbinService) Enforce(ctx context.Context, sub, obj, act string) (bool, error) {
	ok, err := s.enforcer.Enforce(sub, obj, act)
	if err != nil {
		return false, fmt.Errorf("enforce error: %w", err) //TODO: Create Error with utils
	}
	return ok, nil
}
