package nodeservice

import (
	"GalaxyEmpireWeb/logger"
	"GalaxyEmpireWeb/utils"
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var log = logger.GetLogger()
var nodeService *NodeService

type NodeService struct {
	db  *gorm.DB
	rdb *redis.Client
}

func NewService(db *gorm.DB, rdb *redis.Client) *NodeService {
	return &NodeService{
		db:  db,
		rdb: rdb,
	}
}

func InitService(db *gorm.DB, rdb *redis.Client) {
	nodeService = NewService(db, rdb)
}

func GetService() *NodeService {
	if nodeService == nil {
		log.Fatal("[service]NodeService is not initialized")
	}

	return nodeService

}

func (service *NodeService) RegisterNode(ctx context.Context, name string) error {
	traceID := utils.TraceIDFromContext(ctx)
	log.Info("[service] Register node",
		zap.String("name", name),
		zap.String("traceID", traceID),
	)

	// 设置节点信息
	result := service.rdb.HSet(ctx, "node", name, "1")
	if err := result.Err(); err != nil {
		log.Warn("[service] Register node failed",
			zap.String("traceID", traceID),
			zap.Error(err),
		)
		return err
	}

	// 使用 Sorted Set 记录节点过期时间
	expirationTime := float64(time.Now().Add(time.Hour).Unix())
	service.rdb.ZAdd(ctx, "node_expiration", redis.Z{
		Score:  expirationTime,
		Member: name,
	})

	return nil
}

func (service *NodeService) GetActiveNodes(ctx context.Context) ([]string, error) {
	traceID := utils.TraceIDFromContext(ctx)
	// 获取当前时间戳
	currentTime := float64(time.Now().Unix())

	// 清除过期的节点
	service.rdb.ZRemRangeByScore(ctx, "node_expiration", "-inf", fmt.Sprintf("%f", currentTime))

	log.Info("[service] remove expired node",
		zap.String("traceID", traceID),
	)

	// 查询所有未过期的节点
	nodeNames, err := service.rdb.ZRangeByScore(ctx, "node_expiration", &redis.ZRangeBy{
		Min:    fmt.Sprintf("%f", currentTime),
		Max:    "+inf",
		Offset: 0,
		Count:  -1,
	}).Result()

	if err != nil {
		return nil, err
	}

	// 过滤并返回有效的节点名
	var activeNodes []string
	for _, name := range nodeNames {
		if exists := service.rdb.HExists(ctx, "node", name).Val(); exists {
			activeNodes = append(activeNodes, name)
		}
	}

	return activeNodes, nil
}
