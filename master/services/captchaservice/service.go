package captchaservice

import (
	"GalaxyEmpireWeb/logger"
	"GalaxyEmpireWeb/utils"
	"context"
	"time"

	"github.com/dchest/captcha"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var log = logger.GetLogger()
var captchaService *CaptchaService

type CaptchaService struct {
	store captcha.Store
}

func NewCaptchaService(rdb *redis.Client) *CaptchaService {
	store := NewRedisCaptchaStore(rdb, 10*time.Minute)
	return &CaptchaService{store: store}
}
func InitCaptchaService(rdb *redis.Client) {
	captchaService = NewCaptchaService(rdb)
}
func GetCaptchaService() *CaptchaService {
	if captchaService == nil {
		log.DPanic("captchaService is nil")
	}
	return captchaService

}

func (s *CaptchaService) GenerateCaptcha() string {
	return captcha.New()
}

func (s *CaptchaService) VerifyCaptcha(ctx context.Context, captchaID, userInput string) bool {
	traceID := utils.TraceIDFromContext(ctx)
	log.Info("[service]VerifyCaptcha",
		zap.String("traceID", traceID),
		zap.String("captchaID", captchaID),
		zap.String("userInput", userInput),
	)
	result := captcha.VerifyString(captchaID, userInput)
	log.Info("[service]VerifyCaptcha", zap.String("traceID", traceID), zap.Bool("result", result))
	return result
}
