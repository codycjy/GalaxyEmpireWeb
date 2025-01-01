package user

import (
	"GalaxyEmpireWeb/api"
	"GalaxyEmpireWeb/services/userservice"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type updateBalanceResponse struct {
	UserID  uint   `json:"user_id"`
	Succeed bool   `json:"succeed"`
	Message string `json:"message"`
}
type updateBalanceRequest struct {
	UserID uint   `json:"user_id" binding:"required"`
	Amount int64  `json:"amount" binding:"required"`
	Reason string `json:"reason" binding:"required"`
}

// UpdateBalance godoc
// @Summary Admin update user balance
// @Description Admin API to directly modify user balance
// @Tags admin
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body updateBalanceRequest true "Update balance request"
// @Success 200 {object} updateBalanceResponse
// @Failure 401 {object} api.ErrorResponse "Unauthorized"
// @Failure 400 {object} api.ErrorResponse "Bad Request"
// @Failure 500 {object} api.ErrorResponse "Internal Server Error"
// @Router /admin/user/balance [put]
func UpdateBalance(c *gin.Context) {
	var req updateBalanceRequest
	uuid := c.GetString("traceID")
	log.Info("[user]UpdateBalance", zap.String("traceID", uuid))

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error("[API::Balance]UpdateBalance Failed to bind request body", zap.String("traceID", uuid), zap.Error(err))
		c.JSON(http.StatusBadRequest, api.ErrorResponse{
			Succeed: false,
			Error:   err.Error(),
			Message: "Invalid request body",
		})
		return
	}

	userService := userservice.GetService(c)

	if err := userService.AdminUpdateBalance(c, req.UserID, req.Amount, req.Reason); err != nil {
		c.JSON(err.StatusCode(), api.ErrorResponse{
			Succeed: false,
			Error:   err.Error(),
			Message: err.Msg(),
		})
		return
	}

	c.JSON(http.StatusOK, updateBalanceResponse{
		Succeed: true,
		UserID:  req.UserID,
		Message: "Balance updated successfully",
	})
}
