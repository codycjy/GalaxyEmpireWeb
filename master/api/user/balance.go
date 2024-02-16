package user

import (
	"GalaxyEmpireWeb/api"
	"GalaxyEmpireWeb/models"
	"GalaxyEmpireWeb/services/userservice"
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

// UpdateUser godoc
// @Summary Update a user balance
// @Description Update a user balance
// @Tags user
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} userResponse "Successful response with user data"
// @Failure 400 {object} api.ErrorResponse "Bad Request with error message"
// @Failure 404 {object} api.ErrorResponse "Not Found with error message"
// @Failure 500 {object} api.ErrorResponse "Internal Server Error with error message"
// @Router /user/balance [put]
func UpdateBalance(c *gin.Context) {
	var user *models.User
	uuid := c.GetString("traceID")
	ctx := context.WithValue(context.Background(), "traceID", uuid)
	err := c.ShouldBindJSON(user)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{
			Succeed: false,
			Error:   err.Error(),
			Message: "Failed to bind user",
		})
		return
	}
	userservice, err := userservice.GetService(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ErrorResponse{
			Succeed: false,
			Error:   err.Error(),
			Message: "User service not initialized",
		})

		return
	}
	serviceErr := userservice.UpdateBalance(ctx, user)
	if serviceErr != nil {
		c.JSON(serviceErr.StatusCode(), api.ErrorResponse{
			Succeed: false,
			Error:   serviceErr.Error(),
			Message: serviceErr.Msg(),
		})
		return
	}
	userDTO := user.ToDTO()
	c.JSON(http.StatusOK, userResponse{
		Succeed: true,
		Data:    userDTO,
	})
}
