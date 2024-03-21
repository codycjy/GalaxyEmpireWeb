package orderservice

import (
	"GalaxyEmpireWeb/logger"
	"GalaxyEmpireWeb/models"
	"GalaxyEmpireWeb/services/paymentservice"
	"GalaxyEmpireWeb/utils"
	"context"
	"errors"
	"net/http"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

var log = logger.GetLogger()

var orderServiceInstance *OrderService

func NewService(db *gorm.DB) *OrderService { // TODO: add paymentservice
	return &OrderService{
		db: db,
	}
}
func InitService(db *gorm.DB) error {
	if orderServiceInstance != nil {
		return errors.New("AccountService is already initialized")
	}
	orderServiceInstance = NewService(db)
	log.Info("[service] Account service Initialized")
	return nil
}

type OrderService struct {
	db *gorm.DB
	ps paymentservice.PaymentService
}

func (service *OrderService) CreateOrder(ctx context.Context, account models.Account, price int) (*models.Order, *models.Payment, error) {
	traceID := utils.TraceIDFromContext(ctx)
	order := &models.Order{
		Price:     price,
		UserID:    account.UserID,
		AccountID: account.ID,
	}

	// 将新订单保存到数据库
	result := service.db.Create(order)
	if result.Error != nil {
		log.Error("Failed to create order",
			zap.String("traceID", traceID),
			zap.Error(result.Error),
			zap.Uint("accountID", account.ID),
			zap.Uint("userID", account.UserID),
			zap.Int("price", price),
		)
		return nil, nil, utils.NewServiceError(http.StatusInternalServerError, "Failed to create order", result.Error)
	}
	payment, err := service.ps.CreatePayment(ctx, order)
	if err != nil {

		log.Error("Failed to create payment",
			zap.String("traceID", traceID),
			zap.Error(result.Error),
			zap.Uint("accountID", account.ID),
			zap.Uint("userID", account.UserID),
			zap.Int("price", price),
		)
		return nil, nil, utils.NewServiceError(http.StatusInternalServerError, "Failed to create order", result.Error)
	}

	return order, payment, nil
}

func (service *OrderService) HandleCallback(ctx context.Context, paymentID uint) (uint, error) {
	traceID := utils.TraceIDFromContext(ctx)
	log.Info("[service]HandleCallback",
		zap.Uint("paymentID", paymentID),
		zap.String("traceID", traceID),
	)
	service.ps.HandleCallback(ctx, paymentID) // TODO: process payment
	orderID := uint(1)                        // WARNING: hardcode
	log.Fatal("HandleCallback not implemented")
	service.CompleteOrder(ctx, orderID) // TODO:

	return 0, nil
}

func (service *OrderService) CompleteOrder(ctx context.Context, orderID uint) error { // TODO:
	traceID := utils.TraceIDFromContext(ctx)
	log.Info("[service]CompleteOrder",
		zap.Uint("orderID", orderID),
		zap.String("traceID", traceID))
	return nil
}
