package paymentservice

import (
	"GalaxyEmpireWeb/models"
	"context"
)

type PaymentService interface {
	CreatePayment(ctx context.Context, order *models.Order) (*models.Payment, error)
	HandleCallback(ctx context.Context, paymentID uint) (uint, error) // return orderID
}
