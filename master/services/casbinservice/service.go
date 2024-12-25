package casbinservice

import (
	"GalaxyEmpireWeb/logger"
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// CasbinService 结构体
type Enforcer interface {
	Enforce(ctx context.Context, sub, obj, act string) (bool, error)
	AddPolicy(ctx context.Context, tx *gorm.DB, sub, obj, act string) (bool, error)
	AddPolicies(ctx context.Context, tx *gorm.DB, rules [][]string) (bool, error)
	AddUserToGroup(ctx context.Context, tx *gorm.DB, user, group string) (bool, error)
	ReloadPolicy() error
	Stop()
}

type CasbinService struct {
	enforcer     *casbin.Enforcer
	adapter      *gormadapter.Adapter
	reloadTicker *time.Ticker
	stopCh       chan struct{}
}

const READ = 1 // TODO: change it later
const WRITE = 2

var log = logger.GetLogger()
var casbinEnforcer Enforcer
var reloadInterval = defaultReloadInterval

// Add configuration constant
const (
	defaultReloadInterval = 30 * time.Second // Adjust this value based on your needs
)

func init() {
	if intervalStr := os.Getenv("CASBIN_RELOAD_INTERVAL"); intervalStr != "" {
		if interval, err := time.ParseDuration(intervalStr); err == nil {
			reloadInterval = interval
		}
	}
}

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

	service := &CasbinService{
		enforcer:     enforcer,
		adapter:      adapter,
		reloadTicker: time.NewTicker(reloadInterval),
		stopCh:       make(chan struct{}),
	}

	// Start auto-reload goroutine
	go service.autoReload()

	return service, nil
}

func (s *CasbinService) autoReload() {
	for {
		select {
		case <-s.reloadTicker.C:
			if err := s.enforcer.LoadPolicy(); err != nil {
				log.Info("Auto-reload policy failed", zap.Error(err))
			} else {
				log.Info("Auto-reload policy successful")
			}
		case <-s.stopCh:
			s.reloadTicker.Stop()
			return
		}
	}
}

// Add cleanup method to stop the auto-reload goroutine
func (s *CasbinService) Stop() {
	close(s.stopCh)
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
func (s *CasbinService) AddPolicy(ctx context.Context, tx *gorm.DB, sub, obj, act string) (bool, error) {
	txAdapter, err := gormadapter.NewAdapterByDB(tx)
	if err != nil {
		return false, fmt.Errorf("failed to create transaction adapter: %w", err)
	}

	txEnforcer, err := casbin.NewEnforcer(s.enforcer.GetModel(), txAdapter)
	if err != nil {
		return false, fmt.Errorf("failed to create transaction enforcer: %w", err)
	}

	ok, err := txEnforcer.AddPolicy(sub, obj, act)
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

// AddUserToGroup 将用户添加到组中
func (s *CasbinService) AddUserToGroup(ctx context.Context, tx *gorm.DB, user, group string) (bool, error) {
	txAdapter, err := gormadapter.NewAdapterByDB(tx)
	if err != nil {
		return false, fmt.Errorf("failed to create transaction adapter: %w", err)
	}

	txEnforcer, err := casbin.NewEnforcer(s.enforcer.GetModel(), txAdapter)
	if err != nil {
		return false, fmt.Errorf("failed to create transaction enforcer: %w", err)
	}

	ok, err := txEnforcer.AddGroupingPolicy(user, group)
	if err != nil {
		return false, fmt.Errorf("failed to add user to group: %w", err)
	}
	return ok, nil
}

// RemoveUserFromGroup 将用户从组中��除
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
		return false, fmt.Errorf("enforce error: %w", err)
	}
	return ok, nil
}
func (s *CasbinService) AddPolicies(ctx context.Context, tx *gorm.DB, rules [][]string) (bool, error) {
	// Create a new adapter with the transaction
	txAdapter, err := gormadapter.NewAdapterByDB(tx)
	if err != nil {
		return false, fmt.Errorf("failed to create transaction adapter: %w", err)
	}

	// Create a new enforcer with the transaction adapter
	txEnforcer, err := casbin.NewEnforcer(s.enforcer.GetModel(), txAdapter)
	if err != nil {
		return false, fmt.Errorf("failed to create transaction enforcer: %w", err)
	}

	ok, err := txEnforcer.AddPolicies(rules)
	if err != nil {
		return false, fmt.Errorf("failed to add policies: %w", err)
	}
	return ok, nil
}

func (s *CasbinService) ReloadPolicy() error {
	return s.enforcer.LoadPolicy()
}
