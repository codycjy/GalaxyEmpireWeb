//go:build !test

package account

import (
	"GalaxyEmpireWeb/api"
	"GalaxyEmpireWeb/models"
	"GalaxyEmpireWeb/services/accountservice"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type accountResponse struct {
	Succeed bool               `json:"succeed"`
	Data    *models.AccountDTO `json:"data"`
	TraceID string             `json:"traceID"`
}
type userAccountResponse struct {
	Succeed bool            `json:"succeed"`
	Data    *models.UserDTO `json:"data"`
	TraceID string          `json:"traceID"`
}

// GetAccountByID godoc
// @Summary Get account by ID
// @Description Get Account by ID
// @Tags account
// @Accept json
// @Produce json
// @Param id path int true "Account ID"
// @Success 200 {object} accountResponse "Successful response with account data"
// @Failure 400 {object} api.ErrorResponse "Bad Request with error message"
// @Failure 404 {object} api.ErrorResponse "Not Found with error message"
// @Failure 500 {object} api.ErrorResponse "Internal Server Error with error message"
// @Router /account/{id} [get]
func GetAccountByID(c *gin.Context) {
	traceID := c.GetString("traceID")
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{
			Succeed: false,
			Error:   err.Error(),
			Message: "Wrong User ID",
			TraceID: traceID,
		})
		return
	}
	accountService, err := accountservice.GetService(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ErrorResponse{
			Succeed: false,
			Error:   err.Error(),
			Message: "Account service not initialized",
			TraceID: traceID,
		})
	}

	account, err := accountService.GetById(c, uint(id), []string{})
	if err != nil {
		c.JSON(http.StatusNotFound, api.ErrorResponse{
			Succeed: false,
			Error:   err.Error(),
			Message: "Account not found",
			TraceID: traceID,
		})
		return
	}
	c.JSON(http.StatusOK, accountResponse{
		Succeed: true,
		Data:    account.ToDTO(),
		TraceID: traceID,
	})

}

// GetAccountByUserID godoc
// @Summary Get account by User ID
// @Description Get Account by User ID
// @Tags account
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} userAccountResponse "Successful response with account data"
// @Failure 400 {object} api.ErrorResponse "Bad Request with error message"
// @Failure 404 {object} api.ErrorResponse "Not Found with error message"
// @Failure 500 {object} api.ErrorResponse "Internal Server Error with error message"
// @Router /account/user/{id} [get]
func GetAccountByUserID(c *gin.Context) {
	traceID := c.GetString("traceID")
	idStr := c.Param("userid")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{
			Succeed: false,
			Error:   err.Error(),
			Message: "Wrong User ID",
			TraceID: traceID,
		})
		return
	}

	accountService, err := accountservice.GetService(c)
	account, err := accountService.GetByUserId(c, uint(id), []string{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ErrorResponse{
			Succeed: false,
			Error:   err.Error(),
			Message: "Account service not initialized",
			TraceID: traceID,
		})
	}
	if err != nil {
		c.JSON(http.StatusNotFound, api.ErrorResponse{
			Succeed: false,
			Error:   err.Error(),
			Message: "Account not found",
			TraceID: traceID,
		})
		return
	}
	accountDTOs := make([]models.AccountDTO, len(*account))
	for i, acc := range *account {
		accountDTOs[i] = *acc.ToDTO()
	}
	user := &models.UserDTO{
		ID:       uint(id),
		Accounts: accountDTOs,
	}

	c.JSON(http.StatusOK, userAccountResponse{
		Succeed: true,
		Data:    user,
		TraceID: traceID,
	})
}
