package paymentservice

import (
	"GalaxyEmpireWeb/config"
	"GalaxyEmpireWeb/logger"
	"GalaxyEmpireWeb/models"
	"GalaxyEmpireWeb/utils"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/checkout/session"
	"github.com/stripe/stripe-go/v81/webhook"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var paymentServiceInstance *paymentService
var log = logger.GetLogger()

type paymentService struct {
	DB *gorm.DB
}

func InitService(db *gorm.DB) {
	stripeKey := config.GetStripeConfig().SecretKey
	if stripeKey == "" {
		log.Fatal("[PaymentService] Stripe secret key is not configured")
	}

	stripe.Key = stripeKey
	log.Info("[PaymentService] Initialized with Stripe configuration",
		zap.String("keyPrefix", stripeKey[:7]+"..."))

	paymentServiceInstance = NewService(db)

	paymentServiceInstance.StartPaymentStatusChecker()
}

func NewService(db *gorm.DB) *paymentService {
	return &paymentService{
		DB: db,
	}
}

func GetService() *paymentService {
	if paymentServiceInstance == nil {
		log.Fatal("Payment Service Not Initialized")
	}
	return paymentServiceInstance
}

func (ps *paymentService) CreateCheckoutSession(ctx context.Context, priceID string) (*stripe.CheckoutSession, *utils.ServiceError) {
	traceID := utils.TraceIDFromContext(ctx)
	userID := utils.UserIDFromContext(ctx)
	log.Info("[PaymentService] CreateCheckoutSession Start",
		zap.String("traceID", traceID),
		zap.Uint("userID", userID),
		zap.String("priceID", priceID))

	expire_duration := time.Hour * 24
	if config.EXPIRE_TEST == "true" {
		expire_duration = time.Second * 10
	}
	params := &stripe.CheckoutSessionParams{
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(1),
			},
		},
		Mode:       stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL: stripe.String("http://localhost:9333/api/v1/ping"),
		CancelURL:  stripe.String("http://localhost:9333/api/v1/ping"),
		ExpiresAt:  stripe.Int64(time.Now().Add(expire_duration).Unix()),
	}

	// Create payment intention record
	intention := &models.PaymentIntention{
		UserID:   userID,
		Amount:   *params.LineItems[0].Quantity * getPriceAmount(priceID),
		Currency: "usd",
		Status:   "pending",
	}
	log.Info("[PaymentService] Creating payment intention",
		zap.String("traceID", traceID),
		zap.Uint("userID", userID),
		zap.Int64("amount", intention.Amount),
		zap.String("priceID", priceID))

	if err := ps.DB.Create(intention).Error; err != nil {
		log.Error("[PaymentService] Failed to create payment intention",
			zap.String("traceID", traceID),
			zap.Uint("userID", userID),
			zap.Error(err))
		return nil, utils.NewServiceError(http.StatusInternalServerError, "Failed to create payment intention", err)
	}

	log.Info("[PaymentService] Payment intention created",
		zap.String("traceID", traceID),
		zap.Uint("userID", userID),
		zap.Uint("intentionID", intention.ID))

	// Add metadata to track the intention
	params.Metadata = map[string]string{
		"intention_id": fmt.Sprintf("%d", intention.ID),
	}

	session, err := session.New(params)
	if err != nil {
		log.Error("[PaymentService] Failed to create checkout session",
			zap.String("traceID", traceID),
			zap.Uint("userID", userID),
			zap.Error(err),
			zap.Bool("stripeKeySet", stripe.Key != ""))
		return nil, utils.NewServiceError(http.StatusInternalServerError, "Failed to create checkout session", err)
	}

	// Update the payment intention with session details
	updateFields := map[string]interface{}{
		"session_id": session.ID,
	}

	// Only add payment_intent if it exists
	if session.PaymentIntent != nil {
		updateFields["payment_intent"] = session.PaymentIntent.ID
	}

	if err := ps.DB.Model(&models.PaymentIntention{}).
		Where("id = ?", intention.ID).
		Updates(updateFields).Error; err != nil {
		log.Error("[PaymentService] Failed to update payment intention with session details",
			zap.String("traceID", traceID),
			zap.Uint("userID", userID),
			zap.Error(err))
		return nil, utils.NewServiceError(http.StatusInternalServerError, "Failed to update payment intention", err)
	}

	log.Info("[PaymentService] Checkout session created",
		zap.String("traceID", traceID),
		zap.Uint("userID", userID),
		zap.String("sessionID", session.ID))

	return session, nil
}

func (ps *paymentService) HandleWebhook(ctx context.Context, payload []byte, signature string) *utils.ServiceError {
	traceID := utils.TraceIDFromContext(ctx)
	log.Info("[PaymentService] HandleWebhook Start",
		zap.String("traceID", traceID),
		zap.String("signature", signature),
		zap.String("Payload Size", fmt.Sprintf("%d bytes", len(payload))))

	webhookSecret := config.GetStripeConfig().WebhookSecret
	if webhookSecret == "" {
		log.Warn("[PaymentService] Webhook secret is not configured",
			zap.String("traceID", traceID))
	}

	var event stripe.Event
	var err error

	if os.Getenv("ENV") == "development" && webhookSecret == "" {
		// In development without webhook secret, just parse the JSON
		err = json.Unmarshal(payload, &event)
	} else {
		// In production, always verify the signature
		event, err = webhook.ConstructEvent(payload, signature, webhookSecret)
	}

	if err != nil {
		log.Error("[PaymentService] Webhook verification failed",
			zap.String("traceID", traceID),
			zap.Error(err),
			zap.String("env", os.Getenv("ENV")))
		return utils.NewServiceError(http.StatusBadRequest, "Webhook verification failed", err)
	}

	log.Info("[PaymentService] Webhook event received",
		zap.String("traceID", traceID),
		zap.String("eventType", string(event.Type)))

	switch event.Type {
	case "checkout.session.completed":
		var session stripe.CheckoutSession
		err := json.Unmarshal(event.Data.Raw, &session)
		if err != nil {
			log.Error("[PaymentService] Error parsing webhook payload",
				zap.String("traceID", traceID),
				zap.Error(err))
			return utils.NewServiceError(http.StatusBadRequest, "Error parsing webhook payload", err)
		}
		err1 := ps.handleSessionCompleted(ctx, &session)
		return err1

	case "charge.refunded":
		// Handle refunds - deduct from user's balance
		var charge stripe.Charge
		err := json.Unmarshal(event.Data.Raw, &charge)
		if err != nil {
			return utils.NewServiceError(http.StatusBadRequest, "Error parsing webhook payload", err)
		}

		// Start transaction
		tx := ps.DB.Begin()

		// Update user balance
		if err := tx.Exec(`
			UPDATE users 
			SET balance = balance - ? 
			WHERE id = ?`,
			charge.Amount,
			charge.Metadata["user_id"],
		).Error; err != nil {
			tx.Rollback()
			return utils.NewServiceError(http.StatusInternalServerError, "Failed to update balance", err)
		}

		// Update payment status
		if err := tx.Model(&models.Payment{}).
			Where("payment_intent = ?", charge.PaymentIntent).
			Updates(map[string]interface{}{
				"status":        "refunded",
				"refund_status": "completed",
			}).Error; err != nil {
			tx.Rollback()
			return utils.NewServiceError(http.StatusInternalServerError, "Failed to update payment", err)
		}

		if err := tx.Commit().Error; err != nil {
			return utils.NewServiceError(http.StatusInternalServerError, "Failed to commit transaction", err)
		}
	case "checkout.session.expired":
		var session stripe.CheckoutSession
		err := json.Unmarshal(event.Data.Raw, &session)
		if err != nil {
			log.Error("[PaymentService] Error parsing webhook payload",
				zap.String("traceID", traceID),
				zap.Error(err),
				zap.String("sessionID", session.ID),
				zap.String("TraceID", utils.TraceIDFromContext(ctx)))
			return utils.NewServiceError(http.StatusBadRequest, "Error parsing webhook payload", err)
		}
		err1 := ps.handleSessionExpired(ctx, session.ID)
		if err1 != nil {
			return utils.NewServiceError(http.StatusInternalServerError, "Failed to handle session expired", err1)
		}

	default:
		log.Info("[PaymentService] Unhandled event type",
			zap.String("traceID", traceID),
			zap.String("eventType", string(event.Type)))
		// Do not return an error for unhandled event types
		return nil
	}

	return nil
}
func (ps *paymentService) handleSessionCompleted(ctx context.Context, session *stripe.CheckoutSession) *utils.ServiceError {

	// First try to find the payment intention by session ID
	var intention models.PaymentIntention
	traceID := utils.TraceIDFromContext(ctx)
	if err := ps.DB.Where("session_id = ?", session.ID).First(&intention).Error; err != nil {
		log.Error("[PaymentService] Failed to find payment intention by session ID",
			zap.String("traceID", traceID),
			zap.String("sessionID", session.ID),
			zap.Error(err))

		// Fallback to metadata if available
		if intentionID, ok := session.Metadata["intention_id"]; ok {
			if err := ps.DB.First(&intention, intentionID).Error; err != nil {
				return utils.NewServiceError(http.StatusNotFound, "Payment intention not found", err)
			}
		} else {
			return nil
			// Because we cannot find the intention, we don't want Stripe to retry the payment
		}
	}

	// Validate payment amount
	if !ps.validatePaymentAmount(session, &intention) {
		log.Error("[PaymentService] Invalid payment amount",
			zap.String("traceID", traceID),
			zap.String("sessionID", session.ID),
			zap.Uint("intentionID", intention.ID),
			zap.Int64("sessionAmount", session.AmountTotal),
			zap.Int64("intentionAmount", intention.Amount))
		// Because the payment amount is invalid, we don't want Stripe to retry the payment
		return nil
	}

	if ps.isPaymentProcessed(session.PaymentIntent.ID) {
		log.Info("[PaymentService] Payment already processed",
			zap.String("traceID", traceID),
			zap.String("sessionID", session.ID))
		return nil
	}
	// Create payment record
	payment, err := ps.createPaymentRecord(ctx, session, models.PaymentStatusSucceeded)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Info("[PaymentService] Payment already processed",
				zap.String("traceID", traceID),
				zap.String("sessionID", session.ID))
			return nil
		}
		return utils.NewServiceError(http.StatusInternalServerError, "Failed to create payment record", err)
	}

	// Handle successful payment
	if err := ps.fulfillOrder(ctx, payment); err != nil {
		return utils.NewServiceError(http.StatusInternalServerError, "Failed to fulfill order", err)
	}

	log.Info("[PaymentService] Payment fulfilled",
		zap.String("traceID", traceID),
		zap.String("sessionID", session.ID),
		zap.Uint("paymentID", payment.ID))
	return nil
}
func (ps *paymentService) handleSessionExpired(ctx context.Context, sessionID string) error {
	// Find the payment intention by session ID
	var intention models.PaymentIntention
	if err := ps.DB.Where("session_id = ?", sessionID).First(&intention).Error; err != nil {
		log.Error("[PaymentService] Failed to find payment intention by session ID",
			zap.String("sessionID", sessionID),
			zap.Error(err),
			zap.String("traceID", utils.TraceIDFromContext(ctx)))
		return nil
		// Because the session is expired, we don't want stripe to retry the payment
	}

	// Update intention status
	err := ps.DB.Model(&intention).Update("status", "expired").Error
	if err != nil {
		log.Error("[PaymentService] Failed to update payment intention status",
			zap.String("sessionID", sessionID),
			zap.Error(err),
			zap.String("traceID", utils.TraceIDFromContext(ctx)))

		return err
	}
	log.Info("[PaymentService] Payment intention expired",
		zap.String("sessionID", sessionID),
		zap.Uint("intentionID", intention.ID),
		zap.String("traceID", utils.TraceIDFromContext(ctx)))
	return nil
}
func (ps *paymentService) fulfillOrder(ctx context.Context, payment *models.Payment) error {
	return ps.DB.Transaction(func(tx *gorm.DB) error {
		// 锁定用户记录
		var user models.User
		if err := tx.Set("gorm:query_option", "FOR UPDATE").
			First(&user, payment.UserID).Error; err != nil {
			return err
		}

		// 检查支付记录状态，防止重复处理
		var currentPayment models.Payment
		if err := tx.Set("gorm:query_option", "FOR UPDATE").
			Where("id = ? AND fulfilled_at IS NULL", payment.ID).
			First(&currentPayment).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Info("[PaymentService] Payment already fulfilled",
					zap.String("paymentIntent", payment.PaymentIntent),
					zap.String("traceID", utils.TraceIDFromContext(ctx)))
				return nil
			}
			return err
		}

		now := time.Now()

		// 直接更新用户余额
		if err := tx.Model(&user).Update("balance", user.Balance+payment.Amount).Error; err != nil {
			return err
		}

		// 更新支付状态
		if err := tx.Model(&currentPayment).Updates(map[string]interface{}{
			"fulfilled_at": now,
			"status":       models.PaymentStatusSucceeded,
		}).Error; err != nil {
			return err
		}

		// 记录余额变更日志
		balanceLog := &models.BalanceLog{
			UserID:      payment.UserID,
			Amount:      payment.Amount,
			Type:        "deposit",
			Reference:   payment.PaymentIntent,
			Description: fmt.Sprintf("Payment deposit: %s", payment.PaymentIntent),
			Balance:     user.Balance + payment.Amount,
		}

		return tx.Create(balanceLog).Error
	})
}
func getPriceAmount(priceID string) int64 {
	// First try to get price from configuration
	for _, price := range config.GetPrices() {
		if price.StripeID == priceID {
			return price.Amount
		}
	}

	// Fallback: try to get from price configuration
	for _, priceConfig := range config.GetAvailablePrices() {
		if priceConfig.StripeID == priceID {
			return priceConfig.Amount
		}
	}

	log.Error("[PaymentService] Price not found",
		zap.String("priceID", priceID))
	return 0
}

// Add these functions to help with testing

func (ps *paymentService) GetPaymentIntention(id uint) (*models.PaymentIntention, error) {
	var intention models.PaymentIntention
	err := ps.DB.First(&intention, id).Error
	return &intention, err
}

func (ps *paymentService) GetPayment(sessionID string) (*models.Payment, error) {
	var payment models.Payment
	err := ps.DB.Where("session_id = ?", sessionID).First(&payment).Error
	return &payment, err
}

func (ps *paymentService) GetUserBalance(userID uint) (int64, error) {
	var balance int64
	err := ps.DB.Raw("SELECT balance FROM users WHERE id = ?", userID).Scan(&balance).Error
	return balance, err
}

func (ps *paymentService) GetPaymentHistory(ctx context.Context, userID uint) ([]*models.Payment, *utils.ServiceError) {
	var payments []*models.Payment
	if err := ps.DB.Where("user_id = ?", userID).Find(&payments).Error; err != nil {
		return nil, utils.NewServiceError(http.StatusInternalServerError, "Failed to get payment history", err)
	}
	return payments, nil
}

func (ps *paymentService) GetPaymentStatus(ctx context.Context, sessionID string) (*models.Payment, *utils.ServiceError) {
	// First check if payment exists
	var payment models.Payment
	err := ps.DB.Where("session_id = ?", sessionID).First(&payment).Error
	if err == nil {
		return &payment, nil
	}

	// If payment not found, check payment intention
	intention, err := ps.GetPaymentIntentionBySession(sessionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.NewServiceError(http.StatusNotFound, "Payment not found", err)
		}
		return nil, utils.NewServiceError(http.StatusInternalServerError, "Error checking payment status", err)
	}

	// Return a payment object with pending status
	return &models.Payment{
		UserID:    intention.UserID,
		Amount:    intention.Amount,
		Currency:  intention.Currency,
		Status:    models.PaymentStatusPending,
		SessionID: intention.SessionID,
	}, nil
}

func (ps *paymentService) CreateDepositSession(ctx context.Context, amount int64) (*stripe.CheckoutSession, *utils.ServiceError) {
	userID := utils.UserIDFromContext(ctx)
	log.Info("[PaymentService] CreateDepositSession",
		zap.Uint("userID", userID),
		zap.String("traceID", utils.TraceIDFromContext(ctx)),
		zap.Int64("amount", amount))

	if userID == 0 {
		return nil, utils.NewServiceError(http.StatusUnauthorized, "User not authenticated", nil)
	}

	// Verify user exists
	var user models.User
	if err := ps.DB.First(&user, userID).Error; err != nil {
		return nil, utils.NewServiceError(http.StatusNotFound, "User not found", err)
	}

	// Get Stripe price ID from config
	priceID, valid := config.GetStripePrice(amount)
	if !valid {
		log.Error("[PaymentService] Invalid deposit amount",
			zap.Int64("amount", amount),
			zap.String("traceID", utils.TraceIDFromContext(ctx)))
		return nil, utils.NewServiceError(http.StatusBadRequest, "Invalid deposit amount", nil)
	}

	log.Info("[PaymentService] CreateDepositSession",
		zap.String("priceID", priceID),
		zap.String("traceID", utils.TraceIDFromContext(ctx)))

	return ps.CreateCheckoutSession(ctx, priceID)
}

func (ps *paymentService) isPaymentProcessed(paymentIntentID string) bool {
	var payment models.Payment
	return ps.DB.Where("payment_intent = ? AND status = 'completed'", paymentIntentID).First(&payment).Error == nil
}

func (ps *paymentService) createPaymentRecord(ctx context.Context, session *stripe.CheckoutSession, status models.PaymentStatus) (*models.Payment, error) {
	// 开始事务
	tx := ps.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get userID from session metadata or intention
	var userID uint
	if intentionID, ok := session.Metadata["intention_id"]; ok {
		var intention models.PaymentIntention
		// 在事务中使用FOR UPDATE锁定intention记录
		if err := tx.Set("gorm:query_option", "FOR UPDATE").
			First(&intention, intentionID).Error; err != nil {
			tx.Rollback()
			log.Error("[PaymentService] Failed to find intention",
				zap.String("intentionID", intentionID),
				zap.Error(err),
				zap.String("TraceID", utils.TraceIDFromContext(ctx)))
			return nil, err
		}
		userID = intention.UserID
	}

	// Because we are using the webhook, we can't assume the userID is in the context
	if userID == 0 {
		log.Error("[PaymentService] Invalid user ID",
			zap.String("sessionID", session.ID),
			zap.String("TraceID", utils.TraceIDFromContext(ctx)))
		tx.Rollback()
		return nil, fmt.Errorf("invalid user ID")
	}

	// Check if payment already exists with FOR UPDATE lock
	var existingPayment models.Payment
	err := tx.Set("gorm:query_option", "FOR UPDATE").
		Where("session_id = ?", session.ID).
		First(&existingPayment).Error

	if err == nil {
		// Payment exists, commit transaction and return existing payment
		log.Info("[PaymentService] Payment already exists",
			zap.String("sessionID", session.ID))
		tx.Commit()
		return &existingPayment, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		// Unexpected error
		tx.Rollback()
		log.Error("[PaymentService] Error checking existing payment",
			zap.Error(err),
			zap.String("sessionID", session.ID))
		return nil, err
	}

	now := time.Now()
	payment := &models.Payment{
		UserID:         userID,
		Amount:         session.AmountTotal,
		Currency:       string(session.Currency),
		Status:         status,
		PaymentIntent:  session.PaymentIntent.ID,
		SessionID:      session.ID,
		PaidAt:         now,
		IdempotencyKey: session.ID,
		CustomerEmail:  session.CustomerDetails.Email,
	}

	// Create new payment record within the transaction
	if err := tx.Create(payment).Error; err != nil {
		log.Error("[PaymentService] Failed to create payment",
			zap.Error(err),
			zap.String("sessionID", session.ID))
		zap.String("TraceID", utils.TraceIDFromContext(ctx))
		tx.Rollback()
		return nil, err
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		log.Error("[PaymentService] Failed to commit transaction",
			zap.Error(err),
			zap.String("sessionID", session.ID),
			zap.String("TraceID", utils.TraceIDFromContext(ctx)))
		return nil, err
	}
	log.Info("[PaymentService] Payment created",
		zap.String("sessionID", session.ID),
		zap.Uint("paymentID", payment.ID),
		zap.String("TraceID", utils.TraceIDFromContext(ctx)),
		zap.Int64("amount", payment.Amount),
		zap.Int64("PaymentID", int64(payment.ID)))

	return payment, nil
}

func (ps *paymentService) validatePaymentAmount(session *stripe.CheckoutSession, intention *models.PaymentIntention) bool {
	return session.AmountTotal == intention.Amount
}

func (ps *paymentService) GetPaymentIntentionBySession(sessionID string) (*models.PaymentIntention, error) {
	var intention models.PaymentIntention
	err := ps.DB.Where("session_id = ?", sessionID).First(&intention).Error
	if err != nil {
		return nil, err
	}
	return &intention, nil
}

func (ps *paymentService) StartPaymentStatusChecker() {
	go func() {
		ticker := time.NewTicker(5 * time.Minute) // Check every 5 minutes
		defer ticker.Stop()

		for {
			ctx := utils.NewContextWithTraceID()

			select {
			case <-ctx.Done():
				log.Info("[PaymentService] Stopping payment status checker")
				return
			case <-ticker.C:
				log.Info("[PaymentService] Starting payment status checker", zap.String("traceID", utils.TraceIDFromContext(ctx)))
				ps.checkPendingPayments(ctx)
			}
		}
	}()
	log.Info("[PaymentService] Started payment status checker")
}

func (ps *paymentService) checkPendingPayments(ctx context.Context) {
	var intentions []models.PaymentIntention
	// Get pending intentions that are not too old (e.g., less than 24 hours)
	if err := ps.DB.Where("status = ? AND created_at > ?", "pending", time.Now().Add(-24*time.Hour)).Find(&intentions).Error; err != nil {
		log.Error("[PaymentService] Failed to fetch pending payments", zap.Error(err))
		return
	}
	log.Info("[PaymentService] Found pending payments", zap.Int("count", len(intentions)))
	if len(intentions) == 0 {
		log.Info("[PaymentService] No pending payments found")
		return
	}

	for _, intention := range intentions {
		log.Info("[PaymentService] Checking intention", zap.Uint("intentionID", intention.ID), zap.String("sessionID", intention.SessionID))
		if intention.SessionID == "" {
			log.Info("[PaymentService] No session ID found for intention", zap.Uint("intentionID", intention.ID))
			continue
		}

		// Fetch session from Stripe
		session, err := ps.getStripeSession(intention.SessionID)
		if err != nil {
			log.Error("[PaymentService] Failed to fetch session from Stripe",
				zap.String("sessionID", intention.SessionID),
				zap.Error(err))
			continue
		}
		log.Info("[PaymentService] Fetched session from Stripe", zap.String("sessionID", session.ID), zap.String("status", string(session.Status)))

		// Update status based on session
		ps.updatePaymentStatus(ctx, &intention, session)
	}
}

func (ps *paymentService) getStripeSession(sessionID string) (*stripe.CheckoutSession, error) {
	return session.Get(sessionID, nil)
}

func (ps *paymentService) updatePaymentStatus(ctx context.Context, intention *models.PaymentIntention, session *stripe.CheckoutSession) {
	// Start transaction
	tx := ps.DB.Begin()

	switch session.Status {
	case "complete":
		// Create payment record if it doesn't exist
		payment, err := ps.createPaymentRecord(ctx, session, models.PaymentStatusSucceeded)
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				log.Error("[PaymentService] Failed to create payment record",
					zap.String("sessionID", session.ID),
					zap.Error(err))
				tx.Rollback()
				return
			}
		}

		// Update intention status
		if err := tx.Model(intention).Updates(map[string]interface{}{
			"status":       "completed",
			"completed_at": time.Now(),
		}).Error; err != nil {
			log.Error("[PaymentService] Failed to update intention status",
				zap.String("sessionID", session.ID),
				zap.Error(err))
			tx.Rollback()
			return
		}

		// Fulfill the order if payment record was just created
		if payment != nil {
			if err := ps.fulfillOrder(ctx, payment); err != nil {
				log.Error("[PaymentService] Failed to fulfill order",
					zap.String("sessionID", session.ID),
					zap.Error(err))
				tx.Rollback()
				return
			}
		}

	case "expired":
		if err := tx.Model(intention).Updates(map[string]interface{}{
			"status": "expired",
		}).Error; err != nil {
			log.Error("[PaymentService] Failed to update intention status to expired",
				zap.String("sessionID", session.ID),
				zap.Error(err))
			tx.Rollback()
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		log.Error("[PaymentService] Failed to commit transaction",
			zap.String("sessionID", session.ID),
			zap.Error(err))
		tx.Rollback()
	}
}
