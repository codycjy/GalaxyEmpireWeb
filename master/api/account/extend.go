package account

import (
	"GalaxyEmpireWeb/api"
	"GalaxyEmpireWeb/models"
	"GalaxyEmpireWeb/services/paymentservice"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ExtendAccount godoc
// @Summary Extend account time by 31 days
// @Description Extends the account expiry time by 31 days for config.ExtendPrice
// @Tags account
// @Accept json
// @Produce json
// @Param account body models.Account true "Account to extend"
// @Success 200 {object} accountResponse "Account extended successfully"
// @Failure 400 {object} api.ErrorResponse "Bad Request with error message"
// @Failure 401 {object} api.ErrorResponse "Unauthorized"
// @Failure 403 {object} api.ErrorResponse "Forbidden - Account not owned by user"
// @Failure 500 {object} api.ErrorResponse "Internal Server Error with error message"
// @Router /account/extend [post]
func ExtendAccount(c *gin.Context) {
	traceID := c.GetString("traceID")

	var account models.Account
	if err := c.BindJSON(&account); err != nil {
		log.Error("[API::ExtendAccount]Failed to bind account data",
			zap.String("traceID", traceID),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, api.ErrorResponse{
			Succeed: false,
			Error:   err.Error(),
			Message: "Invalid request body",
			TraceID: traceID,
		})
		return
	}

	paymentService := paymentservice.GetService()
	if serviceErr := paymentService.ExtendAccountTime(c, &account); serviceErr != nil {
		c.JSON(serviceErr.StatusCode(), api.ErrorResponse{
			Succeed: false,
			Error:   serviceErr.Error(),
			Message: serviceErr.Msg(),
			TraceID: traceID,
		})
		return
	}

	c.JSON(http.StatusOK, accountResponse{
		Succeed: true,
		TraceID: traceID,
		Data:    account.ToDTO(),
	})
}
