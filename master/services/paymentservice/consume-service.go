package paymentservice

import (
	"GalaxyEmpireWeb/config"
	"GalaxyEmpireWeb/models"
	"GalaxyEmpireWeb/utils"
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

func (ps *paymentService) ExtendAccountTime(ctx context.Context, account *models.Account) *utils.ServiceError {
	userID := utils.UserIDFromContext(ctx)
	traceID := utils.TraceIDFromContext(ctx)

	if userID == 0 {
		log.Error("[ExtendAccountTime] Unauthorized",
			zap.String("traceID", traceID),
		)
		return utils.NewServiceError(http.StatusUnauthorized, "Unauthorized", errors.New("Unauthorized"))
	}
	if account == nil || account.ID == 0 {
		log.Error("[ExtendAccountTime] Invalid account",
			zap.String("traceID", traceID),
		)
		return utils.NewServiceError(http.StatusBadRequest, "Invalid account", nil)
	}

	// Get extend price from config
	extendPrice, exists := config.GetPrice(config.ExtendPrice)
	if !exists || !extendPrice.Available {
		log.Error("[ExtendAccountTime] Account extension price not configured or not available",
			zap.String("traceID", traceID),
		)
		return utils.NewServiceError(http.StatusServiceUnavailable, "Account extension is not available", nil)
	}

	// Start transaction
	tx := ps.DB.Begin()
	if tx.Error != nil {
		log.Error("[ExtendAccountTime] Failed to start transaction",
			zap.String("traceID", traceID),
			zap.Error(tx.Error),
		)
		return utils.NewServiceError(http.StatusInternalServerError, "Failed to start transaction", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Set("gorm:query_option", "FOR UPDATE").
		First(account).Error; err != nil {
		tx.Rollback()
		log.Error("[ExtendAccountTime] Failed to get account",
			zap.String("traceID", traceID),
			zap.Error(err),
		)
		return utils.NewServiceError(http.StatusNotFound, "Account not found", err)
	}
	if account.UserID != userID {
		tx.Rollback()
		log.Error("[ExtendAccountTime] Account not owned by user",
			zap.String("traceID", traceID),
			zap.Uint("accountID", account.ID),
			zap.Uint("userID", userID),
		)
		return utils.NewServiceError(http.StatusForbidden, "Account not owned by user", nil)
	}

	log.Info("[ExtendAccountTime] Account Info",
		zap.String("traceID", traceID),
		zap.Time("expireAt", account.ExpireAt),
		zap.Uint("accountID", account.ID),
		zap.Uint("userID", userID),
	)

	// Lock and get user record
	var user models.User
	if err := tx.Set("gorm:query_option", "FOR UPDATE").
		Preload("Accounts", "id = ?", account.ID).
		First(&user, userID).Error; err != nil {
		tx.Rollback()
		log.Error("[ExtendAccountTime] Failed to get user",
			zap.String("traceID", traceID),
			zap.Error(err),
		)
		return utils.NewServiceError(http.StatusNotFound, "User not found", err)
	}

	// Verify account ownership // Just a double check
	if len(user.Accounts) == 0 {
		tx.Rollback()
		log.Error("[ExtendAccountTime] Account not owned by user",
			zap.String("traceID", traceID),
			zap.Uint("accountID", account.ID),
			zap.Uint("userID", userID),
		)
		return utils.NewServiceError(http.StatusForbidden, "Account not owned by user", nil)
	}

	// Check balance
	if user.Balance < extendPrice.Amount {
		tx.Rollback()
		log.Error("[ExtendAccountTime] Insufficient balance",
			zap.String("traceID", traceID),
			zap.Int64("balance", user.Balance),
			zap.Int64("required", extendPrice.Amount),
		)
		return utils.NewServiceError(http.StatusBadRequest, fmt.Sprintf("Insufficient balance. Required: %d", extendPrice.Amount), nil)
	}

	// Update user balance
	if err := tx.Model(&user).Update("balance", user.Balance-extendPrice.Amount).Error; err != nil {
		tx.Rollback()
		log.Error("[ExtendAccountTime] Failed to update balance",
			zap.String("traceID", traceID),
			zap.Error(err),
		)
		return utils.NewServiceError(http.StatusInternalServerError, "Failed to update balance", err)
	}

	var newExpireTime time.Time
	// Update account expiry time
	if account.ExpireAt.After(time.Now()) {
		newExpireTime = account.ExpireAt.AddDate(0, 0, 31)
	} else {
		newExpireTime = time.Now().AddDate(0, 0, 31)
	}
	if err := tx.Model(account).
		Where("id = ?", account.ID).
		Update("expire_at", newExpireTime).Error; err != nil {
		tx.Rollback()
		log.Error("[ExtendAccountTime] Failed to extend account time",
			zap.String("traceID", traceID),
			zap.Error(err),
		)
		return utils.NewServiceError(http.StatusInternalServerError, "Failed to extend account time", err)
	}

	// Create balance log
	balanceLog := &models.BalanceLog{
		UserID:      userID,
		Amount:      -extendPrice.Amount,
		Type:        "extend_account",
		Reference:   fmt.Sprintf("account_%d", account.ID),
		Description: fmt.Sprintf("Extended account %d for 31 days (%s)", account.ID, extendPrice.Description),
		Balance:     user.Balance,
	}

	if err := tx.Create(balanceLog).Error; err != nil {
		tx.Rollback()
		log.Error("[ExtendAccountTime] Failed to create balance log",
			zap.String("traceID", traceID),
			zap.Error(err),
		)
		return utils.NewServiceError(http.StatusInternalServerError, "Failed to create balance log", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		log.Error("[ExtendAccountTime] Failed to commit transaction",
			zap.String("traceID", traceID),
			zap.Error(err),
		)
		return utils.NewServiceError(http.StatusInternalServerError, "Failed to commit transaction", err)
	}

	log.Info("[ExtendAccountTime] Successfully extended account time",
		zap.String("traceID", traceID),
		zap.Uint("accountID", account.ID),
		zap.Uint("userID", userID),
		zap.Time("newExpireTime", newExpireTime),
	)

	return nil
}
