package captchaservice

import (
	"GalaxyEmpireWeb/utils"
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type redisCaptchaStore struct {
	expire time.Duration
	rdb    *redis.Client
	ctx    context.Context
}

func NewRedisCaptchaStore(rdb *redis.Client, expire time.Duration) *redisCaptchaStore {
	return &redisCaptchaStore{
		expire: expire,
		rdb:    rdb,
		ctx:    utils.NewContextWithTraceID(),
	}
}

func (s *redisCaptchaStore) Set(id string, digits []byte) {
	s.rdb.Set(s.ctx, id, string(digits), s.expire)
}

func (s *redisCaptchaStore) Get(id string, clear bool) (digits []byte) {
	traceID := utils.TraceIDFromContext(s.ctx)
	val, err := s.rdb.Get(s.ctx, id).Result()
	if err != nil {
		log.Error("[redisCaptchaStore]Get",
			zap.Error(err),
			zap.String("traceID", traceID),
			zap.String("id", id),
		)
		return nil
	}
	if clear {
		s.rdb.Del(s.ctx, id)
	}
	return []byte(val)
}
