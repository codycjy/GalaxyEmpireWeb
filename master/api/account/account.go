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

type AccountResponse struct {
	Succeed bool               `json:"succeed"`
	Data    *models.AccountDTO `json:"data"`
}

// GetAccount godoc
// @Summary Get account by ID
// @Description Get Account by ID
// @Tags account
// @Accept json
// @Produce json
// @Param id path int true "Account ID"
// @Success 200 {object} AccountResponse "Successful response with account data"
// @Failure 400 {object} api.ErrorResponse "Bad Request with error message"
// @Failure 404 {object} api.ErrorResponse "Not Found with error message"
// @Failure 500 {object} api.ErrorResponse "Internal Server Error with error message"
// @Router /account/{id} [get]
func GetAccountById(c *gin.Context) {

	accountService, err := accountservice.GetService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ErrorResponse{
			Succeed: false,
			Error:   err.Error(),
			Message: "Account service not initialized",
		})
	}
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{
			Succeed: false,
			Error:   err.Error(),
			Message: "Wrong User ID",
		})
		return
	}

	account, err := accountService.GetById(uint(id), []string{})
	if err != nil {
		c.JSON(http.StatusNotFound, api.ErrorResponse{
			Succeed: false,
			Error:   err.Error(),
			Message: "Account not found",
		})
		return
	}
	c.JSON(http.StatusOK, AccountResponse{
		Succeed: true,
		Data:    account.ToDTO(),
	})

}
