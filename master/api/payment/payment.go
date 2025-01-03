package payment

import (
	"GalaxyEmpireWeb/api"
	"GalaxyEmpireWeb/logger"
	"GalaxyEmpireWeb/services/paymentservice"
	"GalaxyEmpireWeb/utils"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var log = logger.GetLogger()

// @Summary Create a checkout session
// @Description Creates a Stripe checkout session for payment
// @Tags payment
// @Accept json
// @Produce json
// @Param price_id query string true "Stripe Price ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Router /payment/create-checkout [post]
func CreateCheckoutSession(c *gin.Context) {
	priceID := c.Query("price_id")
	if priceID == "" {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{
			Succeed: false,
			TraceID: utils.TraceIDFromContext(c),
			Message: "price_id is required",
		})
		return
	}

	session, err := paymentservice.GetService().CreateCheckoutSession(c, priceID)
	if err != nil {
		c.JSON(err.StatusCode(), api.ErrorResponse{
			Succeed: false,
			TraceID: utils.TraceIDFromContext(c),
			Message: err.Msg(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"sessionId": session.ID,
		"url":       session.URL,
	})
}

// @Summary Handle Stripe webhook
// @Description Handles Stripe webhook events
// @Tags payment
// @Accept json
// @Produce json
// @Success 200 {string} string "ok"
// @Failure 400 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Router /payment/webhook [post]
func HandleWebhook(c *gin.Context) {
	// Read the request body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{
			Succeed: false,
			TraceID: utils.TraceIDFromContext(c),
			Message: "Error reading request body",
		})
		return
	}

	// Get the Stripe signature header
	stripeSignature := c.GetHeader("Stripe-Signature")
	if stripeSignature == "" {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{
			Succeed: false,
			TraceID: utils.TraceIDFromContext(c),
			Message: "Stripe signature is missing",
		})
		return
	}

	err1 := paymentservice.GetService().HandleWebhook(c, body, stripeSignature)
	if err1 != nil {
		svcErr, ok := err.(*utils.ServiceError)
		if ok && svcErr != nil {
			c.JSON(svcErr.StatusCode(), api.ErrorResponse{
				Succeed: false,
				TraceID: utils.TraceIDFromContext(c),
				Message: svcErr.Msg(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, api.ErrorResponse{
			Succeed: false,
			TraceID: utils.TraceIDFromContext(c),
			Message: "Failed to handle webhook",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

// @Summary Get payment history
// @Description Get payment history for the current user
// @Tags payment
// @Accept json
// @Produce json
// @Success 200 {array} models.Payment
// @Failure 400 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Router /payment/history [get]
func GetPaymentHistory(c *gin.Context) {
	userID := utils.UserIDFromContext(c)
	payments, err := paymentservice.GetService().GetPaymentHistory(c.Request.Context(), userID)
	if err != nil {
		c.JSON(err.StatusCode(), api.ErrorResponse{
			Succeed: false,
			TraceID: utils.TraceIDFromContext(c),
			Message: err.Msg(),
		})
		return
	}

	c.JSON(http.StatusOK, payments)
}

// @Summary Get payment status
// @Description Get the status of a specific payment
// @Tags payment
// @Accept json
// @Produce json
// @Param payment_id path string true "Payment ID"
// @Success 200 {object} models.Payment
// @Failure 400 {object} api.ErrorResponse
// @Failure 404 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Router /payment/{payment_id} [get]
func GetPaymentStatus(c *gin.Context) {
	paymentID := c.Param("payment_id")
	if paymentID == "" {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{
			Succeed: false,
			TraceID: utils.TraceIDFromContext(c),
			Message: "Payment ID is required",
		})
		return
	}

	payment, err := paymentservice.GetService().GetPaymentStatus(c.Request.Context(), paymentID)
	if err != nil {
		c.JSON(err.StatusCode(), api.ErrorResponse{
			Succeed: false,
			TraceID: utils.TraceIDFromContext(c),
			Message: err.Msg(),
		})
		return
	}

	c.JSON(http.StatusOK, payment)
}

// @Summary Create a deposit session
// @Description Creates a Stripe checkout session for deposit
// @Tags payment
// @Accept json
// @Produce json
// @Param amount query int true "Deposit amount in cents (1000 for $10, 2000 for $20, 5000 for $50)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Router /payment/deposit [post]
func CreateDepositSession(c *gin.Context) {
	amountStr := c.Query("amount")
	amount, err := strconv.ParseInt(amountStr, 10, 64)
	log.Info("CreateDepositSession", zap.String("amount", amountStr), zap.String("traceID", utils.TraceIDFromContext(c)))
	if err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{
			Succeed: false,
			TraceID: utils.TraceIDFromContext(c),
			Message: "Invalid amount",
		})
		return
	}
	log.Info("CreateDepositSession", zap.Int64("amount", amount), zap.String("traceID", utils.TraceIDFromContext(c)))

	session, svcErr := paymentservice.GetService().CreateDepositSession(c, amount)
	if svcErr != nil {
		log.Info("CreateDepositSession", zap.String("error", svcErr.Msg()), zap.Int("status", svcErr.StatusCode()), zap.String("traceID", utils.TraceIDFromContext(c)))
		c.JSON(svcErr.StatusCode(), api.ErrorResponse{
			Succeed: false,
			TraceID: utils.TraceIDFromContext(c),
			Message: svcErr.Msg(),
		})
		return
	}
	log.Info("CreateDepositSession", zap.String("sessionID", session.ID), zap.String("url", session.URL), zap.String("traceID", utils.TraceIDFromContext(c)))

	c.JSON(http.StatusOK, gin.H{
		"sessionId": session.ID,
		"url":       session.URL,
	})
}
